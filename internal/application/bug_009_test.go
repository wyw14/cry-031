package application

import (
	"context"
	"github.com/wyw/cry031-volunteer/internal/domain"
	"testing"
)

func TestCompletedHandoffCannotBeAcknowledged(t *testing.T) {
	e, s, _ := fixture()
	st, _ := s.Snapshot(context.Background())
	h := st.Handoffs["handoff-railing"]
	h.Status = domain.HandoffCompleted
	st.Handoffs[h.ID] = h
	if err := s.Commit(context.Background(), st.Version, st, nil); err != nil {
		t.Fatal(err)
	}
	if err := e.AcknowledgeHandoff(context.Background(), volunteer(), RequestMeta{}, h.ID); err == nil {
		t.Fatal("completed handoff acknowledged")
	}
}
