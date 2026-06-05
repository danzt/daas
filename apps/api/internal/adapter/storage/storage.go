package storage

import (
	"context"
	"io"
)

// Storage abstracts file storage so we can swap LocalFS for Supabase/S3.
type Storage interface {
	// Save writes content under key and returns a publicly resolvable URL
	// (served via the API's static file handler).
	Save(ctx context.Context, key string, content io.Reader, contentType string) (publicURL string, err error)
	// Delete removes the object at key. Idempotent — missing key is not an error.
	Delete(ctx context.Context, key string) error
}
