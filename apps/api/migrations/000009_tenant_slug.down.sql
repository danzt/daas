-- Migration 000009 down: reverse tenant slug column + PL/pgSQL functions
-- Removes everything added by 000009_tenant_slug.up.sql in reverse order.

DROP INDEX IF EXISTS idx_tenants_slug;

ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_slug_format;

ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_slug_unique;

ALTER TABLE tenants DROP COLUMN IF EXISTS slug;

DROP FUNCTION IF EXISTS lookup_tenant_by_slug(TEXT);

DROP FUNCTION IF EXISTS generate_tenant_slug(TEXT, UUID);

DROP FUNCTION IF EXISTS slugify(TEXT);
