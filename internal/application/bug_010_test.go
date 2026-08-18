package application

import (
	"context"
	"github.com/wyw/cry031-volunteer/internal/domain"
	"testing"
)

func TestDashboardRecapsExcludeCancelled(t *testing.T) {
	e, s, _ := fixture()
	st, _ := s.Snapshot(context.Background())
	a := st.Activities["activity-visit"]
	a.Status = domain.ActivityCancelled
	st.Activities[a.ID] = a
	if err := s.Commit(context.Background(), st.Version, st, nil); err != nil {
		t.Fatal(err)
	}
	d, err := e.LeaderDashboard(context.Background(), captain(), "team-riverside")
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range d.CompletedRecaps {
		if x.ID == a.ID {
			t.Fatal("cancelled activity shown in completed recaps")
		}
	}
}
