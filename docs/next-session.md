# Next session — Phase 1 §Data migration: Pareto #4 (HERRAMIENTAS cluster 237 K)

La sesión anterior (2.0.10) cerró Pareto #3 en un PR:

- **CTBREGIS** (604,579 filas live — el scrape estimaba 572,076) migrada
  vía `NewAccountingRegisterMigrator` → nueva tabla
  `erp_accounting_registers` (migración `071`). `account_id` resuelto vía
  el code-index `accounting/erp_accounts/code` (91.2 % match — 551,110
  filas; 8.8 % orphan mantiene `account_id NULL` y `account_code`
  preservado). `reg_date` `NULL` para 8 filas zero-date. No es waiver:
  la xml-form scrape confirmó 59 referencias vivas en libro_diario,
  proveedores_loc, clientes_local, ordenpago, iva, bancos_local,
  anulaciones, estadisticas — y `CTB_DETALLES.ctbregis_id` sigue siendo
  FK viva del flow nuevo hacia el log legacy.

Post-2.0.10 el gap Phase 1 §Data migration quedó en **307 tablas /
≈ 1.47 M filas** (7.7 % del total Histrix). Las 107 tablas cubiertas son
ya **72 %** del volumen total.

Arrancás con PR de 2.0.10 **mergeable** al cerrar la sesión. Si no se
mergea solo, cerrar el ciclo (ver §"Pre-work si 2.0.10 sigue abierto"
abajo).

## Final goal (ADR 026 — no se pierde de vista)

SDA reemplaza Histrix. El empleado abre SDA y:

1. Tiene UI moderna cubriendo **todo** lo que Histrix hacía (1:1 parity,
   mejor UX).
2. Tiene chat donde el agente es su representante — cap parity chat ↔ UI.
3. Arma su dashboard personal (no hay dashboard global).
4. Arma sus rutinas personales.
5. Detrás, agentes hoardean data: mail ingest, WhatsApp interno,
   tree-RAG con ACL por colección.

## Estado Phase 1 §Data migration (post-2.0.10)

| Segment | Tablas | Filas |
|---|---:|---:|
| Histrix total | 675 | 18.94 M |
| Cubiertas | 107 | ≈ 13.56 M (**72 %**) |
| Waiver masivo (W-004/005/006) | 261 | 3.91 M |
| **Gap remaining** | **307** | **≈ 1.47 M (7.7 %)** |

**Top uncovered post-2.0.10** (row volume):

| Rank | Tabla | Rows (approx) | Cum % del gap |
|---:|---|---:|---:|
| 1 | **HERRAMIENTAS** | 225,355 | **15.4 %** |
| 2 | STKINSPR | 180,412 | 27.7 % |
| 3 | PRODUCTO_ATRIB_VALORES | 153,835 | 38.2 % |
| 4 | PROD_CONTROL_HOMOLOG | 105,683 | 45.4 % |
| 5 | STK_COSTO_HIST | 95,217 | 51.9 % |
| 6 | BCS_IMPORTACION | 84,492 | 57.6 % |
| 7 | EGX300EPE | 79,040 | 63.0 % |
| 8 | REG_MOVIMIENTO_OBS | 72,737 | 67.9 % |
| 9 | REG_CUENTA_CALIFICACION | 58,960 | 71.9 % |
| 10 | TEL_LOG | 34,885 | 74.3 % |

Reproducer completo al final de `docs/parity/data-migration.md`.

## Pre-work si 2.0.10 sigue abierto

Antes de cortar 2.0.11:

```bash
gh pr checks <NRO>          # tu PR 2.0.10
gh pr view <NRO> --json mergeable,state

# Post-merge en main:
git checkout main && git pull origin main
git tag v2.0.10 && git push origin v2.0.10
gh release create v2.0.10 --title "..." --notes-file <...>

# Workstation drift (srv-ia-01):
ssh sistemas@srv-ia-01 "cd /opt/saldivia/repo && git pull origin main"
# Verificar: git rev-parse origin/main == git rev-parse HEAD en workstation.
```

