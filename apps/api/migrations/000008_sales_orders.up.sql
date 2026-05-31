-- ─── Sprint 9a: Sales Orders ──────────────────────────────────────────────────

CREATE TYPE sale_order_status AS ENUM ('draft', 'confirmed', 'invoiced', 'cancelled');

CREATE TABLE sales_orders (
    id                  UUID                PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id           UUID                NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    status              sale_order_status   NOT NULL DEFAULT 'draft',
    customer_name       VARCHAR(200),
    customer_id_type    customer_id_type    NOT NULL DEFAULT 'anonymous',
    customer_id_number  VARCHAR(50),
    notes               TEXT,
    total               NUMERIC(12,2)       NOT NULL DEFAULT 0,
    confirmed_at        TIMESTAMPTZ,
    invoiced_at         TIMESTAMPTZ,
    invoice_id          UUID                REFERENCES internal_invoices(id),
    created_by          UUID                NOT NULL REFERENCES tenant_users(id),
    created_at          TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE TABLE sales_order_lines (
    id          UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id    UUID            NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    product_id  UUID            NOT NULL REFERENCES products(id),
    description VARCHAR(500)    NOT NULL,
    quantity    NUMERIC(12,3)   NOT NULL CHECK (quantity > 0),
    unit_price  NUMERIC(12,2)   NOT NULL CHECK (unit_price > 0),
    subtotal    NUMERIC(12,2)   NOT NULL,
    sort_order  INT             NOT NULL DEFAULT 0
);

CREATE INDEX idx_sales_orders_tenant_date ON sales_orders(tenant_id, created_at DESC);
CREATE INDEX idx_sales_orders_status      ON sales_orders(tenant_id, status);
CREATE INDEX idx_sales_order_lines_order  ON sales_order_lines(order_id);

-- ─── RLS ────────────────────────────────────────────────────────────────────

ALTER TABLE sales_orders      ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales_order_lines ENABLE ROW LEVEL SECURITY;

CREATE POLICY service_role_bypass ON sales_orders
    TO service_role USING (true) WITH CHECK (true);
CREATE POLICY service_role_bypass ON sales_order_lines
    TO service_role USING (true) WITH CHECK (true);

CREATE POLICY tenant_isolation ON sales_orders
    FOR ALL TO authenticated
    USING (tenant_id = (current_setting('app.tenant_id', true))::UUID);
CREATE POLICY tenant_isolation ON sales_order_lines
    FOR ALL TO authenticated
    USING (
        order_id IN (
            SELECT id FROM sales_orders
            WHERE tenant_id = (current_setting('app.tenant_id', true))::UUID
        )
    );
