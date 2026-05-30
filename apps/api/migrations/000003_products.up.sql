-- Ensure service_role exists (local dev compatibility — Supabase creates this automatically)
DO $$ BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'service_role') THEN
    CREATE ROLE service_role;
  END IF;
END $$;

-- Product categories
CREATE TABLE product_categories (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  parent_id UUID REFERENCES product_categories(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, name)
);

CREATE INDEX idx_product_categories_tenant_id ON product_categories(tenant_id);

-- Products
CREATE TABLE products (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  sku VARCHAR(100),
  barcode VARCHAR(100),
  description TEXT,
  category_id UUID REFERENCES product_categories(id) ON DELETE SET NULL,
  is_fiscal BOOLEAN NOT NULL,
  fiscal_price NUMERIC(12,2),
  internal_price NUMERIC(12,2),
  tax_rate NUMERIC(5,2),
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, sku),
  CONSTRAINT fiscal_fields_required CHECK (
    (is_fiscal = TRUE AND fiscal_price IS NOT NULL AND fiscal_price > 0 AND tax_rate IS NOT NULL)
    OR
    (is_fiscal = FALSE AND internal_price IS NOT NULL AND internal_price > 0)
  ),
  CONSTRAINT no_tax_on_internal CHECK (
    is_fiscal = TRUE OR tax_rate IS NULL
  )
);

CREATE INDEX idx_products_tenant_id ON products(tenant_id);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_active ON products(tenant_id, active);

-- RLS
ALTER TABLE product_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE products ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON product_categories
  USING (tenant_id = current_setting('app.tenant_id', TRUE)::UUID);

CREATE POLICY tenant_isolation ON products
  USING (tenant_id = current_setting('app.tenant_id', TRUE)::UUID);

CREATE POLICY service_role_bypass ON product_categories
  TO service_role USING (TRUE);

CREATE POLICY service_role_bypass ON products
  TO service_role USING (TRUE);

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();
