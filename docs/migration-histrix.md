# Histrix → SDA Migration Guide

Migrates historical data from the Histrix legacy MySQL ERP to SDA (PostgreSQL).
Implemented in `tools/cli/cmd/migrate_legacy.go` and `tools/cli/internal/migration/`.

---

## How it works

- MySQL is treated as **read-only** — the connection sets `SET SESSION TRANSACTION READ ONLY` on open (`tools/cli/internal/legacy/connection.go`)
- All legacy integer IDs are mapped to UUIDs via a mapping table in PostgreSQL
- Migration is **idempotent**: every INSERT uses `ON CONFLICT DO NOTHING` plus a mapping table — re-running is safe
- Progress is checkpointed per-batch (500 rows) in `erp_migration_table_progress` — failures can be resumed
- A dry-run that reaches `dry_run_ok` is required before any production run (enforced by the orchestrator)
- Migration runs are recorded in `erp_migration_runs` with status, per-table stats, and error messages

---

## Prerequisites

1. SDA deployed and all migrations applied (`bash deploy/scripts/migrate.sh`)
2. MySQL read access to the Histrix database
3. PostgreSQL write access to the target tenant database
4. `sda` CLI binary built: `make build-cli` (or `go build ./tools/cli/...`)
5. A read-only MySQL user on Histrix (the CLI enforces read-only at the session level, but a DB-level read-only user is safer)

### MySQL DSN format

```
user:password@tcp(host:3306)/database?charset=utf8mb4&parseTime=true
```

`charset=utf8mb4` is required — the driver handles latin1 → utf8mb4 conversion automatically.
`parseTime=true` enables automatic MySQL date → Go `time.Time` parsing.

The DSN can be provided via `--mysql-dsn` flag or `MYSQL_LEGACY_DSN` environment variable.
The PostgreSQL DSN can be provided via `--pg-dsn` flag or `POSTGRES_TENANT_URL` environment variable.

---

## Migration Order

The orchestrator registers migrators in FK dependency order (`cmd/migrate_legacy.go:registerMigrators`). This order is enforced automatically when migrating all domains, but if using `--domains` to migrate a subset, observe the dependencies below.

### Phase 2: Catalogs and Entities (no FK dependencies)

| Legacy table | SDA table | Domain |
|---|---|---|
| `GEN_PROVINCIAS` | `erp_catalogs` (type=province) | catalog |
| `GEN_LOCALIDADES` | `erp_catalogs` (type=city) | catalog |
| `GEN_MONEDAS` | `erp_catalogs` (type=currency) | catalog |
| `FORMAPAGO` | `erp_catalogs` (type=payment_method) | catalog |
| `GEN_IVA` | `erp_catalogs` (type=iva_condition) | catalog |
| `GEN_IVA_ALICUOTAS` | `erp_catalogs` (type=iva_rate) | catalog |
| `GEN_UNIDADES` | `erp_catalogs` (type=unit) | catalog |
| `GEN_COMPROBANTES` | `erp_catalogs` (type=voucher_type) | catalog |
| `GEN_TIPOS_DOCUMENTOS` | `erp_catalogs` (type=document_type) | catalog |
| `GEN_ESTADOCIVIL` | `erp_catalogs` (type=civil_status) | catalog |
| `GEN_NACIONALIDADES` | `erp_catalogs` (type=nationality) | catalog |
| `GEN_OBRASSOCIALES` | `erp_catalogs` (type=health_plan) | catalog |
| `GEN_PARENTESCOS` | `erp_catalogs` (type=kinship) | catalog |
| `GEN_PROFESIONES` | `erp_catalogs` (type=profession) | catalog |
| `GEN_CALLES` | `erp_catalogs` (type=street) | catalog |
| `GEN_TIPO_CONTACTOS` | `erp_catalogs` (type=contact_type) | catalog |
| `GEN_IIBB` | `erp_catalogs` (type=iibb_regime) | catalog |
| `REG_CUENTA` | `erp_entities` (type=customer/supplier) | entity |
| `PERSONAL` | `erp_entities` (type=employee) | entity |

### Phase 3: Financial (depends on catalogs + entities)

| Legacy table | SDA table | Domain |
|---|---|---|
| `CTB_CENTROS` | `erp_cost_centers` | accounting |
| `CTB_CUENTAS` | `erp_accounts` | accounting |
| `CTB_EJERCICIOS` | `erp_fiscal_years` | accounting |
| `CTB_MOVIMIENTOS` | `erp_journal_entries` | accounting |
| `CTB_DETALLES` | `erp_journal_lines` | accounting |
| `BCSCUENT` | `erp_bank_accounts` | accounting |
| `CAJMOVIM` | `erp_cash_registers` | accounting |
| `CARCHEQU` | `erp_checks` | accounting |

Note: after `erp_accounts` and `erp_journal_entries` are migrated, the orchestrator runs setup hooks to build a code index (for account FK resolution by varchar code) and an entry date cache (for journal line date resolution). This happens automatically before `erp_journal_lines` starts.

### Phase 4: Operational (depends on catalogs + entities)

