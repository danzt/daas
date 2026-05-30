package middleware

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// TenantMiddleware reads the tenant_id from context (set by AuthMiddleware)
// and executes SET LOCAL app.tenant_id = $1 on the database connection.
// This activates PostgreSQL RLS policies that filter by app.tenant_id.
//
// The middleware acts as the second isolation barrier (Go is the first;
// RLS is the second). Any query executed after this middleware runs will
// automatically be filtered by the tenant's RLS policy.
type TenantMiddleware struct {
	pool *pgxpool.Pool
}

// NewTenantMiddleware creates a TenantMiddleware backed by the given pgx pool.
func NewTenantMiddleware(pool *pgxpool.Pool) *TenantMiddleware {
	return &TenantMiddleware{pool: pool}
}

// Handle returns the Echo middleware handler function.
// It acquires a connection from the pool, sets the local tenant parameter,
// and stores the connection on the context for use by handlers.
func (m *TenantMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenantID, ok := c.Get(string(ContextKeyTenantID)).(string)
			if !ok || tenantID == "" {
				return problemJSON(c, http.StatusForbidden, "forbidden", "tenant context missing")
			}

			// Acquire a dedicated connection from the pool. This is required
			// because SET LOCAL only persists for the duration of the current
			// transaction/connection — pgxpool would route queries to different
			// connections otherwise.
			conn, err := m.pool.Acquire(c.Request().Context())
			if err != nil {
				return problemJSON(c, http.StatusInternalServerError, "internal-error",
					"failed to acquire database connection")
			}
			defer conn.Release()

			// SET LOCAL is transaction-scoped. We set it at connection level
			// so it applies for the request lifetime.
			_, err = conn.Exec(c.Request().Context(),
				fmt.Sprintf("SET LOCAL app.tenant_id = '%s'", tenantID))
			if err != nil {
				return problemJSON(c, http.StatusInternalServerError, "internal-error",
					"failed to set tenant context")
			}

			// Store the acquired connection on the context so handlers can use
			// the same connection (and therefore the same SET LOCAL session).
			c.Set("db_conn", conn)
			c.Set(string(ContextKeyTenantID), tenantID)

			return next(c)
		}
	}
}
