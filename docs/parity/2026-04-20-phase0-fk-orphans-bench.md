# Phase 0 incident — FK orphans in saldivia_bench

**Detected:** 2026-04-20 during cycle 2.0.21 Phase 1 (DB local
restore from workstation `sda_tenant_saldivia_bench`).

## Summary

`pg_restore` of the migrated tenant DB fails to apply **15 foreign-key
constraints** because the source bench has orphan rows: child rows
whose parent FK target no longer exists. The data restored OK; the
constraints could not be enforced.

This violates Phase 0 of ADR 027 ("data integrity = religion") —
specifically the implicit invariant that every FK is satisfiable.
The migrator did not detect or block these orphans.

## The 15 broken constraints

| Table | FK column | Parent table |
|---|---|---|
| `erp_purchase_order_lines` | `order_id` | `erp_purchase_orders` |
| `erp_invoice_lines` | `invoice_id` | `erp_invoices` |
| `erp_tax_entries` | `invoice_id` | `erp_invoices` |
| `erp_purchase_orders` | `supplier_id` | `erp_entities` |
| `erp_account_movements` | `entity_id` | `erp_entities` |
| `erp_hr_events` | `entity_id` | `erp_entities` |
| `erp_invoices` | `entity_id` | `erp_entities` |
| `erp_treasury_movements` | `entity_id` | `erp_entities` |
| `erp_withholdings` | `entity_id` | `erp_entities` |
| `erp_orders` | `customer_id` | `erp_entities` |
| `erp_units` | `customer_id` | `erp_entities` |
| `erp_quotations` | `customer_id` | `erp_entities` |
| `erp_work_orders` | `asset_id` | `erp_assets` |
| `erp_fuel_logs` | `asset_id` | `erp_assets` |
| `erp_production_inspections` | `step_id` | `erp_production_steps` |

The dominant pattern is `erp_entities` (clientes/proveedores) —
8 of 15 constraints reference it. The migrator either drops entity
rows that are still referenced, or imports child rows whose parent
was never imported.

## Why it does not block this cycle

Cycle 2.0.21 is frontend reinforcement against a local mirror of
the bench. The UI reads through these tables but does not depend on
the constraints being enforced — joins still return rows even when
FKs are rotten. Visible symptoms in UI may include:

- Invoice / order detail pages where the entity badge says
  "(huérfano)" or shows an empty entity link.
- Cost-center detail with no parent journal entry.
- Asset-keyed pages where the asset row is missing.

## What the next migration session must do

- Add a pre-flight check in the migrator: for every FK column in the
  target schema, count rows whose value is non-NULL and is missing
  from the parent. Block the run if any are non-zero.
- For the 8 `erp_entities` cases: audit the entities migrator to
  ensure no entity row that is referenced by any child table is
  ever skipped or overwritten. The bench likely lost entities during
  legacy archive consolidation.
- For `erp_assets` (work_orders, fuel_logs) and `erp_production_steps`
  (production_inspections): same audit.
- Re-run the migration on a fresh tenant DB and confirm
  `pg_restore --create --clean` succeeds with zero FK errors.

This goes on the ADR 027 Phase 0 checklist for cycle 2.0.22.
