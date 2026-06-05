-- ─── Sprint 8 PR-2: Payment Proof Upload (rollback) ──────────────────────────

DROP INDEX IF EXISTS idx_shop_orders_payment_proof;

ALTER TABLE shop_orders
    DROP COLUMN IF EXISTS payment_proof_url,
    DROP COLUMN IF EXISTS payment_proof_filename,
    DROP COLUMN IF EXISTS payment_proof_uploaded_at,
    DROP COLUMN IF EXISTS payment_method_id,
    DROP COLUMN IF EXISTS payment_reference;
