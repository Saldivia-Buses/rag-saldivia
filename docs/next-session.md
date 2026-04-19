# Next session — 2.0.18: más §UI parity con Tier A agotándose

**Goal**: seguir maximizando §UI parity. Los candidatos Tier A (handlers
mounted sin página) están casi agotados — el grueso del ciclo va a ser
Tier B (queries con handler nuevo) + primeros Tier C (generalizaciones
de queries con filter obligatorio o nuevas sub-pages `[id]` de detalle).

Target realista: **8–12 clusters**. El upper bound depende de cuántas
queries Tier B solid queden y cuánto esfuerzo requiere el primer batch
de `[id]` routes.

## Cierre sesión anterior (6 ciclos en una sola sesión)

| Ciclo | Clusters nuevos | Tag | PR |
|---|---:|---|---:|
| 2.0.12 | 1 | v2.0.12 | pre-sesión |
| 2.0.13 | 1 | v2.0.13 | #160 |
| 2.0.14 | 3 | v2.0.14 | #161 |
| 2.0.15 | 6 | v2.0.15 | #162 |
| 2.0.16 | 10 | v2.0.16 | #163 |
| 2.0.17 | 12 | v2.0.17 | #164 |

Totales post-2.0.17: **33 §UI parity clusters tracked** en
`docs/parity/ui-parity.md`, 99 `page.tsx` routes shipped, 1 waiver
(W-009 bulk CSV/XLS import).

Workstation `srv-ia-01` sincronizada en `6d953e0c`.

## Final goal (ADR 026 — no se pierde de vista)

SDA reemplaza Histrix. El empleado abre SDA y:

1. UI moderna cubriendo **todo** lo que Histrix hacía (1:1 parity).
2. Chat donde el agente es su representante — cap parity chat ↔ UI.
3. Dashboard personal + rutinas personales.
4. Agentes background: mail, WhatsApp, tree-RAG con ACL.

## Estado post-2.0.17

- **§UI parity**: 33 clusters explícitos. Aún faltan centenares de
  XML-forms sin cobertura — no se ha rozado el bulk de
  `.intranet-scrape/xml-forms/` (~4,500 files).
- **Backend surface**: la gran mayoría de sqlc `List*` queries con
  tenant+pagination están mounted. El pozo de Tier A barato ya es
  superficial.
- **Arquitectura ERP handler**: los 4 endpoints Tier B de 2.0.17 viven
  bajo `/v1/erp/admin/` (product-sections, products, product-attributes,
  tools) como shortcut. Hay que decidir en un refactor dedicado si
  splitear a `Products` / `Tools` services propios.

## Plan de trabajo 2.0.18

### Pre-work

