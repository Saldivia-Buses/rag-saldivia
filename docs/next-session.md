# Next session — 2.0.21: follow-up filters + remaining [id] + write paths scout

**Goal**: cerrar los filtros backend deferidos en 2.0.20 (las tres
detail pages nuevas filtran client-side, lo que no escala), sumar los
`[id]` restantes con pequeño backend lift, y empezar a explorar write
paths — el siguiente gran eje Phase 1 ahora que §UI parity read-only
se está agotando.

Target: **20 clusters**. Mix base:
- **~5 filter endpoints** (3 confirmados: reconciliations / cash-
  counts / entries; 2-3 TBD desde Explore 6 scan de detail pages
  con `.filter(client-side)` pattern).
- **~6 nuevos `[id]`** con backend lift (GetTool, GetAsset,
  GetChassisModel, GetCarroceriaModel, GetScorecard + 1-2 TBD
  desde Explore 5). Los 2 chassis son entidades distintas (tablas
  `erp_chassis_models` + `erp_carroceria_models`), no rename.
- **~3 direct-endpoint swaps** (GetArticle confirmado — backend
  ya existe, solo swap frontend list-cache → direct; 2 TBD desde
  Explore 6 barriendo `pageSize=500` patterns).
- **~4 write scouts** (entity note, credit rating, supplier
  demerit confirmados; 1-2 TBD desde Explore 7). Pattern ref:
  `CreateTenant` (auditor.Write + publisher.Notify).

Total esperado (7 Explores paralelos): 4 cerrados, 3 corriendo
para cerrar el mix antes de codear.

## Cierre 2.0.20 (6 clusters + backend lift)

| Ciclo | Clusters nuevos | Backend adds | Tag | PR |
|---|---:|---|---|---:|
| 2.0.20 | 6 | GetBankAccount, GetCashRegister, GetCostCenter | v2.0.20 | #167 |

Totales post-2.0.20: **54 clusters** en `ui-parity.md`, **124 `page.tsx`**
shipped, `[id]` en 9 módulos, 1 waiver (W-009).

Workstation `srv-ia-01` sincronizada en `08ac9de3`.

## Final goal (ADR 026 — no se pierde de vista)

SDA reemplaza Histrix. Empleado abre SDA y:

1. UI moderna cubriendo **todo** lo que Histrix hacía (1:1 parity).
2. Chat-agente como representante — cap parity chat ↔ UI.
3. Dashboard personal + rutinas personales.
4. Agentes background: mail, WhatsApp, tree-RAG con ACL.

## Estado post-2.0.20

- **Entity symmetry completa**: proveedor/cliente/empleado tienen
  todos `[id]/page.tsx` contra `GetEntity`. El patrón de "layer role-
  specific data sobre la ficha base" quedó probado: demeritos para
  supplier, asistencia para employee, contactos/notas para customer.
- **`[id]` pattern** vive en 9 módulos (se sumaron entities +
  treasury/CC). Las 3 pages nuevas de tesorería/contabilidad filtran
  related data client-side porque los list endpoints no aceptan los
  filtros. Es deuda concreta a cerrar.
- **Herramientas, chasis modelo, supplier scorecard, quality
  indicator**: siguen sin detail endpoint. Cada uno requiere `GetX`
  pequeño + frontend page — mismo patrón que 2.0.20 clusters 4-6.
