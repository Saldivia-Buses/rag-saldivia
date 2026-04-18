# Next session — Phase 1 §Data migration: Pareto #6 (PRODUCTO_ATRIB_VALORES 154 K + friends)

La sesión anterior (2.0.10) cerró Pareto #3, #4 y #5 en un PR:

- **CTBREGIS** (604,579 filas live) → `erp_accounting_registers` (071).
  Discovery: no es waiver — libro_diario_qry sigue joining via
  CTB_DETALLES.ctbregis_id.
- **HERRAMIENTAS** (389,253) + **HERRMOVS** (11,680) → `erp_tools` +
  `erp_tool_movements` (072). Discovery: no es catálogo de taller, es
  ledger de tags serializados de inventario + ledger de préstamos.
  Mantiene naming "Herramientas" para parity de menú.
- **STKINSPR** (189,863) → `erp_article_costs` (073). Discovery: no
  es stock inspection, es per-supplier cost ledger (artcos por
  artcod+ctacod). supplier_entity_id vía ResolveEntityFlexible.

Post-2.0.10 el gap Phase 1 §Data migration quedó en **304 tablas /
≈ 876 K filas** (4.6 % del total Histrix). Las 110 tablas cubiertas
son ya **75 %** del volumen total.

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
| Cubiertas | 110 | ≈ 14.15 M (**75 %**) |
| Waiver masivo (W-004/005/006) | 261 | 3.91 M |
| **Gap remaining** | **304** | **≈ 876 K (4.6 %)** |

**Top uncovered post-2.0.10** (row volume):

| Rank | Tabla | Rows (approx) | Cum % del gap |
|---:|---|---:|---:|
| 1 | **PRODUCTO_ATRIB_VALORES** | 153,835 | **17.6 %** |
| 2 | PROD_CONTROL_HOMOLOG | 105,683 | 29.6 % |
| 3 | STK_COSTO_HIST | 95,217 | 40.5 % |
| 4 | BCS_IMPORTACION | 84,492 | 50.1 % |
| 5 | EGX300EPE | 79,040 | 59.2 % |
| 6 | REG_MOVIMIENTO_OBS | 72,737 | 67.5 % |
| 7 | REG_CUENTA_CALIFICACION | 58,960 | 74.2 % |
| 8 | TEL_LOG | 34,885 | 78.2 % |
| 9 | COTIZOPMOVIM | 28,626 | 81.4 % |
| 10 | STK_COSTO_REPOSICION_HIST | 28,515 | 84.7 % |

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

## Tarea principal — Pareto #6: PRODUCTO_ATRIB_VALORES (154 K)

**Skills primarios:** `htx-parity` + `migration-health` + `database` +
`backend-go`.

### Contexto de negocio

PRODUCTO_ATRIB_VALORES (153,835 filas) = atributos key-value por
producto. Extensión lógica del dominio de artículos / productos
(erp_articles existente). Target likely `erp_article_attribute_values`.

Naming precaution: el ejercicio de 2.0.10 mostró que los nombres
Histrix engañan (HERRAMIENTAS no era herramientas de taller; STKINSPR
no era stock inspection). Leer XML-forms + probe shape antes de
asumir el dominio.

### Arranque recomendado

1. **XML-form inventory**:
   ```bash
   grep -rlE "PRODUCTO_ATRIB_VALORES|PRODUCTO_ATRIBUTO" \
     .intranet-scrape/xml-forms/ | head -30
   ```

2. **DB shape**:
   ```sql
   DESCRIBE PRODUCTO_ATRIB_VALORES;
   DESCRIBE PRODUCTO_ATRIBUTO_HOMOLOGACION;   -- rank 18, 17.5 K — sister
   SELECT COUNT(*) FROM PRODUCTO_ATRIB_VALORES;
   SELECT * FROM PRODUCTO_ATRIB_VALORES LIMIT 5;
   ```

3. **Domain**: extension natural de `erp_articles` (value per article).
   Migración `074`. Reutilizar el stock default-subsystem lookup.

### Tarea secundaria — PROD_CONTROL_HOMOLOG (106 K) + STK_COSTO_HIST (95 K)

Dos candidatos top-10:
- **PROD_CONTROL_HOMOLOG** — control de homologación de producción
  (está en rank 3 del gap). Relacionado con erp_homologations ya migrado.
- **STK_COSTO_HIST** — **hermana de STKINSPR**; es la historia de
  costos. Naming likely es literal aquí. Target: `erp_article_cost_history`
  con UNIQUE por (article, supplier, fecha).

Si PRODUCTO_ATRIB_VALORES cierra limpio + PROD_CONTROL_HOMOLOG es
sencillo, +260 K filas cubiertas en la sesión.

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

- **Numeración migration** — 2.0.10 dejó `073`. Next libre: **074**.

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

- **Mínimo** (PRODUCTO_ATRIB_VALORES solo): +154 K filas cubiertas,
  gap baja a 303 tablas / ≈ 722 K filas.
- **Ideal** (PRODUCTO_ATRIB_VALORES + PROD_CONTROL_HOMOLOG +
  STK_COSTO_HIST): +355 K filas cubiertas, gap baja a 301 tablas /
  ≈ 521 K filas (2.8 % del total Histrix). Cumulative coverage ~77 %.

Post-PR: `gh pr create --base main --head 2.0.11` y tras merge,
`git pull` en la workstation para volver a cerrar drift + tag +
release.

## Working branch

Antes de código: cortar `2.0.11` desde `main` **post-merge de 2.0.10**
y bumpear `CLAUDE.md` ("Working: 2.0.10" → "Working: 2.0.11") con
`chore: bump working branch 2.0.10 → 2.0.11` (pattern `348f02f4`).

Si 2.0.10 sigue abierto cuando arrancás, NO cortes 2.0.11 encima —
primero cerrá el ciclo (ver §"Pre-work si 2.0.10 sigue abierto").
