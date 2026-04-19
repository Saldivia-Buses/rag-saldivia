# Next session — 2.0.19: más §UI parity con [id] pattern estandarizado

**Goal**: seguir el momentum de 2.0.18. El pattern `[id]/page.tsx` ya
vive en 5 módulos — ahora toca cubrir los endpoints `GET /{id}` que
están mounted pero sin página, más algún endpoint Tier C que necesite
una extensión quirúrgica tipo `GetPriceList`.

Target realista: **6-10 clusters**. Mix esperado:
- 4-6 `[id]` en endpoints ya mounted (receipts, work orders, inspections, …)
- 1-2 clusters con pequeña extensión de backend (sales orders, supplier demerits)
- 1-2 Tier A residuales (accounting entries list, fuel logs, tax-book)

## Cierre 2.0.18 (8 clusters)

| Ciclo | Clusters nuevos | Tag | PR |
|---|---:|---|---:|
| 2.0.18 | 8 | v2.0.18 | #165 |

Totales post-2.0.18: **41 clusters** tracked en `ui-parity.md`,
**110 `page.tsx`** shipped, 1 waiver (W-009).

Workstation `srv-ia-01` sincronizada en `9706ad93`.

## Final goal (ADR 026 — no se pierde de vista)

SDA reemplaza Histrix. Empleado abre SDA y:

1. UI moderna cubriendo **todo** lo que Histrix hacía (1:1 parity).
2. Chat-agente como representante — cap parity chat ↔ UI.
3. Dashboard personal + rutinas personales.
4. Agentes background: mail, WhatsApp, tree-RAG con ACL.

## Estado post-2.0.18

- **`[id]` pattern** estandarizado en 5 módulos (tesorería, compras,
  facturación, mantenimiento, almacén) + módulo `/ventas` nuevo.
- **Backend surface**: varios `Get{Entity}(id)` con detail bundle ya
  mounted que aún no tienen página — receipts (treasury), work orders
  (maintenance), inspections (purchasing), supplier demerits detail.
- **Deuda explícita** de 2.0.17: refactor Admin Tier B (products/tools)
  sigue sin hacerse — no es un cluster §UI parity nuevo, sólo higiene.
