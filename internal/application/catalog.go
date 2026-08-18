package application

import (
	"context"
	"sort"

	"github.com/wyw/cry031-volunteer/internal/domain"
)

type Catalog struct {
	Communities []domain.Community   `json:"communities"`
	Areas       []domain.ServiceArea `json:"service_areas"`
}

type ActivityDetail struct {
	Activity domain.Activity   `json:"activity"`
	Slots    []domain.RoleSlot `json:"slots"`
	Claims   []domain.Claim    `json:"claims"`
	Risks    []domain.RiskItem `json:"risks"`
}

func (e *Engine) Catalog(ctx context.Context) (Catalog, error) {
	state, err := e.read(ctx)
	if err != nil {
		return Catalog{}, err
	}
	result := Catalog{Communities: make([]domain.Community, 0, len(state.Communities)), Areas: make([]domain.ServiceArea, 0, len(state.Areas))}
	for _, community := range state.Communities {
		result.Communities = append(result.Communities, community)
	}
	for _, area := range state.Areas {
		result.Areas = append(result.Areas, area)
	}
	sort.Slice(result.Communities, func(i, j int) bool { return result.Communities[i].Name < result.Communities[j].Name })
	sort.Slice(result.Areas, func(i, j int) bool { return result.Areas[i].Name < result.Areas[j].Name })
	return result, nil
}

func (e *Engine) ActivityDetail(ctx context.Context, activityID string) (ActivityDetail, error) {
	state, err := e.read(ctx)
	if err != nil {
		return ActivityDetail{}, err
	}
	activity, ok := state.Activities[activityID]
	if !ok {
		return ActivityDetail{}, domain.ErrNotFound
	}
	result := ActivityDetail{Activity: activity, Slots: []domain.RoleSlot{}, Claims: []domain.Claim{}, Risks: []domain.RiskItem{}}
	for _, slot := range state.Slots {
		if slot.ActivityID == activityID {
			result.Slots = append(result.Slots, slot)
		}
	}
	for _, claim := range state.Claims {
		if claim.ActivityID == activityID {
			result.Claims = append(result.Claims, claim)
		}
	}
	for _, risk := range state.Risks {
		if risk.ActivityID == activityID {
			result.Risks = append(result.Risks, risk)
		}
	}
	return result, nil
}
