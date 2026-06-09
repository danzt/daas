package middleware

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// ConnQuerier abstracts the single query StorefrontFeatureMiddleware needs.
// It is satisfied by *pgxpool.Conn (via its underlying Conn) and by test mocks.
type ConnQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// poolQuerier wraps *pgxpool.Pool to satisfy ConnQuerier for production use.
type poolQuerier struct {
	pool *pgxpool.Pool
}

func (p *poolQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

// StorefrontFeatureMiddleware checks that the resolved tenant has the
// 'storefront' feature enabled before allowing access to public shop endpoints.
// It must run after PublicTenantMiddleware, which sets ContextKeyTenantID.
type StorefrontFeatureMiddleware struct {
	querier ConnQuerier
}

// NewStorefrontFeatureMiddleware creates a StorefrontFeatureMiddleware backed
// by the given connection pool.
func NewStorefrontFeatureMiddleware(pool *pgxpool.Pool) *StorefrontFeatureMiddleware {
	return &StorefrontFeatureMiddleware{querier: &poolQuerier{pool: pool}}
}

// NewStorefrontFeatureMiddlewareWithQuerier creates a StorefrontFeatureMiddleware
// with an injectable ConnQuerier. Intended for unit tests.
func NewStorefrontFeatureMiddlewareWithQuerier(q ConnQuerier) *StorefrontFeatureMiddleware {
	return &StorefrontFeatureMiddleware{querier: q}
}

// Handle returns the Echo middleware handler function.
//
// Flow:
//  1. Read tenant_id from Echo context (set by PublicTenantMiddleware).
//  2. If empty → 403 {"error":"storefront-disabled"}.
//  3. Query tenant_features for the storefront flag.
//  4. ErrNoRows or enabled=false → 403 {"error":"storefront-disabled"}.
//  5. enabled=true → call next(c).
func (m *StorefrontFeatureMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenantID, _ := c.Get(string(ContextKeyTenantID)).(string)
			if tenantID == "" {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "storefront-disabled",
				})
			}

			var enabled bool
			err := m.querier.QueryRow(c.Request().Context(),
				`SELECT enabled FROM tenant_features WHERE tenant_id=$1 AND feature='storefront'`,
				tenantID,
			).Scan(&enabled)
			if err != nil {
				if err == pgx.ErrNoRows {
					return c.JSON(http.StatusForbidden, map[string]string{
						"error": "storefront-disabled",
					})
				}
				return problemJSON(c, http.StatusInternalServerError, "internal-error",
					"failed to check storefront feature")
			}

			if !enabled {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "storefront-disabled",
				})
			}

			return next(c)
		}
	}
}