Memoria: `feedback_version_tagging.md`. Incluir en release body la
sección 2.0.10 (Pareto #3 CTBREGIS) con métricas.

## Tarea principal — Pareto #4: HERRAMIENTAS + HERRMOVS (~237 K filas)

**Skills primarios:** `htx-parity` + `migration-health` + `database` +
`decisions` (nuevo dominio) + `backend-go`.

### Contexto de negocio

Cluster de herramientas de taller (workshop tools):
- `HERRAMIENTAS` (225 K filas) — catálogo de herramientas (items,
  características, estado).
- `HERRMOVS` (11 K filas) — movimientos (préstamo/devolución, asignación
  a empleado/coche, historial).

El cluster es un **dominio SDA nuevo** — no existe `erp_tools_*`
todavía. La pregunta clave es si se crea:
- (A) **`erp_tools` + `erp_tool_movements`** — dominio dedicado.
- (B) Se fuerza en `erp_stock` como categoría especial — probablemente
  peor: "herramienta prestada a coche 134 por 3 días" no encaja en
  stock movements.

Requiere **ADR corto** antes del migrator (al estilo ADR 023/024 —
nuevo dominio lateral). `decisions` skill.

### Arranque recomendado

1. **XML-form inventory**:
   ```bash
   grep -rlE "HERRAMIENTAS|HERRMOVS" .intranet-scrape/xml-forms/ | \
     xargs -n1 dirname | sort -u
   ```
   Esperable: `taller/`, `herramientas/`, `coches/`. Leer
   `herramientas*.xml` para ver qué dimensiones soporta la UI Histrix.

2. **DB shape en Histrix live**:
   ```sql
   DESCRIBE HERRAMIENTAS;
   DESCRIBE HERRMOVS;
   SELECT COUNT(*) FROM HERRAMIENTAS;
   SELECT COUNT(*) FROM HERRMOVS;
   -- FK discovery
   SELECT DISTINCT tipomov FROM HERRMOVS LIMIT 20;
   -- orphan check
   SELECT COUNT(*) FROM HERRMOVS m LEFT JOIN HERRAMIENTAS h
     ON h.id_herramienta = m.herramienta_id WHERE h.id_herramienta IS NULL;
   ```

3. **Borrador de ADR `028_tools_domain.md`** — una página:
   - Tables nuevas y rationale de dominio separado.
   - FKs a `erp_entities` (empleados prestatarios) y/o
     `erp_maintenance_assets` (vehículos que reciben herramientas).
   - RLS + permisos (`erp.tools.read` / `erp.tools.write`).

4. **Migrator**: mismo patrón que CTBREGIS — reader en
   `tools/cli/internal/legacy/stock.go` (o nuevo archivo), migrator en
   `migrators_plan21b.go` bajo "Phase 4b — Tools". Migración `072` en
   `db/tenant/migrations/`.

### Tarea secundaria (si Pareto #4 cierra con tiempo) — STKINSPR

**STKINSPR** (180 K filas) — inspección de stock. Posible extensión de
`erp_stock` existente o nueva tabla `erp_stock_inspections`. Ver
`.intranet-scrape/xml-forms/stock/` primero. Menos ambigüedad de dominio
que HERRAMIENTAS (probablemente fila suelta en el dominio stock).

## Trampas conocidas (heredadas de 2.0.10)

- **sqlc drift** — el pattern del hand-patch sigue vigente: editá la
  query `.sql`, regenerá, extraé SOLO los blocks nuevos, revertí con
  `git checkout services/erp/internal/repository/`, appendeá a mano.
  Memoria `feedback_sqlc_version_drift`.

- **Phase 0 invariants** — religion. `rows_read = rows_written +
  rows_skipped + rows_duplicate` **y** NO completar con
  `rows_written=0` sobre `rows_read>0`. `migration-health` skill.

- **Lint errcheck pre-existente** en `pkg/server/healthcheck*.go` (3
  errores desde 2.0.7, `cb2f444c`). CI pasa sin ellos. No bloquean.

- **Cold-start migrations** — tres bugs latentes surface sólo en silo
  fresco (psql `:'var'`, cross-DB INSERT, LANGUAGE sql forward-ref).
  Memoria `feedback_migration_cold_start`.

- **Numeración migration** — 2.0.10 dejó `071`. Next libre: **072**.

- **Histrix access requiere VPN activa en Windows**. Pattern
  docker+sshpass en `reference_histrix_access.md`. Passwords:
  workstation `sistemas/Saldivia01!`, Histrix
  `sistemas/SaldiviaAdmin02` + DB `root/m2450e`.

- **Tailscale SSH re-auth** para la workstation — el session start
  puede pedir re-auth via URL (`https://login.tailscale.com/a/...`).
  No bloquea el trabajo de migración (MySQL queries van por WireGuard),
  sólo el drift check post-merge.

- **No hardcodear `tenant_id`** — ADR 022 silo.

## Fuera de scope

- **Phase 1 §UI parity** — sub-order ADR 027 dice "Data migration → UI
  parity". Las pages esperan detrás del data.
- **Phase 2+ (chat, prompts jerárquicos, tree-RAG, ACL)** —
  desbloqueado pero no es prioridad top-down mientras quede Phase 1
  gap abierto.
- **Extensión columnas** — shape mínima es OK; extender cuando una UI
  lo pida.

## Cierre esperado

- **Mínimo** (HERRAMIENTAS solo): +225 K filas cubiertas, gap baja a
  306 tablas / ≈ 1.24 M filas. PR con ADR 028 + migrator.
- **Ideal** (HERRAMIENTAS + HERRMOVS + STKINSPR): +417 K filas
  cubiertas, gap baja a 304 tablas / ≈ 1.05 M filas. Cumulative 79 %
  covered de Histrix.

Post-PR: `gh pr create --base main --head 2.0.11` y tras merge,
`git pull` en la workstation para volver a cerrar drift + tag +
release.

## Working branch

Antes de código: cortar `2.0.11` desde `main` **post-merge de 2.0.10**
y bumpear `CLAUDE.md` ("Working: 2.0.10" → "Working: 2.0.11") con
`chore: bump working branch 2.0.10 → 2.0.11` (pattern `348f02f4`).

Si 2.0.10 sigue abierto cuando arrancás, NO cortes 2.0.11 encima —
primero cerrá el ciclo (ver §"Pre-work si 2.0.10 sigue abierto").
