//go:build integration

package handler_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danzt/daas/api/testhelpers"
)

// setupTestDB starts a testcontainers PostgreSQL instance, applies all
// migrations, and returns a single-connection pool.
//
// Using a single connection (pool_max_conns=1) ensures that set_config calls
// in activateRLS affect the same connection used by ShopService queries —
// RLS is session-scoped, so all queries on the same connection see the config.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := testhelpers.NewPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("setupTestDB: start container: %v", err)
	}
	t.Cleanup(func() { _ = container.Cleanup(ctx) })

	// pool_max_conns=1 guarantees activateRLS and ShopService queries use
	// the same physical connection, making session-scoped set_config reliable.
	cfg, err := pgxpool.ParseConfig(container.DSN + "&pool_max_conns=1")
	if err != nil {
		t.Fatalf("setupTestDB: parse config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("setupTestDB: connect pool: %v", err)
	}
	t.Cleanup(pool.Close)

	// Apply migrations using a temporary standard pool (not the single-conn pool)
	// to avoid blocking — the single-conn pool may be held during migration.
	migPool, err := pgxpool.New(ctx, container.DSN)
	if err != nil {
		t.Fatalf("setupTestDB: connect migration pool: %v", err)
	}
	defer migPool.Close()

	if err := applyAllMigrations(ctx, migPool); err != nil {
		t.Fatalf("setupTestDB: apply migrations: %v", err)
	}

	return pool
}

// applyAllMigrations runs all *.up.sql files in the migrations/ directory
// in lexicographic order.
func applyAllMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	dir, err := findMigrationsDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)

	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", filepath.Base(f), err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("exec %s: %w", filepath.Base(f), err)
		}
	}
	return nil
}

// findMigrationsDir walks up from the working directory to locate migrations/.
func findMigrationsDir() (string, error) {
	if d := os.Getenv("MIGRATIONS_DIR"); d != "" {
		return d, nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	for {
		candidate := filepath.Join(dir, "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not find migrations/ directory; set MIGRATIONS_DIR env var")
}
