package app_test

import (
	"context"
	"sync"
	"testing"

	notifadapter "github.com/danzt/daas/api/internal/adapter/notification"
	"github.com/danzt/daas/api/internal/app"
	"github.com/danzt/daas/api/internal/domain/notification"
)

// TestMultiChannelNotifier_FansOutToAllAdapters verifies that Send delivers
// the message to every registered adapter and that each adapter receives the
// message with its own Channel value (not the caller's original Channel).
func TestMultiChannelNotifier_FansOutToAllAdapters(t *testing.T) {
	email := notifadapter.NewNoopAdapter()
	whatsapp := notifadapter.NewNoopAdapter()

	n := app.NewMultiChannelNotifier(email, "email", whatsapp, "whatsapp")

	msg := notification.Message{
		Channel: "email", // caller sets this; notifier overrides per-adapter
		To:      "test@example.com",
		Subject: "Test",
		Body:    "Test body",
	}

	if err := n.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}

	emailSent := email.Sent()
	if len(emailSent) != 1 {
		t.Fatalf("email adapter received %d messages, want 1", len(emailSent))
	}
	if emailSent[0].Channel != "email" {
		t.Errorf("email adapter got Channel=%q, want %q", emailSent[0].Channel, "email")
	}

	waSent := whatsapp.Sent()
	if len(waSent) != 1 {
		t.Fatalf("whatsapp adapter received %d messages, want 1", len(waSent))
	}
	if waSent[0].Channel != "whatsapp" {
		t.Errorf("whatsapp adapter got Channel=%q, want %q", waSent[0].Channel, "whatsapp")
	}
}

// TestMultiChannelNotifier_NeverReturnsError verifies that Send returns nil
// even when all adapters fail.
func TestMultiChannelNotifier_NeverReturnsError(t *testing.T) {
	failing := &failingAdapter{}
	n := app.NewMultiChannelNotifier(failing, "email", failing, "whatsapp")

	err := n.Send(context.Background(), notification.Message{
		Channel: "email",
		To:      "test@example.com",
		Body:    "hello",
	})
	if err != nil {
		t.Errorf("Send returned %v, want nil (MultiChannelNotifier must never return an error)", err)
	}
}

// TestMultiChannelNotifier_ConcurrentSafe verifies that concurrent Send calls
// do not race (run with -race to detect data races).
func TestMultiChannelNotifier_ConcurrentSafe(t *testing.T) {
	noop := notifadapter.NewNoopAdapter()
	n := app.NewMultiChannelNotifier(noop, "email")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = n.Send(context.Background(), notification.Message{
				Channel: "email",
				To:      "race@example.com",
				Body:    "concurrent",
			})
		}()
	}
	wg.Wait()

	got := noop.Sent()
	if len(got) != 50 {
		t.Errorf("expected 50 messages, got %d", len(got))
	}
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// failingAdapter always returns an error from Send.
type failingAdapter struct{}

func (f *failingAdapter) Send(_ context.Context, _ notification.Message) error {
	return notification.ErrNotImplemented
}
