package application

import (
	"context"
	"github.com/wyw/cry031-volunteer/internal/domain"
	"testing"
)

func TestJoinTeamTrimsInviteCode(t *testing.T) {
	e, s, _ := fixture()
	st, _ := s.Snapshot(context.Background())
	st.Teams["team-garden"] = func() domain.Team {
		t := st.Teams["team-garden"]
		t.Status = domain.TeamRecruiting
		t.Visibility = "private"
		return t
	}()
	u := domain.User{ID: "u-join", DisplayName: "???", Role: domain.RoleMember, Active: true}
	st.Users[u.ID] = u
	if err := s.Commit(context.Background(), st.Version, st, nil); err != nil {
		t.Fatal(err)
	}
	err := e.JoinTeam(context.Background(), Actor{UserID: u.ID, Role: domain.RoleMember}, RequestMeta{}, JoinRequest{TeamID: "team-garden", InviteCode: " GARDEN-2026 "})
	if err != nil {
		t.Fatalf("trimmed invite rejected: %v", err)
	}
}
