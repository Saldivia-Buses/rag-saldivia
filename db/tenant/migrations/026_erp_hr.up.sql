-- 026_erp_hr.up.sql
-- Plan 17 Phase 9: RRHH
-- Replaces ~62 legacy tables: RH_*, RRHH_*, FALTAS, FRANCOS, SINDICATO, etc.

CREATE TABLE IF NOT EXISTS erp_departments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    code        TEXT NOT NULL,
    name        TEXT NOT NULL,
    parent_id   UUID REFERENCES erp_departments(id),
    manager_id  UUID REFERENCES erp_entities(id),
    active      BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(tenant_id, code)
);
CREATE INDEX idx_erp_departments ON erp_departments(tenant_id, active);

CREATE TABLE IF NOT EXISTS erp_employee_details (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        TEXT NOT NULL,
    entity_id        UUID NOT NULL REFERENCES erp_entities(id) UNIQUE,
    department_id    UUID REFERENCES erp_departments(id),
    position         TEXT NOT NULL DEFAULT '',
    hire_date        DATE,
    termination_date DATE,
    union_id         UUID REFERENCES erp_catalogs(id),
    health_plan_id   UUID REFERENCES erp_catalogs(id),
    schedule_type    TEXT NOT NULL DEFAULT 'full_time' CHECK (schedule_type IN ('full_time','part_time','shifts')),
    category_id      UUID REFERENCES erp_catalogs(id),
    encrypted_salary BYTEA,
    metadata         JSONB NOT NULL DEFAULT '{}',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_hr_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    entity_id   UUID NOT NULL REFERENCES erp_entities(id),
    event_type  TEXT NOT NULL CHECK (event_type IN ('absence','leave','accident','transfer','promotion','sanction','overtime','vacation')),
    date_from   DATE NOT NULL,
    date_to     DATE,
    hours       NUMERIC(6,2),
    reason_id   UUID REFERENCES erp_catalogs(id),
    notes       TEXT NOT NULL DEFAULT '',
    user_id     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_hr_events ON erp_hr_events(tenant_id, entity_id, date_from DESC);

CREATE TABLE IF NOT EXISTS erp_training (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    instructor  TEXT NOT NULL DEFAULT '',
    date_from   DATE,
    date_to     DATE,
    status      TEXT NOT NULL DEFAULT 'planned' CHECK (status IN ('planned','in_progress','completed'))
);
CREATE INDEX idx_erp_training ON erp_training(tenant_id);

CREATE TABLE IF NOT EXISTS erp_training_attendees (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    training_id UUID NOT NULL REFERENCES erp_training(id) ON DELETE CASCADE,
    entity_id   UUID NOT NULL REFERENCES erp_entities(id),
    result      TEXT NOT NULL DEFAULT '',
    score       NUMERIC(5,2)
);

CREATE TABLE IF NOT EXISTS erp_attendance (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    entity_id   UUID NOT NULL REFERENCES erp_entities(id),
    date        DATE NOT NULL,
    clock_in    TIMESTAMPTZ,
    clock_out   TIMESTAMPTZ,
    hours       NUMERIC(6,2),
    source      TEXT NOT NULL DEFAULT 'manual' CHECK (source IN ('manual','rfid','biometric')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_erp_attendance ON erp_attendance(tenant_id, entity_id, date DESC);

-- RLS
ALTER TABLE erp_departments ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_departments USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_employee_details ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_employee_details USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_hr_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_hr_events USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_training ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_training USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_training_attendees ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_training_attendees USING (tenant_id = current_setting('app.tenant_id', true));
ALTER TABLE erp_attendance ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_attendance USING (tenant_id = current_setting('app.tenant_id', true));

INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.hr.read',  'Ver RRHH',      'Consultar empleados, novedades, asistencia', 'erp'),
    ('erp.hr.write', 'Gestionar RRHH', 'Registrar novedades, fichadas, cursos',     'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.hr.%'
ON CONFLICT DO NOTHING;
