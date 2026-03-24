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

---

<!-- Instrucciones:
  - Cada tarea completada genera una entrada en [Unreleased] antes de hacer commit
  - Al publicar una release, [Unreleased] se mueve a [vX.Y.Z] con la fecha
  - Categorías: Added | Changed | Deprecated | Removed | Fixed | Security
-->
