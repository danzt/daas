package middleware

import (
	"errors"

	"github.com/labstack/echo/v4"
)

// ErrTenantNotFound is returned when the resolver cannot find an active
// tenant for the provided slug or subdomain.
var ErrTenantNotFound = errors.New("tenant not found")

// ErrNotImplemented is returned by resolver stubs that are not yet
// implemented (e.g. SubdomainResolver, which is planned for S11).
var ErrNotImplemented = errors.New("resolver not implemented")

// TenantResolver resolves the tenant for an incoming HTTP request.
// Implementations may use path params (slug), subdomain, or other signals.
// The resolved tenantID is a UUID string.
type TenantResolver interface {
	Resolve(c echo.Context) (tenantID string, err error)
}

// SubdomainResolver is a placeholder for future subdomain-based resolution
// planned in S11. All calls return ErrNotImplemented until that sprint.
type SubdomainResolver struct{}

// NewSubdomainResolver constructs a SubdomainResolver stub.
func NewSubdomainResolver() *SubdomainResolver { return &SubdomainResolver{} }

// Resolve always returns ("", ErrNotImplemented) until S11 implements
// subdomain-to-tenant lookup.
func (r *SubdomainResolver) Resolve(_ echo.Context) (string, error) {
	return "", ErrNotImplemented
}
