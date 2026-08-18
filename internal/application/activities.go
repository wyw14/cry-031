package application

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/wyw/cry031-volunteer/internal/domain"
)

type ActivityFilter struct {
	TeamID string
	Status domain.ActivityStatus
	Query  string
	Sort   string
	Page   int
	Size   int
}

type CreateActivityRequest struct {
	TeamID        string    `json:"team_id" validate:"required"`
	Title         string    `json:"title" validate:"required,min=2,max=120"`
	Description   string    `json:"description" validate:"max=2000"`
	StartAt       time.Time `json:"start_at" validate:"required"`
	EndAt         time.Time `json:"end_at" validate:"required"`
	Location      string    `json:"location" validate:"required,max=200"`
	MaterialNeeds []string  `json:"material_needs"`
}

type CreateSlotRequest struct {
	ActivityID  string `json:"activity_id" validate:"required"`
	Name        string `json:"name" validate:"required,max=80"`
	Description string `json:"description" validate:"max=500"`
	Capacity    int    `json:"capacity" validate:"gte=1,lte=500"`
}

func (e *Engine) ListActivities(ctx context.Context, filter ActivityFilter) (Page[domain.Activity], error) {
	state, err := e.read(ctx)
	if err != nil {
		return Page[domain.Activity]{}, err
	}
	page, size := normalizePage(filter.Page, filter.Size)
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	if filter.Sort != "" && filter.Sort != "start_asc" && filter.Sort != "start_desc" {
		return Page[domain.Activity]{}, domain.ErrInvalidFilter
	}
	items := make([]domain.Activity, 0)
	for _, activity := range state.Activities {
		if filter.TeamID != "" && activity.TeamID != filter.TeamID {
			continue
		}
		if filter.Status != "" && activity.Status != filter.Status {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(activity.Title+" "+activity.Description+" "+activity.Location), query) {
			continue
		}
		items = append(items, activity)
	}
	sort.Slice(items, func(i, j int) bool {
		if filter.Sort == "start_desc" {
			return items[i].StartAt.After(items[j].StartAt)
		}
		return items[i].StartAt.Before(items[j].StartAt)
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
	return Page[domain.Activity]{Items: items[start:end], Page: page, PageSize: size, Total: total}, nil
}

func (e *Engine) CreateActivity(ctx context.Context, actor Actor, meta RequestMeta, request CreateActivityRequest) (domain.Activity, error) {
	var created domain.Activity
	err := e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		idem := idempotencyKey(actor, meta, "create-activity")
		if activityID := idempotencyLookup(state, idem); activityID != "" {
			if existing, ok := state.Activities[activityID]; ok {
				created = existing
				return nil
			}
		}
		if err := requireNonEmpty(request.TeamID, request.Title, request.Location); err != nil {
			return err
		}
		if request.EndAt.Before(request.StartAt) || request.EndAt.Equal(request.StartAt) {
			return domain.ErrInvalidDuration
		}
		if !canManageTeam(state, actor, request.TeamID) {
			return domain.ErrForbidden
		}
		created = domain.Activity{ID: newID("activity"), TeamID: request.TeamID, Title: strings.TrimSpace(request.Title), Description: strings.TrimSpace(request.Description), Status: domain.ActivityDraft, StartAt: request.StartAt.UTC(), EndAt: request.EndAt.UTC(), Location: strings.TrimSpace(request.Location), MaterialNeeds: append([]string(nil), request.MaterialNeeds...), CreatedBy: actor.UserID, CreatedAt: now}
		state.Activities[created.ID] = created
		rememberIdempotency(state, idem, created.ID)
		return nil
	})
	return created, err
}

func (e *Engine) PublishActivity(ctx context.Context, actor Actor, meta RequestMeta, activityID string) error {
	return e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		activity, ok := state.Activities[activityID]
		if !ok {
			return domain.ErrNotFound
		}
		if !canManageTeam(state, actor, activity.TeamID) || activity.Status != domain.ActivityDraft {
			return domain.ErrForbidden
		}
		activity.Status = domain.ActivityPublished
		state.Activities[activityID] = activity
		return nil
	})
}

