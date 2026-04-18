-- Reverse of 064. Each ALTER will fail if the column still has NULL rows;
-- that's by design — resolve them before rolling back.
DO $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN
        SELECT c.table_name, c.column_name
        FROM information_schema.columns c
        WHERE c.table_schema = 'public'
          AND c.table_name LIKE 'erp_%'
          AND c.is_nullable = 'YES'
          AND c.column_name ~ '_id$'
          AND c.column_name NOT IN ('id', 'tenant_id')
    LOOP
        BEGIN
            EXECUTE format('ALTER TABLE %I ALTER COLUMN %I SET NOT NULL',
                           rec.table_name, rec.column_name);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'could not re-enforce NOT NULL on %.%: %',
                         rec.table_name, rec.column_name, SQLERRM;
        END;
    END LOOP;
END $$;
