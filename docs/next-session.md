# Next session — Phase 1 §Data migration: cleanup del long tail (BCS_IMPORTACION + friends)

La sesión anterior (2.0.10) cerró Paretos #3, #4, #5, #6, #7, #8 y #18 en un PR:

- **CTBREGIS** (604,579) → `erp_accounting_registers` (071).
- **HERRAMIENTAS** (389,253) + **HERRMOVS** (11,680) → `erp_tools` +
  `erp_tool_movements` (072). Discovery: ledger de inventario
  serializado + ledger de préstamos (NO workshop tools catalog).
- **STKINSPR** (189,863) → `erp_article_costs` (073). Discovery:
  per-supplier cost ledger (NOT stock inspection).
- **PRODUCTO_* cluster** (405,805 total) → 6 tablas `erp_product_*` (074).
  Cierra Pareto #6 y #18.
- **PROD_CONTROL_HOMOLOG** (403,028 live — scrape decía 105K, +282 %
  growth) + **STK_COSTO_HIST** (103,799 live) → `erp_production_
  inspection_homologations` + `erp_article_cost_history` (075).
  Cierra Pareto #7 y #8.
- **BCS_IMPORTACION** (91,959 live) → `erp_bank_imports` (076).
  Cierra rank #9 del Pareto original — import bancario staging.

**6 commits, +2,199,966 filas migradas, 8 ranks del Pareto cerrados
en un PR. Cobertura total post-PR: ~80 %.**

Post-2.0.10 el gap quedó en **296 tablas**, ~200 K filas scrape-anchored
(~1 % del total Histrix). Las 118 tablas cubiertas son ~80 % del
volumen total por live COUNT(*). Caveat: las estimates de
`information_schema.table_rows` subestiman bastante — los live counts
que hicimos validar llegan 10-50 % por encima del scrape. Ver
`feedback_live_count_vs_scrape_estimate`.

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

| Segment | Tablas | Filas (live donde medido, scrape si no) |
|---|---:|---:|
| Histrix total | 675 | 18.94 M (scrape estimate) |
| Cubiertas | 118 | ≈ 15.07 M (**~80 %**) |
| Waiver masivo (W-004/005/006) | 261 | 3.91 M |
| **Gap remaining** | **296** | **≈ 200 K scrape** (~1 %) |

**Top uncovered post-2.0.10** (scrape values — live puede ser +20-50 %):

| Rank | Tabla | Rows (scrape) |
|---:|---|---:|
| 1 | **EGX300EPE** | 79,040 |
| 2 | REG_MOVIMIENTO_OBS | 72,737 |
| 3 | REG_CUENTA_CALIFICACION | 58,960 |
| 4 | TEL_LOG | 34,885 |
| 5 | COTIZOPMOVIM | 28,626 |
| 6 | STK_COSTO_REPOSICION_HIST | 28,515 |
| 7 | CARCHEHI | 26,882 |
| 8 | ACCESORIOS_COCHE | 19,671 |
| 9 | EGX_300 | 15,702 |
| 10 | RECLAMOPAGOS | 15,463 |

Estos 10 suman ~380 K filas scraped (~250-500 K live estimate). El
resto son 285 tablas con <15 K filas cada una — candidatas para bulk
waiver W-007 o cierre rápido con migrator genérico.

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

## Tarea principal — long tail cleanup: BCS_IMPORTACION + friends

**Skills primarios:** `htx-parity` + `migration-health` + `database` +
`backend-go`.

### Contexto

El top-5 del Pareto original ya cerró. Los 10 siguientes suman ~450 K
filas scraped, probablemente 200-400 K live. Cada uno es un
migrator sencillo (3-6 columnas), similar shape a los de 2.0.10.

### Candidatos

- **BCS_IMPORTACION** (84,492 scrape) — BCS = banco-caja-sucursal.
  Probablemente importaciones de transacciones bancarias.
- **EGX300EPE** / **EGX_300** (~95 K combinados) — dos tablas con
  prefix EGX, desconocido sin XML scrape.
- **REG_MOVIMIENTO_OBS** (72,737) — observaciones sobre registros de
  movimientos. Simple join a movements.
- **REG_CUENTA_CALIFICACION** (58,960) — calificaciones de cuentas
  corrientes (clientes/proveedores).
- **TEL_LOG** (34,885) — log de telefonía.
- **COTIZOPMOVIM** (28,626) — cotización-operación-movimiento.
- **STK_COSTO_REPOSICION_HIST** (28,515) — historia costos reposición
  (hermana de STK_COSTO_HIST migrada en 2.0.10).
- **CARCHEHI** (26,882) — carbanco-cheque-historia (treasury).
- **ACCESORIOS_COCHE** (19,671) — accesorios de coches.

### Estrategia

Agrupar por dominio para minimizar churn:
- **Accounting/treasury**: BCS_IMPORTACION + CARCHEHI (~111 K combinadas)
- **Current accounts**: REG_MOVIMIENTO_OBS + REG_CUENTA_CALIFICACION
  + COTIZOPMOVIM (~160 K)
- **Stock history**: STK_COSTO_REPOSICION_HIST (~28 K) — mimetic con
  el pattern de STK_COSTO_HIST hecho en 2.0.10
- **Misc**: EGX tables (investigate), TEL_LOG, ACCESORIOS_COCHE

Cada grupo = 1 migración nueva con 1-3 tablas. ~3-5 commits en total.

### Arranque

```bash
# Bulk probe: shape + count + sample de todas
for TBL in BCS_IMPORTACION EGX300EPE EGX_300 REG_MOVIMIENTO_OBS \
  REG_CUENTA_CALIFICACION TEL_LOG COTIZOPMOVIM \
  STK_COSTO_REPOSICION_HIST CARCHEHI ACCESORIOS_COCHE; do
  echo "-- $TBL --"
  echo "DESCRIBE $TBL; SELECT COUNT(*) FROM $TBL; SELECT * FROM $TBL LIMIT 3;"
done > /tmp/work/tail_probe.sql
# Then run via docker+sshpass pattern.
```

Al cerrar el long tail, el gap baja a ~286 tablas con <15 K filas —
eligible para bulk waiver W-007.

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

- **Numeración migration** — 2.0.10 dejó `076`. Next libre: **077**.

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

- **Mínimo** (BCS_IMPORTACION + CARCHEHI solo): +111 K filas cubiertas,
  gap baja a 294 tablas / ≈ 90 K filas scrape.
- **Ideal** (long-tail top-10 cerrado): ~450 K filas cubiertas, gap
  baja a 286 tablas pequeñas (<15 K cada una) — candidatas para bulk
  waiver W-007 "long tail < 15K rows" similar a W-006. Cumulative
  coverage **~83 %** live, con Phase 1 §Data migration prácticamente
  cerrada modulo waivers.

Post-PR: `gh pr create --base main --head 2.0.11` y tras merge,
`git pull` en la workstation para volver a cerrar drift + tag +
release.

## Working branch

Antes de código: cortar `2.0.11` desde `main` **post-merge de 2.0.10**
y bumpear `CLAUDE.md` ("Working: 2.0.10" → "Working: 2.0.11") con
`chore: bump working branch 2.0.10 → 2.0.11` (pattern `348f02f4`).

Si 2.0.10 sigue abierto cuando arrancás, NO cortes 2.0.11 encima —
primero cerrá el ciclo (ver §"Pre-work si 2.0.10 sigue abierto").
