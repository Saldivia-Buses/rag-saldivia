# RAG Saldivia — Contexto de proyecto

## Qué es este proyecto

Overlay sobre **NVIDIA RAG Blueprint v2.5.0** que agrega autenticación, RBAC, multi-colección, frontend SvelteKit 5, CLI, SDK Python, y perfiles de deployment.

- **No es un fork** — incluye el blueprint como git submodule en `vendor/rag-blueprint/` (commit a67a48c, post-v2.3.0)
- **Repo local:** `~/rag-saldivia/` — branch `main`
- **Repo remoto:** https://github.com/Camionerou/rag-saldivia
- **Deploy activo:** instancia Brev `nvidia-enterprise-rag-deb106` (2x RTX PRO 6000 Blackwell)

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

**Perfiles de deployment:**
- `brev-2gpu` — producción actual (2 GPUs, Brev)
- `workstation-1gpu` — desarrollo local con mode switching
- `full-cloud` — sin GPU, todo via API

## Comandos clave

```bash
# Deploy a Brev
ssh nvidia-enterprise-rag-deb106
cd ~/rag-saldivia && make deploy PROFILE=brev-2gpu

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
| `gateway.py` | FastAPI: auth, RBAC, proxy al RAG, SSE streaming |
| `auth/database.py` | SQLite AuthDB: users, areas, api_keys, sessions |
| `auth/models.py` | User, Area, Role dataclasses |
| `config.py` | ConfigLoader: YAML profiles + env merge |
| `providers.py` | Clientes HTTP para RAG Server, Milvus |
| `mode_manager.py` | VLM/LLM mode switching para 1-GPU |
| `collections.py` | CollectionManager: CRUD de colecciones |
| `ingestion_queue.py` | Cola de ingesta con Redis |
| `ingestion_worker.py` | Worker de la cola Redis de ingesta con retry logic y graceful shutdown |
| `cache.py` | Caché HTTP para respuestas frecuentes |
| `watch.py` | File watcher para auto-ingest de directorios |
| `mcp_server.py` | MCP server (no usar por ahora — RAG debe estar running) |

## Estructura del Frontend (`services/sda-frontend/`)

SvelteKit 5 BFF. Rutas principales:
- `/` → redirect a `/chat` si autenticado
- `/(auth)/login` → auth form (grupo de rutas SvelteKit)
- `/(app)/chat` → chat principal con SSE streaming
- `/(app)/chat/[id]` → sesión de chat específica
- `/(app)/collections` → gestión de colecciones
- `/(app)/collections/[name]` → detalle de colección
- `/(app)/upload` → upload de documentos
- `/(app)/settings` → configuración del usuario
- `/(app)/admin/users` → gestión de usuarios (solo admins)
- `/(app)/admin/areas` → gestión de áreas (solo admins)
- `/(app)/admin/permissions` → permisos (solo admins)
- `/(app)/admin/rag-config` → configuración RAG (solo admins)
- `/(app)/admin/system` → estado del sistema (solo admins)
- `/(app)/audit` → audit log
- `/api/auth/*` → BFF endpoints de auth (session POST/DELETE)
- `/api/chat/*` → BFF proxy al gateway (sessions, stream)
- `/api/crossdoc/*` → BFF endpoints de crossdoc (decompose, subquery, synthesize)
- `/api/upload` → BFF endpoint de upload de documentos
- `/api/collections/*` → BFF endpoints de colecciones (GET, POST, DELETE)
- `/api/dev-login` → Login rápido para desarrollo (solo en dev mode)

## Skills obligatorias en este proyecto

- **Features/bugs no triviales** → invocar `superpowers:brainstorming` PRIMERO, siempre, sin excepción
- **Explorar codebase** → `CodeGraphContext` MCP + `repomix` MCP
- **Web / docs externos** → skill `firecrawl` (NUNCA WebSearch/WebFetch)
- **Deploy a Brev** → skill `rag-saldivia:deploy`
- **Ver estado de servicios** → skill `rag-saldivia:status`
- **Operaciones Brev generales** → skill `brev-cli`
- **Commits** → SOLO cuando Enzo los pide explícitamente

## Workflow para cambios NO triviales

```
Research (CGC + Repomix + firecrawl)
  → Spec (superpowers:brainstorming)
    → Plan (superpowers:writing-plans)
      → Implementación (superpowers:subagent-driven-development)
        → Review + Tests (superpowers:requesting-code-review)
          → Commit (solo si Enzo lo pide)
```

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

## Ports (en Brev)

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
- `config/profiles/brev-2gpu.yaml` — perfil de producción
- `services/sda-frontend/src/lib/server/gateway.ts` — BFF client al gateway

## Patrones importantes (aprendidos en producción)

- `detect_types=PARSE_DECLTYPES` en SQLite crashea con timestamps date-only → usar helper `_ts()`
- JWT debe incluir campo `name` — el BFF lo lee para mostrar en UI
- SSE: httpx `StreamingResponse` siempre manda HTTP 200 → verificar status antes de yield
- `docker network connect --alias` falla silenciosamente si container ya está en la red → disconnect primero
- Deploy: `PYTHONPATH` no definida + `set -u` = crash → usar `${PYTHONPATH:-}`
