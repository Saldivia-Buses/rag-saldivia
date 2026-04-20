# Next session — 2.0.21 quality close: DB migrada + E2E + polish loop

**Pivot crítico** (post 2026-04-20): después de shipear 15 clusters
en 2.0.21 sin polish profundo, el usuario reportó que "fallan casi
todos en un montón de puntos". El target de 20 clusters/sesión queda
**subordinado a calidad**: un cluster sin E2E pasando + sin probarse
contra DB migrada NO cuenta como shipeado. Build verde + typecheck
verde no es evidencia suficiente.

Esta sesión NO suma clusters nuevos. Cierra 2.0.21 con E2E coverage
real + polish loop antes del PR.

## Estado al cierre de la sesión anterior (2026-04-20)

**Branch**: `2.0.21` (no merged), 18 commits sobre `main`:
- 15 § clusters shipped (1, 2, 3, 4, 5, 6, 7, 8, 10, 11, 12, 13, 14, 15, 16).
- 3 docs (plan inicial + scale a 12 + scale a 20).
- 1 chore: bump working branch.
- 1 chore: `next.config.ts` rewrites para que `make dev-frontend`
  funcione en local sin Traefik (fix descubierto durante intento de
  testing — login pegaba a Next.js en vez de backend → "Error
  interno").

**Deferred (NO entran en 2.0.21)**: clusters 17-20 (POST credit
rating, supplier demerit, invoice note, work_order_part). Vuelven
sólo después de que 1-16 estén pulidos.

**Linter mods stasheados**: `git stash list` muestra "linter mods
pre-2.0.21" (CLAUDE.md dup block + .gitignore + .claude/settings.json).
Pop al post-merge final.

**Phase 0 incident encontrado y fixeado durante boot**:
- Redis container estaba healthy en docker pero `Connection reset
  by peer` desde el host — restart resolvió.
- Postgres tenant `dev` tenía outbox-drainer en loop con
  "current transaction is aborted" (SQLSTATE 25P02). Restart de
  postgres lo limpió pero la tabla `outbox` puede tener basura.
  Confirmar antes de cargar tenant migrado.

## Goal

Cerrar 2.0.21 con la 1ra release que tiene **E2E suite verde
contra DB migrada con data real**. Es el primer ciclo donde se
acepta un PR sólo si reproduce el flujo del usuario, no sólo
compila.

## Phase 1 — DB migrada local (Phase 0 prerequisite)

Path A (preferido): `pg_dump` desde workstation `srv-ia-01`
(`172.22.100.23`). Memoria `infrastructure-access` cubre VPN +
SSH + creds.

Pasos:
1. Confirmar VPN activa.
2. SSH a workstation: identificar el container Postgres + nombre
   de DB del tenant migrado (probablemente `tenant_saldivia` o
   similar — verificar en runtime).
3. `pg_dump --format=custom --no-owner --no-acl --exclude-schema=audit ...`
   — dump del tenant migrado. **Excluir** audit schema y
   row-level-security policies para no romper el restore local.
4. `scp` el dump a local.
5. En local: crear `tenant_test` DB en el deploy postgres
   (`deploy-postgres-1`). `pg_restore` el dump.
6. Registrar el tenant en `platform.tenants` para que SDA lo
   reconozca.
7. Crear un user de prueba con role admin (o reusar uno existente
   del dump si las creds son recuperables).

Path B (fallback si VPN no anda): `sda migrate-legacy --mysql-dsn=...
--pg-dsn=... --tenant=test` desde el tunnel a Histrix MySQL. Más
lento (~2.6M rows en la tabla #1 del Pareto), pero valida el
migrator de paso. Usa memoria `histrix_access` para el tunnel.

## Phase 2 — E2E baseline run

`apps/web/e2e/` tiene **5 specs ya escritos**: login, erp-mutations,
erp-navigation, modules-accessibility, settings, chat. Playwright
1.59 instalado.

Pasos:
1. `cd apps/web && bunx playwright install` — descargar browsers.
2. `bunx playwright test` contra el dev local (con DB migrada
   activa).
3. Reporte de fails: agrupar por root cause (UI rota, validación
   ausente, error msg malo, race condition, etc.).
4. Triage: cada fail = ticket implícito. Priorizar por bloqueante
   vs cosmético.

**Importante**: no suprimir fails ni skipear tests. Si un test
existente está mal escrito, fixearlo. Si un cluster shipped no
pasa el spec, eso es un bug del cluster — fixear el cluster.

## Phase 3 — E2E coverage para los 15 clusters de 2.0.21

Los specs existentes cubren el módulo ERP general pero no los
clusters específicos del ciclo. Agregar:

| Cluster | Spec a agregar |
|---|---|
| 1 — filter reconciliations | `e2e/treasury-reconciliations.spec.ts`: navegar a cuenta-bancaria detail, verificar Network call con `?bank_account_id=` |
| 2 — filter cash-counts | `e2e/treasury-cash-counts.spec.ts`: ídem para caja detail |
| 3 — filter entries + UI consumer | `e2e/accounting-cost-center.spec.ts`: ver sección "Asientos imputados" |
| 4 — GetTool detail | `e2e/almacen-tool-detail.spec.ts`: ficha + historial |
| 5 — GetAsset detail | `e2e/maintenance-asset-detail.spec.ts`: planes |
| 6 — GetChassisModel | `e2e/manufacturing-chassis-model.spec.ts` |
| 7 — GetCarroceriaModel + BOM | `e2e/manufacturing-carroceria-model.spec.ts` |
| 8 — GetScorecard | `e2e/quality-scorecard.spec.ts`: badge color |
| 10 — GetUnit | `e2e/manufacturing-unit-detail.spec.ts`: pending controls |
| 11 — GetActionPlan | `e2e/quality-action-plan.spec.ts` |
| 12 — GetNC | `e2e/quality-nc.spec.ts` |
| 13 — GetArticle direct | extender `erp-navigation.spec.ts`: assert single fetch |
| 14 — POST entity note | extender `erp-mutations.spec.ts`: textarea → submit → re-fetch |
| 15 — POST entity contact | ídem cluster 14 |
| 16 — POST inventory movement | extender `erp-mutations.spec.ts`: pickers + submit |

Cada spec debe cubrir: **happy path + 1 error path + 1 empty
state**. No optar por "smoke test que sólo verifica que la página
carga" — eso ya está en `erp-navigation.spec.ts`.

## Phase 4 — Polish loop por cluster

Criteria que cada cluster debe cumplir antes de marcarse "done"
(memoria `feedback_quality_over_quantity`):

- **Validación de form**: required, formato (UUID/email/numérico),
  rangos. Mensajes en español, claros.
- **Defaults inteligentes**: fecha = hoy, tipo = más común,
  unidad = "und" para movements, etc.
- **Submit deshabilitado** mientras inválido o `isPending`.
- **Mensaje de error post-fail** claro, no "Error interno". Distinguir
  401 / 403 / 422 / 5xx.
- **Empty state explicativo** ("Sin notas registradas — usá el form
  abajo para crear la primera"), no sólo "Sin datos".
- **Confirmación destructiva** para delete / override / irreversible.
- **Loading state visible** (skeleton, spinner, disabled state).

Para cada cluster que falla algún criterio: fix + nuevo commit
dentro de `2.0.21`. Empujar el spec primero (TDD), después fix.

## Phase 5 — Cierre 2.0.21

**Sólo cuando**:
- Todos los E2E (existentes + nuevos) pasan contra DB migrada.
- Manual smoke por área (compras, tesorería, contabilidad,
  almacén, mantenimiento, manufactura, calidad, ingeniería) hecho
  por el usuario con el tenant migrado.
- Ningún cluster shipped tiene un fail en el polish criteria
  documentado en este doc.

```bash
gh pr create --base main --head 2.0.21 --title "..." --body "..."
gh pr merge 2.0.21 --squash --auto --delete-branch
# Post-merge:
git stash push -m "linter mods" -- CLAUDE.md .gitignore .claude/settings.json
git checkout main && git pull origin main
git stash pop
git tag v2.0.21 && git push origin v2.0.21
gh release create v2.0.21 --title "..." --notes "..."
ssh sistemas@srv-ia-01 "cd /opt/saldivia/repo && git pull origin main && docker compose up -d --build"
```

## Fuera de scope

- **Clusters 17-20** (write scouts deferred): vuelven en 2.0.22
  como primera tanda *con* E2E desde el commit inicial.
- **Cualquier cluster nuevo** mientras 1-16 no estén polished.
- **Refactor estructural** (admin Tier B, fusion ERP). Sigue siendo
  ADR 027 territory para sesiones futuras.
- **Phase 2+** (chat agent, prompts jerárquicos, tree-RAG, ACL).

## Trampas heredadas / pre-existing

- **Redis local crash silencioso** — si `make dev-services` muere
  por "redis is required for token revocation", `docker restart
  deploy-redis-1`. Worth investigar root cause.
- **Postgres outbox-drainer aborted state** — el tenant `dev` puede
  necesitar `TRUNCATE outbox` antes de iniciar otra vez. Confirmar
  que el `dev` tenant no se corrompa al cargar tenant migrado en
  paralelo.
- **NEXT_PUBLIC_API_URL** se inyecta vacío por `make dev-frontend`
  por diseño — el fix `next.config.ts` (commit `chore(dev)`) ya
  agrega rewrites condicionales. **Verificar** que no rompa el
  build de producción (CI debería detectarlo).
- **Mock interfaces must track**: cada nuevo método agregado al
  handler interface rompe `*_test.go` mocks. Validado otra vez
  durante 2.0.21.
- **sqlc regen drift**: editar generated `*.sql.go` a mano (memoria
  `sqlc_version_drift`).
- **cwd persiste entre Bash calls**: usar rutas absolutas o
  `git -C /home/enzo/rag-saldivia`.
- **Linter mods stasheados** (`linter mods pre-2.0.21`): pop al
  post-merge final, NO antes.

## Candidatos sesiones futuras (2.0.22+)

| Orden | Tema | Pre-req |
|---:|---|---|
| 1 | **Write scouts 17-20 con E2E desde día 1** | 2.0.21 cerrado |
| 2 | **Reports** (Libro IVA, mayor contable, tax-book) | 2.0.21 cerrado |
| 3 | **Seamless-day cutover test** (Phase 1 gating final) | E2E coverage estable |
| 4 | **Admin Tier B refactor** (deuda 2.0.17) | scope independiente |
| 5 | **Bulk write paths** (create/update masivo en clusters scouted) | scout de 2.0.22 validado |
