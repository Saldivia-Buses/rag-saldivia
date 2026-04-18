-- 067_restore_journal_lines_double_entry.up.sql
-- Re-add the double-entry invariant dropped by migration 065.
--
-- Migration 065 loops over every CHECK on erp_* tables and drops any whose
-- `pg_get_constraintdef()` output matches substrings like '% > (0)::numeric%'.
-- The intent was to drop simple positivity guards that reject legacy rows
-- with zero / negative values (adjustments, voids, refunds).
--
-- Collateral damage: 019_erp_accounting.up.sql:74 defines
--     CHECK (NOT (debit > 0 AND credit > 0))
-- on erp_journal_lines. That compound predicate renders as
--     CHECK ((NOT ((debit > (0)::numeric) AND (credit > (0)::numeric))))
-- which contains '> (0)::numeric' twice — 065's LIKE matches it and drops
-- it along with the positivity check. The casualty is the double-entry
-- invariant: without it a single journal line may carry both a debit AND
-- a credit, silently corrupting the general ledger.
--
-- This migration restores only that one constraint. 019's simple positivity
-- check (debit >= 0 AND credit >= 0) stays dropped per 065's intent.
--
-- Verified on saldivia_bench (2026-04-18): zero rows currently violate the
-- invariant, so the constraint can be added VALID without a row scan.
-- A follow-up PR should tighten 065's matcher (allow-list by conname or
-- predicate-shape regex) to prevent a repeat on future additive CHECKs.

ALTER TABLE erp_journal_lines
    ADD CONSTRAINT erp_journal_lines_no_both_debit_credit
    CHECK (NOT (debit > 0 AND credit > 0));
