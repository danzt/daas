package storage_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/danzt/daas/api/internal/adapter/storage"
)

func TestLocalFSAdapter_SaveCreatesFile(t *testing.T) {
	dir := t.TempDir()
	adapter := storage.NewLocalFSAdapter(dir, "http://localhost:8080/files")

	url, err := adapter.Save(context.Background(), "payment-proofs/tenant1/order1/proof.jpg", strings.NewReader("data"), "image/jpeg")
	if err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	expectedURL := "http://localhost:8080/files/payment-proofs/tenant1/order1/proof.jpg"
	if url != expectedURL {
		t.Errorf("Save() URL = %q, want %q", url, expectedURL)
	}

	dest := filepath.Join(dir, "payment-proofs/tenant1/order1/proof.jpg")
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Errorf("Save() did not create file at %q", dest)
	}
}

func TestLocalFSAdapter_SavePreservesContent(t *testing.T) {
	dir := t.TempDir()
	adapter := storage.NewLocalFSAdapter(dir, "http://localhost:8080/files")

	content := []byte("hello payment proof content")
	_, err := adapter.Save(context.Background(), "proofs/test.jpg", bytes.NewReader(content), "image/jpeg")
	if err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	dest := filepath.Join(dir, "proofs/test.jpg")
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ReadFile() unexpected error: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("Save() content = %q, want %q", got, content)
	}
}

func TestLocalFSAdapter_DeleteIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	adapter := storage.NewLocalFSAdapter(dir, "http://localhost:8080/files")

	// Deleting a non-existent key should not error.
	if err := adapter.Delete(context.Background(), "nonexistent/file.jpg"); err != nil {
		t.Errorf("Delete() non-existent key should not error, got: %v", err)
	}

	// Save then delete — should succeed.
	key := "proofs/to-delete.jpg"
	if _, err := adapter.Save(context.Background(), key, strings.NewReader("data"), "image/jpeg"); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}
	if err := adapter.Delete(context.Background(), key); err != nil {
		t.Errorf("Delete() existing file returned error: %v", err)
	}

	// Deleting again (already gone) must be idempotent.
	if err := adapter.Delete(context.Background(), key); err != nil {
		t.Errorf("Delete() second call (already deleted) should not error, got: %v", err)
	}
}
