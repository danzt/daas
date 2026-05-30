CREATE TYPE po_status AS ENUM ('draft', 'ordered', 'received', 'cancelled');

CREATE TABLE suppliers (
    id              UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID            NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            VARCHAR(200)    NOT NULL,
    rif             VARCHAR(20),
    contact_name    VARCHAR(200),
    email           VARCHAR(254),
    phone           VARCHAR(50),
    address         TEXT,
    notes           TEXT,
    active          BOOLEAN         NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, rif)
);

CREATE TABLE purchase_orders (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    supplier_id     UUID        NOT NULL REFERENCES suppliers(id),
    status          po_status   NOT NULL DEFAULT 'draft',
    notes           TEXT,
    total           NUMERIC(12,2) NOT NULL DEFAULT 0,
    ordered_at      TIMESTAMPTZ,
    received_at     TIMESTAMPTZ,
    created_by      UUID        NOT NULL REFERENCES tenant_users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE purchase_order_lines (
    id                  UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    po_id               UUID            NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_id          UUID            NOT NULL REFERENCES products(id),
    description         VARCHAR(500)    NOT NULL,
    quantity_ordered    NUMERIC(12,3)   NOT NULL CHECK (quantity_ordered > 0),
    unit_cost           NUMERIC(12,2)   NOT NULL CHECK (unit_cost > 0),
    subtotal            NUMERIC(12,2)   NOT NULL,
    sort_order          INT             NOT NULL DEFAULT 0
);

-- ─── RLS ────────────────────────────────────────────────────────────────────

ALTER TABLE suppliers            ENABLE ROW LEVEL SECURITY;
ALTER TABLE purchase_orders      ENABLE ROW LEVEL SECURITY;
ALTER TABLE purchase_order_lines ENABLE ROW LEVEL SECURITY;

CREATE POLICY service_role_bypass ON suppliers
    TO service_role USING (true) WITH CHECK (true);
CREATE POLICY service_role_bypass ON purchase_orders
    TO service_role USING (true) WITH CHECK (true);
CREATE POLICY service_role_bypass ON purchase_order_lines
    TO service_role USING (true) WITH CHECK (true);

CREATE POLICY tenant_isolation ON suppliers
    FOR ALL TO authenticated
    USING (tenant_id = (current_setting('app.tenant_id', true))::UUID);
CREATE POLICY tenant_isolation ON purchase_orders
    FOR ALL TO authenticated
    USING (tenant_id = (current_setting('app.tenant_id', true))::UUID);
CREATE POLICY tenant_isolation ON purchase_order_lines
    FOR ALL TO authenticated
    USING (
        po_id IN (
            SELECT id FROM purchase_orders
            WHERE tenant_id = (current_setting('app.tenant_id', true))::UUID
        )
    );
