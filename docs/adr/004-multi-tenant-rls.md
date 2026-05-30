# ADR-004: Multi-Tenant — Shared Schema + RLS + Go Middleware

Status: Accepted

## Decision

Implement multi-tenancy with a shared schema where all business tables have tenant_id UUID NOT NULL, a double-barrier approach: Go middleware extracts tenant_id from the Supabase JWT and executes SET LOCAL app.tenant_id before each query, and PostgreSQL RLS policies enforce isolation at the database level as a safety net for any query that omits the tenant filter.
