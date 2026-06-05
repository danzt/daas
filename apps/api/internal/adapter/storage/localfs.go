package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalFSAdapter writes uploaded files to the local filesystem under rootDir
// and exposes them via publicBase (e.g. "http://localhost:8080/files").
// Intended for local development only; replace with SupabaseStorageAdapter in production.
type LocalFSAdapter struct {
	rootDir    string // e.g. "./storage"
	publicBase string // e.g. "http://localhost:8080/files"
}

// NewLocalFSAdapter creates a LocalFSAdapter.
// rootDir is where files are written; publicBase is prepended to the key to form the URL.
func NewLocalFSAdapter(rootDir, publicBase string) *LocalFSAdapter {
	return &LocalFSAdapter{
		rootDir:    rootDir,
		publicBase: publicBase,
	}
}

// Save writes content to {rootDir}/{key}, creating all intermediate directories,
// and returns {publicBase}/{key} as the public URL.
func (a *LocalFSAdapter) Save(_ context.Context, key string, content io.Reader, _ string) (string, error) {
	dest := filepath.Join(a.rootDir, key)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", fmt.Errorf("localfs: create dirs for %q: %w", dest, err)
	}

	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("localfs: create file %q: %w", dest, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(f, content); err != nil {
		return "", fmt.Errorf("localfs: write file %q: %w", dest, err)
	}

	return a.publicBase + "/" + key, nil
}

// Delete removes the file at {rootDir}/{key}.
// A missing file is not treated as an error (idempotent).
func (a *LocalFSAdapter) Delete(_ context.Context, key string) error {
	dest := filepath.Join(a.rootDir, key)
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("localfs: delete %q: %w", dest, err)
	}
	return nil
}
