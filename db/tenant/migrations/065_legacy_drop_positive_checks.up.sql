-- Migration 065 — drop every remaining `> 0` / `>= 0` numeric CHECK on erp_* tables
--
-- Same theme as 063/064: legacy Histrix data routinely carries zero and
-- negative values (adjustments, voids, refunds, measurement errors) that
-- the strict forward-only SDA CHECKs reject. 061 relaxed the hot-path
-- tables; this migration sweeps the rest so we stop the one-at-a-time
-- iteration loop.
--
-- Approach: drop any constraint whose definition contains "> 0" or ">= 0"
-- on erp_* tables. The data integrity these were protecting can be re-
-- enforced after the historical import if the business requires it.

-- PG renders numeric checks as "x > (0)::numeric" — not "x > 0" — so the
-- original regex version of this migration matched nothing. Use a LIKE
-- pattern against the fully-rendered constraint definition.
DO $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN
        SELECT nsp.nspname AS schema_name,
               cls.relname AS table_name,
               con.conname AS con_name
        FROM pg_constraint con
        JOIN pg_class cls       ON cls.oid  = con.conrelid
        JOIN pg_namespace nsp   ON nsp.oid  = cls.relnamespace
        WHERE con.contype = 'c'
          AND nsp.nspname = 'public'
          AND cls.relname LIKE 'erp_%'
          AND (pg_get_constraintdef(con.oid) LIKE '%> (0)::numeric%'
            OR pg_get_constraintdef(con.oid) LIKE '%>= (0)::numeric%'
            OR pg_get_constraintdef(con.oid) LIKE '%> 0)%'
            OR pg_get_constraintdef(con.oid) LIKE '%>= 0)%')
    LOOP
        EXECUTE format('ALTER TABLE %I.%I DROP CONSTRAINT %I',
                       rec.schema_name, rec.table_name, rec.con_name);
    END LOOP;
END $$;
