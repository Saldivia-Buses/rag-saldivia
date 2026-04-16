-- 038_erp_hr_extended.up.sql
-- RRHH profundo: evaluaciones, competencias, saldos de licencia

-- Competency catalog
CREATE TABLE erp_competencies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category    TEXT NOT NULL DEFAULT 'technical'
                CHECK (category IN ('technical','soft','leadership','safety','quality')),
    active      BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(tenant_id, name)
);
ALTER TABLE erp_competencies ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_competencies
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Employee → competency mapping with proficiency level
CREATE TABLE erp_employee_competencies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    entity_id       UUID NOT NULL REFERENCES erp_entities(id),
    competency_id   UUID NOT NULL REFERENCES erp_competencies(id),
    level           INT NOT NULL DEFAULT 1 CHECK (level BETWEEN 1 AND 5),
    certified       BOOLEAN NOT NULL DEFAULT false,
    certified_at    DATE,
    notes           TEXT NOT NULL DEFAULT '',
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, entity_id, competency_id)
);
ALTER TABLE erp_employee_competencies ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_employee_competencies
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Performance evaluations
CREATE TABLE erp_evaluations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    entity_id       UUID NOT NULL REFERENCES erp_entities(id),
    evaluator_id    TEXT NOT NULL,
    period          TEXT NOT NULL,
    eval_type       TEXT NOT NULL DEFAULT 'annual'
                    CHECK (eval_type IN ('annual','probation','project','360')),
    overall_score   NUMERIC(3,1) CHECK (overall_score BETWEEN 1 AND 5),
    strengths       TEXT NOT NULL DEFAULT '',
    weaknesses      TEXT NOT NULL DEFAULT '',
    goals           TEXT NOT NULL DEFAULT '',
    comments        TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft','submitted','reviewed','acknowledged')),
    submitted_at    TIMESTAMPTZ,
    reviewed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_evaluations_entity ON erp_evaluations(tenant_id, entity_id, period);
ALTER TABLE erp_evaluations ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_evaluations
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Evaluation scores per competency
CREATE TABLE erp_evaluation_scores (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    evaluation_id   UUID NOT NULL REFERENCES erp_evaluations(id) ON DELETE CASCADE,
    competency_id   UUID NOT NULL REFERENCES erp_competencies(id),
    score           NUMERIC(3,1) NOT NULL CHECK (score BETWEEN 1 AND 5),
    comments        TEXT NOT NULL DEFAULT ''
);
ALTER TABLE erp_evaluation_scores ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_evaluation_scores
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Leave balances (accrual tracking)
CREATE TABLE erp_leave_balances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    entity_id       UUID NOT NULL REFERENCES erp_entities(id),
    leave_type      TEXT NOT NULL CHECK (leave_type IN ('vacation','sick','personal','maternity','study')),
    year            INT NOT NULL,
    accrued         NUMERIC(6,1) NOT NULL DEFAULT 0,
    used            NUMERIC(6,1) NOT NULL DEFAULT 0,
    balance         NUMERIC(6,1) GENERATED ALWAYS AS (accrued - used) STORED,
    UNIQUE(tenant_id, entity_id, leave_type, year)
);
ALTER TABLE erp_leave_balances ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_leave_balances
    USING (tenant_id = current_setting('app.tenant_id', true));
