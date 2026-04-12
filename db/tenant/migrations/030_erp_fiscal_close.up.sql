-- 030_erp_fiscal_close.up.sql
-- Plan 18 Fase 0: Cierre de ejercicio contable
-- Adds columns for fiscal year close workflow + result account configuration

-- Columns for close workflow
ALTER TABLE erp_fiscal_years
    ADD COLUMN closed_by TEXT,
    ADD COLUMN closed_at TIMESTAMPTZ,
    ADD COLUMN closing_entry_id UUID REFERENCES erp_journal_entries(id),
    ADD COLUMN opening_entry_id UUID REFERENCES erp_journal_entries(id);

-- Result account — the equity account where net income goes on close.
-- Nullable: existing fiscal years created by plan 17 don't have this.
-- Must be set via PATCH /fiscal-years/{id}/result-account before closing.
-- CHECK ensures closed fiscal years always have a result account.
ALTER TABLE erp_fiscal_years
    ADD COLUMN result_account_id UUID REFERENCES erp_accounts(id);

ALTER TABLE erp_fiscal_years
    ADD CONSTRAINT erp_fiscal_years_result_account_required
    CHECK (status = 'open' OR result_account_id IS NOT NULL);

-- Permission
INSERT INTO permissions (id, name, description, category) VALUES
    ('erp.accounting.close', 'Cerrar ejercicio', 'Cierre de ejercicio contable', 'erp')
ON CONFLICT (id) DO NOTHING;

-- Grant to admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions WHERE id = 'erp.accounting.close'
ON CONFLICT DO NOTHING;
