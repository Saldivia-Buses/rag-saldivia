-- 059_erp_inspection_templates.up.sql
-- Quality control inspection templates (catalog of what to inspect).
-- Source: Histrix PROD_CONTROLES ("tabla maestra de controles de calidad en produccion").
-- Separate from erp_production_inspections (events) because templates have no order_id.

CREATE TABLE erp_inspection_templates (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                TEXT NOT NULL,
    section_id               UUID REFERENCES erp_departments(id),
    step_id                  UUID REFERENCES erp_production_steps(id),
    vehicle_section_id       UUID REFERENCES erp_catalogs(id),
    control_name             TEXT NOT NULL,
    model_code               TEXT NOT NULL DEFAULT '',
    control_type             INTEGER NOT NULL DEFAULT 0,
    sort_order               INTEGER NOT NULL DEFAULT 0,
    active                   BOOLEAN NOT NULL DEFAULT true,
    critical                 BOOLEAN NOT NULL DEFAULT false,
    actionable               BOOLEAN NOT NULL DEFAULT true,
    show_in_tech_sheet       BOOLEAN NOT NULL DEFAULT false,
    default_inspector_id     UUID REFERENCES erp_entities(id),
    enabled_user_id          TEXT REFERENCES users(id),
    observations             TEXT NOT NULL DEFAULT '',
    metadata                 JSONB NOT NULL DEFAULT '{}',
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_inspection_templates_tenant ON erp_inspection_templates (tenant_id);
CREATE INDEX idx_inspection_templates_section ON erp_inspection_templates (tenant_id, section_id);
CREATE INDEX idx_inspection_templates_active ON erp_inspection_templates (tenant_id, active);
CREATE INDEX idx_inspection_templates_model ON erp_inspection_templates (tenant_id, model_code);

ALTER TABLE erp_inspection_templates ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_inspection_templates
    USING (tenant_id = current_setting('app.tenant_id', true));

COMMENT ON TABLE erp_inspection_templates IS
    'Master catalog of quality control inspection definitions. Referenced by erp_production_inspections events.';
