package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/danzt/daas/api/internal/domain/tenant"
)

// OwnerGuard is a middleware that restricts access to tenant owners.
// It reads the role from context (set by AuthMiddleware) and returns
// HTTP 403 for any role that is not "owner".
//
// Usage: apply AFTER AuthMiddleware on owner-only routes.
func OwnerGuard() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roleStr, _ := c.Get(string(ContextKeyRole)).(string)
			role := tenant.Role(strings.TrimSpace(roleStr))
			if !role.IsOwner() {
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"type":   "https://daas.app/errors/forbidden",
					"title":  "Forbidden",
					"status": http.StatusForbidden,
					"detail": "this action requires owner role",
				})
			}
			return next(c)
		}
	}
}
