package application

import (
	"context"
	"github.com/wyw/cry031-volunteer/internal/domain"
	"testing"
	"time"
)

func TestRequestCorrectionRejectsFutureTimes(t *testing.T) {
	e, s, clock := fixture()
	st, _ := s.Snapshot(context.Background())
	start := clock.Now().Add(-2 * time.Hour)
	end := clock.Now().Add(-time.Hour)
	st.Records["r-confirmed"] = domain.ServiceRecord{ID: "r-confirmed", ClaimID: "claim-x", ActivityID: "activity-visit", UserID: "u-volunteer", Version: 1, Status: domain.ServiceConfirmed, StartedAt: &start, EndedAt: &end}
	if err := s.Commit(context.Background(), st.Version, st, nil); err != nil {
		t.Fatal(err)
	}
	_, err := e.RequestCorrection(context.Background(), volunteer(), RequestMeta{}, "r-confirmed", "???", clock.Now().Add(time.Hour), clock.Now().Add(2*time.Hour))
	if err == nil {
		t.Fatal("future correction accepted")
	}
}
