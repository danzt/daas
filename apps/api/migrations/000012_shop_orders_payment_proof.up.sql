-- ─── Sprint 8 PR-2: Payment Proof Upload ─────────────────────────────────────

ALTER TABLE shop_orders
    ADD COLUMN payment_proof_url         TEXT,
    ADD COLUMN payment_proof_filename    TEXT,
    ADD COLUMN payment_proof_uploaded_at TIMESTAMPTZ,
    ADD COLUMN payment_method_id         UUID REFERENCES tenant_payment_methods(id) ON DELETE SET NULL,
    ADD COLUMN payment_reference         VARCHAR(160);

CREATE INDEX idx_shop_orders_payment_proof
    ON shop_orders(tenant_id, status, payment_proof_uploaded_at DESC)
    WHERE payment_proof_url IS NOT NULL;
