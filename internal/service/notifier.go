package service

import (
	"context"
	"sync"
	"time"
)

type Notification struct {
	UserID    string    `json:"user_id"`
	Kind      string    `json:"kind"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// LocalNotifier is a deterministic local adapter. It never calls an external message service.
type LocalNotifier struct {
	mu     sync.RWMutex
	events []Notification
}

func (n *LocalNotifier) Notify(ctx context.Context, userID, kind, message string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.events = append(n.events, Notification{UserID: userID, Kind: kind, Message: message, CreatedAt: time.Now().UTC()})
	return nil
}

func (n *LocalNotifier) Events() []Notification {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return append([]Notification(nil), n.events...)
}
