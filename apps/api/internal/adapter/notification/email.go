// Package notification provides adapters for the NotificationService port.
// EmailAdapter delivers messages via the Resend transactional email API.
package notification

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"

	"github.com/danzt/daas/api/internal/domain/notification"
)

// EmailAdapter sends emails through the Resend API.
//
// Constructed in S6 but NOT invoked by any handler yet — wiring is in place so
// that S8 (payment confirmation, order notifications) can flip the switch
// without touching the dependency graph.
type EmailAdapter struct {
	client *resend.Client
	from   string
}

// NewEmailAdapter builds an EmailAdapter with the given Resend API key and
// "from" address. The key is read from RESEND_API_KEY in production; "from"
// must be a verified sender on the Resend account (e.g. noreply@daas.app).
func NewEmailAdapter(apiKey, from string) *EmailAdapter {
	return &EmailAdapter{
		client: resend.NewClient(apiKey),
		from:   from,
	}
}

// Send delivers msg via Resend. msg.To is the recipient address; msg.Body
// is treated as HTML. Returns a wrapped error on transport failure.
func (a *EmailAdapter) Send(ctx context.Context, msg notification.Message) error {
	req := &resend.SendEmailRequest{
		From:    a.from,
		To:      []string{msg.To},
		Subject: msg.Subject,
		Html:    msg.Body,
	}
	if _, err := a.client.Emails.SendWithContext(ctx, req); err != nil {
		return fmt.Errorf("send email via resend: %w", err)
	}
	return nil
}
