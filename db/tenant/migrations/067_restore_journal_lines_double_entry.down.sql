-- 067_restore_journal_lines_double_entry.down.sql
-- Drops the invariant re-added by 067.up. Leaves the table in the post-065
-- (invariant-less) state.

ALTER TABLE erp_journal_lines
    DROP CONSTRAINT IF EXISTS erp_journal_lines_no_both_debit_credit;
