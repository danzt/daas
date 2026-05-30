-- Ensure service_role exists (local dev compatibility — Supabase creates this automatically)
DO $$ BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'service_role') THEN
    CREATE ROLE service_role;
  END IF;
END $$;

-- Stock snapshot per product (denormalized for O(1) reads)
CREATE TABLE product_stock (
  product_id UUID PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  quantity_on_hand NUMERIC(12,3) NOT NULL DEFAULT 0,
  last_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT non_negative_stock CHECK (quantity_on_hand >= 0)
);

CREATE INDEX idx_product_stock_tenant_id ON product_stock(tenant_id);

-- Inventory movements — immutable audit log
CREATE TABLE inventory_movements (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  product_id UUID NOT NULL REFERENCES products(id),
  type VARCHAR(20) NOT NULL CHECK (type IN ('entry', 'exit', 'adjustment')),
  quantity NUMERIC(12,3) NOT NULL,
  unit_cost NUMERIC(12,2),
  reference_type VARCHAR(30) NOT NULL
    CHECK (reference_type IN ('purchase_order', 'sale', 'manual_adjustment', 'opening')),
  reference_id UUID,
  notes TEXT,
  created_by UUID NOT NULL REFERENCES tenant_users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inventory_movements_tenant ON inventory_movements(tenant_id);
CREATE INDEX idx_inventory_movements_product ON inventory_movements(product_id);
CREATE INDEX idx_inventory_movements_created ON inventory_movements(tenant_id, created_at DESC);

-- Auto-initialize stock row when a product is created
CREATE OR REPLACE FUNCTION init_product_stock()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO product_stock (product_id, tenant_id, quantity_on_hand)
  VALUES (NEW.id, NEW.tenant_id, 0)
  ON CONFLICT (product_id) DO NOTHING;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER initialize_product_stock
  AFTER INSERT ON products
  FOR EACH ROW EXECUTE FUNCTION init_product_stock();

-- Backfill stock rows for products created before this migration
INSERT INTO product_stock (product_id, tenant_id, quantity_on_hand)
SELECT id, tenant_id, 0 FROM products
ON CONFLICT (product_id) DO NOTHING;

-- RLS
ALTER TABLE product_stock ENABLE ROW LEVEL SECURITY;
ALTER TABLE inventory_movements ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON product_stock
  USING (tenant_id = current_setting('app.tenant_id', TRUE)::UUID);

CREATE POLICY tenant_isolation ON inventory_movements
  USING (tenant_id = current_setting('app.tenant_id', TRUE)::UUID);

CREATE POLICY service_role_bypass ON product_stock
  TO service_role USING (TRUE);

CREATE POLICY service_role_bypass ON inventory_movements
  TO service_role USING (TRUE);