```bash
git checkout -b 2.0.18 main
sed -i 's/Working:\*\* `2.0.17`/Working:\*\* `2.0.18`/' CLAUDE.md
git commit -am "chore: bump working branch 2.0.17 → 2.0.18

[resumen 2.0.17 + plan 2.0.18]

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

### Investigación inicial (agente Explore)

Antes de decidir scope, dispatch un agente paralelo con:

1. **Re-mapear Tier A residual**: después de 2.0.17 hay que recontar
   los endpoints `GET` mounted sin page. El de 2.0.17 encontró ~16,
   pero cubrimos 8 — los restantes pueden ser casi todos dropdown-only
   (nc-origins, hr/departments, catalogs/types, …). Lo que queda real:
   - `/v1/erp/accounting/ledger` (requiere account_id — Tier C)
   - `/v1/erp/invoicing/tax-book` (requiere period — Tier C)
   - `/v1/erp/maintenance/assets` (YA consumido en `mantenimiento/equipos/page.tsx`)
   - `/v1/erp/purchasing/suppliers/{id}/demerits` (detail — requiere `[id]` page)
2. **Re-mapear Tier B viable**: el agente de 2.0.17 listó 13 Tier B,
   descartamos 7 como duplicados (ListInspections, ListInvoiceNotes,
   ListCalendarEvents, ListWorkOrders, ListFuelLogs, ListEntries /
   ListJournalEntries, ListReceipts mounted vía `ListReceipts`).
   Quedan viables: **ListMaintenanceAssets wrapper for dedicated
   catalog page (si no ya cubierta), ListCatalogTypes (thin, skippable),
   y el resto probablemente Tier C.**
3. **Tier C worth unlocking**: candidatos con filter obligatorio pero
   fácil de generalizar:
   - `ListTreasuryMovements` — ya mounted en service, pero ¿la page
     existe? Treasury page (`/administracion/tesoreria/page.tsx`) ya
     consume movements. Skip.
   - `ListStatementLines` — requiere `reconciliation_id`. Natural
     `[id]` route bajo reconciliaciones.
   - `ListPriceListItems` — requiere `price_list_id`. Natural `[id]`
     bajo `/compras/listas-precios/[id]/items`.
   - `ListQuotationLines` — requiere `quotation_id`. Abre `/ventas`
     módulo (hoy vacío).
   - `ListPurchaseOrderLines` — requiere `order_id`. Sub-page de
     purchase orders.
   - `ListInvoiceLines` — requiere `invoice_id`. Sub-page de invoice
     detail (ya hay invoice list en `/administracion/facturacion`).
   - `ListMaintenancePlans` — requiere `asset_id`. Sub-page bajo
     `/mantenimiento/equipos/[id]/planes`.
   - `ListWorkOrderParts` — requiere `work_order_id`. Sub-page bajo
     work order detail.

### Batches propuestos

**Batch 1 — Primera fila de detail sub-pages** (abrir patrón `[id]`):

1. **Reconciliación → líneas de extracto**
   `/administracion/tesoreria/reconciliaciones/[id]/page.tsx` — lista
   StatementLines + match status. Trae el entity picker para
   selección de cuenta bancaria (ya hay).
2. **Lista de precios → items**
   `/compras/listas-precios/[id]/page.tsx` — lista PriceListItems con
   article_code + description + price.
3. **Order de compra → líneas**
   `/compras/ordenes/[id]/page.tsx` — lista PurchaseOrderLines con
   article, qty, unit_price, total.
4. **Factura → líneas + tax entries**
   `/administracion/facturacion/[id]/page.tsx` — lista InvoiceLines +
   la lista ya-mounted TaxEntries.

Cada uno es un `[id]/page.tsx` nuevo + probablemente un link desde la
página list correspondiente al detail. El shape se repite, pero
estandarizar el pattern `[id]` abre un segundo eje de coverage.

**Batch 2 — Módulo `/ventas` desde cero (si queda tiempo)**:

5. **Cotizaciones list** `/ventas/cotizaciones` — consume
   GET /v1/erp/sales/quotations (ya mounted).
6. **Cotización → líneas + options**
   `/ventas/cotizaciones/[id]/page.tsx` — abre QuotationLines +
   QuotationOptions.
7. **Recetas de producción** (si queda) —
   `/produccion/recetas` sobre `ListProductionOrderRecipes` si
   existe sin handler.

**Batch 3 — Refactor shortcut de 2.0.17**:

8. **Split Admin's Tier B endpoints**: mover product-sections /
   products / product-attributes / tools a services/handlers
   dedicados (`/v1/erp/products/*`, `/v1/erp/tools/*`). Los 4 pages
   frontend ya existen; sólo cambia el URL y el mount point. Es un
   refactor trivial pero limpia deuda de 2.0.17.

### Cierre esperado

Post-2.0.18:

- ≥ 8 clusters nuevos trackados en ui-parity.md → total ≥ 41.
- Pattern `[id]/page.tsx` abierto en al menos 3 módulos
  (tesorería, compras, facturación).
- Opcional: refactor de Admin → Products/Tools dedicado.
- Phase 0 gates verdes.

## Candidatos para sesiones futuras (lookahead — NO 2.0.18)

Ordenados por importancia para cerrar Phase 1:

| Orden | Tema | Notas |
|---:|---|---|
| 1 | **Write paths en clusters read-only existentes** | Create/update en reclamos, importaciones, carchehi (partial), calificaciones, notas. Cada uno requiere audit + NATS + entity pickers. |
| 2 | **Módulos vacíos con backend listo** | `/ventas` está vacío; `/compras/abastecimiento`, `/compras/comex` sub-features; `/rrhh/legajos` detalle. |
| 3 | **Reports (Phase 1 §Reports parity)** | ADR 027 Phase 1 §Reports es un eje separado de §UI parity. Libro IVA, mayor contable, tax-book necesitan UI + filters + export. |
| 4 | **Cutover seamless-day test** | Phase 1 gating final — una persona trabaja un día completo en SDA sin abrir Histrix. Gap discovery. |
| 5 | **Write paths para crear/editar catálogos** | bodegas, cajas, centros-producción, bank-accounts, cash-registers — hoy todos read-only. |

## Trampas heredadas

- **Agents hallucinate**: ambos agentes paralelos de 2.0.17
  mintieron sobre varios handlers ("ya mounted" cuando no lo
  estaban, y viceversa). Siempre verificar con grep antes de
  construir. Pattern: el agente dice `maintenance.go:42` pero puede
  estar apuntando a un route diferente — confirmar con
  `r.Get(".*endpoint"` específico.
- **Nested git repos**: `apps/web/.git` es un repo separado (leftover
  de setup). `cd apps/web` + `git stash` NO afecta el repo outer.
  Para comandos git del repo principal, usar `git -C /home/enzo/rag-saldivia`.
- **cwd persiste entre Bash calls**: si un comando hace `cd apps/web`,
  los siguientes Bash siguen ahí. Volver con `cd /home/enzo/rag-saldivia`
  o usar rutas absolutas.
- **Registry `FileText` warning**: `src/lib/modules/registry.ts:22`
  tiene `'FileText' is defined but never used` — lint warning
  pre-existente desde main. No bloquea, ignorable.
- **Admin handler bloat**: los 4 endpoints Tier B bolted en 2.0.17
  acumulan `Products` y `Tools` fuera de su domain natural. Si no
  se splitean en 2.0.18, anotar en ADR como deuda explícita.

## Fuera de scope 2.0.18

- **Phase 2+** (chat agent, prompts jerárquicos, tree-RAG, ACL): sigue
  detrás de Phase 1 §UI parity en el orden top-down.
- **ADR 027 §UI parity row 1 tick**: cerrar "Every XML-form has SDA
  equivalent or waiver" requiere cubrir ~4,500 forms. No en este
  ciclo. Cada cluster sigue agregando filas a `ui-parity.md`.
- **W-009 file-upload** (bulk CSV/XLS bank import): sigue waived.
- **Refactor de Admin Tier B endpoints**: opcional en Batch 3, puede
  diferirse a un ciclo dedicado.

## Post-PR cierre ciclo

```bash
gh pr create --base main --head 2.0.18 --title "..." --body "..."
# Post-merge:
git checkout main && git pull origin main
git tag v2.0.18 && git push origin v2.0.18
gh release create v2.0.18 --title "..." --notes "..."
ssh sistemas@srv-ia-01 "cd /opt/saldivia/repo && git pull origin main"
```

Release body incluye:
- Tabla de clusters shipped (shape idéntico a 2.0.16 / 2.0.17).
- ADR 027 deltas (clusters tracked, endpoints nuevos).
- Link al PR.
