package application

import (
	"context"
	"io"
	"time"

	"github.com/wyw/cry031-volunteer/internal/domain"
)

type Store interface {
	Snapshot(context.Context) (domain.State, error)
	Commit(context.Context, uint64, domain.State, []domain.AuditEvent) error
}

type Clock interface {
	Now() time.Time
}

type Notifier interface {
	Notify(context.Context, string, string, string) error
}

type AttachmentStore interface {
	Save(context.Context, string, string, io.Reader, int64) (string, error)
}

type Actor struct {
	UserID string
	Role   domain.UserRole
}

func (a Actor) IsAdmin() bool { return a.Role == domain.RoleAdmin }

type RequestMeta struct {
	RequestID      string
	IdempotencyKey string
}
