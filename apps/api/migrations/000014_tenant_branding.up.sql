-- S9-T2: tenant storefront branding fields
ALTER TABLE tenants
    ADD COLUMN branding_logo_url      TEXT,
    ADD COLUMN branding_primary_color VARCHAR(7),
    ADD COLUMN branding_banner_url    TEXT,
    ADD COLUMN branding_store_name    VARCHAR(120),
    ADD COLUMN branding_tagline       VARCHAR(200);
