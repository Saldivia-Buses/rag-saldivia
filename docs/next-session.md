# Next session — Phase 1 §Data migration: Pareto #7 (PROD_CONTROL_HOMOLOG 106 K + friends)

La sesión anterior (2.0.10) cerró Paretos #3, #4, #5, #6 y #18 en un PR:

- **CTBREGIS** (604,579) → `erp_accounting_registers` (071).
- **HERRAMIENTAS** (389,253) + **HERRMOVS** (11,680) → `erp_tools` +
  `erp_tool_movements` (072). Discovery: ledger de inventario
  serializado + ledger de préstamos (NO workshop tools catalog).
- **STKINSPR** (189,863) → `erp_article_costs` (073). Discovery:
  per-supplier cost ledger (NOT stock inspection).
- **PRODUCTO_* cluster** (405,805 total) → 6 nuevas tablas
  `erp_products`, `erp_product_sections`, `erp_product_attributes`,
  `erp_product_attribute_options`, `erp_product_attribute_values`,
  `erp_product_attribute_homologations` (074). Cierra Pareto #6
  (PRODUCTO_ATRIB_VALORES 353,936) y #18 (PRODUCTO_ATRIBUTO_HOMOLOGACION
  47,189) de una sola migración.

**4 commits, +1,600+ filas migradas, 5 ranks del Pareto cerrados en un PR.**

Post-2.0.10 el gap Phase 1 §Data migration quedó en **298 tablas /
≈ 470 K filas** (2.5 % del total Histrix). Las 116 tablas cubiertas
son ya **77 %** del volumen total.

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
| Cubiertas | 116 | ≈ 14.56 M (**77 %**) |
| Waiver masivo (W-004/005/006) | 261 | 3.91 M |
| **Gap remaining** | **298** | **≈ 470 K (2.5 %)** |

**Top uncovered post-2.0.10** (row volume — estimates del scrape; live
tiende a ser ~+20-50 %):

| Rank | Tabla | Rows (scrape) | Cum % del gap |
|---:|---|---:|---:|
| 1 | **PROD_CONTROL_HOMOLOG** | 105,683 | **22.5 %** |
| 2 | STK_COSTO_HIST | 95,217 | 42.8 % |
| 3 | BCS_IMPORTACION | 84,492 | 60.7 % |
| 4 | EGX300EPE | 79,040 | 77.5 % |
| 5 | REG_MOVIMIENTO_OBS | 72,737 | 93.0 % |
| 6 | REG_CUENTA_CALIFICACION | 58,960 | 105.6 %* |
| 7 | TEL_LOG | 34,885 | 113.0 %* |
| 8 | COTIZOPMOVIM | 28,626 | 119.1 %* |
| 9 | STK_COSTO_REPOSICION_HIST | 28,515 | 125.1 %* |
| 10 | CARCHEHI | 26,882 | 130.8 %* |

\* Cum % excede 100 % porque usamos estimates del scrape; el live count
está ~15 % más bajo (los ranks 1-5 probablemente cubren todo el gap real
de 470 K).

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

## Tarea principal — Pareto #7: PROD_CONTROL_HOMOLOG (~106 K) + STK_COSTO_HIST (~95 K)

**Skills primarios:** `htx-parity` + `migration-health` + `database` +
`backend-go`.

### Contexto de negocio

Dos candidatos del top del gap restante:

- **PROD_CONTROL_HOMOLOG** (105,683 scrape) — probablemente control de
  homologación de producción. Relacionado con erp_homologations ya
  migrado en 2.0.8 + erp_product_attribute_homologations migrado en
  2.0.10. Target likely `erp_production_control_homologations`.
- **STK_COSTO_HIST** (95,217 scrape) — **hermana directa de STKINSPR**
  (migrada en 2.0.10). STKINSPR es el costo actual por (art, prov);
  STK_COSTO_HIST es la **historia** de cambios. Target natural:
  `erp_article_cost_history`. El metadata enricher existing ya consume
  esta tabla — hay que materializarla relacionalmente.

Naming precaution: el ejercicio de 2.0.10 mostró que los nombres
Histrix engañan 3 de 3 veces (HERRAMIENTAS, STKINSPR, PRODUCTO_*).
Leer XML-forms + probe shape antes de asumir el dominio.

### Arranque recomendado

1. **XML-form inventory**:
   ```bash
   grep -rlE "PROD_CONTROL_HOMOLOG|STK_COSTO_HIST" \
     .intranet-scrape/xml-forms/ | head -20
   ```

2. **DB shape + live count**:
   ```sql
   DESCRIBE PROD_CONTROL_HOMOLOG;
   DESCRIBE STK_COSTO_HIST;
   SELECT COUNT(*) FROM PROD_CONTROL_HOMOLOG;
   SELECT COUNT(*) FROM STK_COSTO_HIST;
   SELECT * FROM PROD_CONTROL_HOMOLOG LIMIT 5;
   SELECT * FROM STK_COSTO_HIST LIMIT 5;
   ```

3. **Decisión**: ambos son extensiones naturales de dominios existentes.
   STK_COSTO_HIST: extender stock.sql. PROD_CONTROL_HOMOLOG: extender
   production / quality. Migración `075` (probablemente ambas juntas si
   cierran limpio).

### Tarea secundaria — BCS_IMPORTACION (84 K)

Rank 3 del gap. Probablemente accounting/imports (BCS* es banco-caja).
Puede ir con el trabajo de PROD_CONTROL + STK_COSTO_HIST si hay tiempo.

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

- **Numeración migration** — 2.0.10 dejó `074`. Next libre: **075**.

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

- **Mínimo** (PROD_CONTROL_HOMOLOG solo): +106 K filas cubiertas,
  gap baja a 297 tablas / ≈ 364 K filas.
- **Ideal** (PROD_CONTROL_HOMOLOG + STK_COSTO_HIST + BCS_IMPORTACION):
  +285 K filas cubiertas, gap baja a 295 tablas / ≈ 185 K filas
  (1 % del total Histrix). Cumulative coverage **~80 %**. Esto pondría
  la fase 1 §Data migration muy cerca del cierre completo — quedarían
  ~290 tablas pequeñas (<30 K filas cada una) para cerrar como bulk
  waivers o migradores rápidos.

Post-PR: `gh pr create --base main --head 2.0.11` y tras merge,
`git pull` en la workstation para volver a cerrar drift + tag +
release.

## Working branch

Antes de código: cortar `2.0.11` desde `main` **post-merge de 2.0.10**
y bumpear `CLAUDE.md` ("Working: 2.0.10" → "Working: 2.0.11") con
`chore: bump working branch 2.0.10 → 2.0.11` (pattern `348f02f4`).

Si 2.0.10 sigue abierto cuando arrancás, NO cortes 2.0.11 encima —
primero cerrá el ciclo (ver §"Pre-work si 2.0.10 sigue abierto").
