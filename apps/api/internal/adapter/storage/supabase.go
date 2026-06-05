package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SupabaseStorageAdapter uploads files to Supabase Storage via the REST API.
// Docs: https://supabase.com/docs/reference/javascript/storage-from-upload
type SupabaseStorageAdapter struct {
	projectURL string // e.g. https://vugsdegokxuvwflhdavv.supabase.co
	serviceKey string // service_role JWT (not anon key)
	bucketName string // e.g. "daas-assets"
	httpClient *http.Client
}

// NewSupabaseStorageAdapter creates a SupabaseStorageAdapter.
// projectURL is the Supabase project URL (e.g. https://<ref>.supabase.co).
// serviceKey must be the service_role JWT — not the anon key.
// bucketName is the target bucket (must exist and be public for read access).
func NewSupabaseStorageAdapter(projectURL, serviceKey, bucketName string) *SupabaseStorageAdapter {
	return &SupabaseStorageAdapter{
		projectURL: strings.TrimRight(projectURL, "/"),
		serviceKey: serviceKey,
		bucketName: bucketName,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Save uploads content to Supabase Storage under the given key and returns the
// public URL. The bucket must be public (or have RLS that allows unauthenticated
// reads). Files are upserted — re-uploading the same key is idempotent.
func (s *SupabaseStorageAdapter) Save(ctx context.Context, key string, content io.Reader, contentType string) (string, error) {
	body, err := io.ReadAll(content)
	if err != nil {
		return "", fmt.Errorf("supabase storage: read content: %w", err)
	}

	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.projectURL, s.bucketName, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("supabase storage: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true") // idempotent re-upload

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("supabase storage: upload: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		rb, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("supabase storage: upload: status %d: %s", resp.StatusCode, string(rb))
	}

	// Decode response — Key field is set on success but we can reconstruct the
	// public URL directly from our inputs, which is more reliable.
	var result struct {
		Key string `json:"Key"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.projectURL, s.bucketName, key)
	return publicURL, nil
}

// Delete removes the object at key from Supabase Storage. Idempotent — a missing
// object is treated as success (returns nil).
func (s *SupabaseStorageAdapter) Delete(ctx context.Context, key string) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s", s.projectURL, s.bucketName)

	body, err := json.Marshal(map[string][]string{"prefixes": {key}})
	if err != nil {
		return fmt.Errorf("supabase storage: marshal delete body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("supabase storage: build delete request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("supabase storage: delete: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil // idempotent
	}
	if resp.StatusCode >= 300 {
		rb, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase storage: delete: status %d: %s", resp.StatusCode, string(rb))
	}
	return nil
}
