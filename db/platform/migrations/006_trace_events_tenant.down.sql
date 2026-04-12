DROP INDEX IF EXISTS idx_trace_events_tenant;
ALTER TABLE trace_events DROP COLUMN IF EXISTS tenant_id;
