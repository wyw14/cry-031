package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/wyw/cry031-volunteer/internal/domain"
)

type Engine struct {
	store    Store
	clock    Clock
	notifier Notifier
}

func NewEngine(store Store, clock Clock, notifier Notifier) *Engine {
	return &Engine{store: store, clock: clock, notifier: notifier}
}

func (e *Engine) transact(ctx context.Context, actor Actor, meta RequestMeta, fn func(*domain.State, time.Time) error) error {
	state, err := e.store.Snapshot(ctx)
	if err != nil {
		return err
	}
	now := e.clock.Now().UTC()
	if err := fn(&state, now); err != nil {
		return err
	}
	event := domain.AuditEvent{ID: newID("audit"), ActorID: actor.UserID, Action: "domain_change", RequestID: meta.RequestID, CreatedAt: now}
	state.Audit = append(state.Audit, event)
	if err := e.store.Commit(ctx, state.Version, state, []domain.AuditEvent{event}); err != nil {
		return err
	}
	return nil
}

func (e *Engine) read(ctx context.Context) (domain.State, error) { return e.store.Snapshot(ctx) }

func (e *Engine) notify(ctx context.Context, userID, kind, message string) {
	if e.notifier != nil && strings.TrimSpace(userID) != "" {
		_ = e.notifier.Notify(ctx, userID, kind, message)
	}
}

func (e *Engine) addNotice(state *domain.State, userID, kind, message string, now time.Time) {
	id := newID("notice")
	state.Notices[id] = domain.Notice{ID: id, UserID: userID, Kind: kind, Message: message, CreatedAt: now}
}

func mustUser(state *domain.State, userID string) (domain.User, error) {
	user, ok := state.Users[userID]
	if !ok || !user.Active {
		return domain.User{}, domain.ErrForbidden
	}
	return user, nil
}

func mustTeam(state *domain.State, teamID string) (domain.Team, error) {
	team, ok := state.Teams[teamID]
	if !ok {
		return domain.Team{}, domain.ErrNotFound
	}
	return team, nil
}

func mustMembership(state *domain.State, teamID, userID string) (domain.Membership, error) {
	for _, membership := range state.Memberships {
		if membership.TeamID == teamID && membership.UserID == userID {
			return membership, nil
		}
	}
	return domain.Membership{}, domain.ErrNotFound
}

func canManageTeam(state *domain.State, actor Actor, teamID string) bool {
	if user, ok := state.Users[actor.UserID]; ok && user.Active && user.Role == domain.RoleAdmin {
		return true
	}
	membership, err := mustMembership(state, teamID, actor.UserID)
	return err == nil && membership.CanManage()
}

func (e *Engine) CanAccessProfile(ctx context.Context, actor Actor, userID string) (bool, error) {
	state, err := e.read(ctx)
	if err != nil {
		return false, err
	}
	if actor.UserID == userID {
		_, err := mustUser(&state, userID)
		return err == nil, err
	}
	user, ok := state.Users[actor.UserID]
	if !ok {
		return false, nil
	}
	if !user.Active {
		return false, nil
	}
	if !profileReadAllowed(user) {
		return false, nil
	}
	if target, exists := state.Users[userID]; !exists || !target.Active {
		return false, domain.ErrForbidden
	}
	return true, nil
}

func profileReadAllowed(user domain.User) bool {
	if !user.Active {
		return false
	}
	if user.Role != domain.RoleAdmin {
		return false
	}
	return strings.TrimSpace(user.ID) != "" && strings.TrimSpace(user.DisplayName) != ""
}

func newID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(buf)
}

func requireNonEmpty(values ...string) error {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return errors.New("required value is empty")
		}
	}
	return nil
}

func idempotencyKey(actor Actor, meta RequestMeta, operation string) string {
	if strings.TrimSpace(meta.IdempotencyKey) == "" {
		return ""
	}
	return actor.UserID + ":" + operation + ":" + strings.TrimSpace(meta.IdempotencyKey)
}

func idempotencyLookup(state *domain.State, key string) string {
	if key == "" || state.Idempotency == nil {
		return ""
	}
	return state.Idempotency[key]
}

func rememberIdempotency(state *domain.State, key, resourceID string) {
	if key == "" {
		return
	}
	if state.Idempotency == nil {
		state.Idempotency = map[string]string{}
	}
	state.Idempotency[key] = resourceID
}
