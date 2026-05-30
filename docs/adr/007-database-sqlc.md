# ADR-007: Database — Supabase PostgreSQL + sqlc + golang-migrate (No ORM)

Status: Accepted

## Decision

Use Supabase PostgreSQL 17 as the managed database with connection pooling via PgBouncer (transaction mode) and pgxpool in Go; use golang-migrate for schema migrations (SQL files, no ORM); use sqlc to generate type-safe Go code from raw SQL queries. This approach gives full control over executed SQL (critical for multi-tenant RLS correctness) and eliminates the "magic" layer that ORMs introduce.
