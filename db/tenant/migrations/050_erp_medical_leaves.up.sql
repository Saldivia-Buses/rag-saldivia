-- 050_erp_medical_leaves.up.sql
-- Licencias médicas y ausentismo

CREATE TABLE erp_medical_leaves (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    TEXT NOT NULL,
    entity_id    UUID NOT NULL REFERENCES erp_entities(id),
    body_part_id UUID REFERENCES erp_body_parts(id),
    accident_id  UUID REFERENCES erp_work_accidents(id),
    leave_type   TEXT NOT NULL DEFAULT 'illness'
                 CHECK (leave_type IN ('illness','accident','vacation','leave','other')),
    date_from    DATE NOT NULL,
    date_to      DATE NOT NULL,
    working_days INTEGER NOT NULL DEFAULT 0,
    observations TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending','approved','rejected','closed')),
    approved_by  TEXT NOT NULL DEFAULT '',
    approved_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_leaves_entity ON erp_medical_leaves (tenant_id, entity_id);
CREATE INDEX idx_leaves_dates  ON erp_medical_leaves (date_from, date_to);
CREATE INDEX idx_leaves_type   ON erp_medical_leaves (tenant_id, leave_type);
ALTER TABLE erp_medical_leaves ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_medical_leaves
    USING (tenant_id = current_setting('app.tenant_id', true));
