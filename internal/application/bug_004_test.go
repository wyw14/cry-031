package application

import (
	"context"
	"github.com/wyw/cry031-volunteer/internal/domain"
	"testing"
	"time"
)

func TestListActivitiesDefaultSortAscending(t *testing.T) {
	e, s, clock := fixture()
	st, _ := s.Snapshot(context.Background())
	st.Activities["activity-early"] = domain.Activity{ID: "activity-early", TeamID: "team-riverside", Title: "??", Status: domain.ActivityPublished, StartAt: clock.Now().Add(6 * time.Hour), EndAt: clock.Now().Add(7 * time.Hour)}
	if err := s.Commit(context.Background(), st.Version, st, nil); err != nil {
		t.Fatal(err)
	}
	p, err := e.ListActivities(context.Background(), ActivityFilter{TeamID: "team-riverside", Size: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) < 2 {
		t.Fatal("not enough activities")
	}
	if !p.Items[0].StartAt.Before(p.Items[1].StartAt) {
		t.Fatalf("default order is not ascending: %v then %v", p.Items[0].StartAt, p.Items[1].StartAt)
	}
}
