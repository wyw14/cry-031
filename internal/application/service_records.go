package application

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/wyw/cry031-volunteer/internal/domain"
)

type ServiceRecordFilter struct {
	UserID string
	TeamID string
	Status domain.ServiceStatus
	Sort   string
	Page   int
	Size   int
}

func (e *Engine) StartService(ctx context.Context, actor Actor, meta RequestMeta, claimID string) (domain.ServiceRecord, error) {
	var record domain.ServiceRecord
	err := e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		claim, ok := state.Claims[claimID]
		if !ok || claim.Status != domain.ClaimCheckedIn {
			return domain.ErrInvalidState
		}
		if actor.UserID != claim.UserID {
			activity := state.Activities[claim.ActivityID]
			if !canManageTeam(state, actor, activity.TeamID) {
				return domain.ErrForbidden
			}
		}
		for _, existing := range state.Records {
			if existing.ClaimID == claimID && existing.Status != domain.ServiceCorrectionOpen {
				return domain.ErrConflict
			}
		}
		record = domain.ServiceRecord{ID: newID("service"), ClaimID: claim.ID, ActivityID: claim.ActivityID, UserID: claim.UserID, Version: 1, Status: domain.ServiceStarted, StartedAt: &now, CreatedAt: now, UpdatedAt: now}
		state.Records[record.ID] = record
		return nil
	})
	return record, err
}

func (e *Engine) EndService(ctx context.Context, actor Actor, meta RequestMeta, recordID, photoNote string) (domain.ServiceRecord, error) {
	var result domain.ServiceRecord
	err := e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		record, ok := state.Records[recordID]
		if !ok {
			return domain.ErrNotFound
		}
		if actor.UserID != record.UserID && !canManageActivity(state, actor, record.ActivityID) {
			return domain.ErrForbidden
		}
		if record.Status != domain.ServiceStarted || record.StartedAt == nil {
			return domain.ErrInvalidState
		}
		end := now
		record.EndedAt = &end
		duration, err := record.Duration()
		if err != nil {
			return err
		}
		record.DurationMins = duration
		record.PhotoNote = strings.TrimSpace(photoNote)
		record.Status = domain.ServicePendingReview
		record.UpdatedAt = now
		state.Records[recordID] = record
		result = record
		return nil
	})
	return result, err
}

func (e *Engine) ConfirmService(ctx context.Context, actor Actor, meta RequestMeta, recordID string) error {
	var notifyUser string
	err := e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		record, ok := state.Records[recordID]
		if !ok {
			return domain.ErrNotFound
		}
		if !canManageActivity(state, actor, record.ActivityID) || record.Status != domain.ServicePendingReview {
			return domain.ErrForbidden
		}
		record.Status = domain.ServiceConfirmed
		record.ReviewerID = actor.UserID
		record.UpdatedAt = now
		state.Records[recordID] = record
		if record.SupersedesID != "" {
			original := state.Records[record.SupersedesID]
			original.Status = domain.ServiceCorrectionOpen
			original.UpdatedAt = now
			state.Records[original.ID] = original
		}
		e.addNotice(state, record.UserID, "service_confirmed", "你的服务记录已复核", now)
		notifyUser = record.UserID
		return nil
	})
	if err == nil {
		e.notify(ctx, notifyUser, "service_confirmed", "你的服务记录已复核")
	}
	return err
}

func (e *Engine) RequestCorrection(ctx context.Context, actor Actor, meta RequestMeta, recordID, note string, startedAt, endedAt time.Time) (domain.ServiceRecord, error) {
	var corrected domain.ServiceRecord
	err := e.transact(ctx, actor, meta, func(state *domain.State, now time.Time) error {
		original, ok := state.Records[recordID]
		if !ok {
			return domain.ErrNotFound
		}
		if original.Status != domain.ServiceConfirmed {
			return domain.ErrInvalidCorrection
		}
		if actor.UserID != original.UserID && !canManageActivity(state, actor, original.ActivityID) {
			return domain.ErrForbidden
		}
		candidate := domain.ServiceRecord{ID: newID("service"), SupersedesID: original.ID, ClaimID: original.ClaimID, ActivityID: original.ActivityID, UserID: original.UserID, Version: original.Version + 1, Status: domain.ServicePendingReview, StartedAt: &startedAt, EndedAt: &endedAt, PhotoNote: original.PhotoNote, CorrectionNote: strings.TrimSpace(note), CreatedAt: now, UpdatedAt: now}
		duration, err := candidate.Duration()
		if err != nil {
			return err
		}
		candidate.DurationMins = duration
		state.Records[candidate.ID] = candidate
		corrected = candidate
		return nil
	})
	return corrected, err
}

func (e *Engine) ListServiceRecords(ctx context.Context, filter ServiceRecordFilter) (Page[domain.ServiceRecord], error) {
	state, err := e.read(ctx)
	if err != nil {
		return Page[domain.ServiceRecord]{}, err
	}
	page, size := normalizePage(filter.Page, filter.Size)
	if filter.Sort != "" && filter.Sort != "updated_desc" && filter.Sort != "duration_desc" {
		return Page[domain.ServiceRecord]{}, domain.ErrInvalidFilter
	}
	items := make([]domain.ServiceRecord, 0)
	for _, record := range state.Records {
		if filter.UserID != "" && record.UserID != filter.UserID {
			continue
		}
		if filter.Status != "" && record.Status != filter.Status {
			continue
		}
		if filter.TeamID != "" {
			activity, ok := state.Activities[record.ActivityID]
			if !ok || activity.TeamID != filter.TeamID {
				continue
			}
		}
		items = append(items, record)
	}
	sort.Slice(items, func(i, j int) bool {
		if filter.Sort == "duration_desc" {
			return items[i].DurationMins > items[j].DurationMins
		}
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
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
	return Page[domain.ServiceRecord]{Items: items[start:end], Page: page, PageSize: size, Total: total}, nil
}

func (e *Engine) ExportCertificate(ctx context.Context, userID string) ([]byte, error) {
	state, err := e.read(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := mustUser(&state, userID); err != nil {
		return nil, err
	}
	latest := latestConfirmedRecords(state, userID)
	sort.Slice(latest, func(i, j int) bool { return latest[i].ActivityID < latest[j].ActivityID })
	var total int
	output := "service_id,activity_id,duration_minutes,status\n"
	for _, record := range latest {
		total += record.DurationMins
		output += fmt.Sprintf("%s,%s,%d,%s\n", record.ID, record.ActivityID, record.DurationMins, record.Status)
	}
	output += fmt.Sprintf("total,,%d,confirmed\n", total)
	return []byte(output), nil
}

func latestConfirmedRecords(state domain.State, userID string) []domain.ServiceRecord {
	byClaim := map[string]domain.ServiceRecord{}
	for _, record := range state.Records {
		if record.UserID != userID || record.Status != domain.ServiceConfirmed {
			continue
		}
		current, ok := byClaim[record.ClaimID]
		if !ok || record.Version > current.Version {
			byClaim[record.ClaimID] = record
		}
	}
	result := make([]domain.ServiceRecord, 0, len(byClaim))
	for _, record := range byClaim {
		result = append(result, record)
	}
	return result
}

func canManageActivity(state *domain.State, actor Actor, activityID string) bool {
	activity, ok := state.Activities[activityID]
	return ok && canManageTeam(state, actor, activity.TeamID)
}

func correctionPolicyAudit(startedAt, endedAt time.Time) bool {
	return endedAt.After(startedAt)
}
