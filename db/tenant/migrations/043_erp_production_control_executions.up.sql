-- 043_erp_production_control_executions.up.sql
-- Ejecuciones de control de producción y registro de retrabajos

-- Time-tracked execution entries per production control
CREATE TABLE erp_production_control_executions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     TEXT NOT NULL,
    control_id    UUID NOT NULL REFERENCES erp_production_controls(id) ON DELETE CASCADE,
    operator_id   UUID REFERENCES erp_entities(id),
    causal_id     UUID REFERENCES erp_production_control_causals(id),
    started_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at      TIMESTAMPTZ,
    duration      INTERVAL GENERATED ALWAYS AS (ended_at - started_at) STORED,
    exec_type     TEXT NOT NULL DEFAULT 'work'
                  CHECK (exec_type IN ('work','stoppage','rework','inspection')),
    notes         TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_exec_control    ON erp_production_control_executions (tenant_id, control_id);
CREATE INDEX idx_exec_operator   ON erp_production_control_executions (tenant_id, operator_id);
CREATE INDEX idx_exec_started_at ON erp_production_control_executions (started_at);
ALTER TABLE erp_production_control_executions ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_production_control_executions
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Rework log: defects found and corrections applied at a station
CREATE TABLE erp_production_rework (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     TEXT NOT NULL,
    control_id    UUID NOT NULL REFERENCES erp_production_controls(id),
    causal_id     UUID REFERENCES erp_production_control_causals(id),
    reported_by   UUID REFERENCES erp_entities(id),
    corrected_by  UUID REFERENCES erp_entities(id),
    defect_desc   TEXT NOT NULL,
    correction    TEXT NOT NULL DEFAULT '',
    severity      TEXT NOT NULL DEFAULT 'minor'
                  CHECK (severity IN ('minor','major','critical')),
    reported_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_rework_control  ON erp_production_rework (tenant_id, control_id);
CREATE INDEX idx_rework_severity ON erp_production_rework (tenant_id, severity);
ALTER TABLE erp_production_rework ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_production_rework
    USING (tenant_id = current_setting('app.tenant_id', true));
