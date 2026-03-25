# Changelog

Todos los cambios notables de este proyecto se documentan en este archivo.

Formato basado en [Keep a Changelog](https://keepachangelog.com/es/1.1.0/).
Versionado basado en [Semantic Versioning](https://semver.org/lang/es/).

---

## [Unreleased]

### Added

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

<!-- Instrucciones:
  - Cada tarea completada genera una entrada en [Unreleased] antes de hacer commit
  - Al publicar una release, [Unreleased] se mueve a [vX.Y.Z] con la fecha
  - Categorías: Added | Changed | Deprecated | Removed | Fixed | Security
-->
