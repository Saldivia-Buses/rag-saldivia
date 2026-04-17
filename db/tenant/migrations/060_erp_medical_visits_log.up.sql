-- 060_erp_medical_visits_log.up.sql
-- Daily medical visit log (standalone, not linked to formal leaves).
-- Source: Histrix PARTE_MEDICO_DIARIO — has only free-text operator/patient names, no FK to PERSONAL.

CREATE TABLE erp_medical_visits_log (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         TEXT NOT NULL,
    entity_id         UUID REFERENCES erp_entities(id),  -- nullable: legacy data has free-text names
    visit_date        DATE NOT NULL,
    visit_time        TIME,
    operator_username TEXT NOT NULL DEFAULT '',
    patient_name      TEXT NOT NULL DEFAULT '',
    symptoms          TEXT NOT NULL DEFAULT '',
    prescription      TEXT NOT NULL DEFAULT '',
    metadata          JSONB NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_medical_visits_tenant ON erp_medical_visits_log (tenant_id);
CREATE INDEX idx_medical_visits_date ON erp_medical_visits_log (tenant_id, visit_date);
CREATE INDEX idx_medical_visits_entity ON erp_medical_visits_log (tenant_id, entity_id) WHERE entity_id IS NOT NULL;

ALTER TABLE erp_medical_visits_log ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_medical_visits_log
    USING (tenant_id = current_setting('app.tenant_id', true));

COMMENT ON TABLE erp_medical_visits_log IS
    'Daily medical consultation log. Free-form visits without FK to employees (legacy PARTE_MEDICO_DIARIO).';
