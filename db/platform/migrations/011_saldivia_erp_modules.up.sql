-- 011_saldivia_erp_modules.up.sql
-- Registers the Saldivia ERP vertical modules (manufactura, produccion, etc.)
-- in the global module registry, and enables them for the Saldivia tenant.
--
-- The 10 modules below mirror Saldivia Buses' organigram (Rev. 28, 2025) and
-- are referenced by `apps/web/src/lib/modules/registry.ts` MODULE_REGISTRY.
-- Without this seed the sidebar shows "Sin modulos habilitados" because the
-- backend's GET /v1/modules/enabled returns nothing for the tenant.
--
-- Other tenants (if any) can enable individual modules via the platform admin
-- API or by inserting into tenant_modules manually.

INSERT INTO modules (id, name, category, tier_min) VALUES
    ('manufactura',    'Manufactura',           'vertical', 'starter'),
    ('produccion',     'Produccion',            'vertical', 'starter'),
    ('calidad',        'Calidad',               'vertical', 'starter'),
    ('ingenieria',     'Ingenieria',            'vertical', 'starter'),
    ('mantenimiento',  'Mantenimiento',         'vertical', 'starter'),
    ('compras',        'Compras',               'vertical', 'starter'),
    ('administracion', 'Administracion',        'vertical', 'starter'),
    ('rrhh',           'Recursos Humanos',      'vertical', 'starter'),
    ('seguridad',      'Higiene y Seguridad',   'vertical', 'starter')
ON CONFLICT (id) DO NOTHING;

-- Enable every Saldivia ERP module + the existing 'feedback' AI module
-- for any tenant whose slug is 'saldivia' or 'dev' (the workstation default).
-- Idempotent: ON CONFLICT skips rows that already exist.
INSERT INTO tenant_modules (tenant_id, module_id, enabled, enabled_by)
SELECT t.id, m.id, true, 'migration:011_saldivia_erp_modules'
FROM tenants t
CROSS JOIN (VALUES
    ('manufactura'),
    ('produccion'),
    ('calidad'),
    ('ingenieria'),
    ('mantenimiento'),
    ('compras'),
    ('administracion'),
    ('rrhh'),
    ('seguridad'),
    ('feedback')
) AS m(id)
WHERE t.slug IN ('saldivia', 'dev')
ON CONFLICT (tenant_id, module_id) DO NOTHING;
