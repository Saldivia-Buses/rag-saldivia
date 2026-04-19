# Next session — 2.0.20: entity-symmetry + remaining [id] wave

**Goal**: seguir el ritmo de 2.0.18/19. El pattern `[id]/page.tsx` vive
en 8 módulos y cubrió proveedor en 2.0.19 — ahora toca cerrar la
simetría del lado entities (cliente, empleado) + algunos detail routes
que quedaron sin consumir después del agotamiento de Tier A/B.

Target realista: **6-10 clusters**. Mix esperado:
- 2-3 entity detail pages (cliente, empleado/legajo, vehículo taller).
- 3-5 `[id]` en endpoints ya mounted que aún no tienen página.
- 1-2 Tier A residuales (si quedan listas sin page).

## Cierre 2.0.19 (7 clusters)

| Ciclo | Clusters nuevos | Tag | PR |
|---|---:|---|---:|
| 2.0.19 | 7 | v2.0.19 | #166 |

Totales post-2.0.19: **48 clusters** tracked en `ui-parity.md`,
**118 `page.tsx`** shipped, 1 waiver (W-009).

Workstation `srv-ia-01` sincronizada en `767c41c3`.

## Final goal (ADR 026 — no se pierde de vista)

SDA reemplaza Histrix. Empleado abre SDA y:

1. UI moderna cubriendo **todo** lo que Histrix hacía (1:1 parity).
2. Chat-agente como representante — cap parity chat ↔ UI.
3. Dashboard personal + rutinas personales.
4. Agentes background: mail, WhatsApp, tree-RAG con ACL.

## Estado post-2.0.19

- **`[id]` pattern** vive en 8 módulos: tesorería, compras, facturación,
  mantenimiento, almacén, ventas, calidad, producción. Pendiente
  entity-symmetry (cliente/empleado) + algún detail de catálogos.
- **Backend surface**: sigue habiendo detail endpoints mounted sin
  página (p.ej. `GetEntity` funciona para cliente y empleado — ya se
  usó para proveedor en 2.0.19). Revisar candidatos concretos al
  arrancar con agente Explore.
