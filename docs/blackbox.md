# Black Box — Sistema de logging y reconstrucción

## Qué es

El "black box" es el sistema de logging de RAG Saldivia, análogo a la caja negra de los aviones. Registra **todos** los eventos del sistema (frontend y backend) en la tabla `events` de SQLite, permitiendo reconstruir exactamente qué pasó en caso de fallo.

## Cómo funciona

### 1. Escritura de eventos

**Backend (server):**
```typescript
import { log } from "@rag-saldivia/logger/backend"

log.info("auth.login", { email }, { userId })
log.error("rag.error", { error: "upstream 503" }, { userId, sessionId })
```

**Frontend (browser):**
```typescript
import { clientLog } from "@rag-saldivia/logger/frontend"

clientLog.action("client.action", { button: "send" })
clientLog.error(new Error("SSE disconnected"))
```

Los eventos del browser se envían en batches a `POST /api/log` y se persisten en la misma tabla.

### 2. Estructura de un evento

```typescript
{
  id: "uuid",
  ts: 1742823138000,           // epoch ms — usa Temporal.Now.instant()
  source: "backend",           // o "frontend"
  level: "INFO",               // TRACE | DEBUG | INFO | WARN | ERROR | FATAL
  type: "rag.query",           // ver lista completa en packages/shared/schemas.ts
  userId: 42,
  sessionId: "uuid",
  payload: { query: "...", collection: "tecpia" },
  sequence: 1234               // monotónico, para replay ordenado
}
```

### 3. Tipos de eventos registrados

| Categoría | Tipos |
|---|---|
| Auth | `auth.login`, `auth.logout`, `auth.failed`, `auth.password_changed` |
| RAG | `rag.query`, `rag.query_crossdoc`, `rag.error`, `rag.stream_started` |
| Usuarios | `user.created`, `user.updated`, `user.deleted`, `user.area_assigned` |
| Ingesta | `ingestion.started`, `ingestion.completed`, `ingestion.failed` |
| Colecciones | `collection.created`, `collection.deleted` |
| Admin | `admin.config_changed`, `admin.profile_switched` |
| Frontend | `client.action`, `client.navigation`, `client.error` |
| Sistema | `system.start`, `system.error`, `system.warning` |

## Consultar eventos

### Desde la UI

Ir a `/audit` (requiere rol `area_manager` o `admin`).

Filtros disponibles: nivel, tipo de evento.

### Desde la CLI

```bash
# Últimos 50 eventos
rag audit log

# Solo errores
rag audit log --level ERROR

# Queries RAG de hoy
rag audit log --type rag.query --limit 100

# Exportar todo
rag audit export > backup-events.json
```

### Desde la API

```bash
# Eventos filtrados
GET /api/audit?level=ERROR&limit=50

# Replay desde fecha
GET /api/audit/replay?from=2026-03-24

# Export completo
GET /api/audit/export
```

## Reconstrucción (Replay)

Después de un crash o incidente, reconstituí el estado del sistema:

```bash
rag audit replay 2026-03-24
```

Output de ejemplo:

```
=== Black Box Replay ===
Total eventos: 1247
Errores: 3
Warnings: 12
Usuarios únicos: 8
Queries RAG: 234

=== Timeline (más reciente primero) ===
  2026-03-24 14:32:19  system.error              Error: upstream 503 (RAG Server)
  2026-03-24 14:32:18  rag.query        [user=5]  Query: "¿Qué dice el contrato sobre..."
  2026-03-24 14:31:05  auth.login       [user=5]  Login: enzo@empresa.com
  ...

=== Errores registrados ===
  2026-03-24 14:32:19  rag.error: upstream 503
            → El RAG Server no está corriendo. Verificá con: make status
```

## Sugerencias automáticas de errores

El sistema mapea errores conocidos a mensajes accionables:

| Error | Sugerencia |
|---|---|
| `ECONNREFUSED :8081` | "El RAG Server no está corriendo..." |
| `jwt expired` | "Token expirado. El cliente debería hacer refresh..." |
| `SQLITE_BUSY` | "La DB está bloqueada. Verificá que no haya dos workers..." |
| `collection not found` | "La colección no existe en Milvus. Verificá con: rag collections list" |

Ver lista completa en `packages/logger/src/suggestions.ts`.

## Diferencias con el gateway Python (main)

| Aspecto | gateway.py (main) | Black Box (esta branch) |
|---|---|---|
| Storage | Solo logs en archivo | SQLite tabla `events` + consola |
| Frontend | No hay | clientLog con batching |
| Replay | No existe | `rag audit replay` |
| Sugerencias | No hay | Error → acción clara |
| Timestamps | `_ts()` hack por bug SQLite | Temporal API directo |
