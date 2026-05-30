-- ─── Sprint 4: Internal Invoicing ───────────────────────────────────────────

-- Atomic counter for generating correlatives per tenant per year.
-- We use SELECT ... FOR UPDATE on this row instead of per-tenant sequences
-- so we avoid DDL during runtime and keep migration simple.
CREATE TABLE invoice_correlative_seq (
    tenant_id UUID NOT NULL,
    year      INT  NOT NULL,
    last_seq  INT  NOT NULL DEFAULT 0,
    PRIMARY KEY (tenant_id, year)
);

CREATE TYPE customer_id_type AS ENUM ('cedula', 'rif', 'passport', 'anonymous');
CREATE TYPE invoice_status    AS ENUM ('draft', 'issued', 'cancelled');

CREATE TABLE internal_invoices (
    id                 UUID             PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id          UUID             NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    correlative        VARCHAR(20)      UNIQUE,
    customer_name      VARCHAR(200),
    customer_id_type   customer_id_type NOT NULL DEFAULT 'anonymous',
    customer_id_number VARCHAR(50),
    status             invoice_status   NOT NULL DEFAULT 'draft',
    subtotal           NUMERIC(12,2)    NOT NULL DEFAULT 0,
    total              NUMERIC(12,2)    NOT NULL DEFAULT 0,
    notes              TEXT,
    issued_at          TIMESTAMPTZ,
    created_by         UUID             NOT NULL REFERENCES tenant_users(id),
    created_at         TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE internal_invoice_lines (
    id          UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id  UUID          NOT NULL REFERENCES internal_invoices(id) ON DELETE CASCADE,
    product_id  UUID          NOT NULL REFERENCES products(id),
    description VARCHAR(500)  NOT NULL,
    quantity    NUMERIC(12,3) NOT NULL CHECK (quantity > 0),
    unit_price  NUMERIC(12,2) NOT NULL CHECK (unit_price > 0),
    subtotal    NUMERIC(12,2) NOT NULL,
    sort_order  INT           NOT NULL DEFAULT 0
);

-- Indexes
CREATE INDEX idx_internal_invoices_tenant_date ON internal_invoices(tenant_id, created_at DESC);
CREATE INDEX idx_internal_invoices_status      ON internal_invoices(tenant_id, status);
CREATE INDEX idx_internal_invoice_lines_invoice ON internal_invoice_lines(invoice_id);

-- RLS — invoice_correlative_seq
ALTER TABLE invoice_correlative_seq ENABLE ROW LEVEL SECURITY;
CREATE POLICY service_role_bypass ON invoice_correlative_seq
    TO service_role USING (true);
CREATE POLICY tenant_isolation ON invoice_correlative_seq
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- RLS — internal_invoices
ALTER TABLE internal_invoices ENABLE ROW LEVEL SECURITY;
CREATE POLICY service_role_bypass ON internal_invoices
    TO service_role USING (true);
CREATE POLICY tenant_isolation ON internal_invoices
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- RLS — internal_invoice_lines (via parent invoice)
ALTER TABLE internal_invoice_lines ENABLE ROW LEVEL SECURITY;
CREATE POLICY service_role_bypass ON internal_invoice_lines
    TO service_role USING (true);
CREATE POLICY tenant_isolation ON internal_invoice_lines
    USING (invoice_id IN (
        SELECT id FROM internal_invoices
        WHERE tenant_id::text = current_setting('app.tenant_id', true)
    ));
