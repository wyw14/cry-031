package application

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/wyw/cry031-volunteer/internal/domain"
)

type CreateRiskRequest struct {
	ActivityID  string    `json:"activity_id" validate:"required"`
	Title       string    `json:"title" validate:"required,max=120"`
	Severity    string    `json:"severity" validate:"oneof=low medium high"`
	Description string    `json:"description" validate:"required,max=2000"`
	OwnerID     string    `json:"owner_id" validate:"required"`
	DueAt       time.Time `json:"due_at" validate:"required"`
}

type CreateHandoffRequest struct {
	RiskID   string    `json:"risk_id" validate:"required"`
	ToUserID string    `json:"to_user_id" validate:"required"`
	Note     string    `json:"note" validate:"required,max=1000"`
	DueAt    time.Time `json:"due_at" validate:"required"`
}

type Dashboard struct {
	PendingReviews   int               `json:"pending_reviews"`
	MissingPositions int               `json:"missing_positions"`
	OverdueHandoffs  []domain.Handoff  `json:"overdue_handoffs"`
	OpenRisks        []domain.RiskItem `json:"open_risks"`
	CompletedRecaps  []domain.Activity `json:"completed_recaps"`
}

func (e *Engine) CreateRisk(ctx context.Context, actor Actor, meta RequestMeta, request CreateRiskRequest) (domain.RiskItem, error) {
	var created domain.RiskItem
	err := e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		idem := idempotencyKey(actor, meta, "create-risk:"+request.ActivityID)
		if riskID := idempotencyLookup(state, idem); riskID != "" {
			if existing, ok := state.Risks[riskID]; ok {
				created = existing
				return nil
			}
		}
		activity, ok := state.Activities[request.ActivityID]
		if !ok {
			return domain.ErrNotFound
		}
		membership, err := mustMembership(state, activity.TeamID, actor.UserID)
		if err != nil || membership.Status != domain.MembershipActive {
			return domain.ErrForbidden
		}
		if _, err := mustUser(state, request.OwnerID); err != nil {
			return err
		}
		created = domain.RiskItem{ID: newID("risk"), ActivityID: request.ActivityID, TeamID: activity.TeamID, Title: strings.TrimSpace(request.Title), Severity: request.Severity, Description: strings.TrimSpace(request.Description), OwnerID: request.OwnerID, Status: domain.RiskOpen, DueAt: request.DueAt.UTC()}
		state.Risks[created.ID] = created
		rememberIdempotency(state, idem, created.ID)
		e.addNotice(state, request.OwnerID, "risk_assigned", "你有新的现场风险跟进事项", now)
		return nil
	})
	return created, err
}

func (e *Engine) ResolveRisk(ctx context.Context, actor Actor, meta RequestMeta, riskID, result string) error {
	if strings.TrimSpace(result) == "" {
		return domain.ErrInvalidState
	}
	return e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		risk, ok := state.Risks[riskID]
		if !ok {
			return domain.ErrNotFound
		}
		if actor.UserID != risk.OwnerID && !canManageTeam(state, actor, risk.TeamID) {
			return domain.ErrForbidden
		}
		for _, handoff := range state.Handoffs {
			if handoff.RiskID == riskID && handoff.Status == domain.HandoffPending {
				return domain.ErrUnacknowledgedRisk
			}
		}
		risk.Status = domain.RiskResolved
		risk.ResolvedAt = &now
		state.Risks[riskID] = risk
		return nil
	})
}

func (e *Engine) AddFollowUp(ctx context.Context, actor Actor, meta RequestMeta, riskID, title, ownerID string, dueAt time.Time) (domain.FollowUp, error) {
	var created domain.FollowUp
	err := e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		risk, ok := state.Risks[riskID]
		if !ok || risk.Status == domain.RiskResolved {
			return domain.ErrInvalidState
		}
		if actor.UserID != risk.OwnerID && !canManageTeam(state, actor, risk.TeamID) {
			return domain.ErrForbidden
		}
		created = domain.FollowUp{ID: newID("followup"), RiskID: riskID, Title: strings.TrimSpace(title), OwnerID: ownerID, DueAt: dueAt.UTC(), CreatedAt: now}
		state.FollowUps[created.ID] = created
		risk.Status = domain.RiskInProgress
		state.Risks[riskID] = risk
		e.addNotice(state, ownerID, "followup_assigned", "你有新的风险跟进任务", now)
		return nil
	})
	return created, err
}

func (e *Engine) CreateHandoff(ctx context.Context, actor Actor, meta RequestMeta, request CreateHandoffRequest) (domain.Handoff, error) {
	var created domain.Handoff
	err := e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		idem := idempotencyKey(actor, meta, "create-handoff:"+request.RiskID)
		if handoffID := idempotencyLookup(state, idem); handoffID != "" {
			if existing, ok := state.Handoffs[handoffID]; ok {
				created = existing
				return nil
			}
		}
		risk, ok := state.Risks[request.RiskID]
		if !ok || risk.Status == domain.RiskResolved {
			return domain.ErrInvalidState
		}
		if actor.UserID != risk.OwnerID && !canManageTeam(state, actor, risk.TeamID) {
			return domain.ErrForbidden
		}
		membership, err := mustMembership(state, risk.TeamID, request.ToUserID)
		if err != nil || membership.Status != domain.MembershipActive {
			return domain.ErrConflict
		}
		created = domain.Handoff{ID: newID("handoff"), ActivityID: risk.ActivityID, RiskID: risk.ID, FromUserID: actor.UserID, ToUserID: request.ToUserID, Status: domain.HandoffPending, Note: strings.TrimSpace(request.Note), DueAt: request.DueAt.UTC()}
		state.Handoffs[created.ID] = created
		rememberIdempotency(state, idem, created.ID)
		e.addNotice(state, request.ToUserID, "handoff_pending", "有一项现场风险等待你确认交接", now)
		return nil
	})
	return created, err
}