| Legacy table | SDA table | Domain |
|---|---|---|
| `STK_DEPOSITOS` | `erp_warehouses` | stock |
| `ARTICULO` | `erp_articles` | stock |
| `STK_MOVIMIENTOS` | `erp_stock_movements` | stock |
| `COMPRAS` | `erp_purchase_orders` | purchasing |
| `COMPRAS_DET` | `erp_purchase_order_lines` | purchasing |
| `COTIZACIONES` | `erp_quotations` | purchasing |
| `PROD_CENTROS` | `erp_production_centers` | production |
| `PROD_ORDENES` | `erp_production_orders` | production |

### Phase 5: HR (depends on entities)

| Legacy table | SDA table | Domain |
|---|---|---|
| `PERSONAL` (detail fields) | `erp_employee_details` | hr |
| `AUSENCIAS` | `erp_absences` | hr |
| `CAPACITACIONES` | `erp_trainings` | hr |

---

## Running the Migration

### Step 1: Dry run (required before production)

```bash
sda migrate-legacy \
  --mysql-dsn="histrix_user:pass@tcp(histrix-host:3306)/saldivia?charset=utf8mb4&parseTime=true" \
  --pg-dsn="postgres://sda:pass@pg-host:5432/sda_tenant_saldivia?sslmode=require" \
  --tenant=saldivia \
  --dry-run
```

The dry run:
1. Runs `RunPreValidation` — queries legacy tables for known constraint violations
2. Prints a pre-validation report with counts by resolution type
3. If any `fix_manual` issues exist: **exits with error and blocks the production run**
4. If all clear: runs all transformations in read-only mode (no PostgreSQL INSERTs), counts `read/would_write/skipped`
5. Records `dry_run_ok` in `erp_migration_runs`

Pre-validation rules checked:

| Rule | Legacy table | Issue | Resolution |
|------|---|---|---|
| `invalid_dates_zero` | `CTB_MOVIMIENTOS` | `fecha_movimiento = '0000-00-00'` | `auto_fix` (mapped to NULL) |
| `journal_unbalanced` | `CTB_MOVIMIENTOS` | debit/credit delta > 0.01 | `fix_manual` — **blocks prod** |
| `checks_zero_amount` | `CARCHEQU` | `carimp <= 0` | `skip` |
| `invoice_dates_zero` | `IVAVENTAS` | `feciva = '0000-00-00'` | `auto_fix` |
| `stock_movements_zero_qty` | `STK_MOVIMIENTOS` | `cantidad = 0` | `skip` |
| `orphan_journal_lines` | `CTB_DETALLES` | missing parent in `CTB_MOVIMIENTOS` | `skip` |
| `importe_not_numeric` | `CTB_DETALLES` | non-numeric varchar importe | `skip` |

To fix `fix_manual` issues, correct the data in MySQL and re-run dry-run until `ALL CLEAR`.

### Step 2: Production run

After a `dry_run_ok` exists for the tenant:

```bash
sda migrate-legacy \
  --mysql-dsn="histrix_user:pass@tcp(histrix-host:3306)/saldivia?charset=utf8mb4&parseTime=true" \
  --pg-dsn="postgres://sda:pass@pg-host:5432/sda_tenant_saldivia?sslmode=require" \
  --tenant=saldivia
```

The orchestrator will refuse to run without a prior `dry_run_ok`. To bypass (e.g., for a known tenant in controlled conditions):

```bash
sda migrate-legacy ... --skip-dry-run-for=saldivia
```

### Migrate specific domains only

```bash
sda migrate-legacy ... --domains=catalog,entity
sda migrate-legacy ... --domains=accounting
```

Domain names: `catalog`, `entity`, `accounting`, `stock`, `purchasing`, `production`, `hr`

### Resume a failed run

If a run fails mid-way, the last successful batch key is checkpointed. Resume:

```bash
sda migrate-legacy \
  --tenant=saldivia \
  --resume \
  --resume-run-id=<uuid from erp_migration_runs>
```

The run UUID is printed at migration start and stored in `erp_migration_runs.id`. Only runs with `status='failed'` can be resumed.

---

## Post-Migration Verification

Run the built-in validation:

```bash
sda migrate-legacy \
  --mysql-dsn="..." \
  --pg-dsn="..." \
  --tenant=saldivia \
  --validate-only
```

This runs `PostMigrationValidation` which checks:

| Check | MySQL query | PostgreSQL query | Expected |
|-------|---|---|---|
| `entities_customer` | `COUNT(*) FROM REG_CUENTA WHERE subsistema_id='02'` | `COUNT(*) FROM erp_entities WHERE type='customer'` | equal |
| `entities_supplier` | `COUNT(*) FROM REG_CUENTA WHERE subsistema_id='01'` | `COUNT(*) FROM erp_entities WHERE type='supplier'` | equal |
| `entities_employee` | `COUNT(*) FROM PERSONAL` | `COUNT(*) FROM erp_entities WHERE type='employee'` | equal |
| `accounts` | `COUNT(DISTINCT id_ctbcuenta) FROM CTB_CUENTAS` | `COUNT(*) FROM erp_accounts` | equal |
| `journal_entries` | `COUNT(*) FROM CTB_MOVIMIENTOS` | `COUNT(*) FROM erp_journal_entries` | equal |
| `warehouses` | `COUNT(*) FROM STK_DEPOSITOS` | `COUNT(*) FROM erp_warehouses` | equal |
| `journal_debit_balance` | `SUM(CAST(importe AS DECIMAL)) FROM CTB_DETALLES WHERE doh=0` | `SUM(debit) FROM erp_journal_lines` | diff < 0.01 |

