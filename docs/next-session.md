# Next session — 2.0.13: close reclamos gaps + second §UI parity cluster

**Goal**: ship 1 more Phase 1 §UI parity cluster + polish the reclamos
cluster to match Histrix feature depth. Expected close: 3-4 commits,
one PR, standard cycle cadence.

## Cierre ciclo 2.0.12 — completado 2026-04-19

- PR #159 squash-merged como `0905b673` en main.
- Tag `v2.0.12` pushed, GitHub release publicada.
- Workstation `srv-ia-01` sincronizada en `0905b673`.
- Todos los tests + build + type-check verdes.

Logros:

- **ADR 027 §Data migration ✅ fully ticked**. Migration `080` cerró
  STK_COSTOS → `erp_stock_cost_movements` (15,066 rows). Business-data
  gap: 1 → **0**. El item ADR 027 "Every legacy Histrix table …
  migrated or waived" quedó `[x]`.
- **Primer §UI parity cluster shipped**: `/administracion/reclamos`
  covers 3 XML-forms (`reclamos/reclamopagos*.xml`) consuming
  `erp_payment_complaints`. Endpoints GET/POST/PATCH bajo
  `/v1/erp/accounts/complaints`, scoped a `erp.accounts.read/write`,
  con NATS + strict audit.
- Nuevo doc `docs/parity/ui-parity.md` — living log de XML-form → SDA
  page coverage. First entry: el cluster reclamos.

## Final goal (ADR 026 — no se pierde de vista)

SDA reemplaza Histrix. El empleado abre SDA y:

1. UI moderna cubriendo **todo** lo que Histrix hacía (1:1 parity).
2. Chat donde el agente es su representante — cap parity chat ↔ UI.
3. Dashboard personal + rutinas personales.
4. Agentes background: mail, WhatsApp, tree-RAG con ACL.

## Estado post-2.0.12

| Dimensión | Antes | Después | Nota |
|---|---:|---:|---|
| Covered tables (live) | 126 | **127** | ~82% de filas Histrix |
| Business-data gap | 1 | **0** | Todo migrated or waived |
| Waivers activos | 5 | 5 | W-004/005/006/007/008 |
| SDA pages XML-parity tracked | 0 | **1** | reclamos |
| ADR 027 §Data migration items | 0/4 `[x]` | 1/4 `[x]` | 3 pendientes dependen de cutover live + archive read endpoint |
| ADR 027 §UI parity items | 0/4 `[x]` | 0/4 `[x]` | Primer cluster NO tickea el item global — necesita todas las XML-forms cubiertas |

## Plan de trabajo 2.0.13

Recomendación: dos ejes en una sola cycle — cerrar el cluster reclamos
al nivel Histrix + abrir un segundo cluster. Esto demuestra que la
iteración dentro de un cluster es barata (agrega depth sin más tablas
DB).

### Pre-work

