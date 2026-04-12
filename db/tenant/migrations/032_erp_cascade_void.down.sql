-- 032_erp_cascade_void.down.sql
ALTER TABLE erp_invoices
    DROP COLUMN IF EXISTS voided_by,
    DROP COLUMN IF EXISTS void_reason;

ALTER TABLE erp_journal_entries
    DROP COLUMN IF EXISTS reversed_by;

DROP TRIGGER IF EXISTS trg_journal_no_cancel ON erp_journal_entries;
DROP FUNCTION IF EXISTS erp_prevent_journal_cancel();

ALTER TABLE erp_journal_entries DROP CONSTRAINT erp_journal_entries_entry_type_check;
ALTER TABLE erp_journal_entries ADD CONSTRAINT erp_journal_entries_entry_type_check
    CHECK (entry_type IN ('manual', 'auto', 'adjustment'));

-- Restore original strict immutability function
CREATE OR REPLACE FUNCTION erp_prevent_financial_mutation() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        IF OLD.status IN ('posted', 'confirmed', 'paid', 'reversed') THEN
            RAISE EXCEPTION 'cannot delete financial record with status %', OLD.status;
        END IF;
        RETURN OLD;
    END IF;
    IF OLD.status IN ('posted', 'confirmed', 'paid', 'reversed') THEN
        RAISE EXCEPTION 'cannot modify financial record with status %', OLD.status;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DELETE FROM role_permissions WHERE permission_id = 'erp.invoicing.void';
DELETE FROM permissions WHERE id = 'erp.invoicing.void';
