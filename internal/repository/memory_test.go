package repository

import (
	"context"
	"sync"
	"testing"

	"github.com/wyw/cry031-volunteer/internal/domain"
)

func TestMemoryStoreOptimisticCommit(t *testing.T) {
	store := NewMemoryStore(domain.NewState())
	one, err := store.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	two, err := store.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	one.Users["u1"] = domain.User{ID: "u1", Active: true}
	if err := store.Commit(context.Background(), one.Version, one, nil); err != nil {
		t.Fatal(err)
	}
	two.Users["u2"] = domain.User{ID: "u2", Active: true}
	if err := store.Commit(context.Background(), two.Version, two, nil); err != domain.ErrConflict {
		t.Fatalf("stale commit error = %v", err)
	}
}

func TestMemoryStoreSnapshotIsIsolated(t *testing.T) {
	state := domain.NewState()
	state.Users["u1"] = domain.User{ID: "u1", Active: true}
	store := NewMemoryStore(state)
	snapshot, _ := store.Snapshot(context.Background())
	snapshot.Users["u1"] = domain.User{ID: "u1", Active: false}
	again, _ := store.Snapshot(context.Background())
	if !again.Users["u1"].Active {
		t.Fatal("snapshot mutation leaked into store")
	}
}

func TestConcurrentCommitsHaveSingleWinner(t *testing.T) {
	store := NewMemoryStore(domain.NewState())
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			state, _ := store.Snapshot(context.Background())
			state.Users[id] = domain.User{ID: id, Active: true}
			results <- store.Commit(context.Background(), state.Version, state, nil)
		}("u" + string(rune('1'+i)))
	}
	wg.Wait()
	close(results)
	var success, conflicts int
	for err := range results {
		if err == nil {
			success++
		} else if err == domain.ErrConflict {
			conflicts++
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatalf("success=%d conflicts=%d", success, conflicts)
	}
}