```bash
git checkout -b 2.0.13 main
sed -i 's/Working:\*\* `2.0.12`/Working:\*\* `2.0.13`/' CLAUDE.md
git commit -am "chore: bump working branch 2.0.12 → 2.0.13

[resumen 2.0.12 + plan 2.0.13]

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

### Commit 1 — Reclamos: entity picker

**Scope:** reemplazar el input numérico `ctacod` por un picker modal
contra `/v1/erp/entities?type=supplier&search=<q>`. El backend ya
existe (services/erp/internal/handler/entities.go). Pattern conocido:

- Modal dialog con search input debounced + table de resultados
  (nombre + ctacod + CUIT).
- Al seleccionar: carga `entity_id` (UUID) + `entity_legacy_code`
  (ctacod) en el form — el backend ya acepta ambos en el mismo
  request body (ver handler `CreateComplaint`).
- Component candidate path: `apps/web/src/components/erp/entity-picker.tsx`
  (reusable — habrá muchos clusters que lo necesiten).

**Touched files (expected):**
- `apps/web/src/components/erp/entity-picker.tsx` (nuevo)
- `apps/web/src/app/(modules)/administracion/reclamos/page.tsx` (wire-up)
- `apps/web/src/lib/erp/queries.ts` (add `entitiesSearch` key)
- `apps/web/src/lib/erp/types.ts` (add `EntitySearchResult` if needed)

**Reuso futuro:** mismo componente va a servir para bcs_importacion
(cluster del siguiente commit) y cualquier cluster que necesite
seleccionar cliente/proveedor.

### Commit 2 — Reclamos: saldo aggregate per proveedor

**Scope:** incorporar el saldo total del proveedor junto a cada fila
del reclamo, replicando lo que `reclamopagos_ing.xml` hace con el
`sum(REG_MOVIMIENTOS.saldo_movimiento)` agrupado por `regcuenta_id`.

El endpoint `/v1/erp/accounts/balances` ya existe y devuelve
`{entity_name, entity_type, direction, open_balance}` por entidad.

**Opciones de wire-up (elegir antes de codear):**

1. **Cliente-side join** — fetch `balances` + `complaints` por
   separado, matchearlos en el render por `entity_legacy_code` o
   `entity_id`. Más simple, un ida y vuelta menos al backend. Si el
   set de proveedores es chico (y lo es, ~100s en total) el match
   es O(n·m) pero tolerable.
2. **Nuevo endpoint `/complaints/with-balance`** — `ListComplaints`
   LEFT JOIN de `erp_payment_complaints` con la agregación de
   `erp_account_movements.balance`. Más clean SQL, un single roundtrip,
   pero agrega superficie backend a mantener.

Decision: empezar por (1) — el volumen no justifica un endpoint nuevo
y el match client-side es 20 LOC.

**Touched files:**
- `apps/web/src/app/(modules)/administracion/reclamos/page.tsx`
  (agregar query a balances + render columna "Saldo")

### Commit 3 — Segundo cluster §UI parity: `bancos_local/bcs_importacion`

**Target data:** `erp_bank_imports` (migración `076`, 2.0.10, 91,959
rows live). Reconciliación de extractos bancarios importados desde
CSV/XLS vs REG_MOVIMIENTOS.

**Histrix XML-forms a cubrir (3):**

- `bancos_local/bcs_importacion_qry.xml` — vista principal (list +
  filters: account, processed state, date range).
- `bancos_local/bcsmovim_importacion_auto_ins.xml` — ingesta de archivo
  (bulk insert desde CSV). Puede quedar out-of-scope de este PR si
  requiere file-upload infra; alternativa: waiver con pointer a un
  futuro cluster de ingesta.
- `bancos_local/bcsmovim_importacion_auto_mov_ins.xml` — conciliación
  fila-por-fila. Es el 90% del valor — toggle `processed` + link a
  REG_MOVIMIENTOS match.

**Backend status:** `sda list ListBankImports` ya existe
(`services/erp/db/queries/treasury.sql`). **Hay que agregar:**

- `UpdateBankImportProcessed` — toggle `processed` flag + set
  `treasury_movement_id` if match elegido.
- (Opcional) `MatchBankImportToMovement` — transaction que marca
  ambos lados.

**Frontend:**

- `apps/web/src/app/(modules)/administracion/tesoreria/importaciones/page.tsx`
  (nuevo route — bajo `tesoreria` en vez de crear su propio módulo).
- Table con filters (account dropdown, date range, processed state
  tabs) + toggle fila.
- Registry entry: agregar `/administracion/tesoreria/importaciones` a
  `administracion` subnav o como sub-path de `/tesoreria`.

**Touched files (expected):**
- `services/erp/db/queries/treasury.sql` + hand-patch
- `services/erp/internal/service/treasury.go` — nuevo service methods
- `services/erp/internal/handler/treasury.go` — nuevo endpoints
- `services/erp/cmd/main.go` — si cambian routes
- `apps/web/src/app/(modules)/administracion/tesoreria/importaciones/page.tsx`
- `apps/web/src/lib/erp/queries.ts` + `types.ts`
- `apps/web/src/lib/modules/registry.ts`
- `docs/parity/ui-parity.md` — agregar cluster

### Cierre esperado

Post-2.0.13:

- Reclamos UI: feature-complete vs Histrix (picker + saldo).
- Segundo cluster §UI parity live (bcs_importacion) con 1-2 XML-forms
  cubiertas directamente; el tercer form (file-upload) waived o
  diferido.
- `ui-parity.md`: 2 clusters tracked, 4-5 XML-forms → SDA pages.
- Phase 0 gates verdes.

## Candidatos para clusters futuros (lookahead — NO esta sesión)

Ordenados por tamaño de data migrada + valor de negocio:

| Orden | Cluster | XML-forms | Data backing | Notas |
|---:|---|---:|---|---|
| 1 | **bcs_importacion** | 3 | erp_bank_imports (92 K) | Plan 2.0.13 |
| 2 | **carchehi** (check history) | 3 | erp_check_history (29 K) | Treasury |
| 3 | **reclamos + calificación** sub-view en proveedores | 1 | erp_entity_credit_ratings (136 K) | Embebe calificación dentro de proveedores/[id] |
| 4 | **evolutivo_costo + presup** (budget/chart) | 15 | erp_stock_cost_movements (15 K) | Big cluster, requiere chart lib |
| 5 | **invoice notes** | ~5 forms | erp_invoice_notes (73 K) | Sub-view en facturas |

## Trampas heredadas

- **Dialog asChild** — en este codebase el componente Dialog **NO**
  acepta `asChild` en DialogTrigger. Pattern correcto: `<Button
  onClick={() => setOpen(true)}>` fuera del Dialog, y el Dialog recibe
  `open={open} onOpenChange={(v) => !v && setOpen(false)}`. Ver
  `apps/web/src/app/(modules)/administracion/reclamos/page.tsx` como
  referencia — así lo usé en 2.0.12.
- **Textarea existe** — `apps/web/src/components/ui/textarea.tsx` ya
  está. No crear uno nuevo.
- **sqlc drift** — editar `.sql` + hand-patch `.sql.go`/`models.go`
  quirúrgicamente. Memoria `feedback_sqlc_version_drift`.
- **Phase 0 invariants** — cualquier nuevo endpoint escritura debe
  declarar capability + permission check + strict audit antes de
  commit + NATS event post-commit. Ver el shape de
  `CreateComplaint` en `services/erp/internal/service/current_accounts.go`.
- **Pre-existing lint noise** — `pkg/server/healthcheck.go` + test
  tienen 3 errcheck warnings que vienen de main. No son de este
  scope — ignorar a menos que se toque ese archivo.
- **Tailscale SSH re-auth** — workstation a veces pide re-auth en
  `https://login.tailscale.com/a/...`. No bloquea MySQL que va por
  WireGuard.

