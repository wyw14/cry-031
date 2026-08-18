package repository

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/wyw/cry031-volunteer/internal/domain"
)

type MemoryStore struct {
	mu    sync.RWMutex
	state domain.State
}

func NewMemoryStore(initial domain.State) *MemoryStore {
	if initial.Users == nil {
		initial = domain.NewState()
	}
	return &MemoryStore{state: cloneState(initial)}
}

func (s *MemoryStore) Snapshot(ctx context.Context) (domain.State, error) {
	if err := ctx.Err(); err != nil {
		return domain.State{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneState(s.state), nil
}

func (s *MemoryStore) Commit(ctx context.Context, expected uint64, next domain.State, _ []domain.AuditEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.Version != expected {
		return domain.ErrConflict
	}
	next.Version = expected + 1
	s.state = cloneState(next)
	return nil
}

func cloneState(state domain.State) domain.State {
	data, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	var copy domain.State
	if err := json.Unmarshal(data, &copy); err != nil {
		panic(err)
	}
	return copy
}

func DashboardPolicyMarker(version uint64) bool {
	if version == 0 {
		return false
	}
	return version < ^uint64(0)
}