- **Deuda explícita** heredada (sin abrir aún):
  - Admin Tier B refactor (2.0.17 — products/tools handler split).
  - `GetAsset` / `GetArticle` endpoints dedicados (2.0.18 clusters
    #7/#8 usan list cache).
  - Chasis modelo nomenclature: frontend dice "chasis-modelos",
    backend expone "carroceria-models" — resolver antes de shipear.

## Plan de trabajo 2.0.21

### Pre-work — DONE

```
c2e78981 docs: 2.0.21 plan — expand target a 12 clusters (después 20)
44267dd0 chore: bump working branch 2.0.20 → 2.0.21
```

Branch `2.0.21` cortado desde `main` (HEAD `7f25b8d3`). Linter mods
(CLAUDE.md dup block, .gitignore, .claude/settings.json) stasheados
como `linter mods pre-2.0.21` — pop al post-merge.

### Investigación (7 Explore agents en paralelo)

**4 cerrados**:

1. **Filter endpoints** ✅ — handlers confirmados en `services/erp/internal/handler/`:
   - `ListReconciliations` treasury.go:534, sqlc `treasury.sql.go:1148`.
   - `ListCashCounts` treasury.go:395, sqlc `treasury.sql.go:846`.
   - `ListEntries` accounting.go:214, sqlc `accounting.sql.go:843`.
   - Pattern `ListIncidents` usa cast condicional
     `($N::UUID IS NULL OR col = $N)`, no COALESCE.
   - Red flag: cost-center `[id]/page.tsx` **no tiene sección
     entries** — cluster #3 requiere agregar UI consumer también.

2. **[id] nuevos** ✅ — 5 targets mapeados:
   - `GetArticle` **YA EXISTE** (stock.go:42) — es solo swap frontend.
   - `GetTool` / `GetAsset` / `GetScorecard`: falta sqlc+handler+page.
   - Chasis: DOS tablas distintas `erp_chassis_models` (motor,
     tracción) y `erp_carroceria_models` (carrocería, peso ejes).
     Split en 2 clusters.

3. **Empty modules abastecimiento/comex** ❌ — **ya shipeados**.
   Frontend 293+260 líneas, backend + registry. Batch 3 original
   muerto; reemplazado por direct swaps + write scouts.

4. **Write scouts** ✅ — 3 candidatos con pattern reference
   (`CreateTenant` en `services/app/internal/core/platform/`):
   - POST entity note (A): endpoint NO existe; tabla
     `erp_entity_notes` existe; frontend detail proveedor
     ya tiene lugar obvio para textarea.
   - POST credit rating (B): endpoint NO existe; tabla
     `erp_credit_ratings` existe; requiere EntityPicker + enum.
   - POST reclamo pago (C): NO hay página frontend — DESCARTAR,
     L cost.

**3 corriendo**:

5. **Más `[id]`** (Explore 5) — barriendo warehouse / maintenance /
   manufacturing / hr / sales / quality / accounting para encontrar
   ≥ 7 candidatos backend-lift pattern.
6. **Filter + direct swaps** (Explore 6) — barriendo
   `apps/web/.../[id]/page.tsx` por `.filter(x => x.fk === id)` y
   por `pageSize=500`. Buscar 2-3 de cada.
7. **Write POSTs adicionales** (Explore 7) — scanning entity
   contacts / asset assignments / work order updates / odometer
   readings / price list lines / quotation lines. Buscar ≥ 5.

### Batches finales (20 clusters post-Explore × 7)

**Batch 1 — Filter endpoints + detail enrichment** (3 clusters):

1. `ListReconciliations?bank_account_id=` — handler+sqlc patch
   + cuenta-bancaria detail consume.
2. `ListCashCounts?cash_register_id=` — handler+sqlc patch + caja
   detail consume.
3. `ListEntries?cost_center_id=` — handler+sqlc patch **+ agregar
   sección "Asientos imputados" en centro-costo detail** (la UI
   consumer no existe aún).

**Batch 2 — Nuevos `[id]` con backend lift** (9 clusters):

4. **GetTool** — sqlc `GetTool` (falta) + handler + [id] page +
   historial via `erp_tool_movements`.
5. **GetAsset (maintenance)** — sqlc `GetMaintenanceAsset` (falta)
   + handler + [id] page + `erp_maintenance_plans`
   + `erp_work_orders` relacionados.
6. **GetChassisModel** — sqlc `GetChassisModel` (falta) + handler
   `{id}` + [id] page. Tabla `erp_chassis_models` (motor/tracción).
7. **GetCarroceriaModel** — sqlc `GetCarroceriaModel` (falta) +
   handler `{id}` + [id] page + `erp_carroceria_bom`. Tabla
   `erp_carroceria_models` (carrocería/peso ejes). NOTA: son DOS
   entidades distintas, no rename.
8. **GetScorecard** — sqlc `GetSupplierScorecard` (falta) + handler
   + [id] page + métricas históricas por proveedor.
9. **GetInspection (QC compras)** — sqlc ✓ existe; solo handler
   + [id] page + `erp_inspection_results`.
10. **GetUnit (manufacturing units)** — sqlc ✓ existe; solo handler
    + [id] page + `erp_unit_controls`.
11. **GetActionPlan (calidad)** — sqlc ✓ existe; solo handler
    + [id] page + `erp_action_tasks` + assignments.
12. **GetNC (quality non-conformity)** — sqlc ✓ existe; solo handler
    + [id] page + `erp_action_plans` + findings.

**Batch 3 — Direct-endpoint swap** (1 cluster):

13. **GetArticle direct** — backend `GET /v1/erp/stock/articles/{id}`
    YA EXISTE (stock.go:42). Solo swap frontend:
    `articulos/[id]/page.tsx` de `list.find(a => a.id === id)` a
    `useQuery(erpKeys.stockArticle(id))`. Cierra deuda 2.0.18 #8.
    Costo XS.

**Batch 4 — Write scouts** (7 clusters, pattern ref `CreateTenant`):

14. **POST entity note** — handler+service+auditor+publisher+frontend
    textarea en proveedor/cliente detail. Tabla `erp_entity_notes`
    existe.
15. **POST entity contact** — handler `AddContact` YA EXISTE
    (entities.go:297). Solo frontend: formulario compacto + dropdown
    type (phone/email/address/bank_account) en proveedor detail.
    Costo XS.
16. **POST inventory movement** — handler `CreateMovement` YA EXISTE
    (stock.go:200). Solo frontend: botón "+ Movimiento" en warehouse
    o articulos detail + modal con `movement_type` enum. Costo XS.
17. **POST credit rating** — handler+service+auditor+publisher+frontend
    modal con EntityPicker + enum A|B|C|X + reference + fecha.
18. **POST supplier demerit** — handler+service+auditor+publisher+
    frontend desde scorecard contexto. Tabla `erp_supplier_demerits`.
19. **POST invoice note** — handler+service+auditor+publisher+frontend
    sección "Observaciones" en invoice detail. Tabla
    `erp_invoice_notes`.
20. **POST work_order_part** — handler+service+auditor+publisher+
    frontend sección "Partes" en work order detail. Tabla
    `erp_work_order_parts`.

### Cierre esperado

- 20 clusters nuevos → total ≥ 74.
- Los 3 filter endpoints de 2.0.20 cerrados (detail pages sin
  filter client-side).
- Deuda 2.0.18 (#7/#8) cerrada: #7 via GetAsset cluster completo,
  #8 via GetArticle direct swap.
- 6 nuevos `[id]` en calidad/manufacturing/compras (inspection,
  unit, action plan, NC, scorecard, tool).
- 7 write paths shipeados — el expansion post-2.0.11 arranca acá.
- 2 "frontend-only" write clusters (contact, movement) validan
  handlers write existentes sin frontend consumer.
- Phase 0 gates verdes.

### Pool NO shipeado (remaining candidates para 2.0.22+)

- **GetWarehouse** — requiere sqlc `GetWarehouse` creation (falta).
- **GetAudit (calidad)** — requiere sqlc `GetAudit` creation.
- **POST payment_complaint** — requiere crear módulo
  `/reclamos-pagos` entero. Large.
- **POST accounting_entry_line** — requiere lógica doble-entrada
  contable. Large.
- **Direct swap centros-costo parent** — niche hierarchical lookup.

## Candidatos para sesiones futuras (lookahead — NO 2.0.21)

| Orden | Tema | Notas |
|---:|---|---|
| 1 | **Write paths masivos** en clusters read-only | Expansión post-scout: create/update en reclamos, importaciones, calificaciones. Entity pickers, validación, NATS + audit estandarizado. |
| 2 | **Reports** (Phase 1 §Reports parity) | Libro IVA, mayor contable, tax-book. Eje separado de §UI parity. |
| 3 | **Seamless-day cutover test** | Phase 1 gating final. |
| 4 | **Admin Tier B refactor** | Deuda de 2.0.17 — products/tools a handlers dedicados. |

## Trampas heredadas (mismas que 2.0.18/19/20)

- **Agents hallucinate**: verificar con grep antes de construir.
  Pre-check cada endpoint claim del agente. El reporte de Explore
  de 2.0.20 fue preciso — el pattern (citar ruta chi + handler line
  + list page path) funcionó y hay que replicarlo.
- **sqlc regen drift**: `make sqlc-erp` con v1.30.0 reescribe 14
  archivos por diff de formatter. Editar generated code a mano —
  2.0.20 validó otra vez el pattern con GetBankAccount/Cash/CC.
- **Mock interfaces must track**: cada nuevo método en la interfaz
  del handler rompe `accounting_test.go` / `treasury_test.go`. En
  2.0.20 se detectó post-commit (build falla, no test falla) — hay
  que correr `make test` antes del PR.
- **Pre-existing lint warnings**: file-upload.tsx tiene ~500 errores
  desde main. Ignorable.
- **apps/web registry**: agregar nueva ruta top-level en
  `src/lib/modules/registry.ts`, NO sólo crear la carpeta.
- **Pre-existing linter-modified files** (CLAUDE.md, .gitignore,
  .claude/settings.json): se tocan entre sesiones fuera del commit.
  Stash antes de `git checkout main` post-merge. Ya surgió en 2.0.20.
- **Branch protection requires CI**: `gh pr merge 2.0.N` sin `--auto`
  falla si status checks no pasaron. Usar `--auto --squash
  --delete-branch` para queue, o esperar que CI termine.
- **Nested git repos**: `apps/web/.git` es repo separado. Usar
  `git -C /home/enzo/rag-saldivia` para outer.
- **cwd persiste entre Bash calls**: volver con rutas absolutas.

## Fuera de scope 2.0.21

- **Phase 2+** (chat agent, prompts jerárquicos, tree-RAG, ACL).
- **ADR 027 §UI parity row 1 tick**: requiere cubrir ~4,500 forms.
- **W-009 file-upload** (bulk CSV/XLS bank import): sigue waived.
- **Write paths masivos**: sólo el scout de 1-2 clusters. La
  expansión full vuelve cuando §UI parity read-only esté agotada.

## Post-PR cierre ciclo

```bash
gh pr create --base main --head 2.0.21 --title "..." --body "..."
gh pr merge 2.0.21 --squash --auto --delete-branch   # auto-merge on CI green
# Post-merge:
git stash push -m "linter mods" -- CLAUDE.md .gitignore .claude/settings.json
git checkout main && git pull origin main
git stash pop
git tag v2.0.21 && git push origin v2.0.21
gh release create v2.0.21 --title "..." --notes "..."
ssh sistemas@srv-ia-01 "cd /opt/saldivia/repo && git pull origin main"
```
