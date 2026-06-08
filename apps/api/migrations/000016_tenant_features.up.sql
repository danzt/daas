CREATE TABLE tenant_features (
  tenant_id  UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  feature    VARCHAR(50) NOT NULL CHECK (feature IN ('storefront', 'fiscal_invoicing')),
  enabled    BOOLEAN     NOT NULL DEFAULT FALSE,
  enabled_at TIMESTAMPTZ,
  PRIMARY KEY (tenant_id, feature)
);

ALTER TABLE tenant_features ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON tenant_features
  USING (tenant_id = current_setting('app.tenant_id', TRUE)::UUID);

CREATE POLICY service_role_bypass ON tenant_features
  TO service_role USING (TRUE);

INSERT INTO tenant_features (tenant_id, feature, enabled, enabled_at)
SELECT id, 'storefront', TRUE, NOW() FROM tenants
ON CONFLICT DO NOTHING;

INSERT INTO tenant_features (tenant_id, feature, enabled, enabled_at)
SELECT id, 'fiscal_invoicing', TRUE, NOW() FROM tenants
ON CONFLICT DO NOTHING;
