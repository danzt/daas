package notification

import (
	"testing"
)

func TestEmailAdapter_Compiles(t *testing.T) {
	// We don't make real HTTP calls in unit tests; integration sends are
	// deferred to S8 when the adapter is actually invoked. This test just
	// verifies construction works and the type satisfies the interface.
	a := NewEmailAdapter("test-key", "noreply@daas.app")
	if a == nil {
		t.Fatal("NewEmailAdapter returned nil")
	}
	if a.from != "noreply@daas.app" {
		t.Errorf("from address not stored: got %q", a.from)
	}
	if a.client == nil {
		t.Error("Resend client not initialized")
	}
}
