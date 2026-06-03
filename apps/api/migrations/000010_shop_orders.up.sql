-- ─── Sprint 7 PR-2: Shop Orders ────────────────────────────────────────────────

CREATE TYPE shop_order_status AS ENUM (
    'pending',      -- created by customer, awaiting payment proof
    'paid',         -- payment verified by tenant
    'fulfilled',    -- tenant prepared/dispatched the order
    'delivered',    -- customer received
    'cancelled'     -- voided by tenant or customer
);

CREATE TABLE shop_orders (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id           UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    -- Customer info (no account required)
    customer_name       VARCHAR(200) NOT NULL,
    customer_email      VARCHAR(200) NOT NULL,
    customer_phone      VARCHAR(50),
    -- Shipping
    shipping_address    TEXT NOT NULL,
    shipping_city       VARCHAR(100),
    shipping_notes      TEXT,
    -- Money
    subtotal            NUMERIC(12,2) NOT NULL CHECK (subtotal >= 0),
    shipping_cost       NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (shipping_cost >= 0),
    total               NUMERIC(12,2) NOT NULL CHECK (total >= 0),
    -- Lifecycle
    status              shop_order_status NOT NULL DEFAULT 'pending',
    notes               TEXT,
    -- Magic access token so the customer (no account) can fetch their order
    access_token        VARCHAR(64) NOT NULL UNIQUE,
    -- Timestamps
    paid_at             TIMESTAMPTZ,
    fulfilled_at        TIMESTAMPTZ,
    delivered_at        TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE shop_order_lines (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id        UUID NOT NULL REFERENCES shop_orders(id) ON DELETE CASCADE,
    product_id      UUID NOT NULL REFERENCES products(id),
    -- Snapshot of product info at time of purchase
    name            VARCHAR(200) NOT NULL,
    unit_price      NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0),
    quantity        INT NOT NULL CHECK (quantity > 0),
    subtotal        NUMERIC(12,2) NOT NULL,
    is_fiscal       BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order      INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_shop_orders_tenant_status ON shop_orders(tenant_id, status, created_at DESC);
CREATE INDEX idx_shop_orders_access_token ON shop_orders(access_token);
CREATE INDEX idx_shop_order_lines_order ON shop_order_lines(order_id);

-- RLS
ALTER TABLE shop_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE shop_order_lines ENABLE ROW LEVEL SECURITY;

-- service_role bypass (Go backend)
CREATE POLICY service_role_bypass ON shop_orders
    TO service_role USING (true) WITH CHECK (true);
CREATE POLICY service_role_bypass ON shop_order_lines
    TO service_role USING (true) WITH CHECK (true);

-- authenticated users (tenant admin) see only their tenant's orders
CREATE POLICY tenant_isolation ON shop_orders
    FOR ALL TO authenticated
    USING (tenant_id = (current_setting('app.tenant_id', true))::UUID);
CREATE POLICY tenant_isolation ON shop_order_lines
    FOR ALL TO authenticated
    USING (
        order_id IN (
            SELECT id FROM shop_orders
            WHERE tenant_id = (current_setting('app.tenant_id', true))::UUID
        )
    );
