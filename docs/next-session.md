# Next session — Cerrar Phase 1 §Data migration

**Goal**: dejar ADR 027 §Phase 1 §Data migration ✅ en esta sesión.
Hoy estamos a **~80 % covered** (119 tablas, ~15 M filas). El long tail
son ~295 tablas de las cuales las 10 más grandes suman ~564 K filas
live — el resto (~285) tienen <15 K cada una y son candidatas a
waiver bulk. Plan = 2-3 migradores cortos + W-007 + strike.

## Cierre ciclo 2.0.10 — completado 2026-04-19

- PR #157 squash-merged como `cd13b7f8` en main.
- Tag `v2.0.10` pushed, GitHub release publicada.
- Workstation `srv-ia-01` sincronizada en `cd13b7f8`.
- Todos los tests + build + lint verdes antes y después del merge.

## Final goal (ADR 026 — no se pierde de vista)

SDA reemplaza Histrix. El empleado abre SDA y:

1. UI moderna cubriendo **todo** lo que Histrix hacía (1:1 parity).
2. Chat donde el agente es su representante — cap parity chat ↔ UI.
3. Dashboard personal + rutinas personales.
4. Agentes background: mail, WhatsApp, tree-RAG con ACL.

## Estado post-2.0.10 (probado live 2026-04-19)

| Segment | Tablas | Filas (live donde medido) |
|---|---:|---:|
| Histrix total | 675 | 18.94 M (scrape) |
| Cubiertas | 119 | ≈ 15.16 M (**~80 %**) |
| Waiver masivo (W-004/005/006) | 261 | 3.91 M |
| **Gap remaining** | **295** | **≈ 115 K** (scrape) |

Scrape subestima — live counts son +20-50 % más en promedio. Ver
`feedback_live_count_vs_scrape_estimate.md`.

## El long tail probado live (2026-04-19)

Shape + live count de las 10 tablas más grandes del gap post-2.0.10,
con domain classification + target SDA sugerido:

### Grupo A — Current-accounts / treasury (~237 K live)

| Tabla | Live | Scrape | Shape clave |
|---|---:|---:|---|
| **REG_CUENTA_CALIFICACION** | **136,064** | 58,960 (+131 %) | `id_regcalificacion` AI PRI, `regcuenta_id` FK, `calificacion` VARCHAR(40), `fecha_calificacion` DATETIME, `referencia_calificacion` |
| **REG_MOVIMIENTO_OBS** | **72,737** | 72,737 | `id_regmovimientoobs` AI PRI, `fec_observacion`, `hora_observacion`, `observacion` LONGTEXT, `regmovim_id` FK, `login`, `tabla_origen`, `ctacod` |
| **CARCHEHI** | **28,763** | 26,882 | composite PK (`carint`, `siscod`, `succod`); 35-col cheque-historia: `carimp`, `carfec`, `carbco`, `carnro`, `cartip`, `ctacod`, `movnro`, `fecha_emision`, `cartera_id` |

**Target SDA**:
- `erp_entity_credit_ratings` — REG_CUENTA_CALIFICACION. FK
  `regcuenta_id → erp_entities` via ResolveEntityFlexible. Trivial.
- `erp_invoice_notes` — REG_MOVIMIENTO_OBS. FK `regmovim_id` via
  `ResolveRegMovim` (Phase 6 index). Preserve observation longtext.
- `erp_check_history` — CARCHEHI. Composite PK → hashCode() for
  legacy_id. FK `ctacod` via entity resolver. Migración con el
  mayor número de columnas del grupo pero shape limpio.

### Grupo B — Stock / production extensions (~176 K live)

