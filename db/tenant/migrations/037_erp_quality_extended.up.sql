-- 037_erp_quality_extended.up.sql
-- Calidad extendida: supplier scorecards, quality KPIs, risk register

-- Supplier quality scorecards (aggregated from inspections + demerits)
CREATE TABLE erp_supplier_scorecards (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    supplier_id     UUID NOT NULL REFERENCES erp_entities(id),
    period          TEXT NOT NULL,
    total_receipts  INT NOT NULL DEFAULT 0,
    accepted_qty    NUMERIC(14,4) NOT NULL DEFAULT 0,
    rejected_qty    NUMERIC(14,4) NOT NULL DEFAULT 0,
    total_demerits  INT NOT NULL DEFAULT 0,
    quality_score   NUMERIC(5,2),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, supplier_id, period)
);
ALTER TABLE erp_supplier_scorecards ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_supplier_scorecards
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Risk register
CREATE TABLE erp_quality_risks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    category        TEXT NOT NULL DEFAULT 'process'
                    CHECK (category IN ('process','product','supplier','environmental','safety')),
    probability     TEXT NOT NULL DEFAULT 'medium'
                    CHECK (probability IN ('low','medium','high','critical')),
    impact          TEXT NOT NULL DEFAULT 'medium'
                    CHECK (impact IN ('low','medium','high','critical')),
    status          TEXT NOT NULL DEFAULT 'identified'
                    CHECK (status IN ('identified','mitigated','accepted','closed')),
    mitigation      TEXT NOT NULL DEFAULT '',
    responsible_id  UUID REFERENCES erp_entities(id),
    review_date     DATE,
    user_id         TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE erp_quality_risks ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_quality_risks
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Quality indicators / KPIs tracking (periodic snapshots)
CREATE TABLE erp_quality_indicators (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    period          TEXT NOT NULL,
    indicator_type  TEXT NOT NULL
                    CHECK (indicator_type IN ('nc_rate','resolution_rate','effectiveness_rate','audit_score','defect_ppm','supplier_quality')),
    value           NUMERIC(10,4) NOT NULL,
    target          NUMERIC(10,4),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, period, indicator_type)
);
ALTER TABLE erp_quality_indicators ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_quality_indicators
    USING (tenant_id = current_setting('app.tenant_id', true));
