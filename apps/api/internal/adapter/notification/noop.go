package notification

import (
	"context"
	"sync"

	"github.com/danzt/daas/api/internal/domain/notification"
)

// NoopAdapter implements notification.NotificationService for test environments.
// It records every Send call in memory without performing any real I/O.
// Callers retrieve recorded messages via Sent() for assertion.
//
// NoopAdapter is safe for concurrent use — all state mutations are guarded by
// a sync.Mutex to support parallel test goroutines.
type NoopAdapter struct {
	mu   sync.Mutex
	sent []notification.Message
}

// NewNoopAdapter constructs a NoopAdapter with an empty message log.
func NewNoopAdapter() *NoopAdapter { return &NoopAdapter{} }

// Send records msg in the adapter's internal log and returns nil.
// It implements notification.NotificationService.
func (a *NoopAdapter) Send(_ context.Context, msg notification.Message) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sent = append(a.sent, msg)
	return nil
}

// Sent returns a snapshot copy of all messages sent so far, in the order
// they were received. The copy ensures callers cannot mutate internal state.
func (a *NoopAdapter) Sent() []notification.Message {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]notification.Message, len(a.sent))
	copy(out, a.sent)
	return out
}
