# Next session — Phase 1 §Data migration: Pareto #5 (STKINSPR 180 K + friends)

La sesión anterior (2.0.10) cerró Pareto #3 y #4 en un PR:

- **CTBREGIS** (604,579 filas live — el scrape estimaba 572,076) migrada
  vía `NewAccountingRegisterMigrator` → nueva tabla
  `erp_accounting_registers` (migración `071`). `account_id` resuelto vía
  code-index accounting (91.2 % match; 8.8 % orphan con `account_id NULL`
  + `account_code` preservado).
- **HERRAMIENTAS** (389,253) + **HERRMOVS** (11,680) migradas vía
  `NewToolMigrator` + `NewToolMovementMigrator` → nuevas tablas
  `erp_tools` + `erp_tool_movements` (migración `072`). Discovery: pese
  al naming, HERRAMIENTAS **no es catálogo de herramientas de taller**
  sino **ledger de tags serializados de inventario** (1 fila = 1 unidad
  física recibida con barcode único). HERRMOVS es ledger de préstamos
  (empleados retiran / devuelven / rompen items). CONCHERR (enum 4
  filas) inlineado como `concept_code SMALLINT`. Mantiene naming
  "Herramientas" para parity con menú Histrix.

Post-2.0.10 el gap Phase 1 §Data migration quedó en **305 tablas /
≈ 1.07 M filas** (5.6 % del total Histrix). Las 109 tablas cubiertas
son ya **74 %** del volumen total.

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
| Cubiertas | 109 | ≈ 13.96 M (**74 %**) |
| Waiver masivo (W-004/005/006) | 261 | 3.91 M |
| **Gap remaining** | **305** | **≈ 1.07 M (5.6 %)** |

**Top uncovered post-2.0.10** (row volume):

| Rank | Tabla | Rows (approx) | Cum % del gap |
|---:|---|---:|---:|
| 1 | **STKINSPR** | 180,412 | **16.9 %** |
| 2 | PRODUCTO_ATRIB_VALORES | 153,835 | 31.3 % |
| 3 | PROD_CONTROL_HOMOLOG | 105,683 | 41.2 % |
| 4 | STK_COSTO_HIST | 95,217 | 50.1 % |
| 5 | BCS_IMPORTACION | 84,492 | 58.0 % |
| 6 | EGX300EPE | 79,040 | 65.4 % |
| 7 | REG_MOVIMIENTO_OBS | 72,737 | 72.2 % |
| 8 | REG_CUENTA_CALIFICACION | 58,960 | 77.8 % |
| 9 | TEL_LOG | 34,885 | 81.0 % |
| 10 | COTIZOPMOVIM | 28,626 | 83.7 % |

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

Memoria: `feedback_version_tagging.md`. Release body incluye ambas
sections: Pareto #3 CTBREGIS + Pareto #4 HERRAMIENTAS/HERRMOVS.

## Tarea principal — Pareto #5: STKINSPR (180 K)

**Skills primarios:** `htx-parity` + `migration-health` + `database` +
`backend-go`.

### Contexto de negocio

STKINSPR (180,412 filas) es probablemente **stock inspection** — algún
tipo de control de calidad / inspección al recibir mercadería. Hay que
leer XML-forms para confirmar.

### Arranque recomendado

1. **XML-form inventory**:
   ```bash
   grep -rlE "STKINSPR\b" .intranet-scrape/xml-forms/ | head -20
   ```
   Áreas probables: `almacen/`, `recepcion/`, `calidad/`, `controlcalidad/`.

2. **DB shape en Histrix live**:
   ```sql
   DESCRIBE STKINSPR;
   SELECT COUNT(*) FROM STKINSPR;
   -- FKs (artcod? siscod? id_alguna_cosa?)
   SELECT * FROM STKINSPR LIMIT 5;
   ```

3. **Decisión de dominio**: casi seguro extiende `erp_stock` existente
   (`erp_stock_inspections`?) — no requiere ADR nuevo. Migración `073`.

### Tarea secundaria — PRODUCTO_ATRIB_VALORES (154 K)

Atributos de productos — probablemente extension de `erp_articles` con
key-value. Si cierra limpio junto con STKINSPR, +334 K filas en el PR
(31 % del gap).

## Trampas conocidas (heredadas de 2.0.10)

- **sqlc drift** — el pattern del hand-patch sigue vigente: editá la
  query `.sql`, regenerá, extraé SOLO los blocks nuevos, revertí con
  `git checkout services/erp/internal/repository/`, appendeá a mano.
  Memoria `feedback_sqlc_version_drift`.

- **Live count vs scrape estimate** — el Pareto usa
  `information_schema.table_rows` (estimate). CTBREGIS en 2.0.10
  creció 5.7 % vs scrape; HERRAMIENTAS creció 73 %. Confirmá
  COUNT(*) live antes de escribir la Phase 0 invariant en el commit.
  Memoria `feedback_live_count_vs_scrape_estimate`.

- **Phase 0 invariants** — religion. `rows_read = rows_written +
  rows_skipped + rows_duplicate` **y** NO completar con
  `rows_written=0` sobre `rows_read>0`. `migration-health` skill.

- **Lint errcheck pre-existente** en `pkg/server/healthcheck*.go` (3
  errores desde 2.0.7, `cb2f444c`). CI pasa sin ellos. No bloquean.

- **Cold-start migrations** — tres bugs latentes surface sólo en silo
  fresco (psql `:'var'`, cross-DB INSERT, LANGUAGE sql forward-ref).
  Memoria `feedback_migration_cold_start`.

- **Numeración migration** — 2.0.10 dejó `072`. Next libre: **073**.

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

- **Mínimo** (STKINSPR solo): +180 K filas cubiertas, gap baja a 304
  tablas / ≈ 886 K filas.
- **Ideal** (STKINSPR + PRODUCTO_ATRIB_VALORES + PROD_CONTROL_HOMOLOG):
  +440 K filas cubiertas, gap baja a 302 tablas / ≈ 626 K filas
  (3.3 % del total Histrix). Cumulative coverage ~77 %.

Post-PR: `gh pr create --base main --head 2.0.11` y tras merge,
`git pull` en la workstation para volver a cerrar drift + tag +
release.

## Working branch

Antes de código: cortar `2.0.11` desde `main` **post-merge de 2.0.10**
y bumpear `CLAUDE.md` ("Working: 2.0.10" → "Working: 2.0.11") con
`chore: bump working branch 2.0.10 → 2.0.11` (pattern `348f02f4`).

Si 2.0.10 sigue abierto cuando arrancás, NO cortes 2.0.11 encima —
primero cerrá el ciclo (ver §"Pre-work si 2.0.10 sigue abierto").
