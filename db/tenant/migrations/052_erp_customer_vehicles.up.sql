-- 052_erp_customer_vehicles.up.sql
-- Workshop module: customer vehicle registry and incident tracking
-- Maps to Histrix: MANT_VEHICULOS (12 vehicles), MANT_TIPO_NOVEDAD, MANT_NOVEDADES (26,082 events)

-- erp_customer_vehicles (MANT_VEHICULOS)
CREATE TABLE erp_customer_vehicles (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id             TEXT NOT NULL,
  owner_id              UUID REFERENCES erp_entities(id),
  driver_id             UUID REFERENCES erp_entities(id),
  manufacturing_unit_id UUID REFERENCES erp_manufacturing_units(id),
  plate                 TEXT NOT NULL DEFAULT '',
  chassis_serial        TEXT NOT NULL DEFAULT '',
  body_serial           TEXT NOT NULL DEFAULT '',
  internal_number       INTEGER,
  brand                 TEXT NOT NULL DEFAULT '',
  model_year            INTEGER,
  seating_capacity      INTEGER NOT NULL DEFAULT 0,
  fuel_type             TEXT NOT NULL DEFAULT 'diesel'
                        CHECK (fuel_type IN ('diesel','gasolina','gnc','electric','hybrid')),
  color                 TEXT NOT NULL DEFAULT '',
  purchase_date         DATE,
  purchase_price        NUMERIC(14,2),
  warranty_months       INTEGER NOT NULL DEFAULT 0,
  destination           TEXT NOT NULL DEFAULT '',
  observations          TEXT NOT NULL DEFAULT '',
  active                BOOLEAN NOT NULL DEFAULT true,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_vehicles_owner ON erp_customer_vehicles (tenant_id, owner_id);
CREATE INDEX idx_vehicles_plate ON erp_customer_vehicles (tenant_id, plate);
ALTER TABLE erp_customer_vehicles ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_customer_vehicles
    USING (tenant_id = current_setting('app.tenant_id', true));

-- erp_vehicle_incident_types (MANT_TIPO_NOVEDAD)
CREATE TABLE erp_vehicle_incident_types (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  TEXT NOT NULL,
  name       TEXT NOT NULL,
  active     BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, name)
);
ALTER TABLE erp_vehicle_incident_types ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_vehicle_incident_types
    USING (tenant_id = current_setting('app.tenant_id', true));

-- erp_vehicle_incidents (MANT_NOVEDADES)
CREATE TABLE erp_vehicle_incidents (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id        TEXT NOT NULL,
  vehicle_id       UUID NOT NULL REFERENCES erp_customer_vehicles(id),
  incident_type_id UUID REFERENCES erp_vehicle_incident_types(id),
  incident_date    DATE NOT NULL,
  location         TEXT NOT NULL DEFAULT '',
  responsible      TEXT NOT NULL DEFAULT '',
  notes            TEXT NOT NULL DEFAULT '',
  status           TEXT NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending','in_progress','resolved')),
  resolved_at      TIMESTAMPTZ,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_incidents_vehicle ON erp_vehicle_incidents (tenant_id, vehicle_id);
CREATE INDEX idx_incidents_date    ON erp_vehicle_incidents (incident_date);
CREATE INDEX idx_incidents_status  ON erp_vehicle_incidents (tenant_id, status);
ALTER TABLE erp_vehicle_incidents ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON erp_vehicle_incidents
    USING (tenant_id = current_setting('app.tenant_id', true));

-- Permissions
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.maintenance.read',  'Ver taller',         'Consultar vehículos e incidentes',   'erp'),
    ('erp.maintenance.write', 'Gestionar taller',   'Registrar vehículos e incidentes',   'erp')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id LIKE 'erp.maintenance.%'
ON CONFLICT DO NOTHING;
