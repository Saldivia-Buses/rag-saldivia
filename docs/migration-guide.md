# Guia de Migracion: Histrix MySQL → SDA PostgreSQL

> **Plan 21** — Data Migration  
> Este documento describe como ejecutar la migracion de datos historicos.

## Pre-requisitos

1. MySQL legacy accesible (read-only)
2. SDA PostgreSQL con migraciones 001-035 aplicadas
3. CLI `sda` compilado (`go build ./tools/cli/...`)
4. Tenant creado en SDA (slug debe coincidir con `--tenant`)

## Ejecucion paso a paso

### 1. Dry-run (obligatorio)

```bash
sda migrate-legacy \
  --dry-run \
  --tenant=saldivia \
  --mysql-dsn="user:pass@tcp(host:3306)/histrix?charset=utf8mb4&parseTime=true" \
  --pg-dsn="postgres://sda:pass@host:5432/sda_tenant_saldivia?sslmode=disable"
```

El dry-run ejecuta:
- Pre-validacion contra constraints de Plan 17+18
- Transformaciones de prueba (sin escribir en PostgreSQL)
- Reporte de issues bloqueantes

### 2. Resolver issues bloqueantes

Si el dry-run reporta issues `fix_manual`:
1. Corregir datos en MySQL legacy
2. Re-correr dry-run hasta que diga "all clear"

### 3. Migracion en prod

```bash
sda migrate-legacy \
  --tenant=saldivia \
  --skip-dry-run-for=saldivia \
  --mysql-dsn="user:pass@tcp(host:3306)/histrix?charset=utf8mb4&parseTime=true" \
  --pg-dsn="postgres://sda:pass@host:5432/sda_tenant_saldivia?sslmode=disable"
```

### 4. Validacion post-migracion

```bash
sda migrate-legacy \
  --validate-only \
  --tenant=saldivia \
  --mysql-dsn="..." \
  --pg-dsn="..."
```

### 5. Resume (si falla)

```bash
sda migrate-legacy \
  --resume \
  --resume-run-id=<uuid-del-run-fallido> \
  --tenant=saldivia \
  --mysql-dsn="..." \
  --pg-dsn="..."
```

## Migracion por dominios

Se puede migrar un subconjunto de dominios:

```bash
sda migrate-legacy --dry-run --tenant=saldivia --domains=catalog,entity
```

Dominios disponibles: `catalog`, `entity`, `accounting`, `treasury`, `invoicing`, `stock`, `purchasing`, `sales`, `production`, `hr`

## Checklist post-migracion

**Obligatorio antes de apagar Histrix:**

- [ ] Setear `result_account_id` en `erp_fiscal_years` con `status='open'` via `PATCH /fiscal-years/{id}/result-account`
- [ ] Run manual de anulacion de factura de prueba (draft) para verificar cascade void
- [ ] Verificar counts por dominio (output del `--validate-only`)
- [ ] Verificar sumas de control financieras (balance general)
- [ ] Verificar reconciliacion con datos reales en al menos 1 cuenta bancaria

**Diferido hasta Plan 19 (AFIP):**

- [ ] Subir certificados AFIP via `PUT /v1/erp/afip/config`
- [ ] Correr `POST /v1/erp/afip/test-auth` por cada tenant
- [ ] Verificar last_invoice_a/b/c vs AFIP

## Troubleshooting

### "tenant requires --dry-run first"
Correr `--dry-run` primero. El sistema no permite migracion prod sin un dry-run exitoso previo.

### "resolve ... not found (migrate dependency first)"
Un registro referencia a un dominio que aun no fue migrado. Verificar el orden de migracion.

### "unknown invoice status / entry type"
Un valor legacy no tiene mapping en `transformer.go`. Agregar el mapping y re-correr.

### Falla a mitad de una tabla
El progreso se guarda per-batch. Usar `--resume --resume-run-id=<uuid>` para continuar desde donde fallo.

## Rollback

One-way. Si algo sale catastroficamente mal:

1. Borrar datos migrados del tenant en PostgreSQL
2. Borrar registros de `erp_legacy_mapping` para el tenant
3. Re-correr la migracion completa
