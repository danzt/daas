package notification_test

import (
	"context"
	"errors"
	"testing"

	notifadapter "github.com/danzt/daas/api/internal/adapter/notification"
	"github.com/danzt/daas/api/internal/domain/notification"
)

// ─── WhatsAppAdapter.Send ─────────────────────────────────────────────────────

// TestWhatsAppAdapter_MissingCredentials_ReturnsNotImplemented verifies that
// Send returns ErrNotImplemented when phoneNumberID or token is empty.
func TestWhatsAppAdapter_MissingCredentials_ReturnsNotImplemented(t *testing.T) {
	cases := []struct {
		name          string
		phoneNumberID string
		token         string
	}{
		{"both empty", "", ""},
		{"phone empty", "", "some-token"},
		{"token empty", "123456", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := notifadapter.NewWhatsAppAdapter(tc.phoneNumberID, tc.token)
			err := adapter.Send(context.Background(), notification.Message{
				Channel: "whatsapp",
				To:      "+58412000000",
				Body:    "hello",
			})
			if !errors.Is(err, notification.ErrNotImplemented) {
				t.Errorf("want ErrNotImplemented, got %v", err)
			}
		})
	}
}

// TestWhatsAppAdapter_WrongChannel_ReturnsError verifies that Send rejects
// messages whose Channel is not "whatsapp".
func TestWhatsAppAdapter_WrongChannel_ReturnsError(t *testing.T) {
	// Use valid credentials so the channel check is reached.
	adapter := notifadapter.NewWhatsAppAdapter("phone-id", "token")
	err := adapter.Send(context.Background(), notification.Message{
		Channel: "email",
		To:      "+58412000000",
		Body:    "hello",
	})
	if err == nil {
		t.Fatal("expected error for wrong channel, got nil")
	}
	if errors.Is(err, notification.ErrNotImplemented) {
		t.Errorf("expected channel-error, not ErrNotImplemented")
	}
}

// ─── buildWhatsAppText (via exported helpers) ──────────────────────────────────

// TestBuildWhatsAppText_WithSubject verifies that a non-empty Subject is
// prepended as bold text followed by two newlines.
func TestBuildWhatsAppText_WithSubject(t *testing.T) {
	got := notifadapter.BuildWhatsAppTextForTest(notification.Message{
		Subject: "Pedido confirmado",
		Body:    "Tu pedido fue confirmado.",
	})
	want := "*Pedido confirmado*\n\nTu pedido fue confirmado."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestBuildWhatsAppText_NoSubject verifies that an empty Subject produces just
// the body text without any prefix.
func TestBuildWhatsAppText_NoSubject(t *testing.T) {
	got := notifadapter.BuildWhatsAppTextForTest(notification.Message{
		Subject: "",
		Body:    "Solo el cuerpo.",
	})
	want := "Solo el cuerpo."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestBuildWhatsAppText_TruncatesLongBody verifies that text exceeding 4096
// runes is truncated and terminated with "...".
func TestBuildWhatsAppText_TruncatesLongBody(t *testing.T) {
	// Build a body of 5000 'a' runes — well above the 4096 limit.
	body := string(make([]rune, 5000))
	for i := range []rune(body) {
		_ = i
	}
	// Simple approach: just use a long ASCII string.
	longBody := ""
	for i := 0; i < 5000; i++ {
		longBody += "a"
	}

	got := notifadapter.BuildWhatsAppTextForTest(notification.Message{Body: longBody})
	runeCount := len([]rune(got))
	if runeCount != 4096 {
		t.Errorf("truncated text has %d runes, want 4096", runeCount)
	}
	if got[len(got)-3:] != "..." {
		t.Errorf("truncated text does not end with '...', got %q", got[len(got)-10:])
	}
}

// ─── stripHTML ────────────────────────────────────────────────────────────────

// TestStripHTML_RemovesTags verifies that HTML tags are stripped from the body.
func TestStripHTML_RemovesTags(t *testing.T) {
	input := "<p>Hola <strong>mundo</strong></p>"
	got := notifadapter.StripHTMLForTest(input)
	want := "Hola mundo"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestStripHTML_DecodesEntities verifies that common HTML entities are decoded.
func TestStripHTML_DecodesEntities(t *testing.T) {
	input := "AT&amp;T &lt;test&gt; &quot;quoted&quot; &#39;apostrophe&#39; &nbsp;space"
	got := notifadapter.StripHTMLForTest(input)
	want := `AT&T <test> "quoted" 'apostrophe'  space`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// ─── NormalizePhone ────────────────────────────────────────────────────────────

// TestNormalizePhone verifies E.164 formatting for various phone number styles.
// NormalizePhone strips formatting characters (spaces, dashes, parens) and
// ensures the result starts with '+'. Numbers without a leading '+' get one
// prepended — the caller is responsible for providing the full country code.
func TestNormalizePhone(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"+58 412-000 0000", "+584120000000"},
		// Local Venezuelan format preserves all digits including trunk '0'.
		// Callers should store E.164 (+58...) in the database.
		{"(0412) 123-4567", "+04121234567"},
		{"+1 (800) 555-0100", "+18005550100"},
		{"+584120000000", "+584120000000"},
		{"584120000000", "+584120000000"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := notifadapter.NormalizePhone(tc.input)
			if got != tc.want {
				t.Errorf("NormalizePhone(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