- **Deuda de 2.0.18**: no hay `GetAsset` ni `GetArticle` dedicados
  (clusters #7/#8 reutilizan la list cache). Vale agregar cuando haya
  refactor stock/maintenance, no urge.

## Plan de trabajo 2.0.19

### Pre-work

```bash
git checkout -b 2.0.19 main
sed -i 's/Working:\*\* `2.0.18`/Working:\*\* `2.0.19`/' CLAUDE.md
git commit -am "chore: bump working branch 2.0.18 → 2.0.19"
```

### Investigación inicial (agente Explore)

Re-confirmar antes de shipear — patrón de 2.0.18 funcionó:

1. **Receipts `[id]`**: `GET /v1/erp/treasury/receipts/{id}` mounted,
   service `GetReceipt` ya bundlea `{Receipt, Imputations?}`. Parent
   page `/administracion/tesoreria/receipts` o similar.
2. **Work orders `[id]`**: `GetWorkOrder` en maintenance.go:43. Parent
   page `/mantenimiento/preventivo` o `/correctivo` — chequear cuál
   lista work orders y donde linkear.
3. **Inspections `[id]`**: `GetInspection` en purchasing.go:61. Parent
   `/calidad/inspecciones`.
4. **Supplier demerits `{id}`**: `GET /v1/erp/purchasing/suppliers/{id}/demerits`
   mounted, necesita parent page de supplier (puede que no exista aún
   detail page de supplier). Si no existe, puede ser un cluster
   completo: `/compras/proveedores/[id]/demeritos`.
5. **Accounting entries list**: `GET /v1/erp/accounting/entries` mounted,
   sin página. Parent `/administracion/contable`.
6. **Sales orders**: `ListOrders` mounted en sales.go:39, pero NO
   `GetSalesOrder` — agregar backend quirúrgico (patrón PriceList) +
   crear `/ventas/ordenes` list + `[id]` detail.

### Batches propuestos

**Batch 1 — [id] routes con backend ya listo** (4 clusters):

1. Receipt detail (`/administracion/tesoreria/recibos/[id]`)
2. Work order detail (`/mantenimiento/.../ordenes-trabajo/[id]`)
3. Inspection detail (`/calidad/inspecciones/[id]`)
4. Accounting entries list (`/administracion/contable/asientos`)

**Batch 2 — Sales orders (list + detail, requiere backend)**:

5. `/ventas/ordenes/page.tsx` — consume ListOrders existente.
6. `/ventas/ordenes/[id]/page.tsx` — requiere `GetSalesOrder` +
   `ListOrderLines` agregados al backend (patrón PriceList).

**Batch 3 — Stretch si queda tiempo**:

7. Supplier detail `/compras/proveedores/[id]` + demerits sub-section.
8. Fuel logs list `/mantenimiento/combustible`.
9. Tax-book page `/administracion/facturacion/libro-iva` (requiere
   period picker — Tier C).

### Cierre esperado

- ≥ 6 clusters nuevos → total ≥ 47.
- Pattern `[id]` extendido a **ventas** (sales orders) + **calidad**
  (inspecciones) si se shipean — ampliando de 5 a 7 módulos.
- Phase 0 gates verdes.

## Candidatos para sesiones futuras (lookahead — NO 2.0.19)

| Orden | Tema | Notas |
|---:|---|---|
| 1 | **Write paths** en clusters read-only | Create/update en reclamos, importaciones, carchehi, calificaciones, notas. Audit + NATS + entity pickers. |
| 2 | **Módulos vacíos** con backend listo | `/compras/abastecimiento`, `/compras/comex`, `/rrhh/legajos/[id]` detail. |
| 3 | **Reports** (Phase 1 §Reports parity) | Libro IVA, mayor contable, tax-book. Eje separado de §UI parity. |
| 4 | **Seamless-day cutover test** | Phase 1 gating final. |
| 5 | **GetAsset / GetArticle endpoints** | Backend refactor menor — elimina la deuda de clusters #7/#8 de 2.0.18. |
| 6 | **Admin Tier B refactor** | Deuda de 2.0.17 — products/tools a handlers dedicados. |

## Trampas heredadas (mismas que 2.0.18)

- **Agents hallucinate**: verificar con grep antes de construir.
- **Nested git repos**: `apps/web/.git` es repo separado. Usar
  `git -C /home/enzo/rag-saldivia` para outer.
- **cwd persiste entre Bash calls**: `cd apps/web` + comandos → siguientes
  bash siguen ahí. Volver con rutas absolutas.
- **sqlc regen drift**: `make sqlc-erp` con v1.30.0 reescribe 14 archivos
  por diff de formatter. Editar generated code a mano (patrón
  GetPriceList de 2.0.18).
- **Pre-existing lint warnings**: file-upload.tsx tiene ~500 errores
  de react-hooks/immutability desde main. Ignorable.
- **apps/web registry**: agregar nueva ruta top-level en
  `src/lib/modules/registry.ts`, NO sólo crear la carpeta.

## Fuera de scope 2.0.19

- **Phase 2+** (chat agent, prompts jerárquicos, tree-RAG, ACL): sigue
  detrás de §UI parity en el orden top-down.
- **ADR 027 §UI parity row 1 tick**: cerrar "every XML-form has SDA
  equivalent" requiere cubrir ~4,500 forms. No en este ciclo.
- **W-009 file-upload** (bulk CSV/XLS bank import): sigue waived.

## Post-PR cierre ciclo

```bash
gh pr create --base main --head 2.0.19 --title "..." --body "..."
# Post-merge:
git checkout main && git pull origin main
git tag v2.0.19 && git push origin v2.0.19
gh release create v2.0.19 --title "..." --notes "..."
ssh sistemas@srv-ia-01 "cd /opt/saldivia/repo && git pull origin main"
```
