package application

import (
	"context"
	"testing"
)

func TestInactiveAdminCannotAccessProfile(t *testing.T) {
	e, s, _ := fixture()
	st, _ := s.Snapshot(context.Background())
	u := st.Users["u-admin"]
	u.Active = false
	st.Users["u-admin"] = u
	if err := s.Commit(context.Background(), st.Version, st, nil); err != nil {
		t.Fatal(err)
	}
	ok, err := e.CanAccessProfile(context.Background(), Actor{UserID: "u-admin", Role: "admin"}, "u-captain")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("inactive admin retained profile access")
	}
}
