package tenant

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a user's permission level within a tenant.
type Role string

const (
	// RoleOwner has full administrative access: manage users, configure
	// integrations, modify tenant settings.
	RoleOwner Role = "owner"
	// RoleEmployee has read/write access to operational resources (products,
	// inventory, sales) but cannot manage tenant configuration or users.
	RoleEmployee Role = "employee"
)

// IsOwner returns true when the role grants administrative privileges.
func (r Role) IsOwner() bool {
	return r == RoleOwner
}

// UserProfile links a Supabase Auth user to a tenant with a specific role.
// The SupabaseUID is the source of truth for identity; this record adds
// tenant membership and authorization context.
type UserProfile struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	SupabaseUID uuid.UUID
	Email       string
	Role        Role
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
