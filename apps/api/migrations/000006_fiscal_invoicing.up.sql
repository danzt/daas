-- fiscal invoice status lifecycle:
--   draft → pending_fiscal → issued
--                           ↓
--                         failed → (retry) → pending_fiscal
--   any non-cancelled → cancelled
CREATE TYPE fiscal_invoice_status AS ENUM (
    'draft',
    'pending_fiscal',
    'issued',
    'failed',
    'cancelled'
);

CREATE TABLE fiscal_invoices (
    id                  UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id           UUID            NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    -- fields assigned by SENIAT machine on successful send
    fiscal_number       VARCHAR(20),
    machine_serial      VARCHAR(30),
    report_z_number     INT,
    -- customer
    customer_name       VARCHAR(200),
    customer_id_type    customer_id_type NOT NULL DEFAULT 'anonymous',
    customer_id_number  VARCHAR(50),
    -- financial totals
    subtotal_base       NUMERIC(12,2)   NOT NULL DEFAULT 0,  -- base imponible (without tax)
    tax_amount          NUMERIC(12,2)   NOT NULL DEFAULT 0,  -- IVA total
    total               NUMERIC(12,2)   NOT NULL DEFAULT 0,
    -- lifecycle
    status              fiscal_invoice_status NOT NULL DEFAULT 'draft',
    fail_reason         TEXT,
    retry_count         INT             NOT NULL DEFAULT 0,
    notes               TEXT,
    issued_at           TIMESTAMPTZ,
    created_by          UUID            NOT NULL REFERENCES tenant_users(id),
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE TABLE fiscal_invoice_lines (
    id          UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id  UUID            NOT NULL REFERENCES fiscal_invoices(id) ON DELETE CASCADE,
    product_id  UUID            NOT NULL REFERENCES products(id),
    description VARCHAR(500)    NOT NULL,
    quantity    NUMERIC(12,3)   NOT NULL CHECK (quantity > 0),
    unit_price  NUMERIC(12,2)   NOT NULL CHECK (unit_price > 0),  -- fiscal_price
    tax_rate    NUMERIC(5,4)    NOT NULL DEFAULT 0,               -- e.g. 0.1600
    tax_amount  NUMERIC(12,2)   NOT NULL,                         -- quantity * unit_price * tax_rate
    subtotal    NUMERIC(12,2)   NOT NULL,                         -- quantity * unit_price + tax_amount
    sort_order  INT             NOT NULL DEFAULT 0
);

-- Retry queue: rows inserted when a fiscal send fails.
-- A background worker (or manual retry endpoint) polls this table.
CREATE TABLE fiscal_invoice_queue (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id      UUID        NOT NULL REFERENCES fiscal_invoices(id) ON DELETE CASCADE,
    tenant_id       UUID        NOT NULL,
    attempts        INT         NOT NULL DEFAULT 0,
    max_attempts    INT         NOT NULL DEFAULT 3,
    last_error      TEXT,
    next_retry_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (invoice_id)  -- only one queue entry per invoice
);

-- ─── RLS ────────────────────────────────────────────────────────────────────

ALTER TABLE fiscal_invoices      ENABLE ROW LEVEL SECURITY;
ALTER TABLE fiscal_invoice_lines ENABLE ROW LEVEL SECURITY;
ALTER TABLE fiscal_invoice_queue ENABLE ROW LEVEL SECURITY;

-- service_role bypasses all RLS (used by Go backend)
CREATE POLICY service_role_bypass ON fiscal_invoices
    TO service_role USING (true) WITH CHECK (true);
CREATE POLICY service_role_bypass ON fiscal_invoice_lines
    TO service_role USING (true) WITH CHECK (true);
CREATE POLICY service_role_bypass ON fiscal_invoice_queue
    TO service_role USING (true) WITH CHECK (true);

-- authenticated users see only their tenant's data
CREATE POLICY tenant_isolation ON fiscal_invoices
    FOR ALL TO authenticated
    USING (tenant_id = (current_setting('app.tenant_id', true))::UUID);
CREATE POLICY tenant_isolation ON fiscal_invoice_lines
    FOR ALL TO authenticated
    USING (
        invoice_id IN (
            SELECT id FROM fiscal_invoices
            WHERE tenant_id = (current_setting('app.tenant_id', true))::UUID
        )
    );
CREATE POLICY tenant_isolation ON fiscal_invoice_queue
    FOR ALL TO authenticated
    USING (tenant_id = (current_setting('app.tenant_id', true))::UUID);
