-- name: GetTenant :one
SELECT * FROM tenants WHERE id = $1;

-- name: CreateTenant :one
INSERT INTO tenants (name, fiscal_id, country_code)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateTenantStatus :one
UPDATE tenants SET status = $2 WHERE id = $1 RETURNING *;

-- name: GetUserBySupabaseUID :one
SELECT * FROM tenant_users WHERE supabase_uid = $1;

-- name: CreateTenantUser :one
INSERT INTO tenant_users (tenant_id, supabase_uid, email, role)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListTenantUsers :many
SELECT * FROM tenant_users WHERE tenant_id = $1 ORDER BY created_at;

-- name: UpdateTenantUserActive :one
UPDATE tenant_users SET active = $2 WHERE id = $1 AND tenant_id = $3 RETURNING *;

-- name: GetTenantIntegration :one
SELECT * FROM tenant_integrations
WHERE tenant_id = $1 AND integration_type = $2;

-- name: UpsertTenantIntegration :one
INSERT INTO tenant_integrations (tenant_id, integration_type, config_encrypted, active)
VALUES ($1, $2, $3, TRUE)
ON CONFLICT (tenant_id, integration_type)
DO UPDATE SET config_encrypted = $3, active = TRUE, updated_at = NOW()
RETURNING *;
