-- 006_trace_events_tenant.up.sql
-- Add tenant_id to trace_events for tenant isolation (was missing from original schema)
ALTER TABLE trace_events ADD COLUMN IF NOT EXISTS tenant_id TEXT;

-- Backfill from parent trace
UPDATE trace_events SET tenant_id = et.tenant_id
FROM execution_traces et WHERE trace_events.trace_id = et.id AND trace_events.tenant_id IS NULL;

-- Make NOT NULL after backfill
ALTER TABLE trace_events ALTER COLUMN tenant_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_trace_events_tenant ON trace_events(tenant_id, trace_id);
