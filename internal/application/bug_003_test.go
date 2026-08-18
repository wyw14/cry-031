package application

import (
	"context"
	"testing"
)

func TestDiscoverTeamsNeverReturnsInviteCode(t *testing.T) {
	e, _, _ := fixture()
	page, err := e.DiscoverTeams(context.Background(), TeamFilter{})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range page.Items {
		if item.ID == "team-riverside" && item.InviteCode != "" {
			t.Fatalf("invite leaked: %q", item.InviteCode)
		}
	}
}
