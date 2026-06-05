-- S9-T2: rollback tenant storefront branding fields
ALTER TABLE tenants
    DROP COLUMN IF EXISTS branding_logo_url,
    DROP COLUMN IF EXISTS branding_primary_color,
    DROP COLUMN IF EXISTS branding_banner_url,
    DROP COLUMN IF EXISTS branding_store_name,
    DROP COLUMN IF EXISTS branding_tagline;
