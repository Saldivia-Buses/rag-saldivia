-- 032_erp_cascade_void.up.sql
-- Plan 18 Fase 2: Anulación en cascada
-- Rewrite financial mutation trigger to allow controlled state transitions,
-- add 'reversal' entry_type, void columns, journal cancel guard.

-- Replace generic immutability function to allow controlled state transitions.
-- The original blocks ALL updates on posted/confirmed/paid/reversed.
-- Voiding and reversing are legitimate financial operations that change status
-- but don't modify the financial data itself.
CREATE OR REPLACE FUNCTION erp_prevent_financial_mutation() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'cannot delete financial record with status %', OLD.status;
    END IF;
    -- Allow voiding: posted/paid/confirmed → cancelled
    IF NEW.status = 'cancelled' AND OLD.status IN ('posted', 'paid', 'confirmed') THEN
        RETURN NEW;
    END IF;
    -- Allow reversing: posted/confirmed → reversed
    IF NEW.status = 'reversed' AND OLD.status IN ('posted', 'confirmed') THEN
        RETURN NEW;
    END IF;
    -- Block all other modifications
    RAISE EXCEPTION 'cannot modify financial record with status %', OLD.status;
END;
$$ LANGUAGE plpgsql;

-- Add 'reversal' to entry_type CHECK constraint.
-- Reversal entries are created by the cascade void workflow.
ALTER TABLE erp_journal_entries DROP CONSTRAINT erp_journal_entries_entry_type_check;
ALTER TABLE erp_journal_entries ADD CONSTRAINT erp_journal_entries_entry_type_check
    CHECK (entry_type IN ('manual', 'auto', 'adjustment', 'reversal'));

-- Protect journal entries from direct cancel — only reversal is allowed.
-- The generic function allows posted → cancelled for invoices/receipts,
-- but journal entries must go through the reversal workflow.
CREATE OR REPLACE FUNCTION erp_prevent_journal_cancel() RETURNS trigger AS $$
BEGIN
    IF NEW.status = 'cancelled' AND OLD.status IN ('posted', 'confirmed') THEN
        RAISE EXCEPTION 'journal entries cannot be cancelled directly — use reversal';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_journal_no_cancel BEFORE UPDATE ON erp_journal_entries
    FOR EACH ROW EXECUTE FUNCTION erp_prevent_journal_cancel();

-- Columns for void workflow on invoices
ALTER TABLE erp_invoices
    ADD COLUMN voided_by UUID REFERENCES erp_invoices(id),
    ADD COLUMN void_reason TEXT;

-- Column for reversal reference on journal entries
ALTER TABLE erp_journal_entries
    ADD COLUMN reversed_by UUID REFERENCES erp_journal_entries(id);

-- Permission
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.invoicing.void', 'Anular comprobantes', 'Anulacion en cascada de facturas posteadas', 'erp')
ON CONFLICT (id) DO NOTHING;

-- Grant to admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id = 'erp.invoicing.void'
ON CONFLICT DO NOTHING;
