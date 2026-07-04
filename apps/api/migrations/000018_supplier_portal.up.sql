-- Fase 2a — Portal de Proveedores: link público + catálogo del proveedor.
-- portal_token: token del link público del proveedor (NULL = sin link).
-- supplier_catalog_items: catálogo que el proveedor carga desde su portal.
ALTER TABLE suppliers ADD COLUMN portal_token VARCHAR UNIQUE;

CREATE TABLE supplier_catalog_items (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
  name        VARCHAR NOT NULL,
  sku         VARCHAR,
  cost        NUMERIC,
  unit        VARCHAR,
  barcode     VARCHAR,
  description TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_supplier_catalog_supplier ON supplier_catalog_items(supplier_id);

ALTER TABLE supplier_catalog_items ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON supplier_catalog_items
  USING (tenant_id = current_setting('app.tenant_id', TRUE)::UUID);

CREATE POLICY service_role_bypass ON supplier_catalog_items
  TO service_role USING (TRUE);
