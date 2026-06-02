package notification_test

import (
	"context"
	"sync"
	"testing"

	notifadapter "github.com/danzt/daas/api/internal/adapter/notification"
	"github.com/danzt/daas/api/internal/domain/notification"
)

// TestNoopAdapter_RecordsCalls verifies NOTIFY-TS-02:
// Send called N times → Sent() returns N messages in order.
func TestNoopAdapter_RecordsCalls(t *testing.T) {
	adapter := notifadapter.NewNoopAdapter()
	ctx := context.Background()

	msgs := []notification.Message{
		{Channel: "email", To: "a@example.com", Subject: "First", Body: "body1"},
		{Channel: "email", To: "b@example.com", Subject: "Second", Body: "body2"},
		{Channel: "whatsapp", To: "+58412000000", Body: "body3"},
	}

	for _, m := range msgs {
		if err := adapter.Send(ctx, m); err != nil {
			t.Fatalf("Send returned unexpected error: %v", err)
		}
	}

	got := adapter.Sent()
	if len(got) != 3 {
		t.Fatalf("Sent() returned %d messages, want 3", len(got))
	}

	// Verify order and content (NOTIFY-TS-02: in order).
	if got[0].To != "a@example.com" {
		t.Errorf("msg[0].To = %q, want %q", got[0].To, "a@example.com")
	}
	if got[1].Subject != "Second" {
		t.Errorf("msg[1].Subject = %q, want %q", got[1].Subject, "Second")
	}
	if got[2].Channel != "whatsapp" {
		t.Errorf("msg[2].Channel = %q, want %q", got[2].Channel, "whatsapp")
	}
}

// TestNoopAdapter_ConcurrentSafe verifies that 100 goroutines calling Send
// concurrently do not race and that all messages are recorded.
// Run this test with -race to detect sync.Mutex issues.
func TestNoopAdapter_ConcurrentSafe(t *testing.T) {
	adapter := notifadapter.NewNoopAdapter()
	ctx := context.Background()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			msg := notification.Message{
				Channel: "email",
				To:      "concurrent@example.com",
				Subject: "Concurrent",
				Body:    "body",
			}
			if err := adapter.Send(ctx, msg); err != nil {
				t.Errorf("Send error in goroutine: %v", err)
			}
		}()
	}

	wg.Wait()

	sent := adapter.Sent()
	if len(sent) != goroutines {
		t.Errorf("Sent() returned %d messages, want %d", len(sent), goroutines)
	}
}
