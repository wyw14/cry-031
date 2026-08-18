package domain

import (
	"testing"
	"time"
)

func TestActivityCanClaimRejectsCancelledAndExpired(t *testing.T) {
	now := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	activity := Activity{Status: ActivityPublished, StartAt: now.Add(time.Hour), EndAt: now.Add(2 * time.Hour)}
	if err := activity.CanClaim(now); err != nil {
		t.Fatalf("published activity should be claimable: %v", err)
	}
	activity.CancelledAt = &now
	if err := activity.CanClaim(now); err != ErrActivityClosed {
		t.Fatalf("cancelled activity error = %v, want %v", err, ErrActivityClosed)
	}
}

func TestServiceDurationRejectsUnreasonableSpan(t *testing.T) {
	start := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	end := start.Add(25 * time.Hour)
	record := ServiceRecord{StartedAt: &start, EndedAt: &end}
	if _, err := record.Duration(); err != ErrInvalidDuration {
		t.Fatalf("duration error = %v, want %v", err, ErrInvalidDuration)
	}
}

func TestActivityWindowsOverlapAtBoundary(t *testing.T) {
	now := time.Now().UTC()
	left := Activity{StartAt: now, EndAt: now.Add(time.Hour)}
	right := Activity{StartAt: now.Add(time.Hour), EndAt: now.Add(2 * time.Hour)}
	if left.WindowOverlaps(right) {
		t.Fatal("adjacent windows must not overlap")
	}
}
