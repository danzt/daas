package notification

import (
	"context"

	"github.com/danzt/daas/api/internal/domain/notification"
)

// WhatsAppAdapter is a placeholder for the WhatsApp notification channel.
// It implements notification.NotificationService but returns ErrNotImplemented
// for every Send call. Activation is planned for Sprint 9.
//
// To swap it in for a real integration, replace this struct's implementation
// with the chosen WhatsApp Business API client — the interface and wiring
// in main.go remain unchanged.
type WhatsAppAdapter struct{}

// NewWhatsAppAdapter constructs a WhatsAppAdapter stub.
func NewWhatsAppAdapter() *WhatsAppAdapter { return &WhatsAppAdapter{} }

// Send always returns notification.ErrNotImplemented.
// No real delivery occurs in S6 or any sprint before S9.
func (a *WhatsAppAdapter) Send(_ context.Context, _ notification.Message) error {
	return notification.ErrNotImplemented
}
