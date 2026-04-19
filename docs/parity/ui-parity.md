# Phase 1 UI Parity — coverage log

Living inventory of Histrix XML-form → SDA Next.js page coverage.
Feeds ADR 027 Phase 1 §UI parity row 1 ("Every XML-form in
`.intranet-scrape/xml-forms/` has either an `apps/web/src/app/**/page.tsx`
equivalent reachable from the nav, OR a waiver entry in
`docs/parity/waivers.md`").

Unit of work is the **cluster** — one Histrix area/form group maps to
one SDA route. A cluster is "covered" when the SDA page delivers the
same operational surface as the Histrix form(s) it replaces.

## Totals (2026-04-19, post-2.0.14)

| Segment | Count |
|---|---:|
| Histrix XML-forms in `.intranet-scrape/xml-forms/` | ~4,500 files |
| Histrix top-level forms (area/form groups) | 434 |
| SDA `page.tsx` routes shipped | 71 |
| **SDA pages explicitly tracked as XML-form parity** | **5** (this file) |
| XML-form waivers (§UI) | **1** (W-009, see `waivers.md`) |

The 67 existing SDA pages cover the ERP admin surface structurally
(clientes, proveedores, tesorería, facturación, contable, compras,
ventas, producción, mantenimiento, rrhh, seguridad, calidad, etc.)
but most were designed before the parity contract existed. The
explicit XML-form-to-page mapping is what this document tracks.

## Clusters shipped

### Cluster: Reclamos de pagos (2.0.12)

**SDA page:** `apps/web/src/app/(modules)/administracion/reclamos/page.tsx`
→ `/administracion/reclamos` (Administración → Reclamos de pagos).

**Covers Histrix XML-forms:**
- `reclamos/reclamopagos.xml` (abm-mini — main CRUD list)
- `reclamos/reclamopagos_ing.xml` (ing — work-queue grouped by proveedor)
- `reclamos/reclamopagos_ingmov.xml` (update — inline marca toggle)

**Data dependency:** `erp_payment_complaints` (migration `079`,
shipped in 2.0.11; RECLAMOPAGOS → 15,463 rows migrated).

**Backend endpoints added (scoped under `/v1/erp/accounts`):**
- `GET  /complaints?status=0|1|-1&entity_id=&limit=&offset=`
- `POST /complaints` (body: `{date, entity_legacy_code, observation}`)
- `PATCH /complaints/{id}/status` (body: `{status: 0|1}`)

Permissions: `erp.accounts.read` for list, `erp.accounts.write` for
create / status toggle. Every write publishes an `erp_payment_complaints`
NATS event and a strict audit log entry (`erp.payment_complaints.*`).

**Status vs Histrix (post-2.0.13):** parity-complete for the
operational surface the three XML-forms describe.

- Entity picker: shipped in 2.0.13 —
  `apps/web/src/components/erp/entity-picker.tsx` wraps
  `GET /v1/erp/entities?type=supplier&search=…` in a debounced modal
  and feeds `entity_id` + `entity_legacy_code` into the complaint
  create body. Reusable across clusters.
- Saldo aggregate: shipped in 2.0.13 — the reclamos page now runs a
  second query against `/v1/erp/accounts/balances?direction=payable`
  and renders the supplier's open balance next to each complaint row,
  matching the `SUM(saldo_movimiento)` group-by of
  `reclamopagos_ing.xml`.

### Cluster: Cheques históricos (2.0.14)

**SDA page:** `apps/web/src/app/(modules)/administracion/tesoreria/cartera-historica/page.tsx`
→ `/administracion/tesoreria/cartera-historica` (Administración →
Cheques históricos).

**Covers Histrix XML-forms:**
- `cheques/carchehi.xml` — "CARTERA DE CHEQUES HISTORICA". Query view
  with filters (carint, carnro, carfec, caring, carimp, ctanom) and
  columns fecha / nro interno / nro cheque / banco / importe / tipo /
  emisión / acreditación / proveedor.
- `cheques/carchehi_abm.xml` — write form (not covered here; the
  archive is read-only from SDA's perspective — historical cheques are
  migrated out of the active cartera and stay frozen).

**Data dependency:** `erp_check_history` (migration `078`, shipped in
2.0.11; CARCHEHI → 29,006 rows live).

**Backend endpoints added (scoped under `/v1/erp/treasury`):**
- `GET /check-history?entity_id=&date_from=&date_to=&page=&page_size=`

Permission: `erp.treasury.read`. Read-only; no write path.

**Notable gaps vs Histrix:**
- Entity picker: the Histrix form links back to CCTCUENT / CCTMOVIM
  for proveedor lookup. The SDA page surfaces `entity_legacy_code` but
  does not yet resolve it to `entity_name` (would need the same client-
  side join we use in reclamos — follow-up).
- The `O.Pago` / `Recibo` / `Minuta` columns link into CCTMOVIM in
  Histrix. The SDA page shows `movement_no` / `pay_no` / `received_no`
  as raw numbers; drill-down to the receipt / payment record lives in
  the wider treasury cluster.

### Cluster: Importaciones bancarias (2.0.13)

**SDA page:** `apps/web/src/app/(modules)/administracion/tesoreria/importaciones/page.tsx`
→ `/administracion/tesoreria/importaciones` (Administración →
Importaciones bancarias).

**Covers Histrix XML-forms:**
- `bancos_local/bcs_importacion_qry.xml` (query view — filters by
  account, date range, processed state; shows fecha / concepto /
  número / débito / crédito / saldo).
- `bancos_local/bcsmovim_importacion_auto_mov_ins.xml` (per-row
  processed toggle — maps to the "Marcar procesado" / "Reabrir"
  action). Partial coverage: the UI toggles `processed` without
  forcing a `treasury_movement_id` link; the full match flow still
  goes through Histrix.
- `bancos_local/bcsmovim_importacion_auto_ins.xml` — **waived**
  (`W-009`, see `docs/parity/waivers.md`). Bulk CSV/XLS ingest
  requires file-upload infra; the staging table is still populated
  via the existing Histrix pipeline.

**Data dependency:** `erp_bank_imports` (migration `076`, shipped in
2.0.10; BCS_IMPORTACION → 91,959 rows live).

**Backend endpoints added (scoped under `/v1/erp/treasury`):**
- `GET   /imports?account=&processed=&date_from=&date_to=&page=&page_size=`
- `PATCH /imports/{id}` (body: `{processed: 0|1|2,
  treasury_movement_id?: "<uuid>"}`)

Permissions: `erp.treasury.read` for list, `erp.treasury.write` for
toggle. The write publishes an `erp_bank_imports` NATS event and a
strict audit entry (`erp.bank_imports.processed_changed`).

**Notable gaps vs Histrix:**
- Treasury-movement picker: Histrix's `auto_mov_ins` form binds
  `regmovim_id` (movement id) alongside `procesado=1` via a linked
  `bcsmovim_conci_bcsmov_ins.xml` helper. The SDA page does not yet
  include a treasury-movement picker, so toggling "procesado" leaves
  `treasury_movement_id` NULL. Full matching ships in a later cluster
  along with a reconciliation picker.

## How to tick an XML-form here

1. Write the SDA page (`apps/web/src/app/**/page.tsx`) or extend an
   existing one.
2. Wire backend endpoints (prefer reusing sqlc queries attached to
   the migrated table).
3. Add the route to `apps/web/src/lib/modules/registry.ts` so it
   shows in the nav.
4. Add the cluster entry here. Waivers go to `waivers.md`.
5. In the same PR, tick the matching item in ADR 027 if the cluster
   closes a named Phase 1 §UI parity milestone.
