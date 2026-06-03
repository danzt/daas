-- ─── Sprint 8 PR-1: Tenant Payment Methods ──────────────────────────────────

CREATE TYPE payment_method_type AS ENUM (
    'pago_movil',      -- Venezuela mobile payment
    'transfer_bank',   -- bank transfer (any country)
    'zelle',           -- USD via Zelle (US bank)
    'paypal',
    'usdt',            -- crypto (TRC20/ERC20/BEP20 specified in details)
    'cash',            -- efectivo / cash on delivery
    'other'            -- free-form (notes-driven)
);

CREATE TABLE tenant_payment_methods (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    type        payment_method_type NOT NULL,
    label       VARCHAR(120),         -- optional friendly name (e.g. "Banesco · Cuenta Corriente")
    details     JSONB NOT NULL DEFAULT '{}'::jsonb,
    currency    VARCHAR(8),            -- VES, USD, USDT, etc.
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenant_payment_methods_tenant ON tenant_payment_methods(tenant_id, active, sort_order);

ALTER TABLE tenant_payment_methods ENABLE ROW LEVEL SECURITY;

CREATE POLICY service_role_bypass ON tenant_payment_methods
    TO service_role USING (true) WITH CHECK (true);

CREATE POLICY tenant_isolation ON tenant_payment_methods
    FOR ALL TO authenticated
    USING (tenant_id = (current_setting('app.tenant_id', true))::UUID);
