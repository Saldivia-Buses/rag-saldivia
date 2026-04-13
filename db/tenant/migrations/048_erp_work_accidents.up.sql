-- 048_erp_work_accidents.up.sql
-- Registro de accidentes de trabajo

CREATE TABLE erp_work_accidents (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        TEXT NOT NULL,
    entity_id        UUID REFERENCES erp_entities(id),
    accident_type_id UUID REFERENCES erp_accident_types(id),
    body_part_id     UUID REFERENCES erp_body_parts(id),
    section_id       UUID REFERENCES erp_departments(id),
    incident_date    DATE NOT NULL,
    recovery_date    DATE,
    lost_days        INTEGER NOT NULL DEFAULT 0,
    observations     TEXT NOT NULL DEFAULT '',
    reported_by      TEXT NOT NULL DEFAULT '',
    status           TEXT NOT NULL DEFAULT 'open'
                     CHECK (status IN ('open','investigating','closed')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_accidents_entity ON erp_work_accidents (tenant_id, entity_id);
CREATE INDEX idx_accidents_date   ON erp_work_accidents (incident_date);
CREATE INDEX idx_accidents_status ON erp_work_accidents (tenant_id, status);
ALTER TABLE erp_work_accidents ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_work_accidents
    USING (tenant_id = current_setting('app.tenant_id', true));
