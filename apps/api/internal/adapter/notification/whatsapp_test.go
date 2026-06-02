package notification_test

import (
	"context"
	"errors"
	"testing"

	notifadapter "github.com/danzt/daas/api/internal/adapter/notification"
	"github.com/danzt/daas/api/internal/domain/notification"
)

// TestWhatsAppAdapter_ReturnsNotImplemented verifies NOTIFY-TS-03:
// WhatsAppAdapter.Send always returns ErrNotImplemented without panicking.
func TestWhatsAppAdapter_ReturnsNotImplemented(t *testing.T) {
	adapter := notifadapter.NewWhatsAppAdapter()
	ctx := context.Background()

	msg := notification.Message{
		Channel: "whatsapp",
		To:      "+58412000000",
		Body:    "hello",
	}

	err := adapter.Send(ctx, msg)
	if err == nil {
		t.Fatal("Send returned nil error, want ErrNotImplemented")
	}
	if !errors.Is(err, notification.ErrNotImplemented) {
		t.Errorf("Send error = %v, want notification.ErrNotImplemented", err)
	}
}
