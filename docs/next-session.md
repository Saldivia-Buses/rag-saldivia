# Next session — Phase 1 §Data migration: Pareto #3 (CTBREGIS 572 K + HERRAMIENTAS cluster)

La sesión anterior (2.0.9) cerró Pareto #2 en un PR:

- **FICHADAS** (1,465,002 filas) + **PERSONAL_TARJETA** (1,403 filas)
  migrados vía `NewTimeClockEventMigrator` + `NewEmployeeCardMigrator`
  → tablas nuevas `erp_time_clock_events` + `erp_employee_cards`
  (migración `070`). Resolución card+fecha → empleado vía el nuevo
  `Mapper.BuildTarjetaIndex` / `ResolveByTarjetaAtDate` (date-versioned).

Post-2.0.9 el gap Phase 1 §Data migration quedó en **308 tablas /
≈ 2.07 M filas** (11 % del total Histrix). Las 106 tablas cubiertas son
ya **68 %** del volumen total.

Arrancás con PR de 2.0.9 **mergeable** al cerrar la sesión. Si no se
mergea solo, cerrar el ciclo (ver §"Pre-work si 2.0.9 sigue abierto"
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

## Estado Phase 1 §Data migration (post-2.0.9)

| Segment | Tablas | Filas |
|---|---:|---:|
| Histrix total | 675 | 18.94 M |
| Cubiertas | 106 | ≈ 12.96 M (**68 %**) |
| Waiver masivo (W-004/005/006) | 261 | 3.91 M |
| **Gap remaining** | **308** | **≈ 2.07 M (11 %)** |

**Top uncovered post-2.0.9** (row volume):

| Rank | Tabla | Rows (approx) | Cum % del gap |
|---:|---|---:|---:|
| 1 | **CTBREGIS** | 572,076 | **27.7 %** |
| 2 | HERRAMIENTAS | 225,355 | 38.6 % |
| 3 | STKINSPR | 180,412 | 47.3 % |
| 4 | PRODUCTO_ATRIB_VALORES | 153,835 | 54.7 % |
| 5 | PROD_CONTROL_HOMOLOG | 105,683 | 59.8 % |
| 6 | STK_COSTO_HIST | 95,217 | 64.4 % |
| 7 | BCS_IMPORTACION | 84,492 | 68.5 % |
| 8 | EGX300EPE | 79,040 | 72.3 % |
| 9 | REG_MOVIMIENTO_OBS | 72,737 | 75.8 % |
| 10 | REG_CUENTA_CALIFICACION | 58,960 | 78.7 % |

Reproducer completo al final de `docs/parity/data-migration.md`.

## Pre-work si 2.0.9 sigue abierto

Antes de cortar 2.0.10:

```bash
gh pr checks <NRO>          # tu PR 2.0.9
gh pr view <NRO> --json mergeable,state

# Post-merge en main:
git checkout main && git pull origin main
git tag v2.0.9 && git push origin v2.0.9
gh release create v2.0.9 --title "..." --notes-file <...>

# Workstation drift (srv-ia-01):
ssh sistemas@srv-ia-01 "cd /opt/saldivia/repo && git pull origin main"
# Verificar: git rev-parse origin/main == git rev-parse HEAD en workstation.
```

Memoria: `feedback_version_tagging.md`. Incluir en release body la
sección 2.0.9 (Pareto #2 FICHADAS + PERSONAL_TARJETA) con métricas.

## Tarea principal — Pareto #3: CTBREGIS (572 K filas)

**Skills primarios:** `htx-parity` + `migration-health` + `database` +
(posiblemente) `backend-go`.

### Contexto de negocio

- `CTBREGIS` (572,076 filas) es el **registro log** contable histórico
  de Histrix. El flow operativo ya lo cubren `CTB_DETALLES` /
  `CTB_MOVIMIENTOS` (nuevo schema) → `erp_journal_entries` +
  `erp_journal_lines` (migradas desde 2.0.3).
- La duda clave: **¿CTBREGIS tiene consumidores UI activos, o ya está
  reemplazado por el flow nuevo?** Si es lo segundo → **waiver W-008**.
  Si existe pantalla Histrix que lo consulta directamente → migrator.

### Arranque recomendado

1. **Read XML-forms**: `grep -rl "CTBREGIS\b" .intranet-scrape/xml-forms/`
   — ver qué pantallas lo usan. Si sólo aparece en pantallas del
   conjunto `contabilidad/registros*` que ya tienen equivalente SDA en
   journal entries, W-008.
2. **Inspeccionar shape en Histrix live** (vía docker+sshpass):
   ```sql
   DESCRIBE CTBREGIS;
   SELECT COUNT(*) FROM CTBREGIS;
   -- comparar con CTB_MOVIMIENTOS (ya cubierto)
   SELECT COUNT(*) FROM CTB_MOVIMIENTOS;
   -- ¿son redundantes?
   ```
3. **Decisión A vs B**:
   - **A (migrator)**: si hay UI dedicada → nueva tabla
     `erp_accounting_registers` + reader + migrator + migración 071.
   - **B (waiver W-008)**: si flow real sigue vía journal entries →
     waiver corto en `docs/parity/waivers.md` + strike en Pareto
     `data-migration.md`.

### Tarea secundaria (si A cerró bien o B es trivial) — HERRAMIENTAS + HERRMOVS

**Cluster de herramientas** (~237 K filas combinadas):
- `HERRAMIENTAS` (225 K) — catálogo de herramientas
- `HERRMOVS` (11 K) — movimientos

Nuevo dominio SDA: probablemente `erp_tools` o compartir con
`erp_stock` extendido. Requiere **ADR note** antes del migrator (nuevo
surface no trivial). Si la Pareto #3 cierra como waiver (20 min), hay
tiempo de abrir el ADR + diseño de esquema.

## Trampas conocidas (heredadas de 2.0.9)

- **sqlc drift** — el pattern del hand-patch sigue vigente: editá la
  query `.sql`, regenerá, extraé SOLO los blocks nuevos, revertí con
  `git checkout services/erp/internal/repository/`, appendeá a mano.
  Memoria `feedback_sqlc_version_drift`.

- **Phase 0 invariants** — religion. `rows_read = rows_written +
  rows_skipped + rows_duplicate` **y** NO completar con
  `rows_written=0` sobre `rows_read>0`. `migration-health` skill.

- **Lint errcheck pre-existente** en `pkg/server/healthcheck*.go` (3
  errores). Vienen desde 2.0.7 (`cb2f444c`), CI pasa sin ellos. No
  bloquean; se puede ignorar o fix drive-by aparte.

- **Cold-start migrations** — tres bugs latentes surface sólo en silo
  fresco (psql `:'var'`, cross-DB INSERT, LANGUAGE sql forward-ref).
  Memoria `feedback_migration_cold_start`.

- **Numeración migration** — leer el último archivo en
  `db/tenant/migrations/` (2.0.9 dejó 070). Next libre: **071**.

- **Histrix access requiere VPN activa en Windows**. Pattern
  docker+sshpass en `reference_histrix_access.md`. El IP
  `172.22.100.23` del workstation NO es accesible desde WSL sin el
  mismo VPN; usá hostname `srv-ia-01` (resuelve vía DNS del VPN).
  Passwords: workstation `sistemas/Saldivia01!`, Histrix
  `sistemas/SaldiviaAdmin02` + DB `root/m2450e`.

- **No hardcodear `tenant_id`** — ADR 022 silo.

## Fuera de scope

- **Phase 1 §UI parity** — sub-order ADR 027 dice "Data migration → UI
  parity". Las pages esperan detrás del data.
- **Phase 2+ (chat, prompts jerárquicos, tree-RAG, ACL)** —
  desbloqueado pero no es prioridad top-down mientras quede Phase 1
  gap abierto.
- **Extensión columnas** — la shape 2.0.9 (FICHADAS con 9 columnas,
  minimal) está bien para Phase 1; extender cuando una UI lo pida.
- **Merge freeze / cutover runbook** — más adelante en Phase 1
  §Cutover readiness.

## Cierre esperado

- **Mínimo** (B, waiver W-008): CTBREGIS waived + strike en Pareto.
  Gap baja a 307 tablas / ≈ 1.50 M filas. PR corto (1-2 h).
- **Ideal** (A, migrator + bonus HERRAMIENTAS cluster): +798 K filas
  cubiertas, gap baja a 305 tablas / ≈ 1.27 M filas. Pareto row 1
  cierra ~28 %, row 2 cierra ~11 % = 39 % del gap restante en una
  sesión.

Post-PR: `gh pr create --base main --head 2.0.10` y tras merge,
`git pull` en la workstation para volver a cerrar drift + tag +
release.

## Working branch

Antes de código: cortar `2.0.10` desde `main` **post-merge de 2.0.9**
y bumpear `CLAUDE.md` ("Working: 2.0.9" → "Working: 2.0.10") con
`chore: bump working branch 2.0.9 → 2.0.10` (pattern `ce1aacf7`).

Si 2.0.9 sigue abierto cuando arrancás, NO cortes 2.0.10 encima —
primero cerrá el ciclo (ver §"Pre-work si 2.0.9 sigue abierto").
