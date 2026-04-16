-- 049_erp_safety_exposures.up.sql
-- Exposiciones a riesgos laborales y consultas médicas

-- Employee risk exposures
CREATE TABLE erp_employee_risk_exposures (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     TEXT NOT NULL,
    entity_id     UUID NOT NULL REFERENCES erp_entities(id),
    risk_agent_id UUID NOT NULL REFERENCES erp_risk_agents(id),
    section_id    UUID REFERENCES erp_departments(id),
    exposed_from  DATE NOT NULL,
    exposed_until DATE,
    notes         TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, entity_id, risk_agent_id, exposed_from)
);
CREATE INDEX idx_risk_exposure_entity ON erp_employee_risk_exposures (tenant_id, entity_id);
ALTER TABLE erp_employee_risk_exposures ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_employee_risk_exposures
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Medical consultations log
CREATE TABLE erp_medical_consultations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    TEXT NOT NULL,
    entity_id    UUID REFERENCES erp_entities(id),
    patient_name TEXT NOT NULL DEFAULT '',
    consult_date DATE NOT NULL,
    consult_time TIME,
    symptoms     TEXT NOT NULL DEFAULT '',
    prescription TEXT NOT NULL DEFAULT '',
    medic_user   TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_medical_date   ON erp_medical_consultations (tenant_id, consult_date);
CREATE INDEX idx_medical_entity ON erp_medical_consultations (tenant_id, entity_id);
ALTER TABLE erp_medical_consultations ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_medical_consultations
    USING (tenant_id = current_setting('app.tenant_id', true));
