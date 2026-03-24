# RAG Saldivia — Contexto de proyecto

## Qué es este proyecto

Overlay sobre **NVIDIA RAG Blueprint v2.5.0** que agrega autenticación, RBAC, multi-colección, frontend SvelteKit 5, CLI, SDK Python, y perfiles de deployment.

- **No es un fork** — incluye el blueprint como git submodule en `vendor/rag-blueprint/` (commit a67a48c, post-v2.3.0)
- **Repo local:** `~/rag-saldivia/` — branch `main`
- **Repo remoto:** https://github.com/Camionerou/rag-saldivia
- **Deploy activo:** workstation física Ubuntu 24.04 (1x RTX PRO 6000 Blackwell, 96GB VRAM)

## Arquitectura de servicios

```
Usuario → SDA Frontend (puerto 3000, SvelteKit 5 BFF)
             ↓ JWT cookie auth
           Auth Gateway (puerto 9000, FastAPI)
             ↓ Bearer token + RBAC
           RAG Server (puerto 8081, NVIDIA Blueprint)
             ↓
           Milvus (vector DB) + NIMs (embed, rerank, OCR)
             ↓
           Nemotron-3-Super-120B-A12B (LLM, GPU 1)
```

**Perfil de deployment activo:**
- `workstation-1gpu` — producción (1 GPU, workstation física Ubuntu 24.04)

## Comandos clave

```bash
# Deploy en workstation física
cd ~/rag-saldivia && make deploy PROFILE=workstation-1gpu

# Estado del sistema
make status

# Tests
cd ~/rag-saldivia && uv run pytest saldivia/tests/ -v

# Ingestar documentos
make ingest DOCS=/path/to/docs COLLECTION=nombre

# Query
make query Q="pregunta sobre los documentos"

# CLI
make cli ARGS="users list"
make cli ARGS="collections list"
make cli ARGS="audit log"
```

## Estructura del SDK (`saldivia/`)

| Archivo | Responsabilidad |
|---------|----------------|
| `gateway.py` | FastAPI: auth, RBAC, proxy al RAG, SSE streaming, endpoints de preferences/profile/password, `POST/DELETE /admin/users/{id}/areas` |
| `auth/database.py` | SQLite AuthDB: users, areas, api_keys, sessions, ingestion_jobs, ingestion_alerts, preferences, `user_areas` (many-to-many); métodos `get_user_area_ids`, `get_user_areas`, `add_user_area`, `remove_user_area`, `count_users_in_area`; `can_access` y `get_user_collections` operan sobre unión de todas las áreas del usuario |
| `auth/models.py` | User, Area, Role dataclasses |
| `config.py` | ConfigLoader: YAML profiles + env merge + `ingestion_config()` con deep-merge de defaults |
| `providers.py` | Clientes HTTP para RAG Server, Milvus |
| `mode_manager.py` | VLM/LLM mode switching para 1-GPU |
| `collections.py` | CollectionManager: CRUD de colecciones |
| `ingestion_queue.py` | Cola de ingesta con Redis |
| `ingestion_worker.py` | Worker de la cola Redis de ingesta con retry logic y graceful shutdown |
| `watch.py` | File watcher para auto-ingest de directorios |

## Estructura del Frontend (`services/sda-frontend/`)

SvelteKit 5 BFF. Rutas principales:
- `/` → redirect a `/chat` si autenticado
- `/(auth)/login` → auth form (grupo de rutas SvelteKit)
- `/(app)/chat` → chat principal con SSE streaming
- `/(app)/chat/[id]` → sesión de chat específica
- `/(app)/collections` → gestión de colecciones
- `/(app)/collections/[name]` → detalle de colección
- `/(app)/upload` → upload de documentos
- `/(app)/settings` → configuración del usuario (5 secciones: Perfil, Contraseña, Preferencias RAG, Notificaciones, Apariencia+APIKey; form actions: `update_profile`, `update_password`, `update_preferences`, `update_notifications`)
- `/(app)/admin/users` → gestión de usuarios (solo admins): lista con columna Área+N (multi-área), modal crear usuario con multi-select de áreas
- `/(app)/admin/areas` → gestión de áreas (solo admins): CRUD completo — crear, editar nombre/descripción, eliminar con modal de usuarios bloqueantes si hay usuarios activos
- `/(app)/admin/permissions` → permisos (solo admins y `area_manager`): asignación de colecciones a áreas con selector read/write, chips de colecciones con permiso
- `/(app)/admin/rag-config` → configuración RAG (solo admins)
- `/(app)/admin/system` → estado del sistema (solo admins): stats cards (usuarios activos, áreas, colecciones con docs), jobs activos con progress bar, alertas de ingesta, botón actualizar
- `/(app)/audit` → audit log
- `/api/auth/*` → BFF endpoints de auth (session POST/DELETE)
- `/api/chat/*` → BFF proxy al gateway (sessions, stream)
  - `GET /api/chat/sessions` — lista sesiones del usuario
  - `POST /api/chat/sessions` — crea sesión nueva
  - `GET /api/chat/sessions/[id]` — detalle de sesión con mensajes
  - `DELETE /api/chat/sessions/[id]` — elimina sesión
  - `PATCH /api/chat/sessions/[id]` — renombra sesión (body: `{ title }`)
  - `POST /api/chat/sessions/[id]/messages/[msgId]/feedback` — feedback por mensaje (body: `{ rating: 'up' | 'down' }`)