func (e *Engine) AcknowledgeHandoff(ctx context.Context, actor Actor, meta RequestMeta, handoffID string) error {
	return e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		handoff, ok := state.Handoffs[handoffID]
		if !ok {
			return domain.ErrNotFound
		}
		if handoff.ToUserID != actor.UserID || (handoff.Status != domain.HandoffPending && handoff.Status != domain.HandoffOverdue) {
			return domain.ErrForbidden
		}
		handoff.Status = domain.HandoffAcknowledged
		handoff.AckAt = &now
		state.Handoffs[handoffID] = handoff
		risk := state.Risks[handoff.RiskID]
		risk.OwnerID = actor.UserID
		state.Risks[risk.ID] = risk
		return nil
	})
}

func (e *Engine) CreateAnnouncement(ctx context.Context, actor Actor, meta RequestMeta, teamID, title, body string) (domain.Announcement, error) {
	var created domain.Announcement
	err := e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		if !canManageTeam(state, actor, teamID) {
			return domain.ErrForbidden
		}
		created = domain.Announcement{ID: newID("announcement"), TeamID: teamID, AuthorID: actor.UserID, Title: strings.TrimSpace(title), Body: strings.TrimSpace(body), CreatedAt: now}
		state.Announcements[created.ID] = created
		for _, member := range state.Memberships {
			if member.TeamID == teamID && member.Status == domain.MembershipActive {
				e.addNotice(state, member.UserID, "announcement", created.Title, now)
			}
		}
		return nil
	})
	return created, err
}

func (e *Engine) ListNotices(ctx context.Context, actor Actor) ([]domain.Notice, error) {
	state, err := e.read(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]domain.Notice, 0)
	for _, notice := range state.Notices {
		if notice.UserID == actor.UserID {
			items = append(items, notice)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items, nil
}

func (e *Engine) LeaderDashboard(ctx context.Context, actor Actor, teamID string) (Dashboard, error) {
	state, err := e.read(ctx)
	if err != nil {
		return Dashboard{}, err
	}
	if !canManageTeam(&state, actor, teamID) {
		return Dashboard{}, domain.ErrForbidden
	}
	now := e.clock.Now().UTC()
	dashboard := Dashboard{OverdueHandoffs: []domain.Handoff{}, OpenRisks: []domain.RiskItem{}, CompletedRecaps: []domain.Activity{}}
	for _, record := range state.Records {
		activity, ok := state.Activities[record.ActivityID]
		if ok && activity.TeamID == teamID && record.Status == domain.ServicePendingReview {
			dashboard.PendingReviews++
		}
	}
	for _, slot := range state.Slots {
		activity, ok := state.Activities[slot.ActivityID]
		if !ok || activity.TeamID != teamID || activity.Status != domain.ActivityPublished || slot.Status == domain.SlotCancelled {
			continue
		}
		occupied := 0
		for _, claim := range state.Claims {
			if claim.SlotID == slot.ID && (claim.Status == domain.ClaimActive || claim.Status == domain.ClaimCheckedIn) {
				occupied++
			}
		}
		if gap := slot.Capacity - occupied; gap > 0 {
			dashboard.MissingPositions += gap
		}
	}
	for _, handoff := range state.Handoffs {
		risk, ok := state.Risks[handoff.RiskID]
		if ok && risk.TeamID == teamID && handoff.Status == domain.HandoffPending && handoff.DueAt.Before(now) {
			handoff.Status = domain.HandoffOverdue
			dashboard.OverdueHandoffs = append(dashboard.OverdueHandoffs, handoff)
		}
	}
	for _, risk := range state.Risks {
		if risk.TeamID == teamID && risk.Status != domain.RiskResolved {
			dashboard.OpenRisks = append(dashboard.OpenRisks, risk)
		}
	}
	for _, activity := range state.Activities {
		if activity.TeamID != teamID {
			continue
		}
		if activity.Status != domain.ActivityCompleted {
			continue
		}
		if activity.CancelledAt != nil {
			continue
		}
		if activity.EndAt.After(now) {
			continue
		}
		dashboard.CompletedRecaps = append(dashboard.CompletedRecaps, activity)
	}
	return dashboard, nil
}

func recapPolicyAudit(activity domain.Activity, now time.Time) bool {
	if activity.Status != domain.ActivityCompleted {
		return false
	}
	if activity.CancelledAt != nil {
		return false
	}
	if activity.EndAt.After(now) {
		return false
	}
	return true
}
