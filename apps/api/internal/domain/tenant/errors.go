package tenant

import "errors"

// Domain errors for the tenant module. These are pure sentinel errors —
// no HTTP status codes or framework types here. HTTP translation happens
// in the handler layer.

var (
	// ErrTenantNotFound is returned when a tenant lookup yields no results.
	ErrTenantNotFound = errors.New("tenant not found")

	// ErrUserNotFound is returned when a user profile lookup yields no results.
	ErrUserNotFound = errors.New("user not found")

	// ErrDuplicateEmail is returned when an attempt to register an email that
	// already exists within the tenant (or globally in Supabase Auth).
	ErrDuplicateEmail = errors.New("email already registered")

	// ErrInvalidCountryCode is returned when a provided country code is not
	// a valid ISO-3166 alpha-2 uppercase pair.
	ErrInvalidCountryCode = errors.New("invalid country code: must be 2 uppercase letters (ISO-3166 alpha-2)")

	// ErrTenantSuspended is returned when a suspended tenant attempts to
	// perform business operations. Callers should translate this to HTTP 403.
	ErrTenantSuspended = errors.New("tenant is suspended")
)
