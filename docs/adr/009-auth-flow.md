# ADR-009: Auth Flow — Supabase JWT + Go JWKS Verification

Status: Accepted

## Decision

Use Supabase Auth for login/registration; Go middleware verifies JWT signatures using Supabase's JWKS endpoint (cached with 1h TTL), extracts tenant_id from a custom JWT claim set by Go Admin API at registration, and injects the tenant context for every request. The Supabase client on the frontend handles automatic token refresh; Go is the authority for all tenant and business logic — Supabase is infrastructure only.
