# Next session — Phase 1 §Data migration: first migrator from the prioritized gap

Arrancás en `main` con `cb2f444c` ya desplegado en la workstation (el
drift gate cerró verde empíricamente al final de la sesión 2.0.7).
Commit tope: `cb2f444c 2.0.7 — Phase 0 closed (5/5) + Phase 1
data-migration roadmap + prod hardening (#154)`.

## Final goal (ADR 026 — no se pierde de vista)

SDA reemplaza Histrix. El empleado abre SDA y:

1. Tiene UI moderna cubriendo **todo** lo que Histrix hacía (1:1 parity,
   mejor UX).
2. Tiene chat donde el agente es su representante — cap parity chat ↔ UI.
3. Arma su dashboard personal (no hay dashboard global).
4. Arma sus rutinas personales.
5. Detrás, agentes hoardean data: mail ingest, WhatsApp interno,
   tree-RAG con ACL por colección.

La vara: `.intranet-scrape/` — 675 tablas + ~4,500 XML-forms.

## Estado Phase 0 (ADR 027)

| # | Item | Estado |
|---|---|---|
| 1 | Migration integrity | ✅ shipped (2.0.6) |
| 2 | No-op migrators | ✅ shipped (2.0.6) |
| 3 | Orphan tables | ✅ shipped (2.0.6) |
| 4 | Tool capabilities | ✅ shipped (2.0.7) |
| 5 | Workstation drift | ✅ shipped (2.0.7) — empíricamente `make check-prod-drift` = verde |

**Phase 0 cerrado 5/5. Phase 2+ desbloqueado. Pero la prioridad top-down
sigue siendo Phase 1 (parity before polish).**

## Estado Phase 1 §Data migration (prioridad para esta sesión)

Del diff realizado en 2.0.7 con row counts live:

| Segment | Tablas | Filas |
|---|---:|---:|
| Histrix total | 675 | 18.9 M |
| Cubiertas (migrator/reader registrado) | 100 | ~8.85 M (47 %) |
| Waiver masivo (W-004 HTX*, W-005 `*_OLD`, W-006 0-filas) | 261 | 3.9 M infra/deads |
| **Gap real a migrar/waiver uno-a-uno** | **314** | **6.18 M (33 %)** |

**Pareto**: top 10 tablas uncovered = 90.6 % del row volume. Top 20 =
95.8 %. Los 294 restantes tailean hacia miles-de-filas — muchos aún
importan por contenido (nombres de entidades, HR, etc.) pero en
volumen son ruido.

Ranking completo vivo en `docs/parity/data-migration.md`. Regenerable
con la receta al final de ese archivo (SSH+mysql via docker+sshpass —
memoria en `reference_histrix_access.md`).

## Tarea principal — primer migrator de la lista Pareto

**Skills primarios:** `htx-parity` + `migration-health` + `database`.

### Opción A (recomendada) — `STK_ARTICULO_PROCESO_HIST_DETALLE` (2.6 M filas, 42.7 % del gap)

Solo esta tabla cubre 42.7 % del row volume uncovered. Por eso entra primera.

**Pre-work obligatorio (no saltar):**

1. Leer `.intranet-scrape/xml-forms/stock/articulos*.xml` y cualquier
   form que mencione `proceso_hist` o `historia` del artículo —
   contrato de parity.
2. Inspeccionar shape de la tabla en Histrix vivo:
   ```sql
   DESCRIBE STK_ARTICULO_PROCESO_HIST_DETALLE;
   SELECT * FROM STK_ARTICULO_PROCESO_HIST_DETALLE LIMIT 5;
   ```
   (via el pattern docker+sshpass de `reference_histrix_access.md`).
3. Buscar si ya existe tabla SDA target — `grep -r "stk_articulo_proceso\|article_process_hist\|erp_process" db/tenant/migrations/`.
   Probablemente no; hay que crearla.
4. Revisar la estructura padre en migración tree:
   `STK_ARTICULOS` ya está migrada → `erp_articles`. El detalle
   histórico es child-of. La FK más natural es `article_id UUID` con
   `resolve_via erp_legacy_mapping`.

**Deliverables:**

1. Nueva migration `db/tenant/migrations/0NN_erp_article_process_history.up.sql` + down pair. Shape mínimo:
   - `id UUID PK DEFAULT gen_random_uuid()`
   - `tenant_id TEXT NOT NULL` (ADR 022 silo — se dejará al seed; revisar convención en tablas cercanas)
   - FK al artículo (`article_id` → `erp_articles(id)`)
   - columnas del detalle (estimar a partir del schema Histrix)
   - `created_at/updated_at TIMESTAMPTZ`
2. sqlc queries mínimas (al menos una read — Phase 0 "no dead-end writes" sigue vigente).
3. Nuevo reader en `tools/cli/internal/legacy/stock_extended.go`:
   - PKColumn (AI int), Columns list, filtros si aplica.
4. Nuevo migrator en `tools/cli/internal/migration/migrators_*.go` —
   probablemente reutiliza `GenericMigrator` + transformFn que resuelve
   `article_id` via `mapper.Resolve(...)`.
