# Changelog

Todos los cambios notables de este proyecto se documentan aquí.

Formato basado en [Keep a Changelog](https://keepachangelog.com/es/1.1.0/).

## [Unreleased]

### Added
- **Fase 6 — Upload Inteligente (Tasks 8-13)**
- `src/lib/ingestion/hash.ts` — `computeSHA256(file)` via Web Crypto API (sin dependencias externas)
- `src/routes/api/documents/check/+server.ts` — BFF GET proxy a `GET /v1/documents/check`
- `src/lib/components/upload/DuplicateModal.svelte` — modal para archivos ya indexados o con fallo previo; muestra nombre, colección, fecha relativa y páginas
- `src/lib/upload/queue.svelte.ts` — `UploadQueue` con Svelte 5 `$state` runes; paralelismo tier-aware: 2 slots concurrentes para small/tiny, 1 slot para medium/large; emite `CustomEvent('upload:start')` para avanzar la cola
- `src/routes/api/ingestion/[jobId]/alert/+server.ts` — BFF POST proxy a `POST /v1/jobs/{jobId}/alert`
- `src/routes/(app)/admin/system/+page.server.ts` — carga alertas no resueltas via SSR al cargar la página
- `src/routes/api/admin/alerts/+server.ts` — BFF GET proxy a `GET /v1/admin/alerts`
- `src/routes/api/admin/alerts/[alertId]/resolve/+server.ts` — BFF PATCH proxy a `PATCH /v1/admin/alerts/{alertId}/resolve`
- `gatewayListAlerts()`, `gatewayResolveAlert()`, interface `IngestionAlert` en `src/lib/server/gateway.ts`
- Tests: `src/lib/ingestion/hash.test.ts` (3 tests), 2 nuevos tests en `poller.test.ts` (stall exhaustion + backoff retry)
- **Fase 6 — Upload Inteligente (Tasks 1-7)**
- Bloque `ingestion:` en `config/profiles/workstation-1gpu.yaml` — todos los parámetros de ingesta configurables: `parallel_slots_small/large`, `client/server_max_retries`, `retry_backoff_base`, `stall_check_interval`, y tiers `tiny/small/medium/large` con `max_pages`, `poll_interval`, `deadlock_threshold`, `timeout`
- `_INGESTION_DEFAULTS` dict en `saldivia/config.py` — valores por defecto de ingesta (movido antes de la clase)
- `ConfigLoader.ingestion_config()` — método que hace deep-merge de defaults con el perfil activo; `self._config` almacena el config completo
- Constantes de módulo en `saldivia/auth/database.py`: `_STALL_CHECK_STATES`, `_ACTIVE_UI_STATES`, `_SALDIVIA_VERSION`
- Migración SQLite automática: columnas `file_hash`, `retry_count`, `last_checked` en tabla `ingestion_jobs`
- Tabla `ingestion_alerts` — registra jobs fallidos con contexto completo (tier, page_count, file_hash, error, retry_count, progress_at_failure, gateway_version, resolved_at, resolved_by, notes)
- `AuthDB.check_file_hash(file_hash, collection)` — deduplicación por SHA-256 antes de ingestar
- `AuthDB.get_all_active_ingestion_jobs()` — snapshot para stall checker
- `AuthDB.increment_ingestion_retry(job_id)` — incrementa contador y actualiza `last_checked`
- `AuthDB.create_ingestion_alert(...)` — registra alerta de fallo con todos los campos de diagnóstico
- `AuthDB.list_ingestion_alerts(resolved=None)` — lista alertas filtrando por estado de resolución
- `AuthDB.resolve_ingestion_alert(alert_id, resolved_by, notes=None)` — marca alerta como resuelta
- `AuthDB.create_ingestion_job` acepta parámetro opcional `file_hash: str | None`
- `INGEST_CACHE_DIR` — path configurable via env var (default `/tmp/saldivia-ingest`) para cache de archivos durante upload
- `_cleanup_ingest_cache(job_id)` — helper interno en `gateway.py`; limpia cache al completar o fallar permanentemente
- `POST /v1/documents` computa SHA-256 del archivo antes del POST al ingestor y escribe el archivo en cache local para retry server-side
- `GET /v1/documents/check?hash=<sha256>&collection=<collection>` — endpoint de deduplicación client-opt-in; requiere auth; devuelve `{"exists": bool, state, filename, pages, indexed_at, job_id}`
- `GET /v1/admin/alerts` — lista alertas de ingesta con filtrado opcional por `resolved=true/false`; solo admins
- `PATCH /v1/admin/alerts/{alert_id}/resolve` — marca alerta resuelta con campos `resolved_by` y `notes`
- `POST /v1/jobs/{job_id}/alert` — cliente crea alerta al agotar reintentos; verifica ownership del job
- `require_user` — dependency FastAPI que centraliza el guard de autenticación (reemplaza duplicados inline)
- `_run_stall_check(ingestion_cfg)` — verifica jobs activos contra el ingestor; estados `running` no consumen `retry_count`
- `_stall_checker_loop()` — background task asyncio; config de ingesta cargado una vez al startup
- Hook `SessionStart` en `.claude/settings.json` — inyecta Mission Brief obligatorio al inicio de cada sesión
- `docs/roadmap.md` — fuente de verdad del roadmap del proyecto
- `docs/sessions/` — directorio para briefs de sesión
- `docs/superpowers/specs/2026-03-23-ooda-sq-workflow-design.md` — design doc del workflow OODA-SQ
- `docs/superpowers/specs/2026-03-23-planning-general-design.md` — design doc del Mission Brief + Roadmap

### Changed
- `src/lib/ingestion/poller.ts` — backoff exponencial al detectar stall (30s * 2^n); `_reportAlert()` al agotar reintentos
- `src/routes/(app)/admin/system/+page.svelte` — tabla de alertas de ingesta con botón "Resolver"
- `saldivia/gateway.py`: migración de `@app.on_event("startup")` a `@asynccontextmanager lifespan` (elimina DeprecationWarning de FastAPI)
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
