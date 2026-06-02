-- Migration 000009: tenant slug column + PL/pgSQL functions
-- S6-T2: B2C Public Storefront — Sprint 6 Foundation
--
-- Adds a URL-friendly slug to every tenant for public storefront routing.
-- Creates PL/pgSQL helpers:
--   slugify(text)                          — pure text normalization
--   generate_tenant_slug(text, uuid)       — conflict-loop dedup
--   lookup_tenant_by_slug(text)            — SECURITY DEFINER (bypasses RLS for PathResolver pre-auth)

-- Enable unaccent extension for accent-aware slug generation (bundled contrib module).
CREATE EXTENSION IF NOT EXISTS unaccent;

-- 1. slugify(text) — IMMUTABLE pure text transform
--    Normalizes a display name into a URL-safe slug:
--    - Strips accents via unaccent (á → a, Ñ → N → n, etc.)
--    - Lowercases
--    - Replaces all non-alphanumeric runs with a single hyphen
--    - Trims leading/trailing hyphens
--    - Truncates to 63 characters
CREATE OR REPLACE FUNCTION slugify(p_text TEXT)
RETURNS TEXT AS $$
DECLARE
    v_slug TEXT;
BEGIN
    -- Strip accents then lowercase.
    v_slug := lower(unaccent(p_text));
    -- Replace every run of non-alphanumeric characters with a hyphen.
    v_slug := regexp_replace(v_slug, '[^a-z0-9]+', '-', 'g');
    -- Trim leading and trailing hyphens.
    v_slug := trim(both '-' from v_slug);
    -- Collapse any remaining consecutive hyphens (defensive, already handled above).
    v_slug := regexp_replace(v_slug, '-+', '-', 'g');
    -- Truncate to maximum slug length.
    v_slug := substring(v_slug, 1, 63);
    RETURN v_slug;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- 2. generate_tenant_slug(p_name, p_tenant_id) — conflict-resolution loop
--    Returns a unique slug for the given tenant by trying:
--      base_slug → base_slug-2 → base_slug-3 → ... → base_slug-99
--    Raises an exception if all 99 variants are taken (extremely unlikely).
--    When base_slug is empty (e.g. input was '###'), falls back to 'tenant-<uuid>'.
CREATE OR REPLACE FUNCTION generate_tenant_slug(p_name TEXT, p_tenant_id UUID)
RETURNS TEXT AS $$
DECLARE
    v_base      TEXT;
    v_candidate TEXT;
    v_n         INT := 2;
    v_exists    INT;
BEGIN
    v_base := slugify(p_name);
    -- Fallback when slug strips to empty string.
    IF v_base = '' THEN
        v_base := 'tenant-' || replace(p_tenant_id::text, '-', '');
    END IF;

    v_candidate := v_base;
    LOOP
        SELECT COUNT(*) INTO v_exists
        FROM tenants
        WHERE slug = v_candidate AND id != p_tenant_id;

        IF v_exists = 0 THEN
            RETURN v_candidate;
        END IF;

        IF v_n > 99 THEN
            RAISE EXCEPTION 'slug conflict: could not generate unique slug for % after 99 attempts', p_name;
        END IF;

        -- Ensure the suffixed candidate does not exceed 63 chars.
        v_candidate := substring(v_base, 1, 60) || '-' || v_n;
        v_n := v_n + 1;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 3. lookup_tenant_by_slug(p_slug) — SECURITY DEFINER
--    Called by PathResolver BEFORE set_config is called on a connection,
--    so RLS is not yet active. SECURITY DEFINER runs as the function owner
--    (migration role), bypassing RLS entirely for this narrow lookup.
--    Returns the tenant UUID if active, NULL otherwise.
CREATE OR REPLACE FUNCTION lookup_tenant_by_slug(p_slug TEXT)
RETURNS UUID AS $$
DECLARE
    v_result UUID;
BEGIN
    SELECT id INTO v_result
    FROM tenants
    WHERE slug = p_slug AND status = 'active';
    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER STABLE;

-- 4. Add slug column with empty-string default to allow backfill.
ALTER TABLE tenants ADD COLUMN slug VARCHAR(63) NOT NULL DEFAULT '';

-- 5. Backfill existing tenants.
UPDATE tenants SET slug = generate_tenant_slug(name, id) WHERE slug = '';

-- 6. Remove the temporary default so new rows must always provide a slug explicitly.
ALTER TABLE tenants ALTER COLUMN slug DROP DEFAULT;

-- 7. Enforce uniqueness at the DB level (last-resort safety net for concurrent inserts).
ALTER TABLE tenants ADD CONSTRAINT tenants_slug_unique UNIQUE (slug);

-- 8. Enforce slug format: lowercase alphanumeric with interior hyphens, no leading/trailing.
--    Min length 2 (single char slugs are ambiguous), max 63.
ALTER TABLE tenants ADD CONSTRAINT tenants_slug_format
    CHECK (slug ~ '^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$');

-- 9. Explicit B-tree index for O(1) lookup by slug (PathResolver hot path).
--    The UNIQUE constraint creates an index too; this one is explicit for documentation.
CREATE INDEX idx_tenants_slug ON tenants(slug);
