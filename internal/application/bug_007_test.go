package application

import (
	"context"
	"github.com/wyw/cry031-volunteer/internal/domain"
	"testing"
	"time"
)

func TestAddFollowUpRejectsInactiveOwner(t *testing.T) {
	e, s, clock := fixture()
	st, _ := s.Snapshot(context.Background())
	u := st.Users["u-volunteer"]
	u.Active = false
	st.Users[u.ID] = u
	if err := s.Commit(context.Background(), st.Version, st, nil); err != nil {
		t.Fatal(err)
	}
	_, err := e.AddFollowUp(context.Background(), captain(), RequestMeta{}, "risk-railing", "????", "u-volunteer", clock.Now().Add(time.Hour))
	if err == nil {
		t.Fatal("inactive owner accepted")
	}
	_ = domain.ErrForbidden
}