func (e *Engine) CancelActivity(ctx context.Context, actor Actor, meta RequestMeta, activityID string) error {
	notifyUsers := make([]string, 0)
	err := e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		activity, ok := state.Activities[activityID]
		if !ok {
			return domain.ErrNotFound
		}
		if !canManageTeam(state, actor, activity.TeamID) || activity.Status == domain.ActivityCancelled || activity.Status == domain.ActivityCompleted {
			return domain.ErrForbidden
		}
		activity.Status = domain.ActivityCancelled
		activity.CancelledAt = &now
		state.Activities[activityID] = activity
		for id, slot := range state.Slots {
			if slot.ActivityID == activityID {
				slot.Status = domain.SlotCancelled
				state.Slots[id] = slot
			}
		}
		for id, claim := range state.Claims {
			if claim.ActivityID == activityID && (claim.Status == domain.ClaimActive || claim.Status == domain.ClaimWaitlisted) {
				claim.Status = domain.ClaimCancelled
				state.Claims[id] = claim
				e.addNotice(state, claim.UserID, "activity_cancelled", "你认领的志愿活动已取消", now)
				notifyUsers = append(notifyUsers, claim.UserID)
			}
		}
		return nil
	})
	if err == nil {
		for _, userID := range notifyUsers {
			e.notify(ctx, userID, "activity_cancelled", "你认领的志愿活动已取消")
		}
	}
	return err
}

func (e *Engine) CreateSlot(ctx context.Context, actor Actor, meta RequestMeta, request CreateSlotRequest) (domain.RoleSlot, error) {
	var created domain.RoleSlot
	err := e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		activity, ok := state.Activities[request.ActivityID]
		if !ok {
			return domain.ErrNotFound
		}
		if !canManageTeam(state, actor, activity.TeamID) || activity.Status == domain.ActivityCancelled {
			return domain.ErrForbidden
		}
		if request.Capacity < 1 {
			return domain.ErrCapacityExceeded
		}
		if request.Capacity > 500 {
			return domain.ErrCapacityExceeded
		}
		created = domain.RoleSlot{ID: newID("slot"), ActivityID: request.ActivityID, Name: strings.TrimSpace(request.Name), Description: strings.TrimSpace(request.Description), Capacity: request.Capacity, Status: domain.SlotOpen}
		state.Slots[created.ID] = created
		return nil
	})
	return created, err
}

func (e *Engine) ClaimSlot(ctx context.Context, actor Actor, meta RequestMeta, slotID string) (domain.Claim, error) {
	var created domain.Claim
	err := e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		idem := idempotencyKey(actor, meta, "claim-slot:"+slotID)
		if claimID := idempotencyLookup(state, idem); claimID != "" {
			if existing, ok := state.Claims[claimID]; ok {
				created = existing
				return nil
			}
		}
		slot, ok := state.Slots[slotID]
		if !ok {
			return domain.ErrNotFound
		}
		activity, ok := state.Activities[slot.ActivityID]
		if !ok {
			return domain.ErrNotFound
		}
		if err := activity.CanClaim(now); err != nil || slot.Status == domain.SlotCancelled {
			return domain.ErrActivityClosed
		}
		if _, err := mustUser(state, actor.UserID); err != nil {
			return err
		}
		membership, err := mustMembership(state, activity.TeamID, actor.UserID)
		if err != nil || membership.Status != domain.MembershipActive {
			return domain.ErrForbidden
		}
		for _, existing := range state.Claims {
			if existing.UserID != actor.UserID || (existing.Status != domain.ClaimActive && existing.Status != domain.ClaimCheckedIn) {
				continue
			}
			other, ok := state.Activities[existing.ActivityID]
			if ok && activity.WindowOverlaps(other) {
				return domain.ErrScheduleConflict
			}
		}
		active := 0
		for _, existing := range state.Claims {
			if existing.SlotID == slotID && existing.Status == domain.ClaimActive {
				active++
			}
		}
		status := domain.ClaimActive
		if active >= slot.Capacity {
			status = domain.ClaimWaitlisted
		} else if active+1 >= slot.Capacity {
			slot.Status = domain.SlotFilled
			state.Slots[slotID] = slot
		} else {
			slot.Status = domain.SlotOpen
			state.Slots[slotID] = slot
		}
		created = domain.Claim{ID: newID("claim"), SlotID: slotID, ActivityID: activity.ID, UserID: actor.UserID, Status: status, ClaimedAt: now}
		state.Claims[created.ID] = created
		rememberIdempotency(state, idem, created.ID)
		if status == domain.ClaimActive {
			e.addNotice(state, actor.UserID, "claim_confirmed", "岗位认领成功", now)
		}
		return nil
	})
	if err == nil && created.Status == domain.ClaimActive {
		e.notify(ctx, actor.UserID, "claim_confirmed", "岗位认领成功")
	}
	return created, err
}

