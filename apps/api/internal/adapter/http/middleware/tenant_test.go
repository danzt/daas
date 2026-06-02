package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"

	mw "github.com/danzt/daas/api/internal/adapter/http/middleware"
)

// mockConn is a test double that captures Exec calls made by TenantMiddleware.
// It records the SQL string and arguments so we can assert parameterization.
type mockConn struct {
	capturedSQL  string
	capturedArgs []any
}

func (m *mockConn) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	m.capturedSQL = sql
	m.capturedArgs = append(m.capturedArgs, args...)
	return pgconn.CommandTag{}, nil
}

func (m *mockConn) Release() {}

// TestTenantMiddleware_ParameterizedSetConfig verifies SEC-TS-01:
// the middleware MUST use a parameterized query — not fmt.Sprintf string interpolation —
// when setting the tenant_id session variable.
//
// Assertions:
//  1. The SQL sent to the DB is exactly "SELECT set_config('app.tenant_id', $1, true)"
//  2. The tenantID is passed as the $1 parameter, NOT embedded in the SQL string
//  3. fmt.Sprintf interpolation is absent (the tenant ID must NOT appear inside the SQL string)
func TestTenantMiddleware_ParameterizedSetConfig(t *testing.T) {
	const tenantID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	conn := &mockConn{}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Pre-populate context as AuthMiddleware would.
	c.Set(string(mw.ContextKeyTenantID), tenantID)

	// Create middleware with our mock conn injector.
	m := mw.NewTenantMiddlewareWithConn(func(_ context.Context) (mw.ConnExecer, error) {
		return conn, nil
	})

	nextCalled := false
	handler := m.Handle()(func(c echo.Context) error {
		nextCalled = true
		return c.JSON(http.StatusOK, "ok")
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}

	if !nextCalled {
		t.Fatal("next handler was not called — middleware may have returned early")
	}

	// Assert 1: SQL must be the parameterized form.
	const wantSQL = "SELECT set_config('app.tenant_id', $1, true)"
	if conn.capturedSQL != wantSQL {
		t.Errorf("SQL mismatch:\n  got:  %q\n  want: %q", conn.capturedSQL, wantSQL)
	}

	// Assert 2: tenantID must be the first argument, not interpolated into the SQL.
	if len(conn.capturedArgs) == 0 {
		t.Fatal("no arguments were passed to Exec — tenantID must be a bind parameter")
	}
	gotArg, ok := conn.capturedArgs[0].(string)
	if !ok {
		t.Fatalf("first arg is %T, want string", conn.capturedArgs[0])
	}
	if gotArg != tenantID {
		t.Errorf("bind parameter mismatch: got %q, want %q", gotArg, tenantID)
	}

	// Assert 3: tenantID must NOT appear inside the SQL string itself.
	// If fmt.Sprintf were used, the tenant ID would be embedded in the SQL.
	if strings.Contains(conn.capturedSQL, tenantID) {
		t.Errorf("tenantID found embedded in SQL string — this indicates fmt.Sprintf was used instead of parameterized query.\nSQL: %q", conn.capturedSQL)
	}
}

// TestTenantMiddleware_ParameterizedSetConfig_Triangulate uses a different
// tenantID to force real logic (not a hardcoded value) and confirms the
// parameter binding works for any UUID, not just the first test value.
func TestTenantMiddleware_ParameterizedSetConfig_Triangulate(t *testing.T) {
	const tenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"

	conn := &mockConn{}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(mw.ContextKeyTenantID), tenantID)

	m := mw.NewTenantMiddlewareWithConn(func(_ context.Context) (mw.ConnExecer, error) {
		return conn, nil
	})

	_ = m.Handle()(func(c echo.Context) error {
		return c.JSON(http.StatusOK, "ok")
	})(c)

	const wantSQL = "SELECT set_config('app.tenant_id', $1, true)"
	if conn.capturedSQL != wantSQL {
		t.Errorf("SQL mismatch:\n  got:  %q\n  want: %q", conn.capturedSQL, wantSQL)
	}
	if len(conn.capturedArgs) == 0 || conn.capturedArgs[0] != tenantID {
		t.Errorf("expected bind arg %q, got %v", tenantID, conn.capturedArgs)
	}
	if strings.Contains(conn.capturedSQL, tenantID) {
		t.Errorf("tenantID embedded in SQL — fmt.Sprintf detected: %q", conn.capturedSQL)
	}
}
