-- 011_saldivia_erp_modules.down.sql
-- Disables and unregisters the Saldivia ERP vertical modules.
-- 'feedback' is left intact (it predates this migration in 001_init).

DELETE FROM tenant_modules WHERE module_id IN (
    'manufactura', 'produccion', 'calidad', 'ingenieria', 'mantenimiento',
    'compras', 'administracion', 'rrhh', 'seguridad', 'astro'
);

DELETE FROM modules WHERE id IN (
    'manufactura', 'produccion', 'calidad', 'ingenieria', 'mantenimiento',
    'compras', 'administracion', 'rrhh', 'seguridad', 'astro'
);
