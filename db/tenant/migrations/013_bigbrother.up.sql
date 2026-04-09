-- 013_bigbrother.up.sql
-- BigBrother: Network Intelligence Service (Plan 15)

-- Add tenant_id to audit_log for multi-tenant support (fix H10)
ALTER TABLE audit_log ADD COLUMN IF NOT EXISTS tenant_id UUID;

CREATE TABLE IF NOT EXISTS bb_devices (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    ip          INET NOT NULL,
    mac         MACADDR,
    hostname    TEXT,
    vendor      TEXT,
    device_type TEXT NOT NULL DEFAULT 'unknown'
        CHECK (device_type IN ('plc','workstation','server','switch','printer','camera','ap','phone','iot','unknown')),
    os          TEXT,
    model       TEXT,
    location    TEXT,
    status      TEXT NOT NULL DEFAULT 'online'
        CHECK (status IN ('online','offline','degraded','new')),
    first_seen  TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen   TIMESTAMPTZ NOT NULL DEFAULT now(),
    metadata    JSONB NOT NULL DEFAULT '{}'
);
-- Partial unique: only enforce MAC uniqueness when MAC is known
CREATE UNIQUE INDEX IF NOT EXISTS idx_bb_devices_tenant_mac ON bb_devices(tenant_id, mac) WHERE mac IS NOT NULL;
-- Fallback unique for MAC-less devices (different subnets where ARP cant resolve)
-- NOTE: IP unique can conflict with DHCP recycled IPs. Upsert strategy:
--   ON CONFLICT (tenant_id, mac) WHERE mac IS NOT NULL → update IP + last_seen
--   ON CONFLICT (tenant_id, ip) → update MAC if learned
CREATE UNIQUE INDEX IF NOT EXISTS idx_bb_devices_tenant_ip ON bb_devices(tenant_id, ip);
CREATE INDEX IF NOT EXISTS idx_bb_devices_status ON bb_devices(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_bb_devices_type ON bb_devices(tenant_id, device_type);

CREATE TABLE IF NOT EXISTS bb_ports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    device_id   UUID NOT NULL REFERENCES bb_devices(id) ON DELETE CASCADE,
    port        INT NOT NULL,
    protocol    TEXT NOT NULL DEFAULT 'tcp',
    service     TEXT,
    version     TEXT,
    state       TEXT NOT NULL DEFAULT 'open',
    UNIQUE (tenant_id, device_id, port, protocol)
);

CREATE TABLE IF NOT EXISTS bb_capabilities (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    device_id   UUID NOT NULL REFERENCES bb_devices(id) ON DELETE CASCADE,
    capability  TEXT NOT NULL,
    details     JSONB NOT NULL DEFAULT '{}',
    verified_at TIMESTAMPTZ,
    UNIQUE (tenant_id, device_id, capability)
);

CREATE TABLE IF NOT EXISTS bb_plc_registers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    device_id       UUID NOT NULL REFERENCES bb_devices(id) ON DELETE CASCADE,
    protocol        TEXT NOT NULL CHECK (protocol IN ('modbus','opcua')),
    address         TEXT NOT NULL,
    name            TEXT,
    data_type       TEXT,
    last_value      TEXT,
    last_value_numeric FLOAT,
    last_read       TIMESTAMPTZ,
    writable        BOOLEAN NOT NULL DEFAULT false,
    safety_tier     TEXT NOT NULL DEFAULT 'unclassified'
        CHECK (safety_tier IN ('unclassified','safe','controlled','critical')),
    min_value       FLOAT,
    max_value       FLOAT,
    max_writes_per_min INT DEFAULT 1,
    classified_by   UUID,
    classified_at   TIMESTAMPTZ,
    UNIQUE (tenant_id, device_id, protocol, address),
    CHECK (min_value IS NULL OR max_value IS NULL OR min_value <= max_value)
);

CREATE TABLE IF NOT EXISTS bb_computer_info (
    device_id     UUID NOT NULL REFERENCES bb_devices(id) ON DELETE CASCADE,
    tenant_id     UUID NOT NULL,
    PRIMARY KEY (tenant_id, device_id),
    os_version    TEXT,
    username      TEXT,
    cpu           TEXT,
    ram_gb        FLOAT,
    disk_total_gb FLOAT,
    disk_free_gb  FLOAT,
    software      JSONB NOT NULL DEFAULT '[]',
    last_scan     TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS bb_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    device_id   UUID REFERENCES bb_devices(id) ON DELETE SET NULL,
    -- ON DELETE SET NULL: preserves event history when device is removed
    event_type  TEXT NOT NULL
        CHECK (event_type IN ('discovered','went_offline','came_online','ip_changed',
               'plc_value_changed','exec_completed','scan_completed','credential_changed',
               'safety_tier_changed','pending_write_expired','pending_write_requested',
               'scan_mode_changed')),
    details     JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_bb_events_time ON bb_events(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bb_events_device ON bb_events(tenant_id, device_id, created_at DESC) WHERE device_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS bb_credentials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    device_id       UUID REFERENCES bb_devices(id) ON DELETE CASCADE,
    cred_type       TEXT NOT NULL
        CHECK (cred_type IN ('ssh_key','ssh_password','winrm','snmp_community')),
    encrypted_dek   BYTEA NOT NULL,
    encrypted_data  BYTEA NOT NULL,
    key_version     INT NOT NULL DEFAULT 1,
    key_fingerprint TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    rotated_at      TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_bb_credentials_tenant ON bb_credentials(tenant_id);
CREATE INDEX IF NOT EXISTS idx_bb_credentials_device ON bb_credentials(device_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bb_credentials_unique
    ON bb_credentials(tenant_id, device_id, cred_type);

-- Two-person rule for critical PLC writes
CREATE TABLE IF NOT EXISTS bb_pending_writes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    device_id       UUID NOT NULL REFERENCES bb_devices(id),
    register_addr   TEXT NOT NULL,
    value           TEXT NOT NULL,
    requestor_id    UUID NOT NULL,
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,
    status          TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending','approved','expired','rejected')),
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bb_pending_active
    ON bb_pending_writes(tenant_id, device_id, register_addr) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_bb_pending_expires
    ON bb_pending_writes(tenant_id, expires_at) WHERE status = 'pending';