## Fuera de scope

- **ADR 027 §Data migration items restantes (3/4)** — dependen de
  cutover rehearsal live o del `erp_legacy_archive` read endpoint.
  Ambas son tareas ops / backend-go de sesión dedicada, no shipping
  §UI parity.
- **Phase 2+ (chat, prompts jerárquicos, tree-RAG, ACL)** —
  desbloqueado pero sigue detrás de Phase 1 §UI parity en el
  orden top-down.
- **File upload para bcs_importacion_auto_ins** — requiere infra
  de ingest/upload que no existe todavía. Waivable (o diferido a una
  sesión posterior con shape más grande).
- **ADR 027 §UI parity row 1 tick** — tickear "Every XML-form has SDA
  equivalent or waiver" solo va a pasar cuando estén todos los
  ~4,500 forms cubiertos o waived. Por ahora, **cada cluster agrega
  una fila a `ui-parity.md`** y no toca el checkbox global.

## Post-PR cierre ciclo

```bash
gh pr create --base main --head 2.0.13 --title "..." --body "..."
# Post-merge:
git checkout main && git pull origin main
git tag v2.0.13 && git push origin v2.0.13
gh release create v2.0.13 --title "..." --notes "..."
ssh sistemas@srv-ia-01 "cd /opt/saldivia/repo && git pull origin main"
```

Memoria `feedback_version_tagging.md`. Release body incluye:

- Resumen de los dos bloques (reclamos depth + bcs_importacion
  cluster).
- Estado post: N clusters en ui-parity.md.
- Link al PR + cualquier gap documentado.
