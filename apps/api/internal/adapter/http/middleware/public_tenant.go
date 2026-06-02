package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// PublicTenantMiddleware resolves the tenant from the request (via
// TenantResolver), acquires a dedicated database connection, sets the tenant
// context via set_config, and stores both on the Echo context for handlers.
//
// This middleware is used exclusively on the public storefront route group
// (/t/:tenantSlug/shop/v1). It does NOT require authentication.
type PublicTenantMiddleware struct {
	resolver TenantResolver
	acquire  connAcquirer
}

// NewPublicTenantMiddleware creates a PublicTenantMiddleware using a production
// pgxpool.Pool to acquire connections.
func NewPublicTenantMiddleware(resolver TenantResolver, pool *pgxpool.Pool) *PublicTenantMiddleware {
	return &PublicTenantMiddleware{
		resolver: resolver,
		acquire: func(ctx context.Context) (ConnExecer, error) {
			return pool.Acquire(ctx)
		},
	}
}

// NewPublicTenantMiddlewareWithAcquire creates a PublicTenantMiddleware with
// an injectable connAcquirer. Intended for unit tests.
func NewPublicTenantMiddlewareWithAcquire(resolver TenantResolver, acquire func(ctx context.Context) (ConnExecer, error)) *PublicTenantMiddleware {
	return &PublicTenantMiddleware{
		resolver: resolver,
		acquire:  acquire,
	}
}

// Handle returns the Echo middleware handler function.
//
// Flow:
//  1. Resolve the tenant via the injected TenantResolver.
//  2. On ErrTenantNotFound → return 404 {"error":"tenant_not_found"}.
//  3. Acquire a dedicated DB connection from the pool.
//  4. Set app.tenant_id via parameterized set_config (activates RLS).
//  5. Store the connection + tenantID on the Echo context (same keys as TenantMiddleware).
//  6. Set Cache-Control: no-store on the response.
//  7. Call next(c); defer conn.Release() to return the connection to the pool.
func (m *PublicTenantMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenantID, err := m.resolver.Resolve(c)
			if err != nil {
				if errors.Is(err, ErrTenantNotFound) {
					return c.JSON(http.StatusNotFound, map[string]string{
						"error": "tenant_not_found",
					})
				}
				return problemJSON(c, http.StatusInternalServerError, "internal-error",
					"failed to resolve tenant")
			}

			conn, err := m.acquire(c.Request().Context())
			if err != nil {
				return problemJSON(c, http.StatusInternalServerError, "internal-error",
					"failed to acquire database connection")
			}
			defer conn.Release()

			// Activate RLS for the resolved tenant. Parameterized query —
			// tenantID is never interpolated into the SQL string.
			_, err = conn.Exec(c.Request().Context(),
				"SELECT set_config('app.tenant_id', $1, true)", tenantID)
			if err != nil {
				return problemJSON(c, http.StatusInternalServerError, "internal-error",
					"failed to set tenant context")
			}

			// Store connection and tenantID for downstream handlers.
			// Same context keys as TenantMiddleware so handlers are agnostic.
			c.Set("db_conn", conn)
			c.Set(string(ContextKeyTenantID), tenantID)

			// Public catalog responses must not be cached by intermediaries
			// (CDNs, shared proxies) since they are tenant-scoped.
			c.Response().Header().Set("Cache-Control", "no-store")

			return next(c)
		}
	}
}
