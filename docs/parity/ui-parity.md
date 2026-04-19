# Phase 1 UI Parity — coverage log

Living inventory of Histrix XML-form → SDA Next.js page coverage.
Feeds ADR 027 Phase 1 §UI parity row 1 ("Every XML-form in
`.intranet-scrape/xml-forms/` has either an `apps/web/src/app/**/page.tsx`
equivalent reachable from the nav, OR a waiver entry in
`docs/parity/waivers.md`").

Unit of work is the **cluster** — one Histrix area/form group maps to
one SDA route. A cluster is "covered" when the SDA page delivers the
same operational surface as the Histrix form(s) it replaces.

## Totals (2026-04-19, post-2.0.12)

| Segment | Count |
|---|---:|
| Histrix XML-forms in `.intranet-scrape/xml-forms/` | ~4,500 files |
| Histrix top-level forms (area/form groups) | 434 |
| SDA `page.tsx` routes shipped | 67 |
| **SDA pages explicitly tracked as XML-form parity** | **1** (this file) |

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

**Notable gaps vs Histrix:**
- Entity picker: Histrix uses a modal `ayuda_ex` popup against
  `CCTCUENT` / `REG_CUENTA`. The SDA page currently takes the raw
  `ctacod` number in the create form and stores it as
  `entity_legacy_code`. An entity-UUID resolver is a follow-up —
  the backend accepts `entity_id` already, so dropping in a picker
  is additive.
- Saldo aggregate: `reclamopagos_ing.xml` joins REG_MOVIMIENTOS and
  sums `saldo_movimiento` per proveedor so the operator sees the
  outstanding balance alongside each complaint. Not included in the
  first cut — will land when `GET /accounts/balances` is wired into
  the reclamos view (same endpoint that feeds `/cuentas`).

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
