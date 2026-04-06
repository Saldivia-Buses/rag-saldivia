CREATE TABLE IF NOT EXISTS contacts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    user_id     UUID NOT NULL,
    name        TEXT NOT NULL,
    birth_date  DATE NOT NULL,
    birth_time  TIME,
    birth_time_known BOOLEAN NOT NULL DEFAULT true,
    city        TEXT NOT NULL,
    nation      TEXT NOT NULL DEFAULT 'Argentina',
    lat         DOUBLE PRECISION NOT NULL,
    lon         DOUBLE PRECISION NOT NULL,
    alt         DOUBLE PRECISION NOT NULL DEFAULT 25.0,
    utc_offset  INTEGER NOT NULL DEFAULT -3,
    relationship TEXT,
    notes       TEXT,
    kind        TEXT NOT NULL DEFAULT 'persona',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_contacts_tenant_name ON contacts(tenant_id, lower(name));
CREATE INDEX idx_contacts_tenant ON contacts(tenant_id);
CREATE INDEX idx_contacts_user ON contacts(tenant_id, user_id);