| Tabla | Live | Scrape | Shape clave |
|---|---:|---:|---|
| **STK_COSTO_REPOSICION_HIST** | **109,123** | 28,515 (+282 %) | `id_costoreposicion_hist` AI PRI, `costoreposicion_id`, `regcuenta_id` FK, `moneda_id`, `cotizacion` DEC(14,4), `costo_proveedor`, `origen`, `incoterm_id`, `gasto_importacion`, `flete_local_ars`, `modificado` TIMESTAMP, `descuento_1/_2` |
| **ACCESORIOS_COCHE** | **37,909** | 19,671 (+93 %) | `id_accesorio` AI PRI, `nrofab` MUL, `artcod`, `artdes` LONGTEXT, `fecha`, `cotizacion_id`, `estado`, `ficha_id`, `precio_adicional`, `cantidad`, `aprobado`, `precio_unitario`, `prdseccion_id`, `muestra_fv/_ft`, `fc_estado_acc_id` |
| **COTIZOPMOVIM** | **28,573** | 28,626 | `idCotiz` MUL, `idSeccion`, `descripcion`, `idMovim` AI PRI |

**Target SDA**:
- `erp_article_replacement_cost_history` — hermana de
  `erp_article_cost_history` (075). Extiende ese dominio con la shape
  más rica (moneda, origen, incoterm). FK `regcuenta_id` via entity.
- `erp_unit_accessories` — ACCESORIOS_COCHE. Asocia a producción
  unit (`nrofab`) + artículo (`artcod` → default-sub lookup) +
  cotización / ficha. Preservar longtext description.
- `erp_quotation_section_items` — COTIZOPMOVIM. Simple. Resolvr
  `idCotiz` contra cotizaciones ya migradas.

### Grupo C — Industrial telemetry (~95 K live)

| Tabla | Live | Scrape | Shape clave |
|---|---:|---:|---|
| **EGX300EPE** | **79,376** | 79,040 | `insertkey` PRI VARCHAR(200); `id_nave` + `id_csv` + `date_time` DATETIME + 8 columnas DECIMAL(10,3) (potencia_activa/aparente/reactiva, demanda_*, intensidades) |
| **EGX_300** | **15,992** | 15,702 | `id_nave` INT(1), `id_csv` INT, `date` DATETIME, `intensidad1/2/3` DEC(10,3). Sin PK declarada (tabla vieja). |

**Decision point**: ambas tablas son **lecturas del medidor
Schneider PowerLogic EGX300** (contador eléctrico). Son time-series
de energía industrial — datos históricos, no operacionales para la
ERP. **Dos opciones**:

- **(A)** Migrar a `erp_power_meter_readings` (single table, columna
  `meter_model` discriminando EGX300EPE vs EGX_300). ~95 K rows.
- **(B)** **Waive via W-008** — telemetría industrial, no business
  data. No hay XML-form Histrix que las consulte operacionalmente
  (verificar con `grep`). Revisit cuando haya una UI de monitoreo
  industrial (Phase 3+ si llega).

Recomendado **B** — no tiene sentido migrar 100 K filas de sensor
histórico a la ERP. Waiver corto, strike.

### Grupo D — Misc (~50 K live)

| Tabla | Live | Scrape | Notes |
|---|---:|---:|---|
| **TEL_LOG** | **34,885** | 34,885 | Log de llamadas VoIP — `fecha`, `extension`, `numero`, `duracion`. Idem EGX: telemetría, no business. **Waive**. |
| **RECLAMOPAGOS** | **15,463** | 15,463 | Reclamos de pagos — ctacod + observacion longtext. Pequeño pero business-relevant. `erp_payment_complaints` si cabe, sino waive. |

### Long tail restante (~285 tablas < 15K cada una)

**W-007 "sub-15K-row long tail"** bulk waiver. Same shape as W-006
(zero-row) y W-004 (HTX infra). Scope: todas las tablas de
`.intranet-scrape/db-tables.txt` con `table_rows < 15000`, que no
estén en (a) los 3 waivers existentes, (b) los migradores ya
registrados, (c) los migradores de 2.0.11. Row total: ≤ ~200 K.

**Rationale**: estas tablas son el long-tail estadístico — cada una
<15K filas individualmente, colectivamente <200K. El costo de migrar
(un migrator custom por cada una) excede ampliamente el valor. La
política default es: si cualquier Phase 1 UI pide una de esas
tablas, strike del waiver y escribir el migrator entonces.

## Plan de trabajo 2.0.11

### Pre-work