- **Deuda explícita** heredada:
  - Admin Tier B refactor (2.0.17 — products/tools handler split).
  - `GetAsset` / `GetArticle` endpoints dedicados (2.0.18 clusters
    #7/#8 usan list cache).

## Plan de trabajo 2.0.20

### Pre-work

```bash
git checkout -b 2.0.20 main
sed -i 's/Working:\*\* `2.0.19`/Working:\*\* `2.0.20`/' CLAUDE.md
git commit -am "chore: bump working branch 2.0.19 → 2.0.20"
```

### Investigación inicial (agente Explore)

Verificar antes de shipear — patrón de 2.0.18/19 funcionó:

1. **Cliente detail**: `/v1/erp/entities/{id}` ya mounted (consumido
   en 2.0.19 para proveedor). ¿Dónde vive la list de clientes?
   (posibles: `/administracion/clientes`, `/ventas/clientes`, o
   embedded en otra page). Confirmar y crear `[id]` ahí.
2. **Empleado/legajo detail**: `/rrhh/legajos/page.tsx` existe
   (revisar shape actual). Detail muestra entity data + campos HR
   (cargo, fecha ingreso, legajo_number, attendance reciente).
3. **Vehículo taller detail**: `/mantenimiento/taller/vehiculos`
   existe. ¿Hay `GetCustomerVehicle` mounted? Si sí, detail con
   vehicle data + VehicleIncidents filtrado por vehicle_id.
4. **Catálogos detail** (si hay endpoints):
   - Herramientas: `GET /admin/tools/{id}`? Historial de uso.
   - Chasis modelo: detail con unidades relacionadas.
   - Cost center: detail con entries filtrados.
5. **Bank account / cash register detail**: `/administracion/tesoreria/cuentas-bancarias` y `/administracion/tesoreria/cajas`
   tienen lists. ¿Get endpoints mounted? Detail show metadata + saldo
   + movimientos recientes.
6. **Quality indicators detail**: `/calidad/indicadores` lista KPIs.
   Detail podría mostrar histórico del indicador por período.

### Batches propuestos

**Batch 1 — Entity symmetry** (2-3 clusters):

1. Cliente detail — simétrico a proveedor de 2.0.19. Contactos +
   notas + historial de ventas/facturas.
2. Empleado/legajo detail — entity + campos HR + attendance reciente.
3. Vehículo taller detail — vehicle data + incident history.

**Batch 2 — Detail routes con backend listo** (2-4 clusters):

4. Bank account detail — saldo + movimientos filtrados.
5. Cash register detail — balance + cash counts.
6. Cost center detail — entries filtrados por CC.
7. Quality indicator detail — time-series por período.

**Batch 3 — Stretch** (si queda tiempo):

8. Herramientas detail — usage history.
9. Chasis modelo detail — unidades relacionadas.
10. Supplier scorecard detail — métricas históricas.

### Cierre esperado

- ≥ 6 clusters nuevos → total ≥ 54.
- Entity detail symmetry cerrada (proveedor + cliente + empleado).
- Phase 0 gates verdes.

## Candidatos para sesiones futuras (lookahead — NO 2.0.20)

| Orden | Tema | Notas |
|---:|---|---|
| 1 | **Write paths** en clusters read-only | Create/update en reclamos, importaciones, calificaciones, notas. Audit + NATS + entity pickers. Primer gate hacia seamless-day. |
| 2 | **Reports** (Phase 1 §Reports parity) | Libro IVA, mayor contable, tax-book. Eje separado de §UI parity. |
| 3 | **Empty modules con backend listo** | `/compras/abastecimiento`, `/compras/comex`. |
| 4 | **Seamless-day cutover test** | Phase 1 gating final. |
| 5 | **GetAsset / GetArticle direct endpoints** | Backend refactor menor — elimina la deuda de clusters 2.0.18 #7/#8. |
| 6 | **Admin Tier B refactor** | Deuda de 2.0.17 — products/tools a handlers dedicados. |

## Trampas heredadas (mismas que 2.0.18/19)

- **Agents hallucinate**: verificar con grep antes de construir.
  Pre-check cada endpoint claim del agente.
- **Nested git repos**: `apps/web/.git` es repo separado. Usar
  `git -C /home/enzo/rag-saldivia` para outer.
- **cwd persiste entre Bash calls**: volver con rutas absolutas.
- **sqlc regen drift**: `make sqlc-erp` con v1.30.0 reescribe 14
  archivos por diff de formatter. Editar generated code a mano
  (patrón GetPriceList de 2.0.18, GetOrder de 2.0.19).
- **Pre-existing lint warnings**: file-upload.tsx tiene ~500 errores
  desde main. Ignorable.
- **apps/web registry**: agregar nueva ruta top-level en
  `src/lib/modules/registry.ts`, NO sólo crear la carpeta.
- **Page may query wrong endpoint**: cluster #3 de 2.0.19 descubrió
  que `/calidad/inspecciones/page.tsx` consumía audits en vez de
  inspections. Abrir siempre el parent list page antes de asumir que
  la ruta/endpoint match.
- **Branch protection requires CI**: `gh pr merge 2.0.N` sin `--auto`
  falla si status checks no pasaron. Usar `--auto --squash
  --delete-branch` para queue, o esperar que CI termine.

## Fuera de scope 2.0.20

- **Phase 2+** (chat agent, prompts jerárquicos, tree-RAG, ACL).
- **ADR 027 §UI parity row 1 tick**: requiere cubrir ~4,500 forms.
- **W-009 file-upload** (bulk CSV/XLS bank import): sigue waived.
- **Write paths**: sigue deferido. Phase 1 §Data migration quedó
  completa en 2.0.11; write paths son el próximo eje Phase 1 después
  que §UI parity read-only se agote.

## Post-PR cierre ciclo

```bash
gh pr create --base main --head 2.0.20 --title "..." --body "..."
gh pr merge 2.0.20 --squash --auto --delete-branch   # auto-merge on CI green
# Post-merge:
git checkout main && git pull origin main
git tag v2.0.20 && git push origin v2.0.20
gh release create v2.0.20 --title "..." --notes "..."
ssh sistemas@srv-ia-01 "cd /opt/saldivia/repo && git pull origin main"
```
