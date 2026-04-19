# Phase 1 UI Parity — coverage log

Living inventory of Histrix XML-form → SDA Next.js page coverage.
Feeds ADR 027 Phase 1 §UI parity row 1 ("Every XML-form in
`.intranet-scrape/xml-forms/` has either an `apps/web/src/app/**/page.tsx`
equivalent reachable from the nav, OR a waiver entry in
`docs/parity/waivers.md`").

Unit of work is the **cluster** — one Histrix area/form group maps to
one SDA route. A cluster is "covered" when the SDA page delivers the
same operational surface as the Histrix form(s) it replaces.

## Totals (2026-04-19, post-2.0.15)

| Segment | Count |
|---|---:|
| Histrix XML-forms in `.intranet-scrape/xml-forms/` | ~4,500 files |
| Histrix top-level forms (area/form groups) | 434 |
| SDA `page.tsx` routes shipped | 77 |
| **SDA pages explicitly tracked as XML-form parity** | **11** (this file) |
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

### Cluster: Auditorías de calidad (2.0.15)

**SDA page:** `apps/web/src/app/(modules)/calidad/auditorias/page.tsx`
→ `/calidad/auditorias`.

**Covers Histrix XML-forms:** auditoría-list views under the Histrix
calidad / control de calidad area (read-only).

**Data dependency:** `erp_audits`. Backend: existing
`GET /v1/erp/quality/audits` (shipped previously, just not wired to
UI until now).

Columns: número / fecha / tipo / alcance / puntaje / estado (badge).

### Cluster: Documentos controlados (2.0.15)

**SDA page:** `apps/web/src/app/(modules)/calidad/documentos-controlados/page.tsx`
→ `/calidad/documentos-controlados`.

**Covers Histrix XML-forms:** `calidad2/cal_registros.xml` —
documentos controlados del SGC.

**Data dependency:** `erp_controlled_documents`. Backend: existing
`GET /v1/erp/quality/documents`.

Columns: código / título / revisión / fecha aprobación / badge
estado. Status tabs: Aprobados / Borradores / Obsoletos / Todos.

### Cluster: Planes de acción (2.0.15)

**SDA page:** `apps/web/src/app/(modules)/calidad/planes-accion/page.tsx`
→ `/calidad/planes-accion`.

**Covers Histrix XML-forms:** `calidad2/cal_plan_accion_total_listado.xml`
— listado de planes de acción.

**Data dependency:** `erp_quality_action_plans`
(`ListActionPlansRow`). Backend: existing
`GET /v1/erp/quality/action-plans`.

Columns: descripción / inicio planificado / fecha objetivo / cierre
/ badge estado. Status tabs: Abiertos / En curso / Cerrados / Todos.

### Cluster: Indicadores de calidad (2.0.15)

**SDA page:** `apps/web/src/app/(modules)/calidad/indicadores/page.tsx`
→ `/calidad/indicadores`.

**Covers Histrix XML-forms:** `calidad2/programas_qry.xml` —
indicadores / KPIs por período.

**Data dependency:** `erp_quality_indicators`. Backend added in
2.0.15: `GET /v1/erp/quality/indicators` (new handler on top of the
existing sqlc query).

Columns: período / indicador / valor / objetivo / cumplimiento
(badge OK / Bajo). Período-range filter defaults to the current year.

### Cluster: Movimientos de stock (2.0.15)

**SDA page:** `apps/web/src/app/(modules)/administracion/almacen/movimientos/page.tsx`
→ `/administracion/almacen/movimientos`.

**Covers Histrix XML-forms:** `stock/qry/stkmovimientos_qry.xml` —
movimientos de stock por artículo / almacén.

**Data dependency:** `erp_stock_movements`. Backend: existing
`GET /v1/erp/stock/movements` (handler already took `article_id` as
optional; no backend change needed).

Columns: fecha / cód. art. / artículo / tipo (badge) / cantidad /
costo unit. / notas. Type tabs: Todos / Ingresos / Egresos /
Transferencias.

### Cluster: Costos de artículos (2.0.15)

**SDA page:** `apps/web/src/app/(modules)/administracion/almacen/costos/page.tsx`
→ `/administracion/almacen/costos`.

**Covers Histrix XML-forms:** `stock/costos/` — ledger de costos por
proveedor.

**Data dependency:** `erp_article_costs` (STKINSPR, 189,863 rows
live). Backend added in 2.0.15: the sqlc query was generalized from
a single-article shape to optional article / supplier filters with
LEFT JOINs on `erp_articles` + `erp_entities` (ships article_name +
supplier_name). New endpoint
`GET /v1/erp/stock/article-costs?article_id=&supplier_code=&page=&page_size=`.

Columns: cód. art. / artículo / cód. prov. / proveedor / costo /
últ. actualización / fecha factura.

### Cluster: Notas de comprobantes (2.0.14)

**SDA page:** `apps/web/src/app/(modules)/administracion/facturacion/notas/page.tsx`
→ `/administracion/facturacion/notas` (Administración → Notas de
comprobantes).

**Covers Histrix XML-forms:**
- `clientes/qry/regmovim_obs_qry.xml` — abm-mini that backs the
  "Observaciones" helper inside each REG_MOVIMIENTOS row. The SDA
  page exposes the same data as a standalone list (subsistema /
  comprobante / observación / usuario), read-only. Write/edit stays
  in Histrix for now — the form requires GEN_TIPO_CONTACTOS which is
  not migrated.

**Data dependency:** `erp_invoice_notes` (migration `077`, shipped in
2.0.11; REG_MOVIMIENTO_OBS → 72,737 rows live).

**Backend endpoints added (scoped under `/v1/erp/invoicing`):**
- `GET /invoice-notes?invoice_id=&date_from=&date_to=&page=&page_size=`

Permission: `erp.invoicing.read`. Read-only. The sqlc query already
supported optional invoice + date filters from the original 2.0.11
migration PR.

**Notable gaps vs Histrix:**
- Create / edit: the Histrix form is an abm-mini (insert/update/
  delete). SDA defers the write path — the picker for
  GEN_TIPO_CONTACTOS (tipo_contacto) and the invoice-picker both need
  separate work. Tracked as follow-up, not waived.
- Drill-down to the source comprobante: Histrix links
  `regmovim_id` back into `cc_notas_venta.xml`. SDA shows
  `movement_voucher_class-movement_no` as a label; deep link into
  facturación will land when facturación has an `/invoices/[id]`
  route.

### Cluster: Calificaciones de cuentas (2.0.14)

**SDA page:** `apps/web/src/app/(modules)/compras/calificaciones/page.tsx`
→ `/compras/calificaciones` (Compras → Calificaciones).

**Covers Histrix XML-forms:**
- `compras/calificacion_prov.xml` — "Proveedores aprobados", qry that
  computes each proveedor's rating from MOVDEMERITO + OCPRECIB. SDA
  serves the **persisted rating-event history** (one row per rating
  change) instead of recomputing on the fly — the migrated
  REG_CUENTA_CALIFICACION already holds the outputs.

**Data dependency:** `erp_entity_credit_ratings` (migration `077`,
shipped in 2.0.11; REG_CUENTA_CALIFICACION → 136,064 rows live).

**Backend endpoints added (scoped under `/v1/erp/entities`):**
- `GET /credit-ratings?entity_id=&rating=&page=&page_size=`

Permission: `erp.entities.read`. Read-only. The sqlc query was
generalized from the original single-entity shape to accept an
optional entity filter + optional rating filter, and LEFT JOINs
`erp_entities` so each row ships with entity_name + entity_type.

**Notable gaps vs Histrix:**
- The Histrix form RECOMPUTES the rating on render from
  MOVDEMERITO. The SDA page shows the persisted history only — if the
  business ever wants a live recompute view, that's a new query
  against MOVDEMERITO-era tables (not migrated at this point; Grupo A
  rank 1 only ships the rating history).
- Rating-change trigger: the Histrix form doesn't write ratings
  directly — they're an output of the demeritos / recibos flow. SDA
  will grow a write path when that flow enters scope.

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