- `/api/crossdoc/*` → BFF endpoints de crossdoc (decompose, subquery, synthesize)
- `/api/upload` → BFF endpoint de upload de documentos
- `/api/collections/*` → BFF endpoints de colecciones (GET, POST, DELETE)
- `/api/dev-login` → Login rápido para desarrollo (solo en dev mode)

## Planificación General

### Mission Brief (inicio de cada sesión)
Antes de tocar código: INTEL (firecrawl) + SITUATION (tests/bugs/deuda) + MISSION (prioridades) + EXECUTION (tasks).
Output: `docs/sessions/YYYY-MM-DD-brief.md`
Ver formato completo en `docs/development-workflow.md`.

### Roadmap Macro
`docs/roadmap.md` — fuente de verdad de fases. Se actualiza **solo cuando Enzo lo pide**.

---

## Workflow — OODA-SQ (ver `docs/development-workflow.md`)

El método estándar del proyecto. Cada iteración no trivial sigue:

```
OBSERVE → ORIENT → DECIDE → ACT (Implement → Simplify → Review → Docs)
```

### Tools por fase

| Fase | Tools principales |
|------|------------------|
| OBSERVE | `firecrawl` + `mcp__CodeGraphContext` + `mcp__repomix` + subagente `Explore` |
| ORIENT | skill `superpowers:brainstorming` + subagente `Plan` |
| DECIDE | skill `superpowers:writing-plans` + subagente `plan-writer` |
| IMPLEMENT | TDD directo en sesión + skill `superpowers:test-driven-development` + subagente `debugger` |
| SIMPLIFY | skill `simplify` + `mcp__CodeGraphContext__find_dead_code` + complexity tools |
| REVIEW | skill `requesting-code-review` + subagentes `gateway-reviewer` / `frontend-reviewer` / `security-auditor` |
| DOCS | subagente `doc-writer` + skill `revise-claude-md` + skill `changelog-generator` |

### Reglas fijas
- **State of the Art check** → `firecrawl search` al arrancar la sesión + antes de cada feature nueva
- **Web / docs externos** → skill `firecrawl` (NUNCA WebSearch/WebFetch)
- **Deploy a workstation** → skill `rag-saldivia:deploy`
- **Ver estado de servicios** → skill `rag-saldivia:status`
- **Commits** → SOLO cuando Enzo los pide explícitamente
- **Trivial (≤3 líneas)** → solo GATE 3 + 4 + 6

## Tests

```bash
# Todos los tests
cd ~/rag-saldivia && uv run pytest saldivia/tests/ -v

# Test específico
uv run pytest saldivia/tests/test_gateway.py -v

# Con coverage
uv run pytest saldivia/tests/ --cov=saldivia -v
```

Tests activos: `test_gateway.py`, `test_auth.py`, `test_config.py`, `test_mode_manager.py`, `test_providers.py`, `test_collections.py`, `test_gateway_extended.py` (39 tests pasan)

## Ports (en workstation)

| Puerto | Servicio |
|--------|----------|
| 3000 | SDA Frontend (SvelteKit) |
| 8081 | RAG Server (Blueprint) |
| 8082 | NV-Ingest |
| 9000 | Auth Gateway (Saldivia) |

## Archivos críticos — leer antes de modificar

- `saldivia/gateway.py` — 25KB, el corazón del sistema de auth
- `saldivia/auth/database.py` — SQLite AuthDB con helpers `_ts()`
- `config/.env.saldivia` — variables de entorno del overlay
- `config/profiles/workstation-1gpu.yaml` — perfil de producción (workstation física)
- `services/sda-frontend/src/lib/server/gateway.ts` — BFF client al gateway
- `services/sda-frontend/src/lib/types/preferences.ts` — `UserPreferences` interface + `DEFAULT_PREFERENCES`
- `services/sda-frontend/src/lib/stores/preferences.svelte.ts` — `PreferencesStore` reactivo (Svelte 5 runes)

## Patrones importantes (aprendidos en producción)

- `detect_types=PARSE_DECLTYPES` en SQLite crashea con timestamps date-only → usar helper `_ts()`
- JWT debe incluir campo `name` — el BFF lo lee para mostrar en UI
- SSE: httpx `StreamingResponse` siempre manda HTTP 200 → verificar status antes de yield
- `docker network connect --alias` falla silenciosamente si container ya está en la red → disconnect primero
- Deploy: `PYTHONPATH` no definida + `set -u` = crash → usar `${PYTHONPATH:-}`
