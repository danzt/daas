package notification

import (
	"context"
	"errors"
)

// ErrNotImplemented is returned by notification adapters that are not yet
// wired (e.g. WhatsAppAdapter, which is planned for S9).
var ErrNotImplemented = errors.New("notification channel not implemented")

// Message is the channel-agnostic notification payload.
// The Channel field selects the transport; adapters map the remaining fields
// to their SDK-specific request types.
type Message struct {
	// Channel selects the delivery transport: "email" | "whatsapp" | "noop".
	Channel string

	// To is the recipient: an email address for email channel,
	// or an E.164 phone number for WhatsApp channel.
	To string

	// Subject is used by email adapters. Ignored by other channels.
	Subject string

	// Body is the message content: HTML for email, plain text for WhatsApp.
	Body string

	// TemplateID is an optional identifier for pre-defined templates in the
	// adapter's system (e.g. Resend template ID). Empty string = no template.
	TemplateID string

	// Data holds template variables for future server-side rendering.
	// Populated by the caller; rendered by the adapter when TemplateID is set.
	Data map[string]any
}

// NotificationService is the port for sending notifications.
// Implementations (adapters) handle the actual delivery:
//   - EmailAdapter  — sends via Resend SDK (activated in S8)
//   - WhatsAppAdapter — stub, returns ErrNotImplemented (activated in S9)
//   - NoopAdapter   — records calls in memory for tests
type NotificationService interface {
	Send(ctx context.Context, msg Message) error
}
