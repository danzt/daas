package tenant_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/danzt/daas/api/internal/domain/tenant"
)

func TestRole_IsOwner(t *testing.T) {
	tests := []struct {
		name     string
		role     tenant.Role
		expected bool
	}{
		{
			name:     "owner role returns true",
			role:     tenant.RoleOwner,
			expected: true,
		},
		{
			name:     "employee role returns false",
			role:     tenant.RoleEmployee,
			expected: false,
		},
		{
			name:     "empty role returns false",
			role:     tenant.Role(""),
			expected: false,
		},
		{
			name:     "unknown role returns false",
			role:     tenant.Role("superadmin"),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.role.IsOwner()
			if got != tc.expected {
				t.Errorf("Role(%q).IsOwner() = %v, want %v", tc.role, got, tc.expected)
			}
		})
	}
}

func TestTenant_IsActive(t *testing.T) {
	base := tenant.Tenant{
		ID:          uuid.New(),
		Name:        "Acme Corp",
		CountryCode: "VE",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name     string
		status   tenant.Status
		expected bool
	}{
		{
			name:     "active status returns true",
			status:   tenant.StatusActive,
			expected: true,
		},
		{
			name:     "suspended status returns false",
			status:   tenant.StatusSuspended,
			expected: false,
		},
		{
			name:     "cancelled status returns false",
			status:   tenant.StatusCancelled,
			expected: false,
		},
		{
			name:     "empty status returns false",
			status:   tenant.Status(""),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			te := base
			te.Status = tc.status
			got := te.IsActive()
			if got != tc.expected {
				t.Errorf("Tenant.IsActive() with status %q = %v, want %v", tc.status, got, tc.expected)
			}
		})
	}
}
