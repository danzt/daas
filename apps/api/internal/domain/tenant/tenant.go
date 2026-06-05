// Package tenant contains the domain entities and business rules for tenant
// management. This package has ZERO imports from Echo, pgx, or any external
// framework — it is pure Go.
package tenant

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the lifecycle state of a tenant.
type Status string

const (
	// StatusActive means the tenant is operational.
	StatusActive Status = "active"
	// StatusSuspended means the tenant has been temporarily suspended.
	// All business endpoints return HTTP 403 for suspended tenants.
	StatusSuspended Status = "suspended"
	// StatusCancelled means the tenant account has been permanently closed.
	StatusCancelled Status = "cancelled"
)

// Tenant is the root aggregate for a business account.
// Each tenant is fully isolated from others via RLS and Go middleware.
type Tenant struct {
	ID          uuid.UUID
	Name        string
	FiscalID    string // RIF (Venezuela), RNC (Dominican Republic), EIN (USA)
	CountryCode string // ISO-3166 alpha-2, e.g. "VE", "DO", "US"
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Storefront branding fields (S9-T2).
	// All are optional — the storefront falls back to DaaS defaults when empty.
	BrandingLogoURL      string `json:"branding_logo_url,omitempty"`
	BrandingPrimaryColor string `json:"branding_primary_color,omitempty"` // hex e.g. #7C3AED
	BrandingBannerURL    string `json:"branding_banner_url,omitempty"`
	BrandingStoreName    string `json:"branding_store_name,omitempty"` // defaults to Name
	BrandingTagline      string `json:"branding_tagline,omitempty"`
}

// IsActive returns true when the tenant is in the active state and allowed
// to perform business operations.
func (t *Tenant) IsActive() bool {
	return t.Status == StatusActive
}
