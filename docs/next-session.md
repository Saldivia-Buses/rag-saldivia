# Next session — 2.0.21 quality close: frontend reinforcement integral

**Pivot crítico** (post 2026-04-20): después de shipear 15 clusters
en 2.0.21 sin polish profundo, el usuario reportó que "fallan casi
todos en un montón de puntos". La suite Playwright actual es floja
(theater tests, mocks ocultos, login hardcoded a dev-seed), hay
mock data en `components/shadcnblocks/`, la información-arquitectura
del sidebar está rota (tesorería no es accesible por UI aunque sus
rutas existen, rutas en lugar equivocado), y los componentes nuevos
tienen drift de types contra el schema real.

El target de 20 clusters/sesión queda **subordinado a calidad**: un
cluster sin E2E pasando + sin probarse contra DB migrada NO cuenta
como shipeado.

Esta sesión NO suma clusters nuevos. Cierra 2.0.21 con un
**reforzamiento integral del frontend**: IA, mocks, types, perms,
performance, a11y, UX, error handling. Más exhaustivo que cualquier
sesión previa porque consolida deuda histórica de los últimos 10
ciclos.

## Estado al cierre de la sesión anterior (2026-04-20)

**Branch**: `2.0.21` (no merged), 20 commits sobre `main`:
- 15 § clusters shipped (1, 2, 3, 4, 5, 6, 7, 8, 10, 11, 12, 13, 14, 15, 16).
- 4 docs (plan inicial → expand 12 → scale 20 → quality close).
- 1 chore: bump working branch.
- 1 chore: `next.config.ts` rewrites para que `make dev-frontend`
  funcione en local sin Traefik (fix descubierto durante intento de
  testing — login pegaba a Next.js en vez de backend → "Error
  interno").

**Deferred (NO entran en 2.0.21)**: clusters 17-20 (POST credit
rating, supplier demerit, invoice note, work_order_part). Vuelven
en 2.0.22 sólo después de que 1-16 estén pulidos.

**Linter mods stasheados**: `git stash list` muestra "linter mods
pre-2.0.21" (CLAUDE.md dup block + .gitignore + .claude/settings.json).
Pop al post-merge final.

**Phase 0 incidents encontrados durante el boot**:
- Redis container healthy en docker pero `Connection reset by peer`
  desde host — restart resolvió, root cause sin investigar.
- Postgres tenant `dev` con outbox-drainer en loop con SQLSTATE
  25P02 ("current transaction is aborted") — restart limpió pero la
  tabla `outbox` puede tener basura.
- `NEXT_PUBLIC_API_URL=` vacío inyectado por `make dev-frontend`
  sin rewrites → fixeado con `next.config.ts`.

**Findings del audit hoy (2026-04-20)**:
- `apps/web/e2e/erp-mutations.spec.ts`: 4 de 7 tests sólo abren
  dialog y le dan **Cancel** (theater tests).
- `MUTATIONS_ENABLED = !!process.env.E2E_MUTATIONS` (línea 15) →
  toda la suite de mutations skipea por default. CI verde sin
  validar nada.
- Tests que sí crean (catalogos, almacen, tesoreria) verifican el
  toast pero **no que el row exista en DB**.
- Login hardcodeado a `admin@sda.local / admin123` — credentials
  del seed dev, NO existen en DB migrada.
- `components/shadcnblocks/data-table-advanced-3.tsx` con John Doe
  / Jane Smith / Bob Johnson / Alice Williams hardcoded (dead code,
  no importado, pero landmine).
- `components/shadcnblocks/sheet-settings-2.tsx`: `john@example.com`.
- `components/shadcnblocks/field-basic-inputs-1.tsx`: placeholder
  `you@example.com`.
- `EntityContact` y `EntityNote` types tenían fields fabricados
  (`name/role/email/phone` vs schema real `type/label/value`) —
  fixeado en cluster 15 pero hay que asumir que **hay más drift
  similar** en otros types.
- IA del sidebar: **tesorería no es accesible desde la UI**. Sus
  rutas existen (`/administracion/tesoreria/*`) pero sin entry en
  el sidebar. Otras rutas viven bajo el path equivocado. Audit
  completo del registry pendiente.

## Goal

Cerrar 2.0.21 con un PR donde **todo el frontend esté reforzado**:
sidebar coherente, sin mock data, types alineados al schema,
permisos respetados, performance razonable, accesibilidad real,
error handling no-genérico, formularios ergonómicos. Es el primer
ciclo donde se acepta merge sólo si el usuario puede operar la app
contra DB migrada con datos reales sin tropezarse.

## North star — McMaster-Carr velocity, mejor UI

Memoria `feedback_mcmaster_velocity_target` es el benchmark
explícito. El estándar:

- **Page load < 300ms** click → first paint útil.
- **Cero waterfalls** de fetch — paralelizar `useQuery`s, prefetch
  en hover.
- **Filtros instantáneos** — sin spinners visibles para acciones
  < 200ms.
- **Búsqueda predictiva** mientras tipeás, no presioná-Enter.
- **Drill-down sin perder contexto** — filtros + scroll position
  preservados al volver desde detail.
- **Soporte teclado total** — `/` focus búsqueda, `Esc` cierra,
  flechas navegan tablas, Enter abre detail.
- **Optimistic UI** en mutations frecuentes (notas, contactos,
  movimientos).
- **Prefetch en hover** sobre rows clicables.
- **Tablas densas pero legibles** — tabular-nums, sticky header,
  filas cliqueables enteras.
- **UI estéticamente mejor que McMaster** — McMaster es funcional
  pero austero (años '90). SDA = velocidad McMaster + estética
  contemporánea (shadcn / Inter / dark mode pulido).

**Performance budget**: cada PR debe demostrar que no empeora
TTFB / TTI / CLS / LCP. Lighthouse ≥ 95 en todas las pages de
`(modules)/**`.

**Anti-patterns explícitos**:
- "Loading…" placeholder genérico.
- `pageSize=500` con filter client-side.
- Modales sobre modales.
- Confirmaciones para acciones reversibles.
- Re-render del shell al navegar tabs internos.

## Phase 1 — DB migrada local (Phase 0 prerequisite)

Path A (preferido): `pg_dump` desde workstation `srv-ia-01`
(`172.22.100.23`). Memoria `infrastructure-access` cubre VPN +
SSH + creds.

Pasos:
1. Confirmar VPN activa.
2. SSH a workstation: identificar el container Postgres + nombre
   de DB del tenant migrado (probablemente `tenant_saldivia`).
3. `pg_dump --format=custom --no-owner --no-acl
   --exclude-schema=audit ...` — dump del tenant migrado.
4. `scp` el dump a local.
5. En local: crear `tenant_test` DB en `deploy-postgres-1`.
   `pg_restore` el dump.
6. Registrar el tenant en `platform.tenants` para que SDA lo
   reconozca. Crear user de prueba con role admin (o reusar uno
   recuperado del dump).

Path B (fallback): `sda migrate-legacy --mysql-dsn=...
--pg-dsn=... --tenant=test` desde el tunnel a Histrix MySQL.
Más lento (~2.6M rows), pero valida el migrator de paso.

## Phase 2 — IA / Sidebar / Routing audit

**Findings actuales**: tesorería tiene rutas (`cuentas-bancarias/[id]`,
`cajas/[id]`, etc.) pero **no aparece en el sidebar**. Hay rutas
"detrás de" otras secciones que tampoco son alcanzables. Y hay
módulos que viven bajo el padre equivocado.

Pasos:
1. **Inventario completo de rutas**: enumerar todo
   `apps/web/src/app/(modules)/**/page.tsx` → lista de rutas vivas.
2. **Inventario completo de sidebar**: leer
   `apps/web/src/lib/modules/registry.ts` → lista de entries del
   sidebar con sus paths.
3. **Diff**: rutas que están en `app/` pero NO en `registry.ts` =
   **rutas huérfanas** (no se llega desde la UI).
4. **Para cada huérfana**: decidir → agregar al sidebar / mover
   bajo otro parent / borrar si es dead code.
5. **Verificar coherencia con Histrix**: cada area de Histrix
   debería tener su contraparte SDA. `.intranet-scrape/xml-forms/`
   tiene 99 area form-groups — el sidebar debería reflejar las que
   tienen rutas SDA listas.
6. **Resolver paths inconsistentes**:
   - Tesorería: ¿va bajo `/administracion/tesoreria/*` o
     `/tesoreria/*`? Decidir y mover. Hoy está bajo
     `/administracion/tesoreria/*` pero no hay link.
   - Cliente vs cliente-detail: `/administracion/clientes/{id}` vs
     `/compras/proveedores/{id}` (mismo entity diferente role) —
     mantener separados pero verificar consistencia visual.
   - Almacén está duplicado en sidebar y dentro de `compras` —
     decidir.
7. **Breadcrumbs**: cada `[id]/page.tsx` debe tener `<Link href=...
   "Volver a X">`. Verificar que apunten al sidebar entry correcto.
8. **Output esperado**: una nueva entrada en cada `registry.ts`
   por cluster huérfano, paths consolidados, breadcrumbs
   coherentes.

## Phase 3 — Mock data extermination

Ya identificados (audit hoy):

| File | Action |
|---|---|
| `components/shadcnblocks/data-table-advanced-3.tsx` | Borrar (dead) |
| `components/shadcnblocks/sheet-settings-2.tsx` | Borrar (dead) |
| `components/shadcnblocks/field-basic-inputs-1.tsx` | Borrar (dead) |
| `components/shadcnblocks/alert-dialog-form-3.tsx` | Verificar uso, borrar si dead |
| `components/shadcnblocks/combobox-grouped-2.tsx` | Verificar uso, borrar si dead |
| `components/reactbits/Antigravity.tsx` | Verificar uso, borrar si dead |
| `components/reactbits/CircularGallery.tsx` | Verificar uso, borrar si dead |

Pasos extra:
1. `grep -rE "John Doe|Jane Doe|@example\.com|Lorem|placeholder.*name"
   apps/web/src/` — ampliar el sweep.
2. Buscar arrays hardcoded de >5 entries en cualquier
   `app/**/page.tsx` (debería usar `useQuery`, no array literal).
3. Storybook / `*.stories.tsx` / `*.fixture.ts`: si existen,
   reemplazar mock props con data real (o eliminar).
4. Seed de dev: si `services/erp/internal/seed*` o similar inserta
   rows ficticios, marcar deprecated. La DB local arranca vacía y
   se llena vía migrator/dump (memoria `feedback_no_mock_data`).
5. Documentar en CLAUDE.md la política "no mock data" para que el
   próximo cambio respete la regla.

## Phase 4 — E2E suite reform completo

La suite actual (5 specs, 761 LOC) **no sirve** para validar
calidad. Reescribir.

Pasos:
1. **Eliminar `E2E_MUTATIONS` skip flag**. Todos los tests corren
   siempre. Si un test es lento, marcarlo `@slow`, no skipear.
2. **Convertir tests que cancelan en tests que crean**: contabilidad
   asiento, calidad NC, manufactura unidad, seguridad accidente —
   los 4 deben crear el row, verificar la response, y cleanup en
   `afterEach`.
3. **Validar side effects**: cada test que crea, después debe leer
   (vía API o navegando al detail) y verificar que el row existe
   con los campos correctos. Toast no es evidencia.
4. **Login con creds del tenant migrado**: parametrizar con env
   vars (TEST_EMAIL/TEST_PASSWORD ya están en
   `e2e/api/helpers.ts`). Default a un user que existe en la DB
   migrada.
5. **Network call assertions**: para los filter endpoints (clusters
   1, 2, 3), `page.waitForRequest(url => url.includes(
   'bank_account_id='))`. Esto valida que el frontend manda el
   filter, no que pasa por casualidad.
6. **Agregar 1 spec nuevo por cluster shipped** (15 nuevos):

| Cluster | Spec |
|---|---|
| 1 — filter reconciliations | `e2e/treasury-reconciliations.spec.ts` |
| 2 — filter cash-counts | `e2e/treasury-cash-counts.spec.ts` |
| 3 — filter entries + UI | `e2e/accounting-cost-center.spec.ts` |
| 4 — GetTool | `e2e/almacen-tool-detail.spec.ts` |
| 5 — GetAsset | `e2e/maintenance-asset-detail.spec.ts` |
| 6 — GetChassisModel | `e2e/manufacturing-chassis-model.spec.ts` |
| 7 — GetCarroceriaModel + BOM | `e2e/manufacturing-carroceria-model.spec.ts` |
| 8 — GetScorecard | `e2e/quality-scorecard.spec.ts` |
| 10 — GetUnit | `e2e/manufacturing-unit-detail.spec.ts` |
| 11 — GetActionPlan | `e2e/quality-action-plan.spec.ts` |
| 12 — GetNC | `e2e/quality-nc.spec.ts` |
| 13 — GetArticle direct | extender `erp-navigation.spec.ts` |
| 14 — POST entity note | extender `erp-mutations.spec.ts` |
| 15 — POST entity contact | extender `erp-mutations.spec.ts` |
| 16 — POST inventory movement | extender `erp-mutations.spec.ts` |

Cada spec cubre: **happy path + 1 error path + 1 empty state +
1 network assertion**. Sin smoke tests vacíos.

## Phase 5 — Type-vs-schema drift audit

`EntityContact` y `EntityNote` ya se descubrieron con fields
fabricados. **Asumir que hay más** — los types frontend NO se
generan automáticamente desde el backend.

Pasos:
1. Listar todos los `interface X` en `apps/web/src/lib/erp/types.ts`.
2. Para cada uno: comparar contra el struct correspondiente en
   `services/erp/internal/repository/models.go` (sqlc-generated).
3. Reportar drift: campos faltantes, fabricados, mal-tipados.
4. Fix: actualizar types frontend al schema real. Re-typecheck +
   re-test.
5. **Long-term**: considerar generar types frontend desde el
   backend (sqlc → JSON schema → ts-types) para que el drift sea
   imposible. Out of scope este ciclo, pero documentar en TODO.

## Phase 6 — Permissions UI

Cada acción de escritura debe respetar `erp.X.write` (o equivalent).
Si el user no tiene el permiso, el botón debe estar **hidden o
disabled**, no mostrar el form que después falla con 403.

Pasos:
1. Listar todas las llamadas a `api.post / api.patch / api.delete`
   en `apps/web/src/`.
2. Para cada una: verificar que el componente que la dispara
   chequea el permiso ANTES de mostrar el botón.
3. Patrón propuesto: hook `useHasPermission("erp.stock.write")`
   que lea del JWT/auth store. Si NO lo hay, hide el botón.
4. Tests: spec por permiso (user sin perm → botón no visible).

## Phase 7 — Performance audit (McMaster-grade)

Norte: page load <300ms, cero waterfalls, prefetch en hover,
optimistic UI. Memoria `feedback_mcmaster_velocity_target`.

Pasos:
1. **Lighthouse baseline**: correr Lighthouse sobre cada page de
   `(modules)/**`. Anotar TTFB / TTI / CLS / LCP. Cualquier page
   con score < 95 va a la lista de fixes.
2. **`grep -rE "page_size=(500|200|100)" apps/web/src/`** → cada
   match es swap obligatorio (endpoint directo o pagination
   server-side).
3. **Waterfall audit**: cada `[id]/page.tsx` con >1 `useQuery` →
   verificar que se disparan en paralelo (no en cascada via
   `enabled` chain). Si hay cascada legítima, considerar mover el
   join al backend.
4. **Prefetch en hover**: implementar handler global en `<Link>`
   para list rows. `onMouseEnter` → `queryClient.prefetchQuery`.
   Patrón reusable.
5. **Optimistic UI** en clusters 14, 15, 16 (notas, contactos,
   movimientos). Mutation con `onMutate` que actualiza la cache
   antes del response, `onError` revierte con explicación.
6. **Cache invalidation audit**: cada mutation debe invalidar
   queries afectadas. Bug: si crear contacto no invalida la entity
   query, el contador no se actualiza.
7. **`enabled: !!id`** debe ser estándar en queries dependientes.
   Buscar queries que corran sin gate.
8. **N+1 patterns**: pages que hacen un query por row de la lista.
9. **Búsqueda predictiva**: cada filtro/búsqueda con debounce
   ~150ms + resultados parciales. No esperar Enter.
10. **Soporte teclado**: implementar atajos globales (`/` busca,
    `Esc` cierra, `g+l` jumplist). Tab order coherente. Tabla con
    flechas + Enter para abrir detail.
11. **Drill-down sin perder contexto**: al volver de detail a list,
    filtros y scroll position se preservan. Usar URL params para
    state.
12. **Tablas densas**: revisar padding, tipografía. Sticky header.
    Filas enteras cliqueables (no sólo el primer celda). Tabular-
    nums en columnas numéricas.

## Phase 8 — Accessibility (a11y)

`modules-accessibility.spec.ts` ya existe pero hay que ver qué
cubre. Probablemente sólo navegación con tab.

Pasos:
1. Leer el spec actual.
2. Agregar: form labels (cada input con `<Label>`), heading
   hierarchy (h1 → h2 → h3 sin saltos), color contrast, focus
   visible, error msgs asociados al input via aria-describedby.
3. Correr `axe-core` (Playwright tiene plugin) sobre cada page de
   `(modules)/**`. Reportar violations.

## Phase 9 — Spanish UI / Error handling

"Error interno" para todo es inaceptable (memoria
`feedback_quality_over_quantity`).

Pasos:
1. `grep -r "Error interno" apps/web/src/` → todos los catch
   genéricos. Reemplazar con switch sobre `err.status`.
2. Inventario de buttons/headers/labels: consistencia
   ("Agregar" vs "Crear" vs "Nuevo" vs "+"). Decidir convención y
   homogeneizar.
3. Empty states: cada tabla con 0 rows debe tener mensaje
   explicativo, no "Sin datos". Idealmente con un CTA.
4. Confirmaciones destructivas: delete, override, reset → modal
   con "¿Estás seguro?" + descripción del impacto.
5. Loading states: skeletons consistentes (no spinner en una page,
   skeleton en otra).

## Phase 10 — Polish loop por cluster

Criteria del feedback memory `feedback_quality_over_quantity`:

- **Validación de form**: required, formato (UUID/email/numérico),
  rangos. Mensajes en español, claros.
- **Defaults inteligentes**: fecha = hoy, tipo = más común,
  unidad = "und" para movements, etc.
- **Submit deshabilitado** mientras inválido o `isPending`.
- **Mensaje de error post-fail** claro, distinguir 401/403/422/5xx.
- **Empty state explicativo**.
- **Confirmación destructiva**.
- **Loading state visible**.

Para cada cluster (1-16) que falla algún criterio: fix + commit
dentro de `2.0.21`. TDD: spec primero, después fix.

## Phase 11 — Cierre 2.0.21

**Sólo cuando**:
- Todos los E2E (existentes + nuevos) pasan contra DB migrada.
- Sidebar coherente (no rutas huérfanas, ningún módulo en parent
  equivocado).
- Mock data eliminado del repo (audit con grep verde).
- Type drift cero contra `models.go`.
- Permisos UI respetados (ningún botón visible sin la perm).
- Performance: Lighthouse ≥ 95 en todas las pages de `(modules)/**`.
  Ningún `page_size>=100` que se pueda evitar. Prefetch en hover
  sobre rows. Optimistic UI en mutations frecuentes. Cero
  waterfalls de fetch en detail pages.
- A11y: axe-core verde en todas las pages de `(modules)/**`.
- Error handling: cero "Error interno" residuales.
- Manual smoke por área (compras, tesorería, contabilidad,
  almacén, mantenimiento, manufactura, calidad, ingeniería) hecho
  por el usuario con tenant migrado.

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
- **Refactor estructural** (admin Tier B, fusion ERP).
- **Phase 2+** (chat agent, prompts jerárquicos, tree-RAG, ACL).

## Trampas heredadas

- **Redis crash silencioso** — investigar root cause.
- **Postgres outbox-drainer aborted state** — `TRUNCATE outbox`
  en tenant `dev` antes de empezar.
- **NEXT_PUBLIC_API_URL vacío + rewrites de dev** — verificar que
  no rompa el build de prod.
- **Mock interfaces must track**: cada nuevo método de handler
  rompe `*_test.go` mocks.
- **sqlc regen drift**: editar generated `*.sql.go` a mano.
- **cwd persiste entre Bash calls**: usar rutas absolutas.
- **Linter mods stasheados**: pop al post-merge final.

## Candidatos sesiones futuras (2.0.22+)

| Orden | Tema | Pre-req |
|---:|---|---|
| 1 | **Write scouts 17-20 con E2E desde día 1** | 2.0.21 cerrado |
| 2 | **Reports** (Libro IVA, mayor contable, tax-book) | 2.0.21 cerrado |
| 3 | **Seamless-day cutover test** | E2E coverage estable |
| 4 | **Admin Tier B refactor** (deuda 2.0.17) | scope independiente |
| 5 | **Bulk write paths** | scout de 2.0.22 validado |
| 6 | **Codegen types frontend desde sqlc** | Phase 5 type drift confirmed solucionable mecánicamente |
