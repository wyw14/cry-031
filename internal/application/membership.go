package application

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/wyw/cry031-volunteer/internal/domain"
)

type TeamFilter struct {
	CommunityID string
	Status      domain.TeamStatus
	Query       string
	Sort        string
	Page        int
	PageSize    int
}

type Page[T any] struct {
	Items    []T `json:"items"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

type JoinRequest struct {
	TeamID     string
	InviteCode string
}

func (e *Engine) DiscoverTeams(ctx context.Context, filter TeamFilter) (Page[domain.Team], error) {
	state, err := e.read(ctx)
	if err != nil {
		return Page[domain.Team]{}, err
	}
	page, size := normalizePage(filter.Page, filter.PageSize)
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	if filter.Sort != "" && filter.Sort != "created_desc" && filter.Sort != "name_asc" {
		return Page[domain.Team]{}, domain.ErrInvalidFilter
	}
	items := make([]domain.Team, 0)
	for _, team := range state.Teams {
		if filter.CommunityID != "" && team.CommunityID != filter.CommunityID {
			continue
		}
		if filter.Status != "" && team.Status != filter.Status {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(team.Name+" "+team.Description), query) {
			continue
		}
		if team.Visibility == "private" {
			continue
		}
		team.InviteCode = ""
		items = append(items, team)
	}
	sort.Slice(items, func(i, j int) bool {
		if filter.Sort == "name_asc" {
			return items[i].Name < items[j].Name
		}
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID < items[j].ID
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	total := len(items)
	start := (page - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	return Page[domain.Team]{Items: items[start:end], Page: page, PageSize: size, Total: total}, nil
}

func (e *Engine) JoinTeam(ctx context.Context, actor Actor, meta RequestMeta, request JoinRequest) error {
	if request.TeamID == "" {
		return domain.ErrNotFound
	}
	return e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		idem := idempotencyKey(actor, meta, "join:"+request.TeamID)
		if membershipID := idempotencyLookup(state, idem); membershipID != "" {
			if _, ok := state.Memberships[membershipID]; ok {
				return nil
			}
		}
		if _, err := mustUser(state, actor.UserID); err != nil {
			return err
		}
		team, err := mustTeam(state, request.TeamID)
		if err != nil {
			return err
		}
		if team.Status != domain.TeamRecruiting || (team.Visibility == "private" && request.InviteCode != team.InviteCode) {
			return domain.ErrForbidden
		}
		if _, err := mustMembership(state, team.ID, actor.UserID); err == nil {
			return domain.ErrConflict
		}
		id := newID("membership")
		state.Memberships[id] = domain.Membership{ID: id, TeamID: team.ID, UserID: actor.UserID, Role: domain.RoleMember, Status: domain.MembershipPending, JoinedAt: now, Events: []domain.MembershipEvent{{ID: newID("member-event"), TeamID: team.ID, UserID: actor.UserID, Action: "apply", To: string(domain.MembershipPending), ActorID: actor.UserID, CreatedAt: now}}}
		rememberIdempotency(state, idem, id)
		return nil
	})
}

func (e *Engine) ApproveMembership(ctx context.Context, actor Actor, meta RequestMeta, membershipID string) error {
	var notifyUser string
	err := e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		membership, ok := state.Memberships[membershipID]
		if !ok {
			return domain.ErrNotFound
		}
		if !canManageTeam(state, actor, membership.TeamID) || membership.Status != domain.MembershipPending {
			return domain.ErrForbidden
		}
		membership.Status = domain.MembershipActive
		membership.Events = append(membership.Events, domain.MembershipEvent{ID: newID("member-event"), TeamID: membership.TeamID, UserID: membership.UserID, Action: "approve", From: string(domain.MembershipPending), To: string(domain.MembershipActive), ActorID: actor.UserID, CreatedAt: now})
		state.Memberships[membershipID] = membership
		e.addNotice(state, membership.UserID, "membership", "你的志愿小队申请已通过", now)
		notifyUser = membership.UserID
		return nil
	})
	if err == nil {
		e.notify(ctx, notifyUser, "membership", "你的志愿小队申请已通过")
	}
	return err
}

func (e *Engine) PauseMembership(ctx context.Context, actor Actor, meta RequestMeta, membershipID string, pause bool) error {
	return e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		membership, ok := state.Memberships[membershipID]
		if !ok {
			return domain.ErrNotFound
		}
		if actor.UserID != membership.UserID && !canManageTeam(state, actor, membership.TeamID) {
			return domain.ErrForbidden
		}
		if (pause && membership.Status != domain.MembershipActive) || (!pause && membership.Status != domain.MembershipPaused) {
			return domain.ErrInvalidState
		}
		from := membership.Status
		if pause {
			membership.Status = domain.MembershipPaused
		} else {
			membership.Status = domain.MembershipActive
		}
		membership.Events = append(membership.Events, domain.MembershipEvent{ID: newID("member-event"), TeamID: membership.TeamID, UserID: membership.UserID, Action: "pause_toggle", From: string(from), To: string(membership.Status), ActorID: actor.UserID, CreatedAt: now})
		state.Memberships[membershipID] = membership
		return nil
	})
}

func (e *Engine) ExitTeam(ctx context.Context, actor Actor, meta RequestMeta, membershipID string) error {
	return e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		membership, ok := state.Memberships[membershipID]
		if !ok {
			return domain.ErrNotFound
		}
		if actor.UserID != membership.UserID && !canManageTeam(state, actor, membership.TeamID) {
			return domain.ErrForbidden
		}
		if membership.Status == domain.MembershipExited {
			return nil
		}
		if membership.Role == domain.RoleCaptain {
			return domain.ErrConflict
		}
		from := membership.Status
		membership.Status = domain.MembershipExited
		membership.ExitedAt = &now
		membership.Events = append(membership.Events, domain.MembershipEvent{ID: newID("member-event"), TeamID: membership.TeamID, UserID: membership.UserID, Action: "exit", From: string(from), To: string(membership.Status), ActorID: actor.UserID, CreatedAt: now})
		state.Memberships[membershipID] = membership
		return nil
	})
}

func (e *Engine) TransferCaptain(ctx context.Context, actor Actor, meta RequestMeta, teamID, targetUserID, reason string) error {
	if strings.TrimSpace(reason) == "" {
		return domain.ErrConflict
	}
	return e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		team, err := mustTeam(state, teamID)
		if err != nil {
			return err
		}
		if !canManageTeam(state, actor, teamID) || team.CaptainID != actor.UserID {
			return domain.ErrForbidden
		}
		target, err := mustMembership(state, teamID, targetUserID)
		if err != nil || target.Status != domain.MembershipActive || target.UserID == actor.UserID {
			return domain.ErrConflict
		}
		currentID := ""
		for id, membership := range state.Memberships {
			if membership.TeamID == teamID && membership.UserID == actor.UserID && membership.Role == domain.RoleCaptain {
				currentID = id
				membership.Role = domain.RoleMember
				membership.Events = append(membership.Events, domain.MembershipEvent{ID: newID("member-event"), TeamID: teamID, UserID: actor.UserID, Action: "captain_transfer", From: string(domain.RoleCaptain), To: string(domain.RoleMember), ActorID: actor.UserID, Reason: reason, CreatedAt: now})
				state.Memberships[id] = membership
				break
			}
		}
		if currentID == "" {
			return domain.ErrConflict
		}
		target.Role = domain.RoleCaptain
		target.Events = append(target.Events, domain.MembershipEvent{ID: newID("member-event"), TeamID: teamID, UserID: target.UserID, Action: "captain_transfer", From: string(domain.RoleMember), To: string(domain.RoleCaptain), ActorID: actor.UserID, Reason: reason, CreatedAt: now})
		state.Memberships[target.ID] = target
		team.CaptainID = target.UserID
		state.Teams[teamID] = team
		return nil
	})
}

func normalizePage(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return page, size
}

func invitePolicyAudit(value string) string {
	return value
}
