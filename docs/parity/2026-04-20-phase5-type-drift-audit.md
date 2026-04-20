# Phase 5 audit — type-vs-schema drift (2.0.21)

**Date:** 2026-04-20.

## Method

Compared each `interface` exported from
`apps/web/src/lib/erp/types.ts` (89 total) against the sqlc-generated
struct in `services/erp/internal/repository/models.go` (172 structs).
Mapped 29 frontend interfaces to their direct backend table struct
(`Account → ErpAccount`, `JournalEntry → ErpJournalEntry`, etc.).
Compared field names with a normalizer that handles Go's `ID` /
`URL` / `JSON` / `BOM` casing.

## Result — no critical drift

Only **15 frontend fields are not present in the backend struct**,
across 8 of the 29 mapped interfaces. Every one is a recognised
**JOIN-derived display field**, not a fabricated/orphan column:

| Interface | Fields outside the table struct | Reason |
|---|---|---|
| `JournalLine` | `account_code`, `account_name` | Joined from `erp_accounts` |
| `TreasuryMovement` | `entity_name` | Joined from `erp_entities` |
| `BankBalance` | `total_in`, `total_out`, `balance` | Computed aggregates |
| `Tool` | `status`, `delivery_date`, `supplier_entity_id` | Computed / latest movement |
| `BankReconciliation` | `bank_name`, `account_number` | Joined from `erp_bank_accounts` |
| `PurchaseOrder` | `supplier_name` | Joined from `erp_entities` |
| `PurchaseOrderLine` | `article_code`, `article_name` | Joined from `erp_articles` |
| `VehicleIncident` | `incident_type_name` | Joined from `erp_incident_types` |

These are the API response shape, not the table shape. They are
correct.

## What this audit does NOT cover

The 60 interfaces with no direct BE struct mapping (`Quotation`,
`WorkOrder`, `EntityContact`, `EntityNote`, `Invoice`, etc.) are
**composite DTOs**: the API computes them by joining several tables.
This script can't validate them by structural comparison alone.

To validate them properly, the next session should:

1. Hit each `GET /v1/erp/...` endpoint against the local
   `sda_tenant_test` mirror, capture the JSON response shape.
2. Diff the JSON keys against the corresponding TS interface fields.
3. Flag any TS field absent from the API response (truly fabricated)
   AND any API field missing from the TS interface (under-typed).

That work is mechanical and can be a single TS script that loops
over a YAML mapping `{interface_name → endpoint_path}`. Out of
scope for 2.0.21; goes on the 2.0.22 backlog.

## Two known fixes already shipped

`EntityContact` and `EntityNote` had fields fabricated against the
real schema (`name/role/email/phone` vs the actual
`type/label/value`). Both were corrected in cluster 15 of cycle
2.0.21 (commit `b0fc8322`-era). The mechanical audit above doesn't
catch those because they're composite DTOs — but the manual fix
during cluster work confirms the pattern is real.

## Decision recorded for 2.0.21

No additional type fixes ship in this cycle. The audit demonstrates
that for direct-table mappings, FE/BE alignment is intact. The
composite-DTO drift survey is a 2.0.22 task.
