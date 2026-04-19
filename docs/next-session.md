# Next session — 2.0.21: follow-up filters + remaining [id] + write paths scout

**Goal**: cerrar los filtros backend deferidos en 2.0.20 (las tres
detail pages nuevas filtran client-side, lo que no escala), sumar los
`[id]` restantes con pequeño backend lift, y empezar a explorar write
paths — el siguiente gran eje Phase 1 ahora que §UI parity read-only
se está agotando.

Target realista: **6-9 clusters**. Mix esperado:
- 2-3 filter endpoints (reconciliations by bank_account, cash-counts
  by cash_register, entries by cost_center) + consumidor frontend
  enriquecido.
- 2-3 nuevos `[id]` (herramienta, supplier scorecard, chasis modelo
  con nomenclature-fix) con backend lift tipo GetPriceList.
- 1-2 scouts de write path (create/update en un cluster pequeño
  — ej. notas de entidad, rating de crédito) con audit + NATS.

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

### Pre-work

```bash
git checkout -b 2.0.21 main
sed -i 's/Working:\*\* `2.0.20`/Working:\*\* `2.0.21`/' CLAUDE.md
git commit -am "chore: bump working branch 2.0.20 → 2.0.21"
```

### Investigación inicial (agente Explore)

Verificar antes de shipear:

1. **Filter endpoints**: para cada uno de los 3 list endpoints
   (reconciliations, cash-counts, journal-entries), confirmar dónde
   vive el query/handler, si ya hay filter params por query string,
   y cuál sería el minimal patch (similar al `vehicle_id` filter
   que ya existe en ListIncidents).
2. **GetTool / GetChassisModel / GetScorecard**: confirmar la tabla
   (sólo lista o hay schema para detail con related rows), y dónde
   vive la list page de cada uno en `apps/web/src/app/`.
3. **Write path scout** — elegir UN cluster pequeño que ya tenga
   list + detail y añada write. Candidatos:
   - Entity credit rating (`/compras/calificacion-proveedores`):
     POST nuevo rating. Form simple (rating + referencia + fecha).
   - Entity notes en proveedor/cliente detail: POST nota desde
     la ficha. Requiere text area + botón. El endpoint
     `POST /v1/erp/entities/{id}/notes` ya existe.
   - Reclamos pagos: ya hay list + detail embebido. Add create.

### Batches propuestos

**Batch 1 — Filter endpoints + detail enrichment** (2-3 clusters):

1. `ListReconciliations?bank_account_id=` — consumir en
   cuenta-bancaria detail para reemplazar filter client-side.
2. `ListCashCounts?cash_register_id=` — consumir en caja detail.
3. `ListEntries?cost_center_id=` — consumir en centro-costo
   detail para mostrar asientos imputados al CC.

**Batch 2 — Nuevos `[id]` con backend lift** (2-3 clusters):

4. Herramienta detail — GetTool + historial de uso (si hay
   movement table relacionada).
5. Supplier scorecard detail — GetScorecard + métricas históricas
   del proveedor.
6. Chasis modelo detail — primero resolver nomenclature (renombrar
   uno de los dos para que match), luego GetChassisModel + BOM.

**Batch 3 — Write path scout** (1-2 clusters):

7. Add note from entity detail (proveedor o cliente) — textarea +
   POST + audit + NATS event.
8. Create credit rating desde calificacion-proveedores — form
   simple + POST + audit.

### Cierre esperado

- ≥ 6 clusters nuevos → total ≥ 60.
- Los 3 filter endpoints de 2.0.20 cerrados (detail pages sin
  filter client-side).
- Primer write path post-2.0.11 shipeado (aunque sea uno chico).
- Phase 0 gates verdes.

## Candidatos para sesiones futuras (lookahead — NO 2.0.21)

| Orden | Tema | Notas |
|---:|---|---|
| 1 | **Write paths masivos** en clusters read-only | Expansión post-scout: create/update en reclamos, importaciones, calificaciones. Entity pickers, validación, NATS + audit estandarizado. |
| 2 | **Reports** (Phase 1 §Reports parity) | Libro IVA, mayor contable, tax-book. Eje separado de §UI parity. |
| 3 | **Empty modules con backend listo** | `/compras/abastecimiento`, `/compras/comex`. |
| 4 | **Seamless-day cutover test** | Phase 1 gating final. |
| 5 | **Admin Tier B refactor** | Deuda de 2.0.17 — products/tools a handlers dedicados. |
| 6 | **GetAsset / GetArticle direct endpoints** | Backend refactor menor — elimina deuda de clusters 2.0.18 #7/#8. |

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
