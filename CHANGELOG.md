# Changelog

Todos los cambios notables de este proyecto se documentan aquí.

Formato basado en [Keep a Changelog](https://keepachangelog.com/es/1.1.0/).

## [Unreleased]

### Added
- Hook `SessionStart` en `.claude/settings.json` — inyecta Mission Brief obligatorio al inicio de cada sesión
- `docs/roadmap.md` — fuente de verdad del roadmap del proyecto
- `docs/sessions/` — directorio para briefs de sesión
- `docs/superpowers/specs/2026-03-23-ooda-sq-workflow-design.md` — design doc del workflow OODA-SQ
- `docs/superpowers/specs/2026-03-23-planning-general-design.md` — design doc del Mission Brief + Roadmap

### Changed
- Workflow de desarrollo actualizado a OODA-SQ + Mission Brief
- `docs/development-workflow.md` reescrito con tools explícitos por fase (MCPs, skills, subagentes)
- `CLAUDE.md` actualizado con sección de planificación general

---

## [0.5.6] — 2026-03-23

### Added
- CORS middleware configurable via `CORS_ORIGINS` env var
- Rate limiting por usuario en `/auth/session` (5 intentos, 1 min lockout)
- Upload limit de 1GB + sanitización de filename en `/v1/documents`
- Tests para `/v1/generate` y `/v1/search` (happy path, 500, timeout, empty)
- Tests para `ingestion_worker` (process_job, retry, shutdown, Redis backoff)
- Tests para `ingestion_queue` con fakeredis (>80% coverage)
- Dockerfile para `ingestion-worker`

### Fixed
- Path traversal en upload de documentos
- Tests con fixtures `admin_user` duplicadas → movidas a `conftest.py`

### Changed
- `PYTHONPATH` con `${PYTHONPATH:-}` para evitar crash con `set -u` en deploy
- Perfil de deployment `workstation-1gpu` como perfil de producción principal

## [0.5.5] — 2026-03-20

### Added
- Audit completo de documentación del proyecto
- `docs/architecture.md`, `docs/deployment.md` actualizados
- `docs/problems-and-solutions.md` con patrones aprendidos en producción

## [0.5.0] — 2026-03-19

### Added
- Frontend SvelteKit 5 BFF (`services/sda-frontend/`)
- Chat con SSE streaming
- Gestión de colecciones desde UI
- Upload de documentos desde UI
- Panel de administración (users, areas, permissions, rag-config, system)
- Audit log desde UI
- Pipeline crossdoc (decompose → subquery → synthesize)

### Changed
- Gateway FastAPI refactorizado con RBAC completo
- Auth basada en JWT con campo `name` para el BFF

## [0.4.0] — 2026-03-18

### Added
- Fase 1 del frontend: fundación SvelteKit 5 + BFF pattern
- Multi-provider config system (OpenRouter, NVIDIA NIMs)
- `saldivia/config.py` con ConfigLoader + profiles YAML

## [0.3.0] — 2026-03-17

### Added
- Gateway FastAPI inicial con auth básica
- SDK Python (`saldivia/`)
- CLI inicial
- Soporte para múltiples colecciones Milvus