5. Registro en orchestrator (con dependency después de `STK_ARTICULOS`).
6. Test dry-run contra Histrix vivo con `--only-table
   STK_ARTICULO_PROCESO_HIST_DETALLE --limit 100 --prod` —
   evidencia de que el Transform no devuelve nil universalmente (zero
   `rows_written=0` completions) y que el invariant
   `rows_read = rows_written + rows_skipped + rows_duplicate` cierra.

### Opción B (warmup) — W-001 fix: REMITOINT migrator + re-point REMDETAL

Más chica (5,125 filas) pero cierra un waiver abierto. Buena para
"primera vez escribiendo un migrator" si querés bajar riesgo antes
de atacar 2.6 M filas.

`docs/parity/waivers.md §W-001` tiene el root cause escrito:
`REMDETAL.idRemito` apunta a `REMITOINT`, no `REMITO`. Hay que:

1. Escribir `NewInternalDeliveryNoteReader` en
   `tools/cli/internal/legacy/invoicing.go` apuntando a `REMITOINT`.
2. `BuildRemitoIntIndex` en la orchestrator hook file.
3. Repointar `NewDeliveryNoteLineMigrator` al índice nuevo.
4. Strike-through W-001 en `waivers.md` + agregar evidencia del dry-run.

Shippable en 1-2 horas. Ideal si querés warm-up + Opción A en sesión
siguiente.

## Trampas conocidas

- **Phase 0 invariants siguen siendo la religión.** Cualquier
  migrator nuevo DEBE cumplir
  `rows_read = rows_written + rows_skipped + rows_duplicate` **y**
  NO completar con `rows_written=0` sobre `rows_read>0`. El skill
  `migration-health` tiene las queries canónicas — corrérlas post-dry-run.
- **FK resolution silenciosa.** El pattern `mapper.ResolveOptional` es
  la fuente #1 de ghost rows: cuando devuelve `nil, nil` por FK no
  encontrada, el migrator a veces emite `return nil, nil` que skippea
  la fila sin incrementar `rows_skipped`. Auditar el Transform nuevo
  contra `ghostrow_test.go`.
- **Histrix access requiere VPN activa en Windows** (como el 2026-04-18).
  SSH sí llega desde WSL (port 22 abierto), MySQL NO (bind-localhost).
  El pattern de docker+sshpass está en
  `~/.claude/projects/-home-enzo-rag-saldivia/memory/reference_histrix_access.md`.
  Las creds están en `c:/Users/enzo/Downloads/Copia de DATOS SISTEMAS -
  HARD.csv` + `reference_db_saldivia.md`.
- **Numeración de migration**. Leer el último archivo en
  `db/tenant/migrations/` antes de elegir número. NO reutilizar números
  aunque parezcan libres.
- **No regenerar sqlc masivo** (memoria: `feedback_sqlc_version_drift`).
  Editar `models.go` quirúrgicamente si necesitás registrar un struct
  nuevo; la CI tiene un check que diff-ea.
- **No escribir `tenant_id` hardcoded**. ADR 022 silo: el valor viene
  del seed por contenedor. Ver cómo lo hace `erp_articles` en
  `db/tenant/migrations/018_erp_stock.up.sql`.

## Tarea secundaria (si cerraste la principal con tiempo)

**Segundo migrator** — `FICHADAS` (1.5 M filas, 23.7 % del gap).

`FICHADADIA` ya está cubierto (es el daily-rollup). `FICHADAS` es el
raw-event stream (cada fichada individual). Skill: `htx-parity` +
`migration-health`. Shape probable: parent HR existente + child rows
con timestamp + tipo + estado.

No atacar si la principal no cerró — separar commit/PR por migrator
(una tabla por PR, más reviewable).

## Fuera de scope

- **Phase 1 §UI parity** — sub-order de ADR 027 dice "Data migration
  → UI parity". Las pages esperan hasta que haya data atrás.
- **Phase 2+ (chat, prompts jerárquicos, tree-RAG, ACL)** —
  desbloqueado pero no es prioridad top-down mientras quede Phase 1
  gap abierto.
- **Merge freeze / cutover runbook** — Phase 1 §Cutover readiness vive
  más adelante; no mezclar.

## Cierre esperado

- **Mínimo** (Opción B, warmup): W-001 cerrado — REMITOINT migrator
  shipped + REMDETAL re-pointed + 5,125 filas recuperadas + strike-
  through en `waivers.md`.
- **Ideal** (Opción A): primer migrator de la lista Pareto shipped —
  `STK_ARTICULO_PROCESO_HIST_DETALLE` completo, 2.6 M filas migradas,
  Phase 0 invariants verde post-run, tick parcial en Phase 1 §Data
  migration row 1.
- **Stretch**: lo anterior + `FICHADAS` (1.5 M filas). Con 2 tablas
  de top-10 cerradas en una sesión, el gap baja de 6.18 M a ~2.12 M
  (65 % de la parte uncovered-with-rows cerrada en una sesión).

Post-PR de cualquiera de los dos: `gh pr create --base main --head
2.0.8` y tras merge, `git pull` en la workstation para volver a
cerrar drift.

## Working branch

Antes de arrancar código, cortar `2.0.8` desde `main` y bumpear
`CLAUDE.md` ("Working: 2.0.7" → "Working: 2.0.8") con el commit
`chore: bump working branch 2.0.7 → 2.0.8` — mismo pattern que
`9277bbd8`.
