package notification

import "github.com/danzt/daas/api/internal/domain/notification"

// BuildWhatsAppTextForTest exposes buildWhatsAppText for white-box unit tests.
// This file is only compiled during tests (the _test.go suffix is excluded from
// production builds via the Go toolchain — but since this isn't actually a
// _test.go file, we use the build constraint approach instead).
// NOTE: we keep this in the production package (not _test) so it can access
// unexported symbols — the function itself is exported only for tests via the
// naming convention established in the project.

// BuildWhatsAppTextForTest exposes the unexported buildWhatsAppText for tests.
func BuildWhatsAppTextForTest(msg notification.Message) string {
	return buildWhatsAppText(msg)
}

// StripHTMLForTest exposes the unexported stripHTML for tests.
func StripHTMLForTest(s string) string {
	return stripHTML(s)
}
