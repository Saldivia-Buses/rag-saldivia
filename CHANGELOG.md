# Changelog

Todos los cambios notables de este proyecto se documentan en este archivo.

Formato basado en [Keep a Changelog](https://keepachangelog.com/es/1.1.0/).
Versionado basado en [Semantic Versioning](https://semver.org/lang/es/).

---

## [Unreleased]

### Added
- Vercel AI SDK (`ai@6`, `@ai-sdk/react@3`) for chat streaming (Plan 14)
- `ai-stream.ts` adapter: transforms NVIDIA SSE to AI SDK Data Stream protocol (Plan 14)
- 6 documentation templates: plan, commit, PR, version, ADR, artifact (Plan 13)
- ADR-012: stack definitivo for 1.0.x series (Plan 13)
- `docs/artifacts/` directory for review/audit results (Plan 13)

### Changed
- ChatInterface migrated from manual SSE to `useChat` from AI SDK (Plan 14)
- `/api/rag/generate` now returns AI SDK Data Stream protocol (Plan 14)
- All 10 Claude Code agents rewritten for TypeScript stack (Plan 13)
- NavRail simplified to 3 links: Chat, Collections, Settings (Plan 13)
- CLAUDE.md rewritten: 591 → 312 lines, only active code described (Plan 13)
- README.md rewritten to reflect core-only codebase (Plan 13)
- ADRs 001, 003, 008, 011 updated with current status notes (Plan 13)
- commitlint: added `style` type, `agents`/`ui`/`release` scopes (Plan 13)

### Removed
- ~100 aspirational files archived to `_archive/` (Plan 13)
- CLI app archived (Plan 13)
- 12 admin components + routes archived (Plan 13)
- 12 aspirational chat components archived (Plan 13)
- Dead `revalidatePath` calls and `/share/` public route (Plan 13)
- `useRagStream` hook archived, replaced by AI SDK `useChat` (Plan 14)
- `detect-artifact.ts` archived with artifacts feature (Plan 14)

## [1.0.0] — 2026-03-27

Primer release del stack TypeScript de RAG Saldivia. Reescritura completa del overlay
sobre NVIDIA RAG Blueprint v2.5.0 — reemplaza el stack Python + SvelteKit con un proceso
único Next.js 16 que incluye UI, autenticación, proxy RAG, admin y CLI TypeScript.

### Highlights

- **Next.js 16** App Router como proceso único — UI + auth + proxy + admin
- **Autenticación JWT** con Redis blacklist para revocación inmediata + RBAC por roles y áreas
- **BullMQ** para cola de ingesta — reemplaza worker manual + tabla SQLite
- **Design system "Warm Intelligence"** — 24 páginas, dark mode, WCAG AA
- **CLI TypeScript** — `rag users/collections/ingest/audit/config/db/status`
- **413+ tests** — lógica, componentes, visual regression, a11y, E2E Playwright
- **Código production-grade** — TypeScript strict, ESLint, commitlint, lint-staged, knip
- **10 ADRs** documentando las decisiones de arquitectura

### Plans completados (Plan 1 → Plan 11)

**Plan 1 — Monorepo TypeScript**
Birth del stack. Turborepo + Bun workspaces + Next.js 15 + Drizzle + JWT auth + CLI base.

**Plan 2 — Testing sistemático**
Primera suite de tests. 270 tests de lógica en verde. Estrategia de testing documentada.

**Plan 3 — Bugfix CodeGraphContext**
Estabilización post-birth. Fixes de imports, build, y MCP.

**Plan 4 — Product Roadmap (Fases 0–2)**
50 features en 3 fases. Design system base, dark mode, 24 páginas, shadcn/ui, design system.

**Plan 5 — Testing Foundation**
Coverage al 95% en lógica pura. 270 tests.

**Plan 6 — UI Testing Suite**
Visual regression con Playwright (22 snapshots). A11y con axe-playwright (WCAG AA).

**Plan 7 — Design System "Warm Intelligence"**
Paleta crema-navy, tokens CSS, 147 tests de componentes, Storybook 8.

**Plan 8 — Optimización + Redis + BullMQ**
~2.516 líneas eliminadas. Next.js 16, Zod 4, Drizzle 0.45, Lucide 1.7.
Redis obligatorio: JWT revocación, cache, Pub/Sub, BullMQ.
10 ADRs. CI paralelo con turbo --affected.

**Plan 9 — Repo Limpio**
46 archivos purgados del remoto. TypeScript a 0 errores.
Dead code eliminado (crossdoc, SSO stub, wrappers). ESLint + husky + commitlint.

**Plan 10 — Testing Completo**
E2E Playwright (5 flujos críticos + smoke Redis).
Visual regression verificada post-upgrades. A11y WCAG AA. Coverage ≥80%.

**Plan 11 — Documentación Perfecta**
README, CONTRIBUTING, SECURITY, LICENSE, CODEOWNERS, issue templates.
ER diagram, API reference (30+ endpoints), JSDoc en funciones críticas.
READMEs de packages. CLAUDE.md actualizado.

### Plan 11 — Documentación (README, CONTRIBUTING, API, packages, JSDoc, CLAUDE, docs/)

#### Added
- `README.md` — reescrito (badges, Quick Start, arquitectura, stack, features v1 y futuras, ≥300 líneas)
- `CONTRIBUTING.md` — setup, tests, commits, PR, patrones de código
- `SECURITY.md`, `LICENSE` (MIT, 2026)
- `.github/CODEOWNERS`, `.github/ISSUE_TEMPLATE/bug_report.md`, `feature_request.md`
- `packages/db/README.md` — diagrama ER Mermaid + lista de queries
- `packages/logger/README.md`, `packages/shared/README.md`, `packages/config/README.md`, `apps/cli/README.md`
- `docs/api.md` — referencia de endpoints HTTP verificados
- `apps/web/src/middleware.ts` — re-export de `proxy` como middleware de Next.js

#### Changed
- `CLAUDE.md`, `docs/architecture.md`, `docs/onboarding.md`, `docs/workflows.md`, `docs/testing.md`, `docs/blackbox.md` — precisión post Plan 8–10 (Next 16, Redis, hooks, sin crossdoc/next-safe-action obsoletos)
- `docs/plans/ultra-optimize-plan11-documentation.md` — checklist del Plan 11 marcada según lo verificado (pendientes explícitos: Quick Start manual, push, `cli.md` / `design-system.md`, etc.)

#### Documentation (JSDoc)
- Funciones críticas: `getRedisClient`, `nextSequence`, `createJwt`, `extractClaims`, `ragFetch`, `getCachedRagCollections`, `startIngestionWorker`, `proxy`, `persistEvent` (logger), `reconstructFromEvents`

### Plan 10 — Testing completo (visual, a11y, cobertura, E2E, smoke Redis)

#### Added
- `apps/web/playwright.e2e.config.ts` y `apps/web/tests/e2e-playwright/auth.spec.ts`, `chat.spec.ts`, `admin-users.spec.ts`, `upload.spec.ts`, `settings.spec.ts`, `redis-smoke.spec.ts` — flujos críticos con Playwright (MOCK_RAG, Redis en CI) — *(Plan 10 F10.4–F10.5)*
- Script `apps/web`: `dev:webpack`, `test:e2e` — Next dev con webpack para Playwright; E2E en CI (`e2e` job) — *(Plan 10)*
- CI: umbral de cobertura de líneas ≥80% en `packages/db`; job `e2e`; Redis + migrate/seed + `NEXT_PUBLIC_DISABLE_REACT_SCAN` en audit de accesibilidad — *(Plan 10 F10.4–F10.5)*
- Badges en `README.md`: cobertura DB y CI — *(Plan 10 F10.3)*

#### Changed
- `--fg-subtle` (light) en `globals.css` para contraste WCAG AA; `login`: landmark `<main>`; `ReactScanProvider` respeta `NEXT_PUBLIC_DISABLE_REACT_SCAN` — *(Plan 10 F10.2)*
- `NavRail` / `SessionList`: `aria-label` en logout y «Nueva sesión» — tests E2E — *(Plan 10)*
- `actionLogout`: revoca JWT en Redis, borra cookie `auth_token` (antes `token`) — *(Plan 10)*

#### Removed
- `apps/web/src/middleware.ts` — Next.js 16 solo usa `proxy.ts` (conflicto middleware+proxy) — *(Plan 10 / convención Next 16)*

#### Baseline visual
- `apps/web/.gitignore`: deja de ignorar `tests/visual/snapshots/` — los 22 PNG de regresión visual quedan versionados (CI y clones tienen baseline) — *(F10.1)*
- `docs/plans/ultra-optimize-plan10-testing.md` — Plan 10; `tokens-palette-*.png` actualizados tras `--fg-subtle` — *(F10.1)*

### Plan 9 — Repo Limpio (completado 2026-03-27)

#### F9.1 — Git purge + `.gitignore`
##### Removed (destrackeado del índice)
- `.playwright-mcp/`, `.superpowers/`, `apps/web/logs/backend.log`, `config/.env.saldivia`, `docs/superpowers/` — artefactos MCP, brainstorming interno, logs y env que no debían versionarse — *(Plan 9 F9.1)*

##### Changed
- `.gitignore`: reglas para `.playwright-mcp/`, `.superpowers/`, `apps/web/logs/`, `config/*.env.*`, `docs/superpowers/`, y archivos de sesión `*.pid`, `*.server-info`, `*.server-stopped` — *(Plan 9 F9.1)*

#### F9.2 — TypeScript sin errores
##### Fixed
- `apps/web/src/app/actions/collections.ts`: `invalidateCollectionsCache()` + `revalidatePath` en lugar de APIs de cache incompatibles con Next.js 16 — *(Plan 9 F9.2)*
- `packages/logger/src/blackbox.ts`: spread condicional en `handleIngestion*` para `exactOptionalPropertyTypes` — *(Plan 9 F9.2)*
- `packages/db/src/ioredis-mock.d.ts`: `declare module "ioredis-mock"` — *(Plan 9 F9.2)*
- `packages/db/tsconfig.json`: excluir `src/test-setup.ts` del compilado — *(Plan 9 F9.2)*

##### Removed (git untrack previo)
- `.turbo/cache/` (408 archivos): `git rm --cached -r .turbo/` — *(Plan 9 F9.2 / cierre Plan 8)*

#### F9.3 — Dead code y actions huérfanas
##### Removed
- Hooks crossdoc (`useCrossdocStream`, `useCrossdocDecompose`), `SplitView`, `CollectionHistory`, SSO (`next-auth`, ruta `[...nextauth]`, `SSOButton` + test), `safe-action.ts`, `form.ts`, `scripts/health-check.ts` — *(Plan 9 F9.3)*
- Server actions sin usos: `actionListAreas`, `actionListSessions`, `actionGetSession`, `actionGetRagParams`, `actionResetOnboarding`, `actionListUsers`, `actionAssignArea`, `actionRemoveArea`, `actionUpdatePassword` (admin en `users.ts`) — la UI usa `loadRagParams` / `actionUpdatePassword` de settings donde aplica — *(Plan 9 F9.3)*

##### Changed
- `apps/web/src/lib/auth/current-user.ts`: `getCurrentUser` deja de exportarse (solo uso interno en `requireUser`) — *(Plan 9 F9.3)*
- `apps/web/src/app/(auth)/login/page.tsx`: flujo solo email/contraseña — *(Plan 9 F9.3)*
- `README.md`, `scripts/README.md`: referencias a `health-check.ts` sustituidas por `rag status` / `/api/health` — *(Plan 9 F9.3)*

#### F9.4 — Dependencias limpias
##### Changed
- `apps/web`: removidos `next-safe-action`, `d3`, `@types/d3`, `next-auth` — sin consumidores tras F9.3 — *(Plan 9 F9.4)*
- `apps/cli`: removido `@rag-saldivia/shared` — sin imports — *(Plan 9 F9.4)*
- `apps/web`: `postcss` declarado como devDependency (uso en `postcss.config.js`) — *(Plan 9 F9.4)*

#### F9.5 — ESLint
##### Added
- `apps/web/eslint.config.js`: flat config con `eslint-config-next/core-web-vitals`, reglas `no-console`, `@typescript-eslint/*`, desactivación acotada de reglas React Compiler ruidosas — *(Plan 9 F9.5)*
- `apps/web/package.json`: script `lint:eslint` — *(Plan 9 F9.5)*

##### Changed
- `apps/web`: ESLint 9.x pinneado; correcciones de unused vars, comillas JSX, exhaustive-deps documentados — *(Plan 9 F9.5)*

#### F9.6 — Husky, commitlint, lint-staged
##### Added
- `commitlint.config.js`, `.lintstagedrc.js`, `.husky/pre-commit` (lint-staged), `.husky/commit-msg` (commitlint con `bunx`) — *(Plan 9 F9.6)*
- `lint-staged` en el root; `prepare`: `husky` — *(Plan 9 F9.6)*

##### Changed
- `.lintstagedrc.js`: solo `apps/web/src/**/*.{ts,tsx}` con ESLint (evita lint masivo en `packages/*` sin config dedicada) — *(Plan 9 F9.6)*

#### F9.7 — `console.log` y calidad
##### Changed
- `packages/db/src/init.ts`, `packages/db/src/seed.ts`: `console.warn` en lugar de `console.log` — *(Plan 9 F9.7)*
- `packages/logger/src/backend.ts`: salida no-error vía `process.stdout.write` — sin `console.log` — *(Plan 9 F9.7)*
- `packages/logger/src/__tests__/logger.test.ts`: mocks de `process.stdout.write` — *(Plan 9 F9.7)*

#### F9.8 — Knip
##### Changed
- `knip.json`: entradas por workspace, `ignore` para tests/storybook/data-table/error-boundary, `ignoreDependencies` (Radix, testing, `@tanstack/react-table`), `ignoreIssues` para exports intencionales — `bunx knip` exit 0 — *(Plan 9 F9.8)*

#### Otros
##### Changed
- `apps/web/src/lib/rag/client.ts`: combinar `AbortSignal` externo con timeout — *(Plan 9)*
- `bun.lock`: resolución tras cambios de dependencias — *(Plan 9)*

---

### Plan 8 — Optimización (Fases 0–8 completadas)

#### Fase 8 — Redis como dependencia requerida + BullMQ (2026-03-27)

##### Added
- `docs/decisions/010-redis-required.md`: ADR-010 — Redis como dependencia del sistema, motivo, primitivas usadas — *(Plan 8 F8.22)*
- `docker-compose.yml`: servicio Redis (`redis:alpine`) con healthcheck — *(Plan 8 F8.22)*
- `packages/db/src/redis.ts`: cliente Redis singleton `getRedisClient()` con fail-fast si `REDIS_URL` no configurado + `_resetRedisForTesting()` — *(Plan 8 F8.23)*
- `packages/db/src/test-setup.ts`: preload de bun:test que activa ioredis-mock para tests unitarios — *(Plan 8 F8.23)*
- `packages/db/bunfig.toml`: preload del test-setup para todos los tests de `packages/db` — *(Plan 8 F8.23)*
- `packages/db/src/__tests__/redis.test.ts`: 4 tests del cliente Redis singleton — *(Plan 8 F8.23)*
- `apps/web/src/app/api/notifications/stream/route.ts`: endpoint SSE via Redis Pub/Sub — elimina polling cada 30s — *(Plan 8 F8.28)*
- `apps/web/src/lib/queue.ts`: BullMQ — definición de `ingestionQueue` (Queue), `createQueueEvents()`, `startIngestionWorker()`, `scheduleToPattern()` — *(Plan 8 F8.30)*
- `.github/workflows/ci.yml`: `services: redis` + `REDIS_URL` en el job `test-logic` — *(Plan 8 F8 cierre)*

##### Changed
- `.env.example`: `REDIS_URL` marcado como `[REQUIRED-DEV]` — *(Plan 8 F8.22)*
- `apps/web/src/app/api/health/route.ts`: verifica Redis via `getRedisClient().ping()` — retorna 503 si está caído — *(Plan 8 F8.22)*
- `packages/db/src/queries/events.ts`: `nextSequence()` usa Redis `INCR events:seq` — elimina variable `_seq` en memoria — *(Plan 8 F8.24)*
- `packages/shared/src/schemas.ts`: `jti: z.string().optional()` en `JwtClaimsSchema` — *(Plan 8 F8.25)*
- `apps/web/src/lib/auth/jwt.ts`: `createJwt` agrega `.setJti(crypto.randomUUID())`; `extractClaims` verifica blacklist Redis antes de retornar claims — *(Plan 8 F8.25)*
- `apps/web/src/proxy.ts`: propaga header `x-user-jti` para que route handlers verifiquen revocación — *(Plan 8 F8.25)*
- `apps/web/src/app/api/auth/logout/route.ts`: escribe `SET revoked:{jti} 1 EX {ttl}` en Redis al hacer logout — *(Plan 8 F8.25)*
- `apps/web/src/workers/external-sync.ts`: master election via Redis `SET NX EX` — evita duplicar sync en múltiples instancias — *(Plan 8 F8.26)*
- `apps/web/src/lib/rag/collections-cache.ts`: `getCachedRagCollections()` via Redis `GET/SET EX` — elimina `unstable_cache`; `invalidateCollectionsCache()` via `DEL` — *(Plan 8 F8.27)*
- `apps/web/src/app/api/rag/collections/route.ts`: llama `invalidateCollectionsCache()` después de POST — *(Plan 8 F8.27)*
- `apps/web/src/hooks/useNotifications.ts`: usa SSE via `EventSource("/api/notifications/stream")` — elimina localStorage y polling — *(Plan 8 F8.28)*
- `packages/logger/src/rotation.ts`: `getLogFileSize/setLogFileSize` via Redis `HSET/HGET log:sizes` — elimina `_sizeCache Map` en memoria — *(Plan 8 F8.29)*
- `apps/web/src/workers/ingestion.ts`: solo contiene lógica de negocio pura (`processJob`, `processScheduledReport`) + arranca BullMQ worker — elimina `workerLoop`, `processWithRetry`, `tryLockJob`, `setInterval`, signal handlers — *(Plan 8 F8.30)*
- `apps/web/src/app/api/upload/route.ts`: usa `ingestionQueue.add()` (BullMQ) en lugar de INSERT en `ingestion_queue` SQLite — *(Plan 8 F8.30)*
- `apps/web/src/app/api/admin/ingestion/stream/route.ts`: usa BullMQ `QueueEvents` en lugar de polling SQLite cada 3s — *(Plan 8 F8.30)*
- `apps/web/src/app/api/admin/ingestion/route.ts`: lee jobs desde BullMQ en lugar de `ingestion_queue` SQLite — *(Plan 8 F8.30)*
- `apps/web/src/app/api/admin/ingestion/[id]/route.ts`: cancela jobs via `job.remove()` de BullMQ — *(Plan 8 F8.30)*
- `apps/web/src/app/(app)/admin/system/page.tsx`: lee jobs activos desde BullMQ — *(Plan 8 F8.30)*
- `apps/web/src/components/admin/SystemStatus.tsx`: usa tipo `ActiveJob` en lugar de `DbIngestionQueueItem` — *(Plan 8 F8.30)*
- `apps/web/src/lib/test-setup.ts`: agrega mock de ioredis con ioredis-mock para tests de web — *(Plan 8 F8.23)*
- `apps/web/src/lib/auth/__tests__/jwt.test.ts`: test `"createJwt incluye jti único por token"` — *(Plan 8 F8.25)*

##### Removed
- `packages/db/src/schema.ts`: tabla `ingestion_queue` eliminada — BullMQ reemplaza toda la funcionalidad — *(Plan 8 F8.30)*
- `packages/db/src/schema.ts`: tipos `DbIngestionQueueItem` y `NewIngestionQueueItem` eliminados — *(Plan 8 F8.30)*

##### Infrastructure
- `ioredis@5.10.1` agregado como dependencia de `packages/db` y `apps/web` — *(Plan 8 F8.23)*
- `ioredis-mock@8.13.1` agregado como devDependency de `packages/db` y `apps/web` — *(Plan 8 F8.23)*
- `bullmq@5.71.1` agregado como dependencia de `apps/web` — *(Plan 8 F8.30)*

**Código eliminado en Fase 8:** ~350 líneas (workarounds in-memory + tabla `ingestion_queue` + `workerLoop` manual)

---

### Plan 8 — Optimización (Fases 1, 2, 3, 4, 5, 6 y 7 completadas)

#### Fase 7 — Mejoras al sistema de logging y Black Box (2026-03-27)

##### Added
- `packages/shared/src/schemas.ts`: `"system.request"` agregado a `EventTypeSchema` — tipo semánticamente correcto para requests HTTP — *(Plan 8 F7.17)*
- `packages/logger/src/backend.ts`: campo `requestId?: string` en `LogContext` para correlación de logs por request — *(Plan 8 F7.17)*
- `apps/web/src/middleware.ts`: archivo entry point de Next.js middleware — re-exporta `proxy` como `middleware` — *(Plan 8 F7.17)*
- `apps/web/src/proxy.ts`: generación de `x-request-id` UUID en cada request (público y autenticado) — *(Plan 8 F7.17)*
- `packages/logger/src/blackbox.ts`: tipo `IngestionEventRecord` y campo `ingestionEvents` en `ReconstructedState` — *(Plan 8 F7.18)*
- `packages/logger/src/blackbox.ts`: handlers para `rag.stream_started`, `rag.stream_completed`, `ingestion.started`, `ingestion.completed`, `ingestion.failed`, `ingestion.stalled` en `EVENT_HANDLERS` — *(Plan 8 F7.18)*
- `packages/logger/src/blackbox.ts`: sección "Ingestas" en `formatTimeline` cuando hay eventos de ingesta — *(Plan 8 F7.18)*
- `packages/logger/src/__tests__/logger.test.ts`: 7 tests nuevos para handlers de ingestion y RAG stream, test de `log.request` con `system.request`, test de sección Ingestas en `formatTimeline` — *(Plan 8 F7.17, F7.18)*
- `packages/db/src/queries/events-cleanup.ts`: función `deleteOldEvents(olderThanDays?)` — elimina eventos más viejos que el cutoff, respeta `LOG_RETENTION_DAYS` env var — *(Plan 8 F7.19)*
- `packages/db/src/__tests__/events.test.ts`: 2 tests nuevos para `deleteOldEvents` — *(Plan 8 F7.19)*
- `.env.example`: variable `LOG_RETENTION_DAYS` documentada (default 90 días) — *(Plan 8 F7.19)*

##### Changed
- `packages/logger/src/backend.ts`: `log.request()` corregido de `"system.warning"` → `"system.request"` — *(Plan 8 F7.17)*
- `packages/logger/src/backend.ts`: `formatPretty` descompuesto en 4 funciones puras exportables: `formatHeader`, `formatContext`, `formatPayloadSummary`, `formatSuggestion` — complejidad ciclomática 29 → < 10 — *(Plan 8 F7.21)*
- `packages/logger/src/backend.ts`: `formatContext` incluye `requestId` truncado (primeros 8 chars) cuando está presente — *(Plan 8 F7.17)*
- `packages/db/src/schema.ts`: índice compuesto `idx_events_query` en `(type, userId, ts)` — convierte analytics queries en index scan O(log n) — *(Plan 8 F7.19)*
- `packages/db/src/index.ts`: exporta `deleteOldEvents` desde `queries/events-cleanup` — *(Plan 8 F7.19)*
- `apps/web/src/workers/ingestion.ts`: integra `deleteOldEvents` en limpieza diaria (setInterval 24h) — *(Plan 8 F7.19)*
- `apps/web/src/workers/external-sync.ts`: todos los `log.info("system.warning", ...)` reemplazados por tipos semánticamente correctos (`ingestion.started`, `ingestion.completed`, `ingestion.failed`, `system.error`) — *(Plan 8 F7.17)*
- `apps/web/src/app/api/audit/export/route.ts`: soporte `?format=json|csv` — CSV generado con `papaparse` (RFC 4180 compliant) — *(Plan 8 F7.20)*
- `apps/web/src/components/audit/AuditTable.tsx`: botones "Exportar CSV" y "Exportar JSON" visibles para admins — *(Plan 8 F7.20)*
- `apps/web/src/components/admin/KnowledgeGapsClient.tsx`: CSV export reemplazado con `Papa.unparse()` — elimina escaping manual propenso a bugs — *(Plan 8 F7.20)*

---

### Plan 8 — Optimización (Fases 1, 2, 3, 4, 5 y 6 completadas)

#### Fase 6 — Error Boundaries y CI paralelo (2026-03-27)

##### Added
- `apps/web/src/components/error-boundary.tsx`: componente `<ErrorBoundary>` reutilizable basado en clase React con estado `hasError`, soporte de `fallback` personalizado y callback `onReset` — *(Plan 8 F6.15)*
- `apps/web/src/app/(app)/chat/error.tsx`: Error Boundary de Next.js App Router para la ruta `/chat` — sanitiza mensajes en producción — *(Plan 8 F6.15)*
- `apps/web/src/app/(app)/admin/error.tsx`: Error Boundary para todo el panel de administración — *(Plan 8 F6.15)*
- `apps/web/src/components/__tests__/error-boundary.test.tsx`: 7 tests de componente para `<ErrorBoundary>` — *(Plan 8 F6.15)*

##### Changed
- `.github/workflows/ci.yml`: jobs separados y paralelos (`type-check`, `test-logic`, `test-components`, `lint`, `coverage`) en lugar de un único job secuencial — *(Plan 8 F6.16)*
- `.github/workflows/ci.yml`: `actions/cache@v4` para `~/.bun/install/cache` en todos los jobs — ahorra 30–60 s por job — *(Plan 8 F6.16)*
- `.github/workflows/ci.yml`: `bunx turbo run test --affected --filter="...[HEAD^1]"` en PRs — solo testea packages afectados; push a dev sigue corriendo la suite completa — *(Plan 8 F6.16)*
- `.github/workflows/ci.yml`: `visual-regression` y `accessibility` pasan a requerir `needs: [test-logic]` — no corren si los tests de lógica fallan — *(Plan 8 F6.16)*
- `turbo.json`: tarea `test:components` agregada al pipeline de Turborepo — *(Plan 8 F6.16)*

---

### Plan 8 — Optimización (Fases 1, 2, 3, 4 y 5 completadas)

#### Fase 5 — Actualizar docs de arquitectura (2026-03-27)

##### Added
- `docs/decisions/009-server-components-first.md`: ADR-009 — documenta la decisión de usar Server Components por defecto y Server Actions para mutaciones (Plan 8 — Fase 2) — *(Plan 8 F5.14)*
- `docs/architecture.md`: sección "Utilidades de stream SSE" — describe las tres funciones públicas de `lib/rag/stream.ts` y sus consumers — *(Plan 8 F5.14)*
- `docs/architecture.md`: sección "Redis (Fase 8 — próxima integración)" — documenta los workarounds SQLite/memoria que serán reemplazados y el motivo por el que Redis será dependencia requerida — *(Plan 8 F5.14)*
- `docs/architecture.md`: ADR-010 agregado a la tabla de ADRs (pendiente Fase 8) — *(Plan 8 F5.14)*

##### Changed
- `docs/architecture.md`: tabla de ADRs actualizada con ADR-008 y ADR-009 — *(Plan 8 F5.14)*
- `docs/architecture.md`: versión Next.js en estructura actualizada (15 → 16), añadidos `safe-action.ts` y `drizzle.config.ts` — *(Plan 8 F5.14)*
- `docs/architecture.md`: sección de workers y Server Actions refleja los cambios de Fases 1–4 — *(Plan 8 F5.14)*

---

### Plan 8 — Optimización (Fases 1, 2, 3 y 4 completadas)

#### Fase 4 — Upgrades de dependencias (2026-03-27)

##### Added
- `apps/web/src/components/collections/DocumentGraphLazy.tsx`: Client Component wrapper que aplica `dynamic` con `ssr: false` para D3 — solución al breaking change de Next.js 16 que prohíbe `dynamic ssr:false` en Server Components — *(Plan 8 F4.9)*

##### Changed
- `next`: 15.5.14 → 16.2.1 — *(Plan 8 F4.9)*
  - `next.config.ts`: `turbopack: { root: __dirname }` para compatibilidad con webpack config custom en Next.js 16
  - `apps/web/package.json`: build script cambiado a `next build --webpack` (webpack config custom, Turbopack tiene limitación de monorepo)
  - `apps/web/tsconfig.json`: `paths` alias `drizzle-orm → ./node_modules/drizzle-orm` para unificar tipos (fix `entityKind` unique symbol)
- `apps/web/src/middleware.ts` → `apps/web/src/proxy.ts`: renombrado + export `middleware` → `proxy` (nueva convención Next.js 16) — *(Plan 8 F4.9)*
- `apps/web/src/app/actions/collections.ts`: `revalidateTag` → `updateTag` (nueva API Next.js 16 para Server Actions) — *(Plan 8 F4.9)*
- `drizzle-orm`: 0.38.4 → 0.45.1 en `packages/db`, `apps/web` y override del root — *(Plan 8 F4.10)*
- `drizzle-kit`: 0.30.0 → 0.31.10 en `packages/db` — *(Plan 8 F4.10)*
- `lucide-react`: 0.475.0 → 1.7.0 — mejor tree-shaking en 1.x — *(Plan 8 F4.11)*
- `zod`: 3.25.0 → 4.3.6 — ~14x mejora de performance en parsing — *(Plan 8 F4.12)*
- `@libsql/client`: 0.14.0 → 0.17.2 — *(Plan 8 F4.13)*
- `next.config.ts`: `@libsql/core` agregado a `serverExternalPackages` (nuevo sub-paquete en 0.17) — *(Plan 8 F4.13)*

##### Fixed
- `apps/web/src/components/auth/SSOButton.tsx`: ícono `Chrome` (removido en lucide 1.x) → `Globe` — *(Plan 8 F4.11)*
- `packages/shared/src/schemas.ts`: `z.record(valueType)` → `z.record(z.string(), valueType)` en 3 schemas (breaking change Zod 4: key schema explícito requerido) — *(Plan 8 F4.12)*
- `apps/web/src/app/(app)/admin/analytics/page.tsx`: queries `sql<T>...as()` reescritas para Drizzle 0.45 (SQL type invariance) — *(Plan 8 F4.13)*
- `apps/web/src/app/(app)/collections/[name]/graph/page.tsx`: eliminado `dynamic ssr:false` en Server Component — usa `DocumentGraphLazy` — *(Plan 8 F4.9)*

---

### Plan 8 — Optimización (Fases 1, 2 y 3 completadas)

#### Fase 3 — Unificación y limpieza de dependencias (2026-03-27)

##### Added
- `packages/db/drizzle.config.ts`: configuración de Drizzle Kit — `schema.ts` es la única fuente de verdad para la DB — *(Plan 8 F3.9)*
- `packages/db/drizzle/0000_hesitant_clint_barton.sql`: migración inicial generada desde `schema.ts` — incluye 27 tablas + FTS5 virtual tables + triggers — *(Plan 8 F3.9)*
- `apps/web/.eslintrc.json` → `apps/web/eslint.config.mjs`: configuración ESLint flat config para Next.js — *(Plan 8 F3.8)*
- Script `db:generate` y `db:push` en `packages/db/package.json` — *(Plan 8 F3.9)*
- Script `lint: tsc --noEmit` en `packages/db`, `packages/logger`, `packages/config`, `packages/shared` — *(Plan 8 F3.8)*
- `eslint`, `eslint-config-next` como devDependencies en `apps/web` — *(Plan 8 F3.8)*
- `@types/node` como devDependency en `packages/db`, `packages/logger`, `packages/config` — *(Plan 8 F3.8)*

##### Changed
- `packages/db/package.json`: `drizzle-orm` sincronizado a `^0.38.4` (igual que `apps/web`) — *(Plan 8 F3.7)*
- `packages/db/src/init.ts`: reemplaza 400 líneas de SQL manual por `migrate(db, { migrationsFolder })` — *(Plan 8 F3.9)*
- `turbo.json`: task `lint` agrega `dependsOn: ["^build"]` para correr en todos los packages — *(Plan 8 F3.8)*
- `apps/web/tsconfig.json`: agrega `exactOptionalPropertyTypes: true` — consistente con `packages/shared` — *(Plan 8 F3.10)*
- `packages/*/tsconfig.json`: `moduleResolution` cambiado de `NodeNext` a `Bundler` (correcto para Bun) — *(Plan 8 F3.8)*

##### Removed
- `zustand@5.0.0`: eliminado de `apps/web/package.json` — sin usos en el codebase — *(Plan 8 F3.10)*
- `dompurify@3.3.3` y `@types/dompurify@3.0.5`: eliminados de `apps/web/package.json` — sin usos — *(Plan 8 F3.10)*

##### Fixed
- `packages/logger/src/blackbox.ts`: corregidos 4 errores de `exactOptionalPropertyTypes` — uso de spread condicional — *(Plan 8 F3.10)*
- `apps/web/src/lib/rag/detect-artifact.ts`, `SourcesPanel.tsx`, `AnnotationPopover.tsx`, `ReportsAdmin.tsx`, `sonner.tsx`, `AnalyticsDashboard.tsx`, `audit/page.tsx`, `api/audit/route.ts`: corregidos errores de `exactOptionalPropertyTypes` — *(Plan 8 F3.10)*

---

### Plan 8 — Optimización (Fases 1 y 2 completadas)

#### Fase 2 — Refactoring de arquitectura React (2026-03-27)

##### Added
- `apps/web/src/components/settings/MemoryClient.tsx`: Client Component que recibe `entries` como prop — sin `useEffect + fetch` — *(Plan 8 F2.4a)*
- `apps/web/src/lib/safe-action.ts`: clientes `authClient` y `adminClient` de `next-safe-action` — *(Plan 8 F2.7)*
- `apps/web/src/lib/form.ts`: helper `createForm` que combina `useForm + zodResolver` — *(Plan 8 F2.8)*
- `apps/web/src/app/actions/auth.ts`: `actionLogout` — Server Action para cerrar sesión — *(Plan 8 F2.7)*
- `apps/web/src/app/actions/projects.ts`: `actionCreateProject`, `actionDeleteProject` — *(Plan 8 F2.7)*
- `apps/web/src/app/actions/webhooks.ts`: `actionCreateWebhook`, `actionDeleteWebhook` — *(Plan 8 F2.7)*
- `apps/web/src/app/actions/reports.ts`: `actionCreateReport`, `actionDeleteReport` — *(Plan 8 F2.7)*
- `apps/web/src/app/actions/external-sources.ts`: `actionCreateExternalSource`, `actionDeleteExternalSource` — *(Plan 8 F2.7)*
- `apps/web/src/app/actions/share.ts`: `actionCreateShare`, `actionRevokeShare` — *(Plan 8 F2.7)*
- `apps/web/src/app/actions/collections.ts`: `actionCreateCollection`, `actionDeleteCollection` — *(Plan 8 F2.7)*
- `next-safe-action@8.1.8`, `react-hook-form@7.72.0`, `@hookform/resolvers@5.2.2`, `nuqs@2.8.9` como dependencias en `apps/web` — *(Plan 8 F2)*

##### Changed
- `settings/memory/page.tsx`: reescrito como Server Component — elimina `"use client"` y `useEffect + fetch` — *(Plan 8 F2.4a)*
- `app/actions/settings.ts`: agrega `actionAddMemory`, `actionDeleteMemory` — *(Plan 8 F2.4a)*
- `admin/webhooks/page.tsx`, `admin/reports/page.tsx`, `admin/external-sources/page.tsx`: fetches server-side + props a Client Components — *(Plan 8 F2.4b)*
- `admin/knowledge-gaps/page.tsx`: extrae lógica de detección de brechas server-side — *(Plan 8 F2.4b)*
- `WebhooksAdmin`, `ReportsAdmin`, `ExternalSourcesAdmin`: eliminan `useEffect`, aceptan `initialData` prop — *(Plan 8 F2.4b)*
- `KnowledgeGapsClient`: eliminan `useEffect` y loading state, acepta `gaps` prop — *(Plan 8 F2.4b)*
- `(app)/layout.tsx`: agrega `Promise.all` para pre-fetch de sessions, projects y changelog — *(Plan 8 F2.4c)*
- `AppShell`, `AppShellChrome`: reciben y propagan `initialSessions`, `initialProjects`, `changelog` — *(Plan 8 F2.4c)*
- `CommandPalette`: acepta `initialSessions` prop, elimina `useEffect + fetch` de sessions — *(Plan 8 F2.4c)*
- `ProjectsPanel`: acepta `initialProjects` prop, elimina `useEffect + fetch` — *(Plan 8 F2.4c)*
- `WhatsNewPanel`: acepta `changelog` prop, elimina `useEffect + fetch` — *(Plan 8 F2.4c)*
- `PromptTemplates`: acepta `templates` prop, elimina `useEffect + fetch` — *(Plan 8 F2.4c)*
- `CollectionSelector`: acepta `availableCollections` prop, elimina fetch — *(Plan 8 F2.4c)*
- `chat/[id]/page.tsx`: pre-fetcha templates y collections server-side — *(Plan 8 F2.4c)*
- `useRagStream.ts`: `stream` y `abort` envueltos en `useCallback` para estabilidad de referencia — *(Plan 8 F2.5)*
- `ChatInterface.tsx`: 5 handlers (`handleSend`, `handleStop`, `handleCopy`, `handleRegenerate`, `handleBookmark`) con `useCallback` — *(Plan 8 F2.5)*
- `SessionList.tsx`: handlers `toggleSelect`, `bulkDelete`, `bulkExport` con `useCallback` — *(Plan 8 F2.5)*
- `AnalyticsDashboard.tsx`: acepta `data` como prop (Server Component pattern); `useMemo` para transformaciones de recharts — *(Plan 8 F2.5)*
- `admin/analytics/page.tsx`: queries analytics server-side — *(Plan 8 F2.5)*
- `collections/[name]/graph/page.tsx`: `DocumentGraph` con `next/dynamic` + skeleton loading — *(Plan 8 F2.6)*
- `NavRail.tsx`: reemplaza `fetch("/api/auth/logout")` con `actionLogout()` — *(Plan 8 F2.7)*
- `ShareDialog.tsx`: reemplaza `fetch("/api/share")` con `actionCreateShare()`/`actionRevokeShare()` — *(Plan 8 F2.7)*
- `ProjectsClient.tsx`: reemplaza fetch mutations con `actionCreateProject`/`actionDeleteProject` — *(Plan 8 F2.7)*
- `CollectionsList.tsx`: reemplaza fetch mutations con `actionCreateCollection`/`actionDeleteCollection` — *(Plan 8 F2.7)*
- `AreasAdmin`, `UsersAdmin`, `WebhooksAdmin`, `ReportsAdmin`, `ExternalSourcesAdmin`, `ProjectsClient`, `CollectionsList`: `useOptimistic` para UI instantánea en delete/create — *(Plan 8 F2.7b)*
- `SettingsClient.tsx`, `login/page.tsx`: migrados a `react-hook-form` con `zodResolver` — *(Plan 8 F2.8)*
- `AreasAdmin`, `UsersAdmin`, `WebhooksAdmin`, `ReportsAdmin`, `ExternalSourcesAdmin`: formularios de creación migrados a `react-hook-form` — *(Plan 8 F2.8)*
- `app/layout.tsx`: agrega `NuqsAdapter` para URL state — *(Plan 8 F2.9)*
- `AuditTable.tsx`: `useState` → `useQueryState("q")` / `useQueryState("level")` — filtros en URL — *(Plan 8 F2.9)*
- `AuditTable.test.tsx`: actualizado para usar `NuqsTestingAdapter` — *(Plan 8 F2.9)*

##### Removed
- `app/api/admin/webhooks/route.ts`: eliminado — reemplazado por Server Actions — *(Plan 8 F2.10)*
- `app/api/admin/reports/route.ts`: eliminado — *(Plan 8 F2.10)*
- `app/api/admin/external-sources/route.ts`: eliminado — *(Plan 8 F2.10)*
- `app/api/admin/knowledge-gaps/route.ts`: eliminado — *(Plan 8 F2.10)*
- `app/api/changelog/route.ts`: eliminado — datos pre-fetched en layout — *(Plan 8 F2.10)*
- `app/api/memory/route.ts`: eliminado — reemplazado por Server Actions — *(Plan 8 F2.10)*
- `app/api/projects/route.ts`: eliminado — *(Plan 8 F2.10)*
- `app/api/chat/sessions/route.ts`: eliminado — datos pre-fetched en layout — *(Plan 8 F2.10)*
- `app/api/share/route.ts`: eliminado — reemplazado por `actionCreateShare` — *(Plan 8 F2.10)*

---

### Plan 8 — Optimización (Fase 1 completada)

#### Added
- `apps/web/src/lib/rag/stream.ts`: utilidades SSE compartidas — `parseSseLine`, `readSseTokens`, `collectSseText` con buffering de líneas parciales y detección de repetición — 2026-03-27 *(Plan 8 F1.1)*
- `apps/web/src/lib/rag/__tests__/stream.test.ts`: 18 tests para las utilidades SSE — 2026-03-27 *(Plan 8 F1.1)*
- `apps/web/src/lib/__tests__/utils.test.ts`: tests de `formatDate`/`formatDateTime` — 2026-03-27 *(Plan 8 F1.7)*
- `docs/decisions/008-sse-reader-extraction.md`: ADR explicando la extracción y por qué vive en `lib/rag/` — 2026-03-27 *(Plan 8 F1)*
- `knip.json`: configuración de workspaces para análisis de dead code — 2026-03-27 *(Plan 8 F1.0)*
- `knip@6.0.6` como devDependency en raíz del monorepo — 2026-03-27 *(Plan 8 F1.0)*

#### Changed
- `hooks/useCrossdocStream.ts`: reemplaza `collectStream` local por `collectSseText` compartido — 2026-03-27 *(Plan 8 F1.1)*
- `hooks/useCrossdocDecompose.ts`: reemplaza `collectSseText` local por la versión compartida — 2026-03-27 *(Plan 8 F1.1)*
- `hooks/useRagStream.ts`: usa `parseSseLine` para extracción de tokens; `sources: Citation[]` en lugar de `unknown[]`; validación Zod con `console.warn` en fallo — 2026-03-27 *(Plan 8 F1.1 + F1.2)*
- `app/api/slack/route.ts`: reemplaza loop SSE inline por `collectSseText` — 2026-03-27 *(Plan 8 F1.1)*
- `app/api/teams/route.ts`: reemplaza loop SSE inline por `collectSseText` — 2026-03-27 *(Plan 8 F1.1)*
- `packages/db/src/queries/sessions.ts`: `addMessage` acepta `sources?: Citation[]` en lugar de `unknown[]` — 2026-03-27 *(Plan 8 F1.2)*
- `lib/export.ts`: `ExportMessage.sources` es `Citation[]`; usa `formatDateTime` centralizado — 2026-03-27 *(Plan 8 F1.2 + F1.7)*
- `app/api/rag/collections/route.ts`: elimina `getCachedRagCollections` duplicada (importa de `collections-cache.ts`) y elimina función dead `ragFetchWithOptions` — 2026-03-27 *(Plan 8 F1.3)*
- `packages/db/src/queries/rate-limits.ts`: `getRateLimit` usa `inArray` en lugar de loop N+1 — 2026-03-27 *(Plan 8 F1.4)*
- `packages/db/src/queries/webhooks.ts`: `listWebhooksByEvent` documenta límite de escala con comentario explícito — 2026-03-27 *(Plan 8 F1.5)*
- `app/api/rag/generate/route.ts`: verificación multi-colección usa `getUserCollections` una sola vez + Set local en lugar de N calls a `canAccessCollection` — 2026-03-27 *(Plan 8 F1.6)*
- `lib/utils.ts`: agrega `formatDate` y `formatDateTime` centralizados — 2026-03-27 *(Plan 8 F1.7)*
- `workers/ingestion.ts`, `components/collections/CollectionHistory.tsx`, `components/admin/KnowledgeGapsClient.tsx`, `components/admin/ReportsAdmin.tsx`, `components/admin/ExternalSourcesAdmin.tsx`, `app/(app)/saved/page.tsx`, `app/(app)/projects/[id]/page.tsx`: reemplazan `toLocaleDateString`/`toLocaleString` inline por `formatDate`/`formatDateTime` — 2026-03-27 *(Plan 8 F1.7)*

**Plan 8 Fase 1 — Extracción de código duplicado: COMPLETO**  
**Duplicaciones eliminadas: SSE reader (×5), getCachedRagCollections (×2), ragFetchWithOptions (dead code), formatDate (×9 instancias). N+1 en getRateLimit corregido. canAccessCollection: N queries → 1.**

---

### Plan 8 — Optimización (Fase 0 completada)

#### Added
- `docs/performance/baseline-plan8.md`: snapshot de métricas pre-optimización — bundle sizes, render analysis, tiempos de CI — 2026-03-27 *(Plan 8 F0)*
- `apps/web`: `@next/bundle-analyzer@16.2.1` como devDependency — 2026-03-27 *(Plan 8 F0.1)*

#### Changed
- `apps/web/next.config.ts`: integración de `withBundleAnalyzer` (activable con `ANALYZE=true`) — 2026-03-27 *(Plan 8 F0.1)*
- `apps/web/src/app/(auth)/login/page.tsx`: `useSearchParams()` envuelto en `<Suspense>` (bug de build Next.js 15) — 2026-03-27

**Plan 8 Fase 0 — Baseline: COMPLETO**
**Métricas capturadas: `/chat` 120 kB, `/chat/[id]` 171 kB, 273 tests en 1.853s, 5 handlers sin memo en ChatInterface**

---

### Plan 6 — UI Testing (completado)

#### Added (F4-F7)
- `apps/web/playwright.config.ts`: config Playwright para visual regression sobre Storybook (6006) — 2026-03-26 *(Plan 6 F4)*
- `apps/web/tests/visual/helpers.ts`: helpers enableDarkMode/enableLightMode + SNAPSHOT_OPTIONS — 2026-03-26 *(Plan 6 F4)*
- `apps/web/tests/visual/design-system.spec.ts`: 20 tests visuales (10 stories × light+dark) — 2026-03-26 *(Plan 6 F4)*
- `apps/web/tests/e2e/` (7 flows Maestro): login-success, login-invalid, logout, new-session, send-message, list-users, collections-list — 2026-03-26 *(Plan 6 F5)*
- `apps/web/playwright.a11y.config.ts`: config Playwright para auditoría a11y en dev server — 2026-03-26 *(Plan 6 F6)*
- `apps/web/tests/a11y/pages.spec.ts`: auditoría WCAG AA en login, chat, collections, admin/users, settings — 2026-03-26 *(Plan 6 F6)*
- `.github/workflows/ci.yml`: jobs component-tests, visual-regression, accessibility — 2026-03-26 *(Plan 6 F7)*
- `axe-playwright@2.2.2`, `@playwright/test@1.58.2` instalados en apps/web — 2026-03-26

#### Changed
- `apps/web/package.json`: scripts test:visual, visual:update, visual:show, test:a11y, test:ui — 2026-03-26 *(Plan 6 F4/F6)*

**Plan 6 — UI Testing Suite: COMPLETO**
**Total tests: 215 (68 lib + 147 componentes) + 20 visuales (baseline pendiente) + 7 E2E flows + 5 a11y auditorías**

---

### Plan 6 — UI Testing (F3 en progreso)

#### Added (F3 — Component tests)
- `src/components/ui/__tests__/button.test.tsx`: 11 tests — render, variantes (6), disabled, onClick, asChild — 2026-03-26 *(Plan 6 F3)*
- `src/components/ui/__tests__/badge.test.tsx`: 7 tests — variantes default/destructive/success/warning/outline/secondary — 2026-03-26 *(Plan 6 F3)*
- `src/components/ui/__tests__/input.test.tsx`: 7 tests — placeholder, onChange, disabled, type, value, border-border, label — 2026-03-26 *(Plan 6 F3)*
- `src/components/ui/__tests__/avatar.test.tsx`: 4 tests — fallback con iniciales, clases accent, rounded-full — 2026-03-26 *(Plan 6 F3)*
- `src/components/ui/__tests__/table.test.tsx`: 5 tests — datos, bg-surface header, hover row, caption, uppercase head — 2026-03-26 *(Plan 6 F3)*
- `src/components/ui/__tests__/skeleton.test.tsx`: 9 tests — Skeleton, SkeletonText, SkeletonAvatar, SkeletonCard, SkeletonTable — 2026-03-26 *(Plan 6 F3)*
- `src/components/ui/__tests__/stat-card.test.tsx`: 7 tests — label, delta+/-, sin delta, deltaLabel, ícono, value string — 2026-03-26 *(Plan 6 F3)*
- `src/components/ui/__tests__/empty-placeholder.test.tsx`: 5 tests — título, ícono, children, className, border-dashed — 2026-03-26 *(Plan 6 F3)*
- `src/lib/component-test-setup.ts`: actualizar con afterEach(cleanup) implícito via patrón por archivo — 2026-03-26 *(Plan 6 F3)*

**Total acumulado F3: 215 tests (68 lib + 147 componentes) — 0 fallos**

#### Added (F3 — Component tests, lote 2)
- `src/components/ui/__tests__/textarea.test.tsx`: 6 tests
- `src/components/ui/__tests__/separator.test.tsx`: 5 tests
- `src/components/ui/__tests__/theme-toggle.test.tsx`: 3 tests
- `src/components/ui/__tests__/data-table.test.tsx`: 6 tests (sorting, filtro, paginación, vacío)
- `src/components/admin/__tests__/UsersAdmin.test.tsx`: 8 tests
- `src/components/admin/__tests__/AreasAdmin.test.tsx`: 6 tests
- `src/components/admin/__tests__/PermissionsAdmin.test.tsx`: 5 tests
- `src/components/admin/__tests__/RagConfigAdmin.test.tsx`: 6 tests
- `src/components/admin/__tests__/SystemStatus.test.tsx`: 4 tests
- `src/components/chat/__tests__/SessionList.test.tsx`: 5 tests
- `src/components/collections/__tests__/CollectionsList.test.tsx`: 5 tests
- `src/components/upload/__tests__/UploadClient.test.tsx`: 5 tests
- `src/components/settings/__tests__/SettingsClient.test.tsx`: 6 tests
- `src/components/audit/__tests__/AuditTable.test.tsx`: 8 tests
- `src/components/projects/__tests__/ProjectsClient.test.tsx`: 6 tests
- `src/components/extract/__tests__/ExtractionWizard.test.tsx`: 6 tests
- `src/components/auth/__tests__/SSOButton.test.tsx`: 1 test

---

### Plan 6 — UI Testing (F2 en progreso)

#### Added
- `apps/web/src/lib/test-setup.ts`: preload global para todos los tests — mocks de next/navigation, next/font, next-themes, next/dynamic — 2026-03-26 *(Plan 6 F2)*
- `apps/web/src/lib/component-test-setup.ts`: preload específico para component tests — GlobalRegistrator (happy-dom) + todos los mocks — 2026-03-26 *(Plan 6 F2)*
- `apps/web/bunfig.toml`: preload de test-setup.ts para tests de lib — 2026-03-26 *(Plan 6 F2)*
- `apps/web/src/components/ui/__tests__/setup-smoke.test.tsx`: smoke test que verifica que @testing-library + happy-dom funcionan — 2026-03-26 *(Plan 6 F2)*
- `@testing-library/react@16.3.2`, `@testing-library/user-event@14.6.1`, `@testing-library/jest-dom@6.9.1`, `happy-dom@20.8.8`, `@happy-dom/global-registrator@20.8.8` — 2026-03-26 *(Plan 6 F2)*

#### Changed
- `apps/web/package.json`: agregar scripts `test:components` y `test:components:watch` con `--preload component-test-setup.ts` — 2026-03-26 *(Plan 6 F2)*

---

### Plan 7 — Design System (F8 en progreso)

#### Changed (F8 — Páginas)
- `apps/web/src/components/extract/ExtractionWizard.tsx`: StepDot navy, Input, EmptyPlaceholder Table2, tokens sin inline styles — 2026-03-26 *(Plan 7 F8.24)*
- `apps/web/src/app/(app)/saved/page.tsx`: EmptyPlaceholder Bookmark, cards bg-surface border-border, tokens — 2026-03-26 *(Plan 7 F8.24)*
- `apps/web/src/app/(app)/settings/memory/page.tsx`: Input, cards bg-surface, Brain icon text-accent, tokens — 2026-03-26 *(Plan 7 F8.24)*
- `apps/web/src/app/(public)/share/[token]/page.tsx`: mensajes bg-accent/bg-surface como ChatInterface, bg-warning-subtle alert, tokens — 2026-03-26 *(Plan 7 F8.24)*
- `apps/web/src/components/admin/IngestionKanban.tsx`: JobCard bg-bg border-border, progress bar bg-accent, error bg-destructive-subtle, bg-success indicator, header con h1 — 2026-03-26 *(Plan 7 F8.9)*
- `apps/web/src/components/admin/AreasAdmin.tsx`: Table shadcn, Input/Button, EmptyPlaceholder, h1 header — 2026-03-26 *(Plan 7 F8.12)*
- `apps/web/src/components/admin/PermissionsAdmin.tsx`: cn() para área activa (bg-accent-subtle), bg-success-subtle/bg-accent-subtle para permisos, Table shadcn — 2026-03-26 *(Plan 7 F8.12)*
- `apps/web/src/components/admin/RagConfigAdmin.tsx`: Button, toggle bg-success, tokens Tailwind, h1 header — 2026-03-26 *(Plan 7 F8.12)*
- `apps/web/src/components/admin/KnowledgeGapsClient.tsx`: EmptyPlaceholder SearchX, Table shadcn, skeleton loading, tokens — 2026-03-26 *(Plan 7 F8.12)*
- `apps/web/src/components/admin/ReportsAdmin.tsx`: Input/Textarea/Button, EmptyPlaceholder FileText, tokens — 2026-03-26 *(Plan 7 F8.12)*
- `apps/web/src/components/admin/WebhooksAdmin.tsx`: Input, EmptyPlaceholder Webhook, event pills cn(), tokens — 2026-03-26 *(Plan 7 F8.12)*
- `apps/web/src/components/admin/IntegrationsAdmin.tsx`: bg-surface-2 code blocks, tokens Tailwind, links text-accent — 2026-03-26 *(Plan 7 F8.12)*
- `apps/web/src/components/admin/ExternalSourcesAdmin.tsx`: Input/Button, EmptyPlaceholder Cloud, tokens — 2026-03-26 *(Plan 7 F8.12)*
- `apps/web/src/components/audit/AuditTable.tsx`: Input, Table shadcn, Badge por nivel, tokens — 2026-03-26 *(Plan 7 F8.12)*
- `apps/web/src/components/settings/SettingsClient.tsx`: tabs con cn(), Input/Button, text-success/text-destructive, PreferenceToggle con tokens — 2026-03-26 *(Plan 7 F8.7)*
- `apps/web/src/components/upload/UploadClient.tsx`: drop zone border-dashed con hover tokens, jobs list bg-surface, text-success/text-destructive — 2026-03-26 *(Plan 7 F8.8)*
- `apps/web/src/components/admin/SystemStatus.tsx`: StatCard + Table shadcn + Badge + Button refresh — 2026-03-26 *(Plan 7 F8.10)*
- `apps/web/src/components/projects/ProjectsClient.tsx`: EmptyPlaceholder, Input/Textarea, cards bg-surface, tokens sin inline styles — 2026-03-26 *(Plan 7 F8.11)*
- `apps/web/src/components/chat/ChatInterface.tsx`: mensajes usuario `bg-accent`, asistente `bg-surface border-border`, input con tokens, `<Button>` send, error `bg-destructive-subtle`, feedback con `cn()` — 2026-03-26 *(Plan 7 F8.3)*
- `apps/web/src/components/chat/SessionList.tsx`: eliminar todos los inline styles — `cn()` para estado activo, tokens Tailwind para destructive/muted/border — 2026-03-26 *(Plan 7 F8.2)*
- `apps/web/src/components/admin/UsersAdmin.tsx`: rediseño completo — `<Table>` shadcn, badges success/destructive/secondary, formulario con `<Input>`, empty state, botones `<Button variant="ghost">`, tokens Tailwind, sin inline styles — 2026-03-26 *(Plan 7 F8.4)*
- `apps/web/src/components/admin/AnalyticsDashboard.tsx`: `<StatCard>`, gráficos con colores navy, tooltips con tokens CSS, loading skeleton, empty state, inline styles eliminados — 2026-03-26 *(Plan 7 F8.5)*
- `apps/web/package.json`: react 19.2.4, react-dom 19.2.4, tailwindcss 4.2.2, typescript 6.0.2, @tailwindcss/postcss 4.2.2 — 2026-03-26 *(chore deps)*

---

### Plan 7 — Design System (en progreso)

#### Added
- `apps/web/.storybook/main.ts` + `preview.ts`: Storybook 8 configurado con @storybook/react-vite, addon-essentials, addon-a11y, addon-themes — 2026-03-26 *(Plan 7 F7)*
- `apps/web/stories/design-system/tokens.stories.tsx`: paleta completa de colores y escala tipográfica — 2026-03-26 *(Plan 7 F7)*
- `apps/web/stories/primitives/button.stories.tsx`: todas las variantes y tamaños — 2026-03-26 *(Plan 7 F7)*
- `apps/web/stories/primitives/badge.stories.tsx`: 6 variantes incluyendo success/warning — 2026-03-26 *(Plan 7 F7)*
- `apps/web/stories/primitives/input.stories.tsx`: estados default, con valor, disabled — 2026-03-26 *(Plan 7 F7)*
- `apps/web/stories/primitives/avatar.stories.tsx`: fallback con iniciales navy — 2026-03-26 *(Plan 7 F7)*
- `apps/web/stories/primitives/table.stories.tsx`: tabla completa con datos mock — 2026-03-26 *(Plan 7 F7)*
- `apps/web/stories/primitives/skeleton.stories.tsx`: SkeletonText, Avatar, Card, Table — 2026-03-26 *(Plan 7 F7)*
- `apps/web/stories/features/stat-card.stories.tsx`: 4 stat cards con deltas — 2026-03-26 *(Plan 7 F7)*
- `apps/web/stories/features/empty-placeholder.stories.tsx`: chat, collections, all variants — 2026-03-26 *(Plan 7 F7)*

#### Added
- `apps/web/src/components/auth/AnimatedBackground.tsx`: fondo animado con orbes CSS (gradiente mesh, sin WebGL) — 2026-03-26 *(Plan 7 F6)*
- `apps/web/src/app/(app)/chat/loading.tsx`: skeleton de carga para /chat — 2026-03-26 *(Plan 7 F6)*
- `apps/web/src/app/(app)/collections/loading.tsx`: skeleton de carga para /collections — 2026-03-26 *(Plan 7 F6)*
- `apps/web/src/app/(app)/admin/users/loading.tsx`: skeleton de carga para /admin/users — 2026-03-26 *(Plan 7 F6)*

#### Changed
- `apps/web/src/app/(auth)/login/page.tsx`: rediseño completo — card glassmorphism, AnimatedBackground, Input/Button components, tokens semánticos — 2026-03-26 *(Plan 7 F6)*
- `apps/web/src/components/chat/SessionList.tsx`: inline styles → tokens Tailwind, bg-surface, border-border — 2026-03-26 *(Plan 7 F6)*

#### Added
- `apps/web/src/components/ui/empty-placeholder.tsx`: componente compuesto para estados vacíos con ícono, título y descripción — 2026-03-26 *(Plan 7 F5)*
- `apps/web/src/components/ui/skeleton.tsx`: shimmer components — Skeleton, SkeletonText, SkeletonAvatar, SkeletonCard, SkeletonTable — 2026-03-26 *(Plan 7 F5)*
- `apps/web/src/components/ui/stat-card.tsx`: tarjeta de estadísticas con valor, delta y ícono — 2026-03-26 *(Plan 7 F5)*
- `apps/web/src/components/ui/data-table.tsx`: tabla avanzada con sorting, filtro y paginación via @tanstack/react-table — 2026-03-26 *(Plan 7 F5)*
- `@tanstack/react-table@8.21.3`: instalado en apps/web — 2026-03-26 *(Plan 7 F5)*

#### Changed
- `apps/web/src/app/(app)/chat/page.tsx`: empty state con EmptyPlaceholder — 2026-03-26 *(Plan 7 F5)*
- `apps/web/src/components/collections/CollectionsList.tsx`: empty state con EmptyPlaceholder, Input component, tokens Tailwind — 2026-03-26 *(Plan 7 F5)*

#### Added
- `apps/web/src/app/(app)/admin/layout.tsx`: layout de admin con `data-density="compact"` aplicado a todas las rutas /admin — 2026-03-26 *(Plan 7 F4)*

#### Changed
- `apps/web/src/components/layout/NavRail.tsx`: rediseño completo — fondo `bg-surface`, iconos con tokens semánticos (`text-fg-muted`, `bg-accent-subtle`), sin colores hardcodeados — 2026-03-26 *(Plan 7 F4)*
- `apps/web/src/components/layout/AppShellChrome.tsx`: `bg-bg` en el contenedor y main, zen indicator con tokens semánticos — 2026-03-26 *(Plan 7 F4)*
- `apps/web/src/components/layout/SecondaryPanel.tsx`: `bg-surface border-border` via clases Tailwind — 2026-03-26 *(Plan 7 F4)*
- `apps/web/src/app/(app)/layout.tsx`: `data-density="spacious"` como default en el contenido de la app — 2026-03-26 *(Plan 7 F4)*

#### Changed
- `apps/web/src/components/ui/button.tsx`: tamaños refinados (h-9/h-8/h-10), `ring-1`, hover states con tokens semánticos — 2026-03-26 *(Plan 7 F3)*
- `apps/web/src/components/ui/badge.tsx`: variantes `success` y `warning` agregadas, forma `rounded-md` más cuadrada — 2026-03-26 *(Plan 7 F3)*
- `apps/web/src/components/ui/input.tsx`: `h-9`, `ring-1`, `border-accent` en focus, `transition-colors` — 2026-03-26 *(Plan 7 F3)*
- `apps/web/src/components/ui/textarea.tsx`: idem input + `resize-y` — 2026-03-26 *(Plan 7 F3)*
- `apps/web/src/components/ui/avatar.tsx`: `AvatarFallback` con `bg-accent-subtle text-accent` — 2026-03-26 *(Plan 7 F3)*
- `apps/web/src/components/ui/table.tsx`: header con `bg-surface`, `TableHead` compact con `h-10 px-3 text-xs uppercase`, `TableRow` con `hover:bg-surface border-border` — 2026-03-26 *(Plan 7 F3)*

#### Added
- `apps/web/src/app/globals.css`: reescritura completa con tokens crema-navy, dark mode cálido (#1a1812), densidad adaptiva, escala tipográfica, aliases shadcn, y `@theme inline` para Tailwind v4 — 2026-03-26 *(Plan 7 F1)*

#### Changed
- `apps/web/src/app/layout.tsx`: agregar Instrument Sans via `next/font/google` con variable CSS `--font-instrument-sans` — 2026-03-26 *(Plan 7 F2)*

---

### Plan 6 — UI Testing (en progreso)

#### Added
- `docs/plans/ultra-optimize-plan6-ui-testing.md`: plan de 7 fases para UI testing completo — component tests, visual regression, Maestro E2E, a11y, CI — 2026-03-26
- `docs/plans/ultra-optimize-plan7-design-system.md`: plan de 8 fases para design system "Warm Intelligence" — tokens crema-navy, Instrument Sans, Storybook, 24 páginas — 2026-03-26
- `docs/superpowers/specs/2026-03-26-design-system-design.md`: spec aprobado del design system — 2026-03-26
- `docs/superpowers/specs/2026-03-26-ui-testing-design.md`: spec aprobado del UI testing — 2026-03-26
- `react-scan@0.5.3`: instalado como devDependency en `apps/web` para baseline de performance — 2026-03-26 *(Plan 6 F1)*
- `apps/web/src/components/dev/ReactScan.tsx`: Client Component que inicializa react-scan solo en `NODE_ENV=development` — 2026-03-26 *(Plan 6 F1)*
- `docs/superpowers/react-scan-baseline.md`: template del reporte baseline de re-renders — completar tras recorrer la app — 2026-03-26 *(Plan 6 F1)*

#### Modified
- `apps/web/src/app/layout.tsx`: agregar `<ReactScanInit />` con dynamic import condicional (solo dev, ssr:false) — 2026-03-26 *(Plan 6 F1)*

---

### Plan 5 — Testing Foundation (2026-03-26)

#### Added
- `docs/plans/ultra-optimize-plan5-testing-foundation.md`: plan de 5 fases para llevar cobertura a 95% en `packages/*` y `apps/web/src/lib/`, con enforcement en CI — 2026-03-26
- `docs/decisions/006-testing-strategy.md`: ADR que codifica metas de cobertura por capa, matriz "tipo de código → test requerido", y enforcement en CI — 2026-03-26 *(Plan 5 F1)*
- `bunfig.toml`: configuración de coverage con `coverageThreshold = 0.80` (sube a 0.95 al completar F3/F4) — 2026-03-26 *(Plan 5 F2)*
- `packages/db/src/__tests__/sessions.test.ts`: 11 tests de sesiones, mensajes y feedback — 2026-03-26 *(Plan 5 F3)*
- `packages/db/src/__tests__/events.test.ts`: 10 tests de writeEvent y queryEvents con todos los filtros — 2026-03-26 *(Plan 5 F3)*
- `packages/db/src/__tests__/memory.test.ts`: 10 tests de setMemory (upsert), getMemory, deleteMemory, getMemoryAsContext — 2026-03-26 *(Plan 5 F3)*
- `packages/db/src/__tests__/annotations.test.ts`: 7 tests de saveAnnotation, listAnnotationsBySession (filtro user+session), deleteAnnotation — 2026-03-26 *(Plan 5 F3)*
- `packages/db/src/__tests__/tags.test.ts`: 9 tests de addTag (idempotente, lowercase), removeTag, listTagsBySession, listTagsByUser — 2026-03-26 *(Plan 5 F3)*
- `packages/db/src/__tests__/shares.test.ts`: 9 tests de createShare (TTL), getShareByToken (expirado/inexistente), revokeShare (protección usuario), listSharesByUser — 2026-03-26 *(Plan 5 F3)*
- `packages/db/src/__tests__/templates.test.ts`: 7 tests de createTemplate, listActiveTemplates (solo activos, orden), deleteTemplate — 2026-03-26 *(Plan 5 F3)*
- `packages/db/src/__tests__/webhooks.test.ts`: 8 tests de createWebhook (secret único), listWebhooksByEvent (wildcards, inactivos), deleteWebhook — 2026-03-26 *(Plan 5 F3)*
- `packages/db/src/__tests__/reports.test.ts`: 8 tests de createReport (calcNextRun), listActiveReports (pasado/futuro), updateLastRun, deleteReport (protección) — 2026-03-26 *(Plan 5 F3)*
- `packages/db/src/__tests__/collection-history.test.ts`: 7 tests de recordIngestionEvent, listHistoryByCollection (orden desc, límite 50) — 2026-03-26 *(Plan 5 F3)*
- `packages/db/src/__tests__/rate-limits.test.ts`: 10 tests de createRateLimit, getRateLimit (prioridad user>area), countQueriesLastHour (tipo, usuario, tiempo) — 2026-03-26 *(Plan 5 F3)*
- `packages/db/src/__tests__/projects.test.ts`: 13 tests de createProject, listProjects, getProject, updateProject, deleteProject (protección), addSessionToProject (idempotente), getProjectBySession — 2026-03-26 *(Plan 5 F3)*
- `packages/db/src/__tests__/search.test.ts`: 9 tests de universalSearch (LIKE fallback) — edge cases, sesiones, templates, saved responses — 2026-03-26 *(Plan 5 F3)*
- `packages/db/src/__tests__/external-sources.test.ts`: 9 tests de createExternalSource, listExternalSources, listActiveSourcesToSync (schedule/lastSync), updateSourceLastSync, deleteExternalSource — 2026-03-26 *(Plan 5 F3)*

#### Changed
- `packages/db/src/connection.ts`: `_injectDbForTesting()` y `_resetDbForTesting()` para inyectar DB en tests sin modificar singleton de producción — 2026-03-26 *(Plan 5 F3)*
- `bunfig.toml`: threshold separado por métrica: `line = 0.90`, `function = 0.50` — schema.ts tiene 100% line coverage — 2026-03-26 *(Plan 5 F3)*
- `apps/web/src/lib/rag/detect-artifact.ts`: función `detectArtifact` extraída de `useRagStream.ts` — lógica pura testeable (marcador explícito + heurísticas código/tabla) — 2026-03-26 *(Plan 5 F4)*
- `apps/web/src/lib/rag/__tests__/detect-artifact.test.ts`: 15 tests de detectArtifact — marcador explícito, heurísticas, casos sin artifact, prioridad del marcador — 2026-03-26 *(Plan 5 F4)*
- `apps/web/src/lib/__tests__/webhook.test.ts`: 8 tests de dispatchWebhook — firma HMAC verificable, headers correctos, manejo silencioso de errores (4xx, 500, timeout, AbortError) — 2026-03-26 *(Plan 5 F4)*

#### Changed
- `apps/web/src/hooks/useRagStream.ts`: `detectArtifact` y `ArtifactData` importados desde `@/lib/rag/detect-artifact` — 2026-03-26 *(Plan 5 F4)*
- `bunfig.toml`: threshold final `line = 0.95` — meta Plan 5 alcanzada — 2026-03-26 *(Plan 5 F4)*
- `.cursor/skills/rag-testing/SKILL.md`: tabla de cobertura actualizada a ~237 tests, limitación de local helpers documentada — 2026-03-26 *(Plan 5 F5)*

### Refactoring: tests de DB con funciones reales (2026-03-26)

#### Added
- `docs/decisions/007-real-functions-over-local-helpers-in-tests.md`: ADR que codifica el patrón de tests con funciones reales + `_injectDbForTesting` — 2026-03-26
- `packages/db/src/__tests__/setup.ts`: SQL completo del schema + helpers `insertUser`, `insertSession`, `insertMessage` compartidos entre todos los test files — 2026-03-26

#### Changed
- `packages/db/src/queries/areas.ts`: `getDb()` movido dentro de cada función (era nivel módulo) — 2026-03-26
- `packages/db/src/queries/users.ts`: `getDb()` movido dentro de cada función — 2026-03-26
- `packages/db/src/queries/sessions.ts`: `getDb()` movido dentro de cada función — 2026-03-26
- `packages/db/src/queries/events.ts`: `getDb()` movido dentro de cada función — 2026-03-26
- `packages/db/src/__tests__/*.test.ts` (17 archivos): reescritos para importar y llamar funciones reales de producción usando `_injectDbForTesting` — cobertura de query files: 0% → 95.20% líneas — 2026-03-26
- `docs/workflows.md`: ADR-007 agregado a la tabla de decisiones — 2026-03-26

#### Fixed
- `packages/db/src/queries/tags.ts`: `removeTag` eliminaba TODOS los tags de la sesión en lugar de solo el especificado — faltaba `and(eq(sessionTags.tag, tag))` en el WHERE — bug expuesto al llamar la función real en tests — 2026-03-26

#### Changed
- `package.json` raíz: script `test:coverage` vía Turborepo — 2026-03-26 *(Plan 5 F2)*
- `packages/*/package.json` + `apps/web/package.json`: script `test:coverage` con `--coverage` — 2026-03-26 *(Plan 5 F2)*
- `turbo.json`: task `test:coverage` con outputs `coverage/**` — 2026-03-26 *(Plan 5 F2)*
- `.github/workflows/ci.yml`: nuevo job `coverage` que corre `bun run test:coverage` en PRs; job `test` separado para pushes rápidos — 2026-03-26 *(Plan 5 F2)*

#### Changed
- `.cursor/skills/rag-testing/SKILL.md`: reescrito con la regla de oro, matriz completa de tests requeridos, metas por capa, tabla de estado de cobertura — 2026-03-26 *(Plan 5 F1)*
- `docs/workflows.md`: sección 2 (testing) reescrita — regla de oro, metas por capa, matriz tipo→test, comandos de coverage, patrón actualizado con `process.env` antes de imports — 2026-03-26 *(Plan 5 F1)*
- `docs/workflows.md`: ADR-006 agregado a la tabla de ADRs en sección 7 — 2026-03-26 *(Plan 5 F1)*

### Mejoras de metodología (2026-03-26)

#### Added
- `docs/decisions/` — nueva carpeta para Architecture Decision Records (ADRs): documenta decisiones arquitectónicas con contexto, opciones consideradas, decisión tomada y consecuencias — 2026-03-26
- `docs/decisions/000-template.md` — template base para nuevos ADRs — 2026-03-26
- `docs/decisions/001-libsql-over-better-sqlite3.md` — por qué `@libsql/client` sobre `better-sqlite3` (compilación nativa, WSL2, Bun) — 2026-03-26
- `docs/decisions/002-cjs-over-esm.md` — por qué CJS sobre ESM en `packages/*` (compatibilidad webpack/Next.js) — 2026-03-26
- `docs/decisions/003-nextjs-single-process.md` — por qué Next.js como proceso único reemplaza Python gateway + SvelteKit — 2026-03-26
- `docs/decisions/004-temporal-api-timestamps.md` — por qué Temporal API sobre `Date.now()` para timestamps — 2026-03-26
- `docs/decisions/005-static-import-logger-db.md` — por qué import estático de `@rag-saldivia/db` en el logger (bug de import dinámico silencioso en webpack) — 2026-03-26

#### Changed
- `docs/workflows.md`: sección 4 (planificación) — agregado checklist de cierre al template de fases: `bun run test`, CHANGELOG actualizado, commit hecho — 2026-03-26
- `docs/workflows.md`: sección 3 (git) — nueva convención de secciones por plan dentro de `[Unreleased]` para hacer navegable el CHANGELOG durante el desarrollo — 2026-03-26
- `docs/workflows.md`: nueva sección 7 — Decisiones de arquitectura (ADRs) con cuándo crear un ADR, formato, convención de nombres y tabla de ADRs existentes — 2026-03-26

### Plan 4 — Product Roadmap (2026-03-25)

#### Added

- `apps/web/src/app/api/extract/route.ts`: extracción estructurada — itera docs de la colección, envía prompt al RAG para extraer campos, retorna JSON — modo mock disponible — 2026-03-25 *(Plan 4 F3.50)*
- `apps/web/src/components/extract/ExtractionWizard.tsx`: wizard 3 pasos (colección → campos → resultados), tabla exportable como CSV — 2026-03-25 *(Plan 4 F3.50)*
- `apps/web/src/app/(app)/extract/page.tsx`: página `/extract` accesible para todos los usuarios — 2026-03-25 *(Plan 4 F3.50)*
- `apps/web/src/components/layout/NavRail.tsx`: ícono `Table2` para `/extract` — 2026-03-25 *(Plan 4 F3.50)*
- `packages/db/src/schema.ts`: tabla `bot_user_mappings` (platform, externalUserId, systemUserId) — 2026-03-25 *(Plan 4 F3.49)*
- `apps/web/src/app/api/slack/route.ts`: handler de slash commands Slack — verifica firma HMAC, resuelve userId via mapping, consulta RAG y responde via response_url — 2026-03-25 *(Plan 4 F3.49)*
- `apps/web/src/app/api/teams/route.ts`: handler de mensajes Teams — respeta RBAC via mapping de usuarios — 2026-03-25 *(Plan 4 F3.49)*
- `apps/web/src/app/(app)/admin/integrations/page.tsx` y `IntegrationsAdmin.tsx`: UI de configuración con URLs y guía de setup — 2026-03-25 *(Plan 4 F3.49)*
- `packages/db/src/schema.ts`: tabla `external_sources` (provider, credentials, collectionDest, schedule, lastSync) — 2026-03-25 *(Plan 4 F3.48)*
- `packages/db/src/queries/external-sources.ts`: `createExternalSource`, `listExternalSources`, `listActiveSourcesToSync`, `updateSourceLastSync`, `deleteExternalSource` — 2026-03-25 *(Plan 4 F3.48)*
- `apps/web/src/workers/external-sync.ts`: worker MVP que detecta fuentes listas para sync y registra logs; implementación OAuth completa pendiente de credenciales reales — 2026-03-25 *(Plan 4 F3.48)*
- `apps/web/src/app/(app)/admin/external-sources/page.tsx` y `ExternalSourcesAdmin.tsx`: UI para configurar fuentes externas — 2026-03-25 *(Plan 4 F3.48)*
- `packages/db/src/schema.ts`: campos `sso_provider` y `sso_subject` en tabla `users` — 2026-03-25 *(Plan 4 F3.47)*
- `apps/web/src/lib/auth/next-auth.ts`: configuración NextAuth v5 con providers Google y Microsoft Entra ID; modo mixto (SSO + password); al primer login SSO crea usuario o vincula cuenta existente; emite JWT propio para compatibilidad RBAC — 2026-03-25 *(Plan 4 F3.47)*
- `apps/web/src/app/api/auth/[...nextauth]/route.ts`: handler de NextAuth — 2026-03-25 *(Plan 4 F3.47)*
- `apps/web/src/components/auth/SSOButton.tsx`: botones Google y Microsoft en página de login (solo visibles si los env vars están configurados) — 2026-03-25 *(Plan 4 F3.47)*
- `.env.example`: variables SSO y NextAuth documentadas — 2026-03-25 *(Plan 4 F3.47)*
- `apps/web/src/app/api/collections/[name]/embeddings/route.ts`: retorna grafo de similitud — intenta obtener docs del RAG server, simula similitud para MVP — 2026-03-25 *(Plan 4 F3.46)*
- `apps/web/src/components/collections/DocumentGraph.tsx`: visualización SVG force-directed sin dependencia de d3-force (simulación propia ligera); zoom, colores por cluster, click en nodo — 2026-03-25 *(Plan 4 F3.46)*
- `apps/web/src/app/(app)/collections/[name]/graph/page.tsx`: página del grafo por colección — 2026-03-25 *(Plan 4 F3.46)*
- `apps/web/src/components/collections/CollectionsList.tsx`: botón "Grafo" por colección — 2026-03-25 *(Plan 4 F3.46)*
- `apps/web/src/workers/ingestion.ts`: `checkProactiveSurface` — cruza keywords del doc nuevo con queries recientes del usuario; si hay match genera evento `proactive.docs_available` — 2026-03-25 *(Plan 4 F3.45)*
- `apps/web/src/app/api/notifications/route.ts`: `proactive.docs_available` agregado a los tipos de notificación — 2026-03-25 *(Plan 4 F3.45)*
- `packages/db/src/schema.ts`: tabla `user_memory` (key, value, source explicit/inferred, UNIQUE per user+key) — 2026-03-25 *(Plan 4 F3.44)*
- `packages/db/src/queries/memory.ts`: `setMemory` (upsert), `getMemory`, `deleteMemory`, `getMemoryAsContext` — 2026-03-25 *(Plan 4 F3.44)*
- `apps/web/src/app/api/rag/generate/route.ts`: inyección de memoria del usuario como system message — 2026-03-25 *(Plan 4 F3.44)*
- `apps/web/src/app/(app)/settings/memory/page.tsx`: UI para ver/agregar/eliminar preferencias de memoria — 2026-03-25 *(Plan 4 F3.44)*
- `packages/db/src/schema.ts`: campo `forked_from` en `chat_sessions` (TEXT nullable, sin FK circular en Drizzle) — 2026-03-25 *(Plan 4 F3.43)*
- `apps/web/src/app/actions/chat.ts`: Server Action `actionForkSession` — copia sesión y mensajes hasta el punto indicado, setea `forked_from` — 2026-03-25 *(Plan 4 F3.43)*
- `apps/web/src/components/chat/ChatInterface.tsx`: botón `GitBranch` en mensajes del asistente para bifurcar desde ese punto — 2026-03-25 *(Plan 4 F3.43)*
- `apps/web/src/components/chat/SessionList.tsx`: badge `GitBranch` en sesiones con `forked_from` no null — 2026-03-25 *(Plan 4 F3.43)*
- `apps/web/src/hooks/useRagStream.ts`: detección de artifacts al finalizar stream — marcador `:::artifact` explícito o heurística (código ≥ 40 líneas, tabla ≥ 5 cols); callback `onArtifact` — 2026-03-25 *(Plan 4 F3.42)*
- `apps/web/src/components/chat/ArtifactsPanel.tsx`: Sheet lateral para código/tabla/documento — botones guardar y exportar; resaltado de código en `<pre>` — 2026-03-25 *(Plan 4 F3.42)*
- `packages/db/src/schema.ts`: tablas `projects`, `project_sessions`, `project_collections` — 2026-03-25 *(Plan 4 F3.41)*
- `packages/db/src/queries/projects.ts`: `createProject`, `listProjects`, `getProject`, `updateProject`, `deleteProject`, `addSessionToProject`, `addCollectionToProject`, `getProjectBySession` — 2026-03-25 *(Plan 4 F3.41)*
- `apps/web/src/app/api/projects/route.ts`: GET/POST/DELETE para proyectos — 2026-03-25 *(Plan 4 F3.41)*
- `apps/web/src/app/(app)/projects/page.tsx` y `[id]/page.tsx`: páginas de proyectos — 2026-03-25 *(Plan 4 F3.41)*
- `apps/web/src/components/projects/ProjectsClient.tsx`: grid de proyectos con crear/eliminar — 2026-03-25 *(Plan 4 F3.41)*
- `apps/web/src/components/layout/panels/ProjectsPanel.tsx`: panel secundario para rutas `/projects` — 2026-03-25 *(Plan 4 F3.41)*
- `apps/web/src/components/layout/NavRail.tsx`: ícono `FolderKanban` para `/projects` — 2026-03-25 *(Plan 4 F3.41)*
- `apps/web/src/app/api/rag/generate/route.ts`: inyección del contexto del proyecto como system message — 2026-03-25 *(Plan 4 F3.41)*
- `apps/web/src/components/chat/DocPreviewPanel.tsx`: panel Sheet lateral para preview de PDF con react-pdf (carga dinámica SSR-safe), paginación, fallback a texto cuando el Blueprint no expone el documento — 2026-03-25 *(Plan 4 F3.40)*
- `apps/web/src/app/api/rag/document/[name]/route.ts`: proxy al RAG server para obtener PDF; retorna 404 con nota si el endpoint no está disponible — 2026-03-25 *(Plan 4 F3.40)*
- `apps/web/src/components/chat/SourcesPanel.tsx`: nombre de cada fuente ahora es botón clic que abre `DocPreviewPanel` con el fragmento resaltado — 2026-03-25 *(Plan 4 F3.40)*
- `packages/db/src/queries/search.ts`: `universalSearch(query, userId, limit)` — busca con FTS5 (sesiones + mensajes) con fallback a LIKE; también busca en templates y saved_responses — 2026-03-25 *(Plan 4 F3.39)*
- `packages/db/src/init.ts`: tablas FTS5 virtuales `sessions_fts` y `messages_fts` con triggers de sincronización automática — 2026-03-25 *(Plan 4 F3.39)*
- `apps/web/src/app/api/search/route.ts`: endpoint `GET /api/search?q=...` con debounce 300ms — 2026-03-25 *(Plan 4 F3.39)*
- `apps/web/src/components/layout/CommandPalette.tsx`: integración de búsqueda universal — grupo "Resultados para X" con tipo (session/message/saved/template) y snippet; aparece cuando query ≥ 2 chars — 2026-03-25 *(Plan 4 F3.39)*

### Added

- `packages/db/src/schema.ts`: tabla `webhooks` (url, events JSON, secret HMAC, active) — 2026-03-25 *(Plan 4 F2.38)*
- `packages/db/src/queries/webhooks.ts`: `createWebhook` (genera secret aleatorio), `listWebhooksByEvent`, `listAllWebhooks`, `deleteWebhook` — 2026-03-25 *(Plan 4 F2.38)*
- `apps/web/src/lib/webhook.ts`: `dispatchWebhook` con firma HMAC-SHA256 en header `X-Signature`; timeout 5s; no interrumpe el flujo principal si falla; `dispatchEvent` busca webhooks activos para el tipo de evento — 2026-03-25 *(Plan 4 F2.38)*
- `apps/web/src/workers/ingestion.ts`: dispatch de `ingestion.completed` al terminar un job — 2026-03-25 *(Plan 4 F2.38)*
- `apps/web/src/app/api/rag/generate/route.ts`: dispatch de `query.completed` al finalizar un stream — 2026-03-25 *(Plan 4 F2.38)*
- `apps/web/src/app/api/admin/webhooks/route.ts`: GET/POST/DELETE para gestión de webhooks — 2026-03-25 *(Plan 4 F2.38)*
- `apps/web/src/components/admin/WebhooksAdmin.tsx`: UI para crear/listar/eliminar webhooks con selector de eventos y copia del secret — 2026-03-25 *(Plan 4 F2.38)*
- `apps/web/src/app/(app)/admin/webhooks/page.tsx`: página `/admin/webhooks` — 2026-03-25 *(Plan 4 F2.38)*
- `packages/db/src/schema.ts`: campo `onboarding_completed` en tabla `users` (default false) — 2026-03-25 *(Plan 4 F2.37)*
- `apps/web/src/components/onboarding/OnboardingTour.tsx`: tour driver.js de 5 pasos (nav, chat, modos de foco, colecciones, settings); se activa al primer login; saltable; llama a `actionCompleteOnboarding` al terminar — 2026-03-25 *(Plan 4 F2.37)*
- `apps/web/src/app/actions/settings.ts`: Server Actions `actionCompleteOnboarding` y `actionResetOnboarding` — 2026-03-25 *(Plan 4 F2.37)*
- `apps/web/src/app/(app)/layout.tsx`: renderiza `OnboardingTour` condicionalmente si `onboardingCompleted === false` — 2026-03-25 *(Plan 4 F2.37)*
- `packages/db/src/__tests__/users.test.ts` y `saved.test.ts`: columna `onboarding_completed` agregada al SQL de test — 2026-03-25 *(bugfix)*
- `packages/db/src/schema.ts`: tabla `rate_limits` (targetType user/area, targetId, maxQueriesPerHour) — 2026-03-25 *(Plan 4 F2.36)*
- `packages/db/src/queries/rate-limits.ts`: `getRateLimit` (user-level primero, luego área), `countQueriesLastHour`, `createRateLimit`, `listRateLimits`, `deleteRateLimit` — 2026-03-25 *(Plan 4 F2.36)*
- `apps/web/src/app/api/rag/generate/route.ts`: check de rate limit antes de procesar — retorna 429 si el usuario superó su límite/hora — 2026-03-25 *(Plan 4 F2.36)*
- `apps/web/src/components/chat/ChatDropZone.tsx`: drop zone sobre el área del chat — overlay al arrastrar, sube el archivo via `/api/upload` y confirma con toast; colecciones efímeras descartadas por viabilidad (Blueprint v2.5.0 no las soporta en Milvus) — 2026-03-25 *(Plan 4 F2.35)*
- `apps/web/src/components/chat/SplitView.tsx`: wrapper de vista dividida — botón Columns2 abre segundo panel con sesión independiente, botón X cierra; cada panel tiene su propio ChatInterface — 2026-03-25 *(Plan 4 F2.34)*
- `packages/db/src/schema.ts`: tabla `scheduled_reports` (query, collection, schedule, destination, email, nextRun) — 2026-03-25 *(Plan 4 F2.33)*
- `packages/db/src/queries/reports.ts`: `createReport`, `listActiveReports`, `listReportsByUser`, `updateLastRun`, `deleteReport` — 2026-03-25 *(Plan 4 F2.33)*
- `apps/web/src/workers/ingestion.ts`: procesador de informes programados — corre cada 5 min, ejecuta query via RAG, guarda en Guardados o env\u00eda por email (si SMTP configurado) — 2026-03-25 *(Plan 4 F2.33)*
- `apps/web/src/app/api/admin/reports/route.ts`: GET/POST/DELETE para informes programados — 2026-03-25 *(Plan 4 F2.33)*
- `apps/web/src/components/admin/ReportsAdmin.tsx`: formulario + lista de informes — 2026-03-25 *(Plan 4 F2.33)*
- `apps/web/src/app/(app)/admin/reports/page.tsx`: p\u00e1gina `/admin/reports` — 2026-03-25 *(Plan 4 F2.33)*
- `packages/db/src/schema.ts`: tabla `collection_history` (collection, userId, action, filename, docCount) — 2026-03-25 *(Plan 4 F2.32)*
- `packages/db/src/queries/collection-history.ts`: `recordIngestionEvent`, `listHistoryByCollection` — 2026-03-25 *(Plan 4 F2.32)*
- `apps/web/src/workers/ingestion.ts`: registrar en `collection_history` al completar job exitosamente — 2026-03-25 *(Plan 4 F2.32)*
- `apps/web/src/components/collections/CollectionHistory.tsx`: timeline de ingestas por colección — 2026-03-25 *(Plan 4 F2.32)*
- `apps/web/src/app/api/collections/[name]/history/route.ts`: endpoint GET para historial de una colección — 2026-03-25 *(Plan 4 F2.32)*
- `apps/web/src/app/api/admin/knowledge-gaps/route.ts`: detecta respuestas del asistente con baja confianza (< 80 palabras + keywords de incertidumbre), retorna hasta 100 gaps — 2026-03-25 *(Plan 4 F2.31)*
- `apps/web/src/components/admin/KnowledgeGapsClient.tsx`: tabla de brechas con link a sesión, exportar CSV, botón "Ingestar documentos" — 2026-03-25 *(Plan 4 F2.31)*
- `apps/web/src/app/(app)/admin/knowledge-gaps/page.tsx`: página `/admin/knowledge-gaps` — 2026-03-25 *(Plan 4 F2.31)*
- `apps/web/src/app/api/admin/analytics/route.ts`: queries de agregación — queries/día, top colecciones, distribución feedback, top usuarios — 2026-03-25 *(Plan 4 F2.30)*
- `apps/web/src/components/admin/AnalyticsDashboard.tsx`: dashboard con recharts — LineChart queries/día, BarChart colecciones, PieChart feedback, tabla top usuarios; stats cards con totales — 2026-03-25 *(Plan 4 F2.30)*
- `apps/web/src/app/(app)/admin/analytics/page.tsx`: página `/admin/analytics` — 2026-03-25 *(Plan 4 F2.30)*
- `apps/web/src/app/api/admin/ingestion/stream/route.ts`: SSE endpoint que emite estado de jobs cada 3s — 2026-03-25 *(Plan 4 F2.29)*
- `apps/web/src/app/api/admin/ingestion/[id]/route.ts`: PATCH con `action: "retry"` para reintentar jobs fallidos — 2026-03-25 *(Plan 4 F2.29)*
- `apps/web/src/components/admin/IngestionKanban.tsx`: kanban de 4 columnas (Pendiente/En progreso/Completado/Error) con barra de progreso, detalle de error expandible, botón retry, indicador SSE en tiempo real — 2026-03-25 *(Plan 4 F2.29)*
- `apps/web/src/app/(app)/admin/ingestion/page.tsx`: página de monitoring de ingesta — 2026-03-25 *(Plan 4 F2.29)*
- `packages/db/src/schema.ts`: tabla `prompt_templates` (title, prompt, focusMode, createdBy, active) — 2026-03-25 *(Plan 4 F2.28)*
- `packages/db/src/queries/templates.ts`: `listActiveTemplates`, `createTemplate`, `deleteTemplate` — 2026-03-25 *(Plan 4 F2.28)*
- `apps/web/src/app/api/admin/templates/route.ts`: GET lista templates activos, POST crea (admin), DELETE elimina (admin) — 2026-03-25 *(Plan 4 F2.28)*
- `apps/web/src/components/chat/PromptTemplates.tsx`: selector de templates como Popover con título y preview del prompt; al elegir setea input + focusMode — 2026-03-25 *(Plan 4 F2.28)*
- `apps/web/src/app/actions/chat.ts`: Server Action `actionCreateSessionForDoc` — crea sesión con system message que restringe el contexto al documento específico — 2026-03-25 *(Plan 4 F2.27)*
- `apps/web/src/components/collections/CollectionsList.tsx`: botón "Chat" por colección + helper `handleChatWithDoc` para crear sesión anclada a un doc — 2026-03-25 *(Plan 4 F2.27)*
- `apps/web/src/app/(app)/collections/page.tsx`: página de colecciones Server Component con lista + metadata — 2026-03-25 *(Plan 4 F2.26)*
- `apps/web/src/components/collections/CollectionsList.tsx`: tabla de colecciones con acciones Chat y Eliminar (solo admin); formulario inline para crear nueva colección — 2026-03-25 *(Plan 4 F2.26)*
- `apps/web/src/app/api/rag/collections/route.ts`: POST para crear colección (solo admin) — 2026-03-25 *(Plan 4 F2.26)*
- `apps/web/src/app/api/rag/collections/[name]/route.ts`: DELETE para eliminar colección (solo admin) — 2026-03-25 *(Plan 4 F2.26)*
- `packages/db/src/schema.ts`: tabla `session_shares` (token UUID 64-char hex, expiresAt) — 2026-03-25 *(Plan 4 F2.25)*
- `packages/db/src/queries/shares.ts`: `createShare`, `getShareByToken`, `getShareWithSession`, `revokeShare` — 2026-03-25 *(Plan 4 F2.25)*
- `apps/web/src/app/api/share/route.ts`: POST crea token, DELETE revoca — 2026-03-25 *(Plan 4 F2.25)*
- `apps/web/src/app/(public)/share/[token]/page.tsx`: página pública read-only sin auth; muestra sesión + aviso de privacidad; 404 si token inválido/expirado — 2026-03-25 *(Plan 4 F2.25)*
- `apps/web/src/middleware.ts`: `/share/` agregado a PUBLIC_ROUTES — 2026-03-25 *(Plan 4 F2.25)*
- `apps/web/src/components/chat/ShareDialog.tsx`: Dialog para generar/copiar/revocar el link de compartir, con aviso de privacidad — 2026-03-25 *(Plan 4 F2.25)*
- `packages/db/src/schema.ts`: tabla `session_tags` (sessionId, tag, PK compuesta) — 2026-03-25 *(Plan 4 F2.24)*
- `packages/db/src/queries/tags.ts`: `addTag`, `removeTag`, `listTagsBySession`, `listTagsByUser` — 2026-03-25 *(Plan 4 F2.24)*
- `apps/web/src/components/chat/SessionList.tsx`: badges de etiquetas por sesión, input inline para agregar tags, filtro por tag en el header, bulk selection con toolbar (exportar/eliminar seleccionadas) — 2026-03-25 *(Plan 4 F2.24)*
- `apps/web/src/app/actions/chat.ts`: Server Actions `actionAddTag`, `actionRemoveTag` — 2026-03-25 *(Plan 4 F2.24)*
- `apps/web/src/components/layout/CommandPalette.tsx`: command palette con `cmdk` — grupos Navegar (chat, colecciones, upload, saved, audit, settings, admin), Apariencia (tema, zen), Sesiones recientes filtradas por texto — 2026-03-25 *(Plan 4 F2.23)*
- `apps/web/src/app/api/chat/sessions/route.ts`: endpoint GET que lista sesiones del usuario para la palette — 2026-03-25 *(Plan 4 F2.23)*
- `apps/web/src/hooks/useGlobalHotkeys.ts`: agregado `Cmd+K` para abrir command palette — 2026-03-25 *(Plan 4 F2.23)*
- `packages/db/src/schema.ts`: tabla `annotations` (selectedText, note, FK a session y message) — 2026-03-25 *(Plan 4 F2.22)*
- `packages/db/src/queries/annotations.ts`: `saveAnnotation`, `listAnnotationsBySession`, `deleteAnnotation` — 2026-03-25 *(Plan 4 F2.22)*
- `apps/web/src/components/chat/AnnotationPopover.tsx`: popover flotante al seleccionar texto en respuestas asistente — opciones Guardar, Preguntar sobre esto, Comentar con nota libre — 2026-03-25 *(Plan 4 F2.22)*
- `apps/web/src/app/actions/chat.ts`: Server Action `actionSaveAnnotation` — 2026-03-25 *(Plan 4 F2.22)*
- `apps/web/src/components/chat/CollectionSelector.tsx`: selector multi-checkbox de colecciones disponibles del usuario, persistido en localStorage; muestra las colecciones activas como Popover junto al input — 2026-03-25 *(Plan 4 F2.21)*
- `apps/web/src/hooks/useRagStream.ts`: acepta `collections?: string[]` para multi-colección; lo incluye como `collection_names` en el body del fetch — 2026-03-25 *(Plan 4 F2.21)*
- `apps/web/src/app/api/rag/generate/route.ts`: verificación de acceso a cada colección en `collection_names`; si alguna está denegada retorna 403 — 2026-03-25 *(Plan 4 F2.21)*
- `apps/web/src/app/api/rag/suggest/route.ts`: endpoint POST que genera 3-4 preguntas de follow-up; modo mock retorna sugerencias hardcodeadas, modo real usa el RAG server con prompt de generación + fallback al mock si falla — 2026-03-25 *(Plan 4 F2.20)*
- `apps/web/src/components/chat/RelatedQuestions.tsx`: chips de preguntas sugeridas debajo de la última respuesta; al clicar pone la pregunta en el input — 2026-03-25 *(Plan 4 F2.20)*
- `apps/web/src/components/chat/SourcesPanel.tsx`: panel colapsable de fuentes bajo cada respuesta asistente — muestra nombre del doc, fragmento (truncado a 2 líneas), relevance score como badge; visible solo cuando `sources.length > 0` — 2026-03-25 *(Plan 4 F2.19)*
- `apps/web/src/components/chat/ChatInterface.tsx`: integración de `SourcesPanel` bajo el contenido de cada mensaje asistente — 2026-03-25 *(Plan 4 F2.19)*

### Changed

- `apps/web/src/components/layout/AppShell.tsx`: reescrito como Server Component puro — delega toda la UI a `AppShellChrome` — 2026-03-25 *(Plan 4 Fase 0d)*

### Added

- `apps/web/src/components/chat/ThinkingSteps.tsx`: steps colapsables del proceso de razonamiento visibles durante streaming — simulación UI-side con timing (paso 1 inmediato, paso 2 a 700ms, paso 3 a 1500ms); se auto-colapsa 1.8s después de que el stream termina; cuando el RAG server exponga eventos SSE de tipo `thinking`, se conectan en `useRagStream` — 2026-03-25 *(Plan 4 F1.5)*
- `apps/web/src/lib/changelog.ts`: `parseChangelog(raw, limit)` extraída a utilidad testeable — 2026-03-25 *(tests Fase 1)*
- `apps/web/src/app/api/changelog/route.ts`: endpoint GET que parsea CHANGELOG.md y retorna las últimas 5 entradas + versión actual del package.json — 2026-03-25 *(Plan 4 F1.18)*
- `apps/web/src/components/layout/WhatsNewPanel.tsx`: Sheet lateral con entradas del CHANGELOG renderizadas con `marked`; `useHasNewVersion()` hook que compara versión actual con `localStorage["last_seen_version"]` — 2026-03-25 *(Plan 4 F1.18)*
- `apps/web/src/components/layout/NavRail.tsx`: logo "R" abre el panel "¿Qué hay de nuevo?" al clic; badge rojo unificado para `unreadCount > 0` o versión nueva no vista — 2026-03-25 *(Plan 4 F1.18)*
- `apps/web/src/components/chat/ChatInterface.tsx`: regenerar respuesta con botón `↻` (pone el último query del usuario en el input) F1.15; copy al portapapeles con ícono Check al confirmar F1.16; stats `{ms}ms · {N} docs` visibles al hover debajo del último mensaje asistente F1.17 — 2026-03-25
- `apps/web/src/hooks/useGlobalHotkeys.ts`: `Cmd+N` → navegar a `/chat`; `j/k` y Esc de sesiones diferidos a Fase 2 (requieren estado centralizado del panel) — 2026-03-25 *(Plan 4 F1.14)*
- `apps/web/src/lib/rag/client.ts`: `detectLanguageHint(text)` — detecta inglés (por palabras clave) y caracteres no-latinos; retorna instrucción "Respond in the same language as the user's message." si aplica — 2026-03-25 *(Plan 4 F1.13)*
- `apps/web/src/app/api/rag/generate/route.ts`: inyección de `detectLanguageHint` como system message cuando el último mensaje del usuario no está en español — 2026-03-25 *(Plan 4 F1.13)*
- `apps/web/src/app/api/notifications/route.ts`: endpoint GET que retorna eventos recientes de tipos `ingestion.completed`, `ingestion.error`, `user.created` (este último solo para admins) — 2026-03-25 *(Plan 4 F1.12)*
- `apps/web/src/hooks/useNotifications.ts`: polling cada 30s, emite toasts con sonner para notificaciones no vistas (gestionado en localStorage), retorna `unreadCount` — 2026-03-25 *(Plan 4 F1.12)*
- `apps/web/src/components/layout/NavRail.tsx`: badge rojo sobre el ícono "R" cuando `unreadCount > 0` — 2026-03-25 *(Plan 4 F1.12)*
- `packages/db/src/__tests__/saved.test.ts`: 13 tests de queries `saved_responses` (saveResponse, listSavedResponses, unsaveResponse, unsaveByMessageId, isSaved) contra SQLite en memoria — 2026-03-25 *(tests Fase 1)*
- `packages/shared/src/__tests__/focus-modes.test.ts`: 6 tests de estructura FOCUS_MODES (cantidad, IDs únicos, labels, systemPrompts, modo ejecutivo) — 2026-03-25 *(tests Fase 1)*
- `packages/shared/package.json`: agregado script `"test": "bun test src/__tests__"` para Turborepo — 2026-03-25 *(tests Fase 1)*
- `apps/web/src/lib/__tests__/export.test.ts`: 8 tests de `exportToMarkdown()` (título, colección, mensajes, fuentes, orden, vacío) — 2026-03-25 *(tests Fase 1)*
- `apps/web/src/lib/__tests__/changelog.test.ts`: 6 tests de `parseChangelog()` (Unreleased, versiones, contenido, límite, vacío, orden) — 2026-03-25 *(tests Fase 1)*
- `apps/web/src/lib/rag/__tests__/detect-language.test.ts`: 13 tests de `detectLanguageHint()` (español no inyecta, inglés inyecta, CJK/cirílico/árabe inyectan) — 2026-03-25 *(tests Fase 1)*
- `apps/web/src/hooks/useZenMode.ts`: hook `useZenMode()` — toggle con `Cmd+Shift+Z`, cierre con `Esc` — 2026-03-25 *(Plan 4 F1.11)*
- `apps/web/src/components/layout/AppShellChrome.tsx`: modo Zen oculta NavRail y SecondaryPanel; badge "ESC para salir" en `fixed bottom-4 right-4` — 2026-03-25 *(Plan 4 F1.11)*
- `packages/db/src/schema.ts`: tabla `saved_responses` (id, userId, messageId nullable, content, sessionTitle, createdAt) — 2026-03-25 *(Plan 4 F1.10)*
- `packages/db/src/queries/saved.ts`: `saveResponse`, `unsaveResponse`, `unsaveByMessageId`, `listSavedResponses`, `isSaved` — 2026-03-25 *(Plan 4 F1.10)*
- `packages/db/src/init.ts`: SQL de creación de tabla `saved_responses` con índice — 2026-03-25 *(Plan 4 F1.10)*
- `apps/web/src/app/actions/chat.ts`: Server Action `actionToggleSaved` (guarda/desuarda por messageId) — 2026-03-25 *(Plan 4 F1.10)*
- `apps/web/src/app/(app)/saved/page.tsx`: página `/saved` — lista de respuestas guardadas con empty state — 2026-03-25 *(Plan 4 F1.10)*
- `apps/web/src/lib/export.ts`: `exportToMarkdown()` (serializa sesión a MD con fuentes), `exportToPDF()` (window.print()), `downloadFile()` — 2026-03-25 *(Plan 4 F1.9)*
- `apps/web/src/components/chat/ExportSession.tsx`: Popover con opciones "Markdown" y "PDF (imprimir)" en el header del chat — 2026-03-25 *(Plan 4 F1.9)*
- `apps/web/src/components/chat/VoiceInput.tsx`: botón micrófono con Web Speech API — transcripción en tiempo real a `lang="es-AR"`, botón MicOff en rojo al grabar, fallback graceful si el browser no soporta SpeechRecognition (no renderiza nada) — 2026-03-25 *(Plan 4 F1.8)*
- `packages/shared/src/schemas.ts`: `FOCUS_MODES` + `FocusModeId` — 4 modos (detallado, ejecutivo, técnico, comparativo) con system prompt para cada uno — 2026-03-25 *(Plan 4 F1.7)*
- `apps/web/src/components/chat/FocusModeSelector.tsx`: selector de modos como pills, persistido en localStorage, `useFocusMode()` hook — 2026-03-25 *(Plan 4 F1.7)*
- `apps/web/src/app/api/rag/generate/route.ts`: prepend de system message según `focus_mode` recibido en el body — 2026-03-25 *(Plan 4 F1.7)*
- `apps/web/src/hooks/useRagStream.ts`: acepta `focusMode` en options y lo envía en el body del fetch — 2026-03-25 *(Plan 4 F1.7)*
- `apps/web/src/components/chat/ChatInterface.tsx`: integración de `ThinkingSteps` (F1.5), feedback shadcn (F1.6), modos de foco (F1.7), voice input (F1.8), ExportSession en header (F1.9), bookmark Guardar respuesta (F1.10) — 2026-03-25

### Fixed

- `apps/web/src/components/ui/theme-toggle.tsx`: hydration mismatch — el server renderizaba el `title` del botón con el tema default mientras el cliente ya conocía el tema guardado en localStorage; fix: `mounted` state con `useEffect` para evitar renderizar contenido dependiente del tema hasta después de la hidratación — 2026-03-25

### Changed

- `apps/web/src/app/globals.css`: design tokens reemplazados con paleta crema-índigo — tokens canónicos `--bg #FAFAF9`, `--sidebar-bg #F2F0F0`, `--nav-bg #18181B`, `--accent #7C6AF5`/`#9D8FF8` (dark), `--fg #18181B`/`#FAFAF9` (dark); aliases de compatibilidad apuntan a los canónicos vía `var()` para que los componentes existentes no requieran cambios; dark mode migrado de `@media prefers-color-scheme` a clase `.dark` en `<html>` (prerequisito de next-themes); directiva `@theme` agrega utilidades Tailwind para los nuevos tokens; agregado `@media print` para export de sesión (Fase 1) — 2026-03-25 *(Plan 4 Fase 0a)*

### Added

- `apps/web/src/components/layout/NavRail.tsx`: barra de íconos 44px, fondo `var(--nav-bg)` siempre oscuro, items con `Tooltip` de shadcn, ThemeToggle + logout al fondo, active state via `usePathname()` — 2026-03-25 *(Plan 4 Fase 0d)*
- `apps/web/src/components/layout/AppShellChrome.tsx`: Client Component wrapper de la shell — concentra estado de UI (zen mode, notificaciones, hotkeys en fases siguientes) — 2026-03-25 *(Plan 4 Fase 0d)*
- `apps/web/src/components/layout/SecondaryPanel.tsx`: panel contextual 168px — renderiza ChatPanel en `/chat`, AdminPanel en `/admin`, nada en otras rutas — 2026-03-25 *(Plan 4 Fase 0d)*
- `apps/web/src/components/layout/panels/ChatPanel.tsx`: panel de sesiones para rutas `/chat` con slot para SessionList (Fase 1) — 2026-03-25 *(Plan 4 Fase 0d)*
- `apps/web/src/components/layout/panels/AdminPanel.tsx`: nav admin con secciones "Gestión" y "Observabilidad" (extensible para Fase 2 sin rediseño) — 2026-03-25 *(Plan 4 Fase 0d)*
- `apps/web/src/components/providers.tsx`: ThemeProvider de next-themes (`attribute="class"`, `defaultTheme="light"`, `storageKey="rag-theme"`) — dark mode via clase `.dark` en `<html>` con script anti-FOUC automático — 2026-03-25 *(Plan 4 Fase 0c)*
- `apps/web/src/components/ui/theme-toggle.tsx`: botón Sun/Moon que alterna light/dark usando `useTheme()` de next-themes — 2026-03-25 *(Plan 4 Fase 0c)*
- `apps/web/components.json`: configuración shadcn/ui (style default, base color stone, Tailwind v4, paths `@/components/ui` y `@/lib/utils`) — 2026-03-25 *(Plan 4 Fase 0b)*
- `apps/web/src/lib/utils.ts`: función `cn()` de `clsx + tailwind-merge` — 2026-03-25 *(Plan 4 Fase 0b)*
- `apps/web/src/components/ui/`: 13 componentes shadcn instalados — button, input, textarea, dialog, popover, table, badge, avatar, separator, tooltip, sheet, sonner, command (cmdk) — 2026-03-25 *(Plan 4 Fase 0b)*
- `apps/web/src/app/layout.tsx`: `<Toaster />` de sonner + `<Providers>` de next-themes + `suppressHydrationWarning` en `<html>` — 2026-03-25 *(Plan 4 Fase 0b/0c)*

- `docs/workflows.md`: nuevo documento que centraliza los 7 flujos de trabajo del proyecto (desarrollo local, testing, git/commits, planificación, features nuevas, deploy, debugging con black box) — 2026-03-25

### Changed

- `CLAUDE.md`: corregido `better-sqlite3` → `@libsql/client`, "14 tablas" → "12 tablas", sección de tests expandida con todos los comandos, planes renombrados correctamente, nota sobre import estático del logger — 2026-03-25
- `docs/architecture.md`: corregido `better-sqlite3` → `@libsql/client`, referencia `ultra-optimize.md` → `ultra-optimize-plan1-birth.md`, documentada auth service-to-service, tabla de tablas actualizada a 12 — 2026-03-25
- `docs/onboarding.md`: comandos de test completos con conteo de tests por suite, nota sobre ubicación de `apps/web/.env.local`, referencia a `docs/workflows.md` — 2026-03-25
- `packages/db/package.json`: agregado script `"test": "bun test src/__tests__"` — Turborepo ahora corre esta suite en `bun run test` — 2026-03-25
- `packages/logger/package.json`: agregado script `"test": "bun test src/__tests__"` — 2026-03-25
- `packages/config/package.json`: agregado script `"test": "bun test src/__tests__"` — 2026-03-25
- `apps/web/package.json`: agregado script `"test": "bun test src/lib"` — 2026-03-25

### Fixed

- Pipeline de tests: `bun run test` desde la raíz ahora ejecuta los 79 tests unitarios via Turborepo — antes los workspaces no tenían script `"test"` y el pipeline completaba silenciosamente sin correr nada — 2026-03-25

### Changed

- `apps/web/src/components/chat/ChatInterface.tsx`: refactor — complejidad reducida de 48 a 22; lógica de fetch + SSE + abort extraída al hook `useRagStream`; `updateLastAssistantMessage` extraída como helper puro
- `apps/web/src/hooks/useRagStream.ts`: nuevo hook que encapsula fetch SSE, lectura del stream, abort controller y callbacks `onDelta`/`onSources`/`onError` — complejidad 19 (autónomo y testeable)
- `packages/logger/src/blackbox.ts`: refactor `reconstructFromEvents` — complejidad reducida de 34 a ~5; cada tipo de evento tiene handler nombrado (`handleAuthLogin`, `handleRagQuery`, `handleError`, `handleUserCreatedOrUpdated`, `handleUserDeleted`, `handleDefault`); despacho via `EVENT_HANDLERS` map

### Fixed

- `packages/db/src/queries/areas.ts`: `removeAreaCollection` ignoraba el parámetro `collectionName` en el WHERE — borraba todas las colecciones del área en lugar de solo la especificada; agregado `and(eq(areaId), eq(collectionName))` y actualizado import de `drizzle-orm` — 2026-03-25 *(encontrado con CodeGraphContext MCP, Plan 3 Fase 1a)*
- `apps/web/src/app/actions/areas.ts`: event types incorrectos en audit log — `actionCreateArea` emitía `"collection.created"`, `actionUpdateArea` emitía `"user.updated"`, `actionDeleteArea` emitía `"collection.deleted"`; corregidos a `"area.created"`, `"area.updated"`, `"area.deleted"` respectivamente — 2026-03-25 *(Plan 3 Fase 2a)*

### Added

- `packages/db/src/__tests__/areas.test.ts`: 8 tests de queries de áreas contra SQLite en memoria — `removeAreaCollection` (selectiva, cross-área, inexistente, última), `setAreaCollections` (reemplaza, vacía), `addAreaCollection` (default read, upsert) — 2026-03-25 *(Plan 3 Fase 1a)*

### Fixed

- `apps/web/src/app/api/auth/login/route.ts`: login con cuenta desactivada retornaba 401 en lugar de 403 — `verifyPassword` devuelve null para inactivos sin distinguir de contraseña incorrecta; agregado `getUserByEmail` check previo para detectar cuenta inactiva — 2026-03-25 *(encontrado en Fase 6e)*
- `apps/web/src/app/api/admin/db/reset/route.ts` y `seed/route.ts`: corregir errores de type-check (initDb inexistente, bcrypt-ts no disponible, null check en insert) — 2026-03-25
- `apps/web/src/lib/auth/jwt.ts`: agregar `iat` y `exp` al objeto retornado desde headers del middleware — 2026-03-25

- `packages/logger/src/backend.ts`: reemplazar lazy-load dinámico `import("@rag-saldivia/db" as any)` por import estático — en webpack/Next.js el dynamic import fallaba silenciosamente y ningún evento backend se persistía — 2026-03-25 *(encontrado en Fase 5)*
- `packages/logger/src/backend.ts`: `persistEvent` pasaba `userId=0` (SYSTEM_API_KEY) a la tabla events que tiene FK constraint a users.id — fix: escribir null cuando userId ≤ 0 — 2026-03-25 *(encontrado en Fase 5)*
- `packages/logger/package.json`: agregar `@rag-saldivia/db` como dependencia explícita del paquete logger — 2026-03-25

### Added

- `apps/web/src/components/chat/SessionList.tsx`: rename de sesión — botón lápiz en hover activa input inline; Enter/botón Confirmar guarda, Escape cancela; el nombre actualiza en la lista inmediatamente — 2026-03-25

- `apps/web/src/app/api/admin/permissions/route.ts`: endpoint POST/DELETE para asignar/quitar colecciones a áreas (necesario para flujos E2E) — 2026-03-25
- `apps/web/src/app/api/admin/users/[id]/areas/route.ts`: endpoint POST/DELETE para asignar/quitar áreas a usuarios (necesario para flujos E2E) — 2026-03-25
- `apps/web/src/app/api/admin/users/route.ts` y `[id]/route.ts`: endpoints GET/POST/DELETE/PATCH para gestión de usuarios desde la CLI — 2026-03-25
- `apps/web/src/app/api/admin/areas/route.ts` y `[id]/route.ts`: endpoints GET/POST/DELETE para gestión de áreas desde la CLI — 2026-03-25
- `apps/web/src/app/api/admin/config/route.ts` y `reset/route.ts`: endpoints GET/PATCH/POST para config RAG desde la CLI — 2026-03-25
- `apps/web/src/app/api/admin/db/migrate/route.ts`, `seed/route.ts`, `reset/route.ts`: endpoints de administración de DB desde la CLI — 2026-03-25

### Fixed

- `apps/web/src/middleware.ts`: agregar soporte para `SYSTEM_API_KEY` como auth de servicio — el CLI recibía 401 en todos los endpoints admin porque el middleware solo verificaba JWTs — 2026-03-25 *(encontrado en Fase 4b)*
- `apps/web/src/lib/auth/jwt.ts`: `extractClaims` leía Authorization header e intentaba verificarlo como JWT incluso cuando el middleware ya había autenticado via SYSTEM_API_KEY; ahora lee `x-user-*` headers del middleware si están presentes — 2026-03-25 *(encontrado en Fase 4b)*
- `apps/cli/src/client.ts`: corregir rutas de ingestion (`/api/ingestion/status` → `/api/admin/ingestion`) — 2026-03-25 *(encontrado en Fase 4d)*
- `apps/cli/src/commands/ingest.ts`: adaptador para respuesta `{ queue, jobs }` del API en lugar de array plano — 2026-03-25 *(encontrado en Fase 4d)*
- `apps/cli/src/commands/config.ts` + `apps/cli/src/index.ts`: agregar parámetro opcional `[key]` a `config get` para mostrar un parámetro específico — 2026-03-25 *(encontrado en Fase 4e)*

- `packages/config/src/__tests__/config.test.ts`: Fase 1d — 14 tests: loadConfig (env mínima, defaults, precedencia de env vars, MOCK_RAG como boolean, perfil YAML, perfil inexistente, error en producción), loadRagParams (defaults correctos, sin undefined), AppConfigSchema (validación: objeto mínimo, jwtSecret corto, logLevel inválido, URL inválida) — 2026-03-25

### Fixed

- `apps/web/src/app/actions/settings.ts`: agregar `revalidatePath("/", "layout")` para invalidar el layout al cambiar el nombre de perfil — 2026-03-25 *(encontrado en Fase 3h)*
- `apps/web/src/app/api/rag/generate/route.ts`: validación de `messages` faltante — body vacío `{}` retornaba 200 en lugar de 400; agregado guard que verifica que `messages` sea array no vacío antes de procesar — 2026-03-25 *(encontrado en Fase 2b)*
- `apps/web/src/app/api/admin/ingestion/[id]/route.ts`: DELETE con ID inexistente retornaba 200 en lugar de 404; agregado SELECT previo para verificar existencia antes del UPDATE — 2026-03-25 *(encontrado en Fase 2c)*

- Branch `experimental/ultra-optimize` iniciada — 2026-03-24
- Plan de trabajo `docs/plans/ultra-optimize.md` con seguimiento de tareas por fase — 2026-03-24
- `scripts/setup.ts`: script de onboarding cero-fricción con preflight check, instalación, migraciones, seed y resumen visual — 2026-03-24
- `.env.example` completamente documentado con todas las variables del nuevo stack — 2026-03-24
- `package.json` raíz mínimo para Bun workspaces con script `bun run setup` — 2026-03-24
- `Makefile`: nuevos targets `setup`, `setup-check`, `reset`, `dev` para el nuevo stack — 2026-03-24
- `.commitlintrc.json`: Conventional Commits enforced con scopes definidos para el proyecto — 2026-03-24
- `.husky/commit-msg` y `.husky/pre-push`: hooks de Git para validar commits y type-check — 2026-03-24
- `.github/workflows/ci.yml`: CI completo (commitlint, changelog check, type-check, tests, lint) en cada PR — 2026-03-24
- `.github/workflows/deploy.yml`: deploy solo en tag `v*` o workflow_dispatch — 2026-03-24
- `.github/workflows/release.yml`: mueve `[Unreleased]` a `[vX.Y.Z]` al publicar release — 2026-03-24
- `.github/pull_request_template.md`: PR template con sección obligatoria de CHANGELOG — 2026-03-24
- `.changeset/config.json`: Changesets para versionado semántico — 2026-03-24
- `turbo.json`: pipeline Turborepo (build → test → lint) con cache — 2026-03-24
- `package.json`: Bun workspaces root con scripts `dev`, `build`, `test`, `db:migrate`, `db:seed` — 2026-03-24
- `packages/shared`: schemas Zod completos (User, Area, Collection, Session, Message, IngestionJob, LogEvent, RagParams, UserPreferences, ApiResponse) — elimina duplicación entre Pydantic + interfaces TS — 2026-03-24
- `packages/db`: schema Drizzle completo (14 tablas), conexión singleton, queries por dominio (users, areas, sessions, events), seed, migración — 2026-03-24
- `packages/db`: tabla `ingestion_queue` reemplaza Redis — locking por columna `locked_at` — 2026-03-24
- `packages/config`: config loader TypeScript con Zod, deep-merge de YAMLs, overrides de env vars, admin-overrides persistentes — reemplaza `saldivia/config.py` — 2026-03-24
- `packages/logger`: logger estructurado (backend + frontend + blackbox + suggestions) con niveles TRACE/DEBUG/INFO/WARN/ERROR/FATAL — 2026-03-24
- `apps/web`: app Next.js 15 iniciada (package.json, tsconfig, next.config.ts) — 2026-03-24
- `apps/web/src/middleware.ts`: middleware de auth + RBAC en el edge — verifica JWT, redirecciona a login, bloquea por rol — 2026-03-24
- `apps/web/src/lib/auth/jwt.ts`: createJwt, verifyJwt, extractClaims, makeAuthCookie (cookie HttpOnly) — 2026-03-24
- `apps/web/src/lib/auth/rbac.ts`: hasRole, canAccessRoute, isAdmin, isAreaManager — 2026-03-24
- `apps/web/src/lib/auth/current-user.ts`: getCurrentUser, requireUser, requireAdmin para Server Components — 2026-03-24
- `apps/web`: endpoints auth (POST /api/auth/login, DELETE /api/auth/logout, POST /api/auth/refresh) — 2026-03-24
- `apps/web`: endpoint POST /api/log para recibir eventos del browser — 2026-03-24
- `apps/web`: página de login con form de email/password — 2026-03-24
- `apps/web`: Server Actions para usuarios (crear, eliminar, activar, asignar área) — 2026-03-24
- `apps/web`: Server Actions para áreas (crear, editar, eliminar con protección si hay usuarios) — 2026-03-24
- `apps/web`: Server Actions para chat (sesiones y mensajes) — 2026-03-24
- `apps/web`: Server Actions para settings (perfil, contraseña, preferencias) — 2026-03-24
- `apps/web/src/lib/rag/client.ts`: cliente RAG con modo mock, timeout, manejo de errores accionables — 2026-03-24
- `apps/web`: POST /api/rag/generate — proxy SSE al RAG Server con verificación de permisos — 2026-03-24
- `apps/web`: GET /api/rag/collections — lista colecciones con cache 60s filtrada por permisos — 2026-03-24
- `apps/web`: AppShell (layout con sidebar de navegación) — 2026-03-24
- `apps/web`: páginas de chat (lista de sesiones + interfaz de chat con streaming SSE + feedback) — 2026-03-24
- `apps/web`: página de admin/users con tabla y formulario de creación — 2026-03-24
- `apps/web`: página de settings con Perfil, Contraseña y Preferencias — 2026-03-24
- `apps/cli`: CLI completa con Commander + @clack/prompts + chalk + cli-table3 — 2026-03-24
- `apps/cli`: `rag status` — semáforo de servicios con latencias — 2026-03-24
- `apps/cli`: `rag users list/create/delete` — gestión de usuarios con wizard interactivo — 2026-03-24
- `apps/cli`: `rag collections list/create/delete` — gestión de colecciones — 2026-03-24
- `apps/cli`: `rag ingest start/status/cancel` — ingesta con barra de progreso — 2026-03-24
- `apps/cli`: `rag config get/set/reset` — configuración RAG — 2026-03-24
- `apps/cli`: `rag audit log/replay/export` — audit log y black box replay — 2026-03-24
- `apps/cli`: `rag db migrate/seed/reset`, `rag setup` — administración de DB — 2026-03-24
- `apps/cli`: modo REPL interactivo (sin argumentos) con selector de comandos — 2026-03-24
- `apps/web`: GET /api/audit — events con filtros (level, type, source, userId, fecha) — 2026-03-24
- `apps/web`: GET /api/audit/replay — black box reconstruction desde fecha — 2026-03-24
- `apps/web`: GET /api/audit/export — exportar todos los eventos como JSON — 2026-03-24
- `apps/web`: GET /api/health — health check público para la CLI y monitoring — 2026-03-24
- `apps/web`: página /audit con tabla de eventos filtrable por nivel y tipo — 2026-03-24
- `docs/architecture.md`: arquitectura completa del nuevo stack (servidor único, DB, auth, caching) — 2026-03-24
- `docs/blackbox.md`: guía del sistema de black box logging y replay — 2026-03-24
- `docs/cli.md`: referencia completa de todos los comandos de la CLI — 2026-03-24
- `docs/onboarding.md`: guía de 5 minutos para nuevos colaboradores — 2026-03-24
- `.gitignore`: agregado `.next/`, `.turbo/`, `logs/`, `data/*.db`, `bun.lockb` — 2026-03-24
- `apps/web/src/lib/auth/__tests__/jwt.test.ts`: tests completos del flujo de auth (JWT, RBAC) — 2026-03-24
- `apps/web/src/app/api/upload/route.ts`: endpoint de upload de archivos con validación de permisos y tamaño — 2026-03-24
- `apps/web/src/app/api/admin/ingestion/route.ts`: listado y cancelación de jobs de ingesta — 2026-03-24
- `apps/web/src/workers/ingestion.ts`: worker de ingesta en TypeScript con retry, locking SQLite, graceful shutdown — 2026-03-24
- `apps/web/src/app/(app)/upload/page.tsx`: página de upload con drag & drop — 2026-03-24
- `apps/web/src/hooks/useCrossdocDecompose.ts`: hook crossdoc portado de patches/ adaptado a Next.js — 2026-03-24
- `apps/web/src/hooks/useCrossdocStream.ts`: orquestación crossdoc (decompose → parallel queries → follow-ups → synthesis) — 2026-03-24
- `apps/web/src/app/(app)/admin/areas/page.tsx`: gestión de áreas con CRUD completo — 2026-03-24
- `apps/web/src/app/(app)/admin/permissions/page.tsx`: asignación colecciones → áreas con nivel read/write — 2026-03-24
- `apps/web/src/app/(app)/admin/rag-config/page.tsx`: config RAG con sliders y toggles — 2026-03-24
- `apps/web/src/app/(app)/admin/system/page.tsx`: estado del sistema con stats cards y jobs activos — 2026-03-24
- `packages/logger/src/rotation.ts`: rotación de archivos de log (10MB, 5 backups) — 2026-03-24
- `CLAUDE.md`: actualizado con el nuevo stack TypeScript — 2026-03-24
- `legacy/`: código del stack original (Python + SvelteKit) movido a carpeta `legacy/` — 2026-03-24
- `legacy/scripts/`: scripts bash y Python del stack original movidos a `legacy/` — 2026-03-24
- `legacy/pyproject.toml` + `legacy/uv.lock`: archivos Python movidos a `legacy/` — 2026-03-24
- `legacy/docs/`: docs del stack viejo movidos a `legacy/` (analysis, contributing, deployment, development-workflow, field-testing, plans-fase8, problems-and-solutions, roadmap, sessions, testing) — 2026-03-24
- `scripts/health-check.ts`: reemplaza health_check.sh — verifica servicios con latencias — 2026-03-24
- `README.md` y `scripts/README.md`: reescritos para el nuevo stack TypeScript — 2026-03-24
- `bun.lock`: lockfile de Bun commiteado para reproducibilidad de dependencias — 2026-03-24
- `scripts/link-libsql.sh`: script que crea symlinks de @libsql en apps/web/node_modules para WSL2 — 2026-03-24
- `scripts/test-login-final.sh`: script de test de los endpoints de auth — 2026-03-24
- `docs/plans/ultra-optimize-plan2-testing.md`: plan de testing granular en 7 fases creado — 2026-03-24
- `apps/web/src/types/globals.d.ts`: declaración de módulo `*.css` para permitir `import "./globals.css"` como side-effect sin error TS2882 — 2026-03-24
- `apps/web/src/lib/auth/__tests__/jwt.test.ts`: Fase 1a/1b — 17 tests: createJwt, verifyJwt (token inválido/firmado mal/expirado), extractClaims (cookie/header/sin token), makeAuthCookie (HttpOnly/Secure en prod), RBAC (getRequiredRole, canAccessRoute) — 2026-03-24
- `packages/db/src/__tests__/users.test.ts`: Fase 1c — 16 tests contra SQLite en memoria: createUser (email normalizado/rol/dup lanza error), verifyPassword (correcta/incorrecta/inexistente/inactivo), listUsers (vacío/múltiples/campos), updateUser (nombre/rol/desactivar), deleteUser (elimina usuario + CASCADE en user_areas) — 2026-03-24
- `packages/logger/src/__tests__/logger.test.ts`: Fase 1e — 24 tests: shouldLog por nivel (5), log.info/warn/error/debug/fatal/request no lanzan (7), output contiene tipo de evento (3), reconstructFromEvents vacío/orden/stats/usuarios/queries/errores (6), formatTimeline (3) — 2026-03-24

### Changed

- `apps/web/tsconfig.json`: excluir `**/__tests__/**` y `**/*.test.ts` del type-check — `bun:test` y asignación a `NODE_ENV` no son válidos en el contexto de `tsc` — 2026-03-24
- `package.json`: agregado `overrides: { "drizzle-orm": "^0.38.0" }` para forzar una sola instancia en la resolución de tipos — 2026-03-24
- `apps/web/package.json`: agregado `drizzle-orm` como dependencia directa para que TypeScript resuelva los tipos desde la misma instancia que `packages/db` — 2026-03-24
- `.gitignore`: agregado `*.tsbuildinfo` — 2026-03-24
- `package.json`: agregado campo `packageManager: bun@1.3.11` requerido por Turborepo 2.x — 2026-03-24
- `packages/db/package.json`: eliminado `type: module` para compatibilidad con webpack CJS — 2026-03-24
- `packages/shared/package.json`: eliminado `type: module` para compatibilidad con webpack CJS — 2026-03-24
- `packages/config/package.json`: eliminado `type: module` para compatibilidad con webpack CJS — 2026-03-24
- `packages/logger/package.json`: eliminado `type: module` para compatibilidad con webpack CJS — 2026-03-24
- `packages/*/src/*.ts`: eliminadas extensiones `.js` de todos los imports relativos (incompatibles con webpack) — 2026-03-24
- `packages/db/src/schema.ts`: agregadas relaciones Drizzle (`usersRelations`, `areasRelations`, `userAreasRelations`, etc.) necesarias para queries con `with` — 2026-03-24
- `packages/shared/src/schemas.ts`: campo `email` del `LoginRequestSchema` acepta `admin@localhost` (sin TLD) — 2026-03-24
- `apps/web/next.config.ts`: configuración completa para compatibilidad con WSL2 y monorepo Bun:
  - `outputFileTracingRoot: __dirname` para evitar detección incorrecta del workspace root
  - `transpilePackages` para paquetes workspace TypeScript
  - `serverExternalPackages` para excluir `@libsql/client` y la cadena nativa del bundling webpack
  - `webpack.externals` con función que excluye `libsql`, `@libsql/*` y archivos `.node` — 2026-03-24

### Fixed

- `apps/cli/package.json`: agregadas dependencias workspace faltantes `@rag-saldivia/logger` y `@rag-saldivia/db` — `audit.ts` importaba `formatTimeline`/`reconstructFromEvents` y `DbEvent` de esos paquetes pero Bun no los encontraba — 2026-03-24
- `packages/logger/package.json`: agregado export `./suggestions` faltante — `apps/cli/src/output.ts` importaba `getSuggestion` de `@rag-saldivia/logger/suggestions` sin que estuviera declarado en `exports` — 2026-03-24
- `apps/web/src/middleware.ts`: agregado `/api/health` a `PUBLIC_ROUTES` — el endpoint retornaba 401 al CLI y a cualquier sistema de monitoreo externo — 2026-03-24 *(encontrado en Fase 0)*
- `apps/web/src/lib/auth/__tests__/jwt.test.ts`: `await import("../rbac.js")` dentro del callback de `describe` lanzaba `"await" can only be used inside an "async" function` — movido al nivel del módulo junto con los demás imports — 2026-03-24 *(encontrado en Fase 1a)*
- `apps/web/src/lib/auth/__tests__/jwt.test.ts`: test `makeAuthCookie incluye Secure en producción` referenciaba `validClaims` definido en otro bloque `describe` — reemplazado por claims inline en el test — 2026-03-24 *(encontrado en Fase 1b)*
- `packages/logger/src/__tests__/logger.test.ts`: mismo patrón `await import` dentro de callbacks `describe` (×3 bloques) — todos los imports movidos al nivel del módulo — 2026-03-24 *(encontrado en Fase 1e)*
- `packages/logger/src/__tests__/logger.test.ts`: tests de formato JSON en producción asumían que cambiar `NODE_ENV` post-import afectaría el logger, pero `isDev` se captura en `createLogger()` al momento del import — tests rediseñados para verificar el output directamente y testear `formatJson` con datos conocidos — 2026-03-24 *(encontrado en Fase 1e)*
- `packages/db/src/queries/users.ts`: reemplazado `Bun.hash()` con `crypto.createHash('sha256')` — `Bun` global no disponible en el contexto `tsc` de `apps/web`; `crypto` nativo es compatible con Node.js y Bun — 2026-03-24
- `apps/web/src/workers/ingestion.ts`: reemplazado `Bun.file()` / `file.exists()` / `file.arrayBuffer()` con `fs/promises` `access` + `readFile` — mismo motivo que `Bun.hash` — 2026-03-24
- `apps/web/src/components/audit/AuditTable.tsx`: eliminado `import chalk from "chalk"` — importado pero nunca usado; chalk es un paquete CLI y no pertenece a un componente React — 2026-03-24
- `apps/web/src/lib/auth/current-user.ts`: `redirect` de `next/navigation` importado estáticamente en lugar de con `await import()` dinámico — TypeScript infiere correctamente que `redirect()` retorna `never`, resolviendo el error TS2322 de `CurrentUser | null` — 2026-03-24
- `packages/logger/src/backend.ts`: corregidos tres errores de tipos: (1) tipo de `_writeToFile` ajustado a `LogFilename` literal union; (2) TS2721 "cannot invoke possibly null" resuelto capturando en variable local antes del `await`; (3) import dinámico de `@rag-saldivia/db` casteado para evitar TS2307 — 2026-03-24
- `packages/logger/src/blackbox.ts`: eliminado `import type { DbEvent } from "@rag-saldivia/db"` — reemplazado por definición inline para cortar la dependencia `logger → db` que causaba TS2307 en el contexto de `apps/web` — 2026-03-24
- `.husky/pre-push`: reemplazado `bun` por ruta dinámica `$(which bun || echo /home/enzo/.bun/bin/bun)` — el PATH de husky en WSL2 no incluye `~/.bun/bin/` y el hook bloqueaba el push — 2026-03-24

- DB: migrado de `better-sqlite3` (requería compilación nativa con node-gyp, falla en Bun) a `@libsql/client` (JS puro, sin compilación, compatible con Bun y Node.js) — 2026-03-24
- DB: creado `packages/db/src/init.ts` con SQL directo (sin drizzle-kit) para inicialización en entornos sin build tools — 2026-03-24
- DB: `packages/db/src/migrate.ts` actualizado para usar `init.ts` en lugar del migrador de drizzle-kit — 2026-03-24
- DB: agregado `bcrypt-ts` como dependencia explícita de `packages/db` — 2026-03-24
- DB: agregado `@libsql/client` como dependencia de `packages/db` reemplazando `better-sqlite3` — 2026-03-24
- DB: conexión singleton migrada a `drizzle-orm/libsql` con `createClient({ url: "file:..." })` — 2026-03-24
- Next.js: Next.js 15.5 detectaba `/mnt/c/Users/enzo/package-lock.json` (filesystem Windows montado en WSL2) como workspace root, ignorando `src/app/`. Resuelto renombrando ese `package-lock.json` abandonado a `.bak` — 2026-03-24
- Next.js: resuelta incompatibilidad entre `transpilePackages` y `serverExternalPackages` al usar los mismos paquetes en ambas listas — 2026-03-24
- Next.js: webpack intentaba bundear `@libsql/client` → `libsql` (addon nativo) → cargaba `README.md` como módulo JS. Resuelto con `webpack.externals` personalizado — 2026-03-24
- Next.js: `@libsql/client` no era accesible en runtime de Node.js (los paquetes de Bun se guardan en `.bun/`, no en `node_modules/` estándar). Resuelto creando symlinks en `apps/web/node_modules/@libsql/` — 2026-03-24
- Next.js: conflicto de instancias de `drizzle-orm` (TypeError `referencedTable` undefined) al excluirlo del bundling. Resuelto manteniéndolo en el bundle de webpack — 2026-03-24
- Next.js: `.env.local` debe vivir en `apps/web/` (el directorio del proyecto), no solo en la raíz del monorepo — 2026-03-24
- Bun workspaces en WSL2: `bun install` en filesystem Windows (`/mnt/c/`) no crea symlinks en `node_modules/.bin/`. Resuelto clonando el repo en el filesystem nativo de Linux (`~/rag-saldivia/`). **En Ubuntu nativo (deployment target) este problema no existe** — 2026-03-24
- `scripts/link-libsql.sh`: workaround específico de WSL2 para crear symlinks de `@libsql` manualmente. **No necesario en Ubuntu nativo ni en producción (workstation Ubuntu 24.04)** — 2026-03-24


---

[1.0.0]: https://github.com/Camionerou/rag-saldivia/releases/tag/v1.0.0

---

<!-- Instrucciones:
  - Cada tarea completada genera una entrada en [Unreleased] antes de hacer commit
  - Al publicar una release, [Unreleased] se mueve a [vX.Y.Z] con la fecha
  - Categorías: Added | Changed | Deprecated | Removed | Fixed | Security
-->
