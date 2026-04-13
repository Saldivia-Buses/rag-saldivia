-- 046_erp_manufacturing_certificates.up.sql
-- Certificados de fabricación: certificación final de cada unidad producida

CREATE TABLE erp_manufacturing_certificates (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           TEXT NOT NULL,
    unit_id             UUID NOT NULL REFERENCES erp_manufacturing_units(id),
    certificate_number  TEXT NOT NULL DEFAULT '',
    cert_type           TEXT NOT NULL DEFAULT 'homologation'
                        CHECK (cert_type IN ('homologation','quality','safety','delivery','warranty')),
    issued_by           UUID REFERENCES erp_entities(id),
    issued_at           TIMESTAMPTZ,
    valid_from          DATE,
    valid_until         DATE,
    authority           TEXT NOT NULL DEFAULT '',
    document_url        TEXT NOT NULL DEFAULT '',
    observations        TEXT NOT NULL DEFAULT '',
    status              TEXT NOT NULL DEFAULT 'draft'
                        CHECK (status IN ('draft','issued','revoked','expired')),
    revoked_at          TIMESTAMPTZ,
    revocation_reason   TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_mfg_certs_unit    ON erp_manufacturing_certificates (tenant_id, unit_id);
CREATE INDEX idx_mfg_certs_status  ON erp_manufacturing_certificates (tenant_id, status);
CREATE INDEX idx_mfg_certs_expiry  ON erp_manufacturing_certificates (valid_until);
ALTER TABLE erp_manufacturing_certificates ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_manufacturing_certificates
    USING (tenant_id = current_setting('app.tenant_id', true));
