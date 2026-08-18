package service

import (
	"context"
	"time"
)

// Scheduler runs local callbacks for deadline scans. Jobs remain deterministic in tests via RunOnce.
type Scheduler struct {
	Interval time.Duration
	Job      func(context.Context) error
}

func (s Scheduler) RunOnce(ctx context.Context) error {
	if s.Job == nil {
		return nil
	}
	return s.Job(ctx)
}

func (s Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = s.RunOnce(ctx)
		}
	}
}
