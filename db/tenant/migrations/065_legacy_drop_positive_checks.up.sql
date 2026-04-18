-- Migration 065 — drop simple single-column positivity CHECKs on erp_* tables
--
-- Same theme as 063/064: legacy Histrix data routinely carries zero and
-- negative values (adjustments, voids, refunds, measurement errors) that
-- the strict forward-only SDA CHECKs reject. 061 relaxed the hot-path
-- tables; this migration sweeps the rest so we stop the one-at-a-time
-- iteration loop.
--
-- Rewrite history (2026-04-18, ADR 027 F-01):
--
-- The original body used LIKE '% > (0)::numeric%' / '% >= 0)%' fragments
-- on pg_get_constraintdef() to find targets. That matched compound
-- predicates too — in particular it dropped erp_journal_lines's
-- `CHECK (NOT (debit > 0 AND credit > 0))` (the double-entry invariant,
-- added by 019_erp_accounting), because the rendered form
-- `CHECK ((NOT ((debit > (0)::numeric) AND (credit > (0)::numeric))))`
-- contains the matched substring. Migration 067 restored that one
-- constraint, but the LIKE-sweep remained armed against any future
-- additive compound CHECK that happens to mention "> 0".
--
-- This version tightens the matcher to a shape-anchored regex: only
-- constraints whose rendered definition is EXACTLY
--     CHECK ((<identifier> > 0))                     (int column)
--     CHECK ((<identifier> >= 0))                    (int column)
--     CHECK ((<identifier> > (0)::numeric))          (numeric column)
--     CHECK ((<identifier> >= (0)::numeric))         (numeric column)
-- are dropped, AND only when the constraint name ends in `_check` (the
-- PG auto-naming convention for anonymous column-level CHECKs). Named
-- constraints like `erp_journal_lines_no_both_debit_credit` and any
-- compound predicate (containing AND / OR / NOT / multiple identifiers)
-- are excluded by construction.
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
          AND con.conname ~ '_check$'
          AND pg_get_constraintdef(con.oid) ~
              '^CHECK \(\([a-z_][a-z0-9_]* >=? (\(0\)::numeric|0)\)\)$'
    LOOP
        EXECUTE format('ALTER TABLE %I.%I DROP CONSTRAINT IF EXISTS %I',
                       rec.schema_name, rec.table_name, rec.con_name);
    END LOOP;
END $$;
