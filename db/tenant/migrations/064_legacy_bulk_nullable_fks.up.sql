-- Migration 064 — bulk-relax every remaining NOT NULL FK on erp_* tables
--
-- Each prior session of the legacy migration hit one-more NOT NULL FK that
-- the Histrix source didn't always resolve (risk_agent_id, communication_id,
-- competency_id, audit_id…). One relaxation per cycle = hours wasted. This
-- migration flips NOT NULL → NULL on every erp_* column ending in `_id`
-- that isn't `id` / `tenant_id`, so the migrator lands what it can and
-- operators reconcile later.
--
-- The down migration re-tightens — data with NULL FKs must be cleaned or
-- reassigned before rolling back.

DO $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN
        SELECT c.table_name, c.column_name
        FROM information_schema.columns c
        WHERE c.table_schema = 'public'
          AND c.table_name LIKE 'erp_%'
          AND c.is_nullable = 'NO'
          AND c.column_name ~ '_id$'
          AND c.column_name NOT IN ('id', 'tenant_id')
    LOOP
        EXECUTE format('ALTER TABLE %I ALTER COLUMN %I DROP NOT NULL',
                       rec.table_name, rec.column_name);
    END LOOP;
END $$;