The journal debit balance check is a financial control: the total debit across all journal lines must match within 0.01 ARS.

Count mismatches are expected when pre-validation `skip` rows existed. Each skip is explained in the pre-validation report.

### Manual spot checks

```sql
-- Verify entity with known CUIT
SELECT * FROM erp_entities WHERE cuit = '20-12345678-9' AND tenant_id = 'saldivia';

-- Verify account plan migrated
SELECT COUNT(*), MAX(code) FROM erp_accounts WHERE tenant_id = 'saldivia';

-- Verify journal entry balance for a specific fiscal year
SELECT SUM(debit) - SUM(credit) AS delta
FROM erp_journal_lines jl
JOIN erp_journal_entries je ON jl.entry_id = je.id
WHERE je.tenant_id = 'saldivia'
  AND je.fiscal_year_id = (SELECT id FROM erp_fiscal_years WHERE tenant_id='saldivia' ORDER BY starts_at DESC LIMIT 1);
-- Expected: 0.00 (double-entry bookkeeping invariant)

-- Check for orphan journal lines (should be 0 in SDA)
SELECT COUNT(*) FROM erp_journal_lines jl
LEFT JOIN erp_journal_entries je ON jl.entry_id = je.id
WHERE jl.tenant_id = 'saldivia' AND je.id IS NULL;
```

---

## Rollback

The migration is **additive only** — it INSERTs data, never modifies existing rows. Rollback means truncating the migrated tables.

```sql
-- Truncate in reverse FK order
TRUNCATE erp_trainings, erp_absences, erp_employee_details;
TRUNCATE erp_production_orders, erp_production_centers;
TRUNCATE erp_purchase_order_lines, erp_purchase_orders, erp_quotations;
TRUNCATE erp_stock_movements, erp_articles, erp_warehouses;
TRUNCATE erp_checks, erp_cash_registers, erp_bank_accounts;
TRUNCATE erp_journal_lines, erp_journal_entries, erp_fiscal_years;
TRUNCATE erp_accounts, erp_cost_centers;
TRUNCATE erp_entities;
TRUNCATE erp_catalogs;

-- Clear migration tracking (so re-run starts fresh)
TRUNCATE erp_migration_table_progress, erp_migration_validation_issues;
DELETE FROM erp_migration_runs WHERE tenant_id = 'saldivia';
```

All statements must be scoped to the target tenant (`WHERE tenant_id = 'saldivia'`) if other tenants share the same database.

---

## Known Issues and Edge Cases

### Encoding: latin1 to utf8mb4

The `go-sql-driver/mysql` driver handles the conversion automatically when `charset=utf8mb4` is in the DSN. No manual encoding conversion is needed. If the legacy DB stores data in a different collation (e.g., `latin1_swedish_ci`), the driver still converts correctly at the protocol level.

### MySQL zero dates (`0000-00-00`)

MySQL allows `0000-00-00` as a date value; PostgreSQL does not. The pre-validation rule `invalid_dates_zero` detects these in `CTB_MOVIMIENTOS` and `IVAVENTAS`. The transformer maps them to `NULL`. This is recorded as `auto_fix` in the pre-validation report — no manual action required.

### NULL handling

`LegacyRow.NullString()` (`tools/cli/internal/legacy/reader.go:63`) returns `nil` for both SQL NULL and empty string. PostgreSQL `TEXT` columns receive `NULL` where MySQL had `''`. This is intentional — SDA uses `NULL` to mean "not provided", not empty string.

### Date-only columns (`parseTime=true`)

MySQL `DATE` columns (no time component) are returned as `time.Time` with midnight UTC. PostgreSQL `date` columns receive the date correctly. The driver's `parseTime=true` is required for this to work — without it, dates arrive as `[]byte`.

### CTB_DETALLES.importe is varchar

The `importe` column in `CTB_DETALLES` is `varchar(45)` in legacy. Non-numeric values are detected by pre-validation rule `importe_not_numeric` and skipped. The migrator uses `shopspring/decimal` to parse the value before inserting into PostgreSQL `numeric`.

### Composite keys in CompositeKeyReader

Tables without a single integer PK use `CompositeKeyReader` (e.g., tables keyed on multiple columns). Resume keys for these tables are serialized as `COL1=val1,COL2=val2` strings in `erp_migration_table_progress.last_legacy_key`.

### Batch size

Fixed at 500 rows per batch (`batchSize = 500` in `orchestrator.go`). Each batch runs in a single PostgreSQL transaction (mapping + data + progress checkpoint). If a batch fails, progress is marked `failed` at the last committed key — resume from there.