```bash
# 1. Cut branch
git checkout -b 2.0.11 main

# 2. Bump CLAUDE.md
sed -i 's/Working:\*\* `2.0.10`/Working:\*\* `2.0.11`/' CLAUDE.md
git commit -am "chore: bump working branch 2.0.10 → 2.0.11

[incluir el resumen del ciclo 2.0.10 + plan 2.0.11]

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

### Commits planeados

1. **Pareto tail Grupo A** (migración 077 — current-accounts/treasury,
   ~237 K filas):
   - `erp_entity_credit_ratings` + `erp_invoice_notes` + `erp_check_history`
   - Reader en `current_accounts.go` o `treasury.go`
   - Migrator con entity + regmovim resolvers (ambos ya construidos)
   - sqlc queries + hand-patch

2. **Pareto tail Grupo B** (migración 078 — stock/production
   extensions, ~176 K filas):
   - `erp_article_replacement_cost_history` (extends erp_article_cost_history)
   - `erp_unit_accessories` + `erp_quotation_section_items`
   - Reader en `stock_extended.go` + `production_extended.go`
   - sqlc queries + hand-patch

3. **Waiver W-007 + W-008** (no migración):
   - `docs/parity/waivers.md` — W-007 long-tail <15K, W-008 EGX+TEL_LOG
     industrial telemetry
   - Lista pinned en `docs/parity/data-migration.md` (top-of-repo
     reproducer regenera)
   - Strike de ranks correspondientes en el Pareto

4. **Optional — RECLAMOPAGOS migrator** si hay tiempo (migración 079,
   15 K filas): `erp_payment_complaints`.

### Cierre esperado

Post-2.0.11:
- Covered tables: 119 → ~125
- Gap tables: 295 → ≤ 10 (o 0 con W-007)
- Covered row share: 80 % → ~82 %
- Phase 1 §Data migration: **✅ cerrada** en ADR 027

## Trampas heredadas

- **Live count vs scrape** — ya documentado en `feedback_live_count_vs_scrape_estimate`.
  Ya probado live en 2.0.10 — valores reales están en este documento.
- **sqlc drift** — editar `.sql` + hand-patch `.sql.go`/`models.go`
  quirúrgicamente. Memoria `feedback_sqlc_version_drift`.
- **Phase 0 invariants** — `rows_read = rows_written + rows_skipped +
  rows_duplicate`. Si un migrator tiene rows_written=0 sobre rows_read>0
  es un BUG.
- **Cold-start migrations** — tres bugs latentes en silo fresco.
  Memoria `feedback_migration_cold_start`.
- **Numeración migration** — 2.0.10 dejó `076`. Next libre: **077**.
- **Histrix access** — VPN Windows + docker+sshpass pattern
  (`reference_histrix_access.md`). Contraseñas en la memoria.
- **Tailscale SSH re-auth** — workstation a veces pide re-auth
  (URL en `https://login.tailscale.com/a/...`). No bloquea trabajo
  de MySQL que va por WireGuard.

## Fuera de scope

- **Phase 1 §UI parity** — ADR 027 sub-order dice Data → UI. Las
  pages esperan detrás del data.
- **Phase 2+ (chat, prompts jerárquicos, tree-RAG, ACL)** —
  desbloqueado pero no es top-down prioridad mientras quede Phase 1
  abierta.
- **End-to-end dry-run contra saldivia tenant** — ops task, no session
  task. El migrador queda shipped en el PR; la validación live la
  hace la siguiente cutover rehearsal.

## Post-PR cierre ciclo

```bash
gh pr create --base main --head 2.0.11 --title "..." --body "..."
# Post-merge:
git checkout main && git pull origin main
git tag v2.0.11 && git push origin v2.0.11
gh release create v2.0.11 --title "..." --notes "..."
ssh sistemas@srv-ia-01 "cd /opt/saldivia/repo && git pull origin main"
```

Memoria `feedback_version_tagging.md`. Release body incluye:
- Resumen de grupos migrados (A/B + waivers W-007/W-008).
- Phase 1 §Data migration = ✅ done → Phase 1 §UI parity desbloqueada.
