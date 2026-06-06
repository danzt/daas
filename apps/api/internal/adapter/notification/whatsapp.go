package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/rs/zerolog/log"

	"github.com/danzt/daas/api/internal/domain/notification"
)

// WhatsAppAdapter delivers WhatsApp messages via the Meta WhatsApp Business
// Cloud API. It uses raw HTTP (no third-party SDK) to keep zero extra
// dependencies. Authentication is done with a System User access token.
//
// Activated in S9. Before that, credentials are not wired and Send returns
// notification.ErrNotImplemented.
type WhatsAppAdapter struct {
	phoneNumberID string // Meta Business phone number ID
	token         string // Bearer token (System User access token)
	httpClient    *http.Client
}

// NewWhatsAppAdapter builds an adapter backed by the given Meta Cloud API
// credentials. When either phoneNumberID or token is empty, Send returns
// notification.ErrNotImplemented so the service degrades gracefully.
func NewWhatsAppAdapter(phoneNumberID, token string) *WhatsAppAdapter {
	return &WhatsAppAdapter{
		phoneNumberID: phoneNumberID,
		token:         token,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
	}
}

// Send delivers a WhatsApp text message to msg.To (E.164 phone number).
// msg.Body is used as the message text (HTML tags stripped, max 4096 chars).
// msg.Subject is prepended as bold text if non-empty: "*{Subject}*\n\n{Body}"
//
// Only processes messages with msg.Channel == "whatsapp". Returns an error for
// any other channel value so MultiChannelNotifier can route correctly.
func (a *WhatsAppAdapter) Send(ctx context.Context, msg notification.Message) error {
	if a.phoneNumberID == "" || a.token == "" {
		return notification.ErrNotImplemented
	}
	if msg.Channel != "whatsapp" {
		return fmt.Errorf("whatsapp adapter: unsupported channel %q", msg.Channel)
	}

	text := buildWhatsAppText(msg)

	payload := map[string]any{
		"messaging_product": "whatsapp",
		"to":                msg.To,
		"type":              "text",
		"text":              map[string]string{"body": text},
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("https://graph.facebook.com/v20.0/%s/messages", a.phoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("whatsapp: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("whatsapp: send: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		rb, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("whatsapp: status %d: %s", resp.StatusCode, string(rb))
	}

	log.Info().Str("to", msg.To).Msg("whatsapp: message sent")
	return nil
}

// buildWhatsAppText strips HTML, combines subject + body, caps at 4096 chars.
func buildWhatsAppText(msg notification.Message) string {
	body := stripHTML(msg.Body)
	var text string
	if msg.Subject != "" {
		text = "*" + msg.Subject + "*\n\n" + body
	} else {
		text = body
	}
	// WhatsApp max message length is 4096 chars.
	if utf8.RuneCountInString(text) > 4096 {
		runes := []rune(text)
		text = string(runes[:4093]) + "..."
	}
	return text
}

// stripHTML removes HTML tags and decodes common entities.
// Uses a simple state machine — no html package dependency.
func stripHTML(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result.WriteRune(r)
		}
	}
	// Decode common HTML entities.
	out := result.String()
	replacer := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
		"&nbsp;", " ",
	)
	return strings.TrimSpace(replacer.Replace(out))
}

// NormalizePhone strips formatting characters from a phone number and ensures
// it starts with '+' for E.164 compliance. Spaces, dashes, parens are removed.
// Example: "+58 412-000 0000" → "+584120000000"
func NormalizePhone(phone string) string {
	var b strings.Builder
	for i, r := range phone {
		if r == '+' && i == 0 {
			b.WriteRune(r)
			continue
		}
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	result := b.String()
	if result != "" && result[0] != '+' {
		result = "+" + result
	}
	return result
}
