//go:build integration

package postgres_test

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

// TestRLSIsolation verifies that PostgreSQL RLS policies prevent tenant A
// from reading tenant B's rows, even when both tenants exist in the same table.
//
// This test requires Docker to be running. Run with:
//
//	go test -tags integration ./internal/adapter/postgres/...
func TestRLSIsolation(t *testing.T) {
	ctx := context.Background()

	container, err := testhelpers.NewPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Cleanup(ctx); err != nil {
			t.Logf("container cleanup error: %v", err)
		}
	})

	pool, err := pgxpool.New(ctx, container.DSN)
	if err != nil {
		t.Fatalf("connect to test DB: %v", err)
	}
	defer pool.Close()

	// Apply migrations from the migrations/ directory.
	if err := applyMigrations(ctx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	// Create two tenants directly (bypassing RLS via service role equivalent).
	var tenantAID, tenantBID string
	err = pool.QueryRow(ctx,
		`INSERT INTO tenants (name, country_code) VALUES ('Tenant A', 'VE') RETURNING id`,
	).Scan(&tenantAID)
	if err != nil {
		t.Fatalf("create tenant A: %v", err)
	}

	err = pool.QueryRow(ctx,
		`INSERT INTO tenants (name, country_code) VALUES ('Tenant B', 'DO') RETURNING id`,
	).Scan(&tenantBID)
	if err != nil {
		t.Fatalf("create tenant B: %v", err)
	}

	// Insert a user for each tenant.
	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_users (tenant_id, supabase_uid, email, role)
         VALUES ($1, gen_random_uuid(), 'owner@tenant-a.com', 'owner')`,
		tenantAID,
	)
	if err != nil {
		t.Fatalf("create user for tenant A: %v", err)
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_users (tenant_id, supabase_uid, email, role)
         VALUES ($1, gen_random_uuid(), 'owner@tenant-b.com', 'owner')`,
		tenantBID,
	)
	if err != nil {
		t.Fatalf("create user for tenant B: %v", err)
	}

	// --- Test: tenants table isolation ---
	t.Run("tenants table: tenant A cannot see tenant B", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()

		// Set RLS context to tenant A.
		_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.tenant_id = '%s'", tenantAID))
		if err != nil {
			t.Fatalf("SET LOCAL app.tenant_id: %v", err)
		}

		var count int
		err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM tenants").Scan(&count)
		if err != nil {
			t.Fatalf("count tenants: %v", err)
		}

		if count != 1 {
			t.Errorf("expected tenant A to see 1 tenant (itself), got %d", count)
		}

		var name string
		err = tx.QueryRow(ctx, "SELECT name FROM tenants LIMIT 1").Scan(&name)
		if err != nil {
			t.Fatalf("select tenant name: %v", err)
		}
		if name != "Tenant A" {
			t.Errorf("expected tenant A to see 'Tenant A', got %q", name)
		}
	})

	// --- Test: tenant_users table isolation ---
	t.Run("tenant_users table: tenant A cannot see tenant B users", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()

		_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.tenant_id = '%s'", tenantAID))
		if err != nil {
			t.Fatalf("SET LOCAL app.tenant_id: %v", err)
		}

		var count int
		err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM tenant_users").Scan(&count)
		if err != nil {
			t.Fatalf("count tenant_users: %v", err)
		}

		if count != 1 {
			t.Errorf("expected tenant A to see 1 user (its own), got %d", count)
		}

		var email string
		err = tx.QueryRow(ctx, "SELECT email FROM tenant_users LIMIT 1").Scan(&email)
		if err != nil {
			t.Fatalf("select user email: %v", err)
		}
		if email != "owner@tenant-a.com" {
			t.Errorf("expected tenant A user email, got %q", email)
		}
	})

	// --- Test: tenant B context sees its own rows, not A's ---
	t.Run("tenant_users table: tenant B sees only its own user", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()

		_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.tenant_id = '%s'", tenantBID))
		if err != nil {
			t.Fatalf("SET LOCAL app.tenant_id: %v", err)
		}

		var email string
		err = tx.QueryRow(ctx, "SELECT email FROM tenant_users LIMIT 1").Scan(&email)
		if err != nil {
			t.Fatalf("select user email for tenant B: %v", err)
		}
		if email != "owner@tenant-b.com" {
			t.Errorf("expected tenant B user email, got %q", email)
		}
	})
}

// applyMigrations reads all *.up.sql files from the migrations/ directory
// and executes them in order. Used only in integration tests.
// The migrations directory is resolved via the MIGRATIONS_DIR env var or
// relative to the process working directory (apps/api/).
func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		// When running via `go test ./...` from apps/api/, the working directory
		// is the package directory. Walk up to find migrations/.
		dir, err := findMigrationsDir()
		if err != nil {
			return err
		}
		migrationsDir = dir
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir %s: %w", migrationsDir, err)
	}

	var upFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, filepath.Join(migrationsDir, e.Name()))
		}
	}
	sort.Strings(upFiles)

	for _, f := range upFiles {
		sql, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("execute migration %s: %w", f, err)
		}
	}
	return nil
}

// findMigrationsDir walks up from the current working directory to find
// the migrations/ directory. This handles `go test` being run from any
// subdirectory of apps/api/.
func findMigrationsDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	for {
		candidate := filepath.Join(dir, "migrations")
		if _, err := os.Stat(candidate); err == nil {
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