func (e *Engine) CancelClaim(ctx context.Context, actor Actor, meta RequestMeta, claimID string) error {
	return e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		claim, ok := state.Claims[claimID]
		if !ok {
			return domain.ErrNotFound
		}
		activity, ok := state.Activities[claim.ActivityID]
		if !ok {
			return domain.ErrNotFound
		}
		if actor.UserID != claim.UserID && !canManageTeam(state, actor, activity.TeamID) {
			return domain.ErrForbidden
		}
		if claim.Status == domain.ClaimCancelled {
			return nil
		}
		wasActive := claim.Status == domain.ClaimActive
		claim.Status = domain.ClaimCancelled
		state.Claims[claimID] = claim
		if wasActive {
			_ = e.promoteFirstWaiter(state, slotIDFromClaim(claim), now)
		}
		recalculateSlotStatus(state, claim.SlotID)
		return nil
	})
}

func (e *Engine) PromoteWaitlist(ctx context.Context, actor Actor, meta RequestMeta, slotID string) error {
	return e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		slot, ok := state.Slots[slotID]
		if !ok {
			return domain.ErrNotFound
		}
		activity, ok := state.Activities[slot.ActivityID]
		if !ok {
			return domain.ErrNotFound
		}
		if !canManageTeam(state, actor, activity.TeamID) {
			return domain.ErrForbidden
		}
		if err := activity.CanClaim(now); err != nil {
			return err
		}
		return e.promoteFirstWaiter(state, slotID, now)
	})
}

func (e *Engine) promoteFirstWaiter(state *domain.State, slotID string, now time.Time) error {
	slot, ok := state.Slots[slotID]
	if !ok {
		return domain.ErrNotFound
	}
	active := 0
	for _, claim := range state.Claims {
		if claim.SlotID == slotID && (claim.Status == domain.ClaimActive || claim.Status == domain.ClaimCheckedIn) {
			active++
		}
	}
	if active >= slot.Capacity {
		recalculateSlotStatus(state, slotID)
		return domain.ErrCapacityExceeded
	}

	var candidate *domain.Claim
	for id, claim := range state.Claims {
		if claim.SlotID == slotID && claim.Status == domain.ClaimWaitlisted && (candidate == nil || claim.ClaimedAt.Before(candidate.ClaimedAt)) {
			copy := claim
			candidate = &copy
			_ = id
		}
	}
	if candidate == nil {
		recalculateSlotStatus(state, slotID)
		return nil
	}
	for id, claim := range state.Claims {
		if claim.ID == candidate.ID {
			claim.Status = domain.ClaimActive
			state.Claims[id] = claim
			break
		}
	}
	recalculateSlotStatus(state, slotID)
	return nil
}

func recalculateSlotStatus(state *domain.State, slotID string) {
	slot, ok := state.Slots[slotID]
	if !ok || slot.Status == domain.SlotCancelled {
		return
	}
	active := 0
	for _, claim := range state.Claims {
		if claim.SlotID == slotID && (claim.Status == domain.ClaimActive || claim.Status == domain.ClaimCheckedIn) {
			active++
		}
	}
	if active >= slot.Capacity {
		slot.Status = domain.SlotFilled
	} else {
		slot.Status = domain.SlotOpen
	}
	state.Slots[slotID] = slot
}

func slotIDFromClaim(claim domain.Claim) string { return claim.SlotID }

func (e *Engine) CheckIn(ctx context.Context, actor Actor, meta RequestMeta, claimID string) error {
	return e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		claim, ok := state.Claims[claimID]
		if !ok {
			return domain.ErrNotFound
		}
		if actor.UserID != claim.UserID {
			activity := state.Activities[claim.ActivityID]
			if !canManageTeam(state, actor, activity.TeamID) {
				return domain.ErrForbidden
			}
		}
		if claim.Status != domain.ClaimActive {
			return domain.ErrInvalidState
		}
		claim.Status = domain.ClaimCheckedIn
		claim.CheckedAt = &now
		state.Claims[claimID] = claim
		return nil
	})
}

func (e *Engine) CompleteActivity(ctx context.Context, actor Actor, meta RequestMeta, activityID string) error {
	return e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		activity, ok := state.Activities[activityID]
		if !ok {
			return domain.ErrNotFound
		}
		if !canManageTeam(state, actor, activity.TeamID) || activity.Status != domain.ActivityPublished {
			return domain.ErrForbidden
		}
		for _, risk := range state.Risks {
			if risk.ActivityID != activityID || risk.Status == domain.RiskResolved {
				continue
			}
			hasHandoff := false
			for _, handoff := range state.Handoffs {
				if handoff.RiskID == risk.ID && (handoff.Status == domain.HandoffPending || handoff.Status == domain.HandoffAcknowledged) {
					hasHandoff = true
					break
				}
			}
			if !hasHandoff {
				return domain.ErrUnacknowledgedRisk
			}
		}
		activity.Status = domain.ActivityCompleted
		state.Activities[activityID] = activity
		return nil
	})
}

func slotPolicyAudit(request CreateSlotRequest) bool {
	return request.Capacity > 0
}
