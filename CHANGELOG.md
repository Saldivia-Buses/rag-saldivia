# Changelog

Todos los cambios notables de este proyecto se documentan aquí.

Formato basado en [Keep a Changelog](https://keepachangelog.com/es/1.1.0/).

## [Unreleased]

### Added
- **Fase 9 — Admin Pro**
- Tabla `user_areas` (many-to-many) en SQLite — migración automática idempotente; copia `area_id` existentes; `PRIMARY KEY (user_id, area_id)` con `ON DELETE CASCADE`
- `AuthDB.get_user_area_ids(user_id)` — lista IDs de áreas asignadas al usuario
- `AuthDB.get_user_areas(user_id)` — lista objetos `Area` completos vía JOIN
- `AuthDB.add_user_area(user_id, area_id)` — asignación idempotente (`INSERT OR IGNORE`)
- `AuthDB.remove_user_area(user_id, area_id)` — desasignación
- `AuthDB.count_users_in_area(area_id)` — conteo de usuarios activos en un área
- `AuthDB.get_user_collections` y `AuthDB.can_access` actualizados para operar sobre la unión de todas las áreas del usuario (multi-área)
- `POST /admin/users/{user_id}/areas` — asigna área a usuario; solo admins; 404 si usuario o área no existe
- `DELETE /admin/users/{user_id}/areas/{area_id}` — desasigna área de usuario; solo admins
- `GET /admin/users` ahora retorna campo `areas: [{id, name}]` por usuario (multi-área)
- Areas admin page (`/(app)/admin/areas`): CRUD completo — tabla con conteo de usuarios, modal crear, modal editar, modal eliminar con lista de usuarios bloqueantes cuando hay activos en el área
- Permissions admin page (`/(app)/admin/permissions`): accesible también para `area_manager`; chips de colecciones con badge read/write; selector de permiso por colección
- System admin page (`/(app)/admin/system`): stats cards (usuarios activos, áreas, colecciones con documentos); tabla de jobs activos con progress bar; tabla de alertas de ingesta; botón actualizar recarga datos
- Users admin page (`/(app)/admin/users`): columna "Área" muestra área principal + badge con cantidad extra si multi-área; modal crear usuario incluye multi-select de áreas iniciales

---

### Added
- **Fase 8 — Settings Pro**
- `src/lib/types/preferences.ts` — `UserPreferences` interface con 11 campos (colección, modo de búsqueda, VDB/reranker top-k, crossdoc, avatar, idioma, notificaciones) + `DEFAULT_PREFERENCES`
- `src/lib/stores/preferences.svelte.ts` — `PreferencesStore` reactivo (Svelte 5 runes); `init()` hidrata desde el server; getters `avatarColor` y `language`
- `gatewayGetPreferences()`, `gatewayUpdatePreferences()`, `gatewayUpdateProfile()`, `gatewayUpdatePassword()` en `src/lib/server/gateway.ts`
- Settings page reescrita con 5 secciones: **Perfil** (nombre, color de avatar, idioma), **Contraseña** (verificación de actual + validación), **Preferencias RAG** (colección, modo, top-k, toggles CrossDoc), **Notificaciones** (ingesta y alertas), **Apariencia + API Key**
- Form actions en `settings/+page.server.ts`: `update_profile` (paralelo via `Promise.all`), `update_password`, `update_preferences`, `update_notifications` — todas con manejo de error `fail(503)`
- `+layout.server.ts` carga `preferences` + alertas de admin en cada navegación via `Promise.all`; genera `notifications[]` para toasts in-app
- `+layout.svelte` hidrata `PreferencesStore` y `CrossdocStore` con prefs del usuario; toasts de alertas con guard de re-disparo (`Set` por sesión de componente)
- Columna `preferences TEXT` en tabla `users` — migration idempotente `ALTER TABLE ... ADD COLUMN`
- `AuthDB.get_user_preferences(user_id)` — merge `{**defaults, **stored}` garantiza compatibilidad con nuevos campos
- `AuthDB.update_user_preferences(user_id, prefs)` — merge parcial (preserva campos no incluidos)
- `AuthDB.update_user_password(user_id, current_pw, new_pw)` — verifica bcrypt en transacción única; retorna `False` si pw incorrecta o usuario sin password
- Gateway: `GET /auth/me/preferences`, `PATCH /auth/me/preferences`, `PATCH /auth/me/profile`, `PATCH /auth/me/password` — auth self-service con guard owner-or-admin
- `UpdatePreferencesRequest` Pydantic con validación de rangos (`vdb_top_k: ge=1, le=100`; `reranker_top_k: ge=1, le=50`; `max_sub_queries: ge=1, le=20`), `Literal` para enums de modo e idioma, `extra="ignore"`, `model_dump(exclude_none=True)` para merge parcial seguro
- Test de cross-user authorization: verifica 403 en los 4 endpoints para token de otro usuario

### Added
- **Fase 7 — Chat Sesiones Pro**
- `src/lib/chat/followups.ts` — función pura `generateFollowUps(content, originalQuery): string[]`; extrae topics de las primeras oraciones de la respuesta y genera hasta 3 preguntas sugeridas con plantillas en español
- `src/lib/chat/export.ts` — `buildMarkdown`, `buildJSON`, `downloadFile`; exporta sesiones de chat a archivo `.md` o `.json` via Blob + anchor click
- `src/lib/components/chat/FeedbackButtons.svelte` — thumbs up/down por mensaje; llama a BFF POST `/api/chat/sessions/{id}/messages/{msgId}/feedback`
- `src/lib/components/chat/FollowUpChips.svelte` — chips de sugerencias post-stream; se muestran cuando termina el streaming y se ocultan al hacer click
- Rutas BFF nuevas: PATCH `/api/chat/sessions/[id]` (rename), POST `/api/chat/sessions/[id]/messages/[msgId]/feedback` (thumbs)
- `gatewayRenameSession()` y `gatewayMessageFeedback()` en `src/lib/server/gateway.ts`
- Tabla `message_feedback` en SQLite — `(message_id, user_id, rating)` con constraint UNIQUE para upsert
- `AuthDB.rename_chat_session(session_id, user_id, title)` — actualiza título (truncado a 80 chars) y `updated_at`
- `AuthDB.upsert_message_feedback(message_id, user_id, rating)` — INSERT OR REPLACE para idempotencia
- Fix en `AuthDB.delete_chat_session`: ahora borra en orden correcto `message_feedback → chat_messages → chat_sessions` para respetar FK
- `ChatMessage.id: Optional[int] = None` como primer campo del dataclass (DB primary key para feedback)
- `Message.id?: number` en interface TypeScript del store de chat
- Inline rename en `HistoryPanel.svelte`: dblclick activa input, Enter confirma, Escape cancela; delete con confirm; pin en localStorage
- `FeedbackButtons` y `FollowUpChips` integrados en `MessageList.svelte`
- PATCH agregado a `allow_methods` en CORS middleware del gateway

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

## [0.10.0] — 2026-03-24

### Added
- **Fase 10 — Admin RAG Config**
- `ConfigLoader.get_rag_params()` — retorna los 10 parámetros RAG configurables con sus valores actuales (priority: base YAML → perfil activo → admin-overrides.yaml)
- `ConfigLoader.update_rag_params(params)` — persiste overrides en `config/admin-overrides.yaml` (gitignored); merge incremental, no reemplaza
- `ConfigLoader.reset_rag_params()` — borra `admin-overrides.yaml` y recarga config en memoria; próximo `get_rag_params()` retorna defaults
- `ConfigLoader.switch_profile(name)` — carga nuevo perfil en memoria (sin restart); valida contra path traversal
- 4 nuevos params en `ENV_MAPPING`: `LLM_TOP_P`, `LLM_TOP_K`, `VDB_TOP_K`, `RERANKER_TOP_K`
- `GET /admin/config` — retorna params RAG actuales; solo admins
- `PATCH /admin/config` — actualiza parámetros RAG; solo admins
- `POST /admin/config/reset` — restaura defaults; solo admins
- `POST /admin/profile` — switch de perfil en memoria; solo admins
- BFF routes `/api/admin/config` (GET/PATCH/POST) y `/api/admin/profile` (POST) con guard admin + manejo de `GatewayError`
- Componente `ConfigSlider.svelte` — slider + input numérico sincronizados (`$bindable()`)
- Componente `ModelSelector.svelte` — select con lista de modelos (`$bindable()`)
- Componente `GuardrailsToggle.svelte` — toggle accesible con `role="switch"` y `aria-checked`
- Componente `ProfileSwitcher.svelte` — selector de perfil con modal de confirmación y warning de restart
- Página `/(app)/admin/rag-config` — 5 secciones colapsables: Generación (temperature, max_tokens, top_p, top_k), Vector DB (vdb_top_k, reranker_top_k), Modelos (llm/embedding/reranker), Guardrails, Perfil activo
- 11 tests nuevos: 5 en `test_config.py` (get/update/merge/reset/switch) y 6 en `test_gateway_extended.py` (auth guard, CRUD endpoints, path traversal)
- Total tests: 180 (desde 169)

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
