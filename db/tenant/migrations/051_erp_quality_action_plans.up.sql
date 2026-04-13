-- 051_erp_quality_action_plans.up.sql
-- Plan 25: Quality Action Plans (CAL_PLAN_ACCION_TOTAL / CAL_PLAN_ACCION_ACCION)

-- erp_nc_origins (CAL_ORIGEN_NCONF — NC origin classification)
CREATE TABLE erp_nc_origins (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  TEXT NOT NULL,
  name       TEXT NOT NULL,
  active     BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, name)
);
ALTER TABLE erp_nc_origins ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_nc_origins
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Add cost_impact to erp_nonconformities (origin_id already exists from migration 027)
ALTER TABLE erp_nonconformities
  ADD COLUMN IF NOT EXISTS cost_impact NUMERIC(14,2) NOT NULL DEFAULT 0;

-- erp_quality_action_plans (CAL_PLAN_ACCION_TOTAL)
CREATE TABLE erp_quality_action_plans (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id          TEXT NOT NULL,
  nonconformity_id   UUID REFERENCES erp_nonconformities(id),
  responsible_id     UUID REFERENCES erp_entities(id),
  section_id         UUID REFERENCES erp_departments(id),
  description        TEXT NOT NULL,
  planned_start      DATE,
  target_date        DATE,
  closed_date        DATE,
  time_savings_hours NUMERIC(8,2) NOT NULL DEFAULT 0,
  cost_savings       NUMERIC(14,2) NOT NULL DEFAULT 0,
  status             TEXT NOT NULL DEFAULT 'draft'
                     CHECK (status IN ('draft','active','closed','cancelled')),
  created_by         TEXT NOT NULL DEFAULT '',
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_action_plans_nc     ON erp_quality_action_plans (tenant_id, nonconformity_id);
CREATE INDEX idx_action_plans_status ON erp_quality_action_plans (tenant_id, status);
CREATE INDEX idx_action_plans_target ON erp_quality_action_plans (target_date);
ALTER TABLE erp_quality_action_plans ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_quality_action_plans
    USING (tenant_id = current_setting('app.tenant_id', true));

-- erp_quality_action_tasks (CAL_PLAN_ACCION_ACCION)
CREATE TABLE erp_quality_action_tasks (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id     TEXT NOT NULL,
  plan_id       UUID NOT NULL REFERENCES erp_quality_action_plans(id) ON DELETE CASCADE,
  description   TEXT NOT NULL,
  leader_id     UUID REFERENCES erp_entities(id),
  planned_start DATE,
  target_date   DATE,
  closed_date   DATE,
  completed     BOOLEAN NOT NULL DEFAULT false,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_action_tasks_plan ON erp_quality_action_tasks (tenant_id, plan_id);
ALTER TABLE erp_quality_action_tasks ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_quality_action_tasks
    USING (tenant_id = current_setting('app.tenant_id', true));
