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

---

<!-- Instrucciones:
  - Cada tarea completada genera una entrada en [Unreleased] antes de hacer commit
  - Al publicar una release, [Unreleased] se mueve a [vX.Y.Z] con la fecha
  - Categorías: Added | Changed | Deprecated | Removed | Fixed | Security
-->
