# 07 — Packages Compartidos

## Estructura

```
packages/
  db/       → ORM + queries + schema (el mas grande)
  shared/   → Zod schemas + tipos (el mas importado)
  config/   → Config loader
  logger/   → Logger estructurado
```

Todos los packages usan CJS (ADR-002) — ESM rompe webpack/Next.js.
Sin imports circulares entre packages.

---

## packages/db (55 archivos)

### Proposito
Toda la interaccion con la base de datos. Schema Drizzle, queries tipadas, conexion SQLite, Redis client, migraciones, seed.

### Estructura interna
```
src/
  schema/           → 6 archivos (core, chat, messaging, events, relations, index)
  queries/          → 21 modulos de queries
  __tests__/        → 19 archivos de test
  connection.ts     → getDb() — instancia Drizzle
  redis.ts          → getRedisClient(), getBullMQConnection()
  init.ts           → Inicializacion de DB
  migrate.ts        → Migraciones Drizzle
  seed.ts           → Seed: admin + demo users + roles + permisos
  test-setup.ts     → Setup de DB para tests
  fts5.ts           → SQLite FTS5 full-text search
  index.ts          → Re-exports
```

### Exports principales
```typescript
// Conexion
export { getDb } from "./connection"
export { getRedisClient, getBullMQConnection } from "./redis"

// Schema (todos los tipos Db* y New*)
export * from "./schema"

// Queries (21 modulos)
export * from "./queries/users"
export * from "./queries/areas"
export * from "./queries/sessions"
// ... etc
```

### Dependencias
- `drizzle-orm` (0.45) — ORM
- `@libsql/client` — SQLite driver
- `ioredis` — Redis client
- `bcrypt` — Password hashing
- `bullmq` — Job queue

### Tests (18 archivos)
Cada modulo de queries tiene su test. Usan DB real en memoria (ADR-007: funciones reales, no mocks).

### Reglas criticas
- `getRedisClient()` NUNCA retorna null — lanza error (ADR-010)
- NO importar logger en redis.ts — dependencia circular (ADR-005)
- `Date.now()` para timestamps, nunca `_ts()` de SQLite (ADR-004)
- Redis con `ioredis-mock` en tests (via bunfig.toml preload)

---

## packages/shared (6 archivos)

### Proposito
Zod schemas y tipos compartidos entre `apps/web` y cualquier otro consumidor (CLI, workers, etc.).

### Estructura interna
```
src/
  schemas/
    core.ts         → Roles, Permissions, LogLevels, Auth, JWT Claims, Pagination
    rag.ts          → RAG params, collection info, chat messages, citations, feedback
    messaging.ts    → Channels, Messages, Reactions, Thread info
    index.ts        → Re-exports
  index.ts          → Main export
  __tests__/
    focus-modes.test.ts
```

### Schemas principales

```typescript
// core.ts
Role = "admin" | "area_manager" | "user"
JwtClaims = { sub, email, name, role, jti?, iat, exp }
PaginationSchema = z.object({ page, limit, sortBy?, sortOrder? })

// rag.ts
RagParamsSchema = z.object({
  temperature, top_p, max_tokens,
  vdb_top_k, reranker_top_k, use_reranker
})
CitationSchema = z.object({ title, content, source, score? })
FocusMode = "detailed" | "executive" | "technical" | "comparative"

// messaging.ts
ChannelType = "public" | "private" | "dm" | "group_dm"
MessageType = "text" | "system" | "file"
```

### Regla
Si un tipo se usa en web Y en otro lugar → va en shared. Si es solo de web → queda en web.

---

## packages/config (3 archivos)

### Proposito
Cargar configuracion desde variables de entorno, archivos YAML, y defaults.

### Estructura interna
```
src/
  loader.ts         → loadConfig(), loadRagParams(), saveRagParams()
  index.ts          → Re-exports
  __tests__/
    config.test.ts
```

### Exports
```typescript
export { loadConfig } from "./loader"
export { loadRagParams, saveRagParams } from "./loader"
export type { AppConfig } from "./loader"
```

### Como funciona
1. Lee `config/platform.yaml`, `config/models.yaml`, etc.
2. Overridea con env vars (`RAG_SERVER_URL`, `JWT_SECRET`, etc.)
3. Aplica defaults para todo lo que no esta configurado
4. Retorna objeto tipado `AppConfig`

---

## packages/logger (8 archivos)

### Proposito
Logger estructurado con niveles, colores, rotacion de archivos, y sistema de black box replay.

### Estructura interna
```
src/
  backend.ts        → log() — funcion principal, formatPretty (complejidad 29)
  frontend.ts       → clientLog() — logging en browser console
  blackbox.ts       → reconstructFromEvents(), formatTimeline() — replay de eventos
  levels.ts         → shouldLog() — filtrado por nivel
  rotation.ts       → writeToLogFile() — rotacion diaria y por tamano
  suggestions.ts    → getSuggestion() — sugerencias AI para errores comunes
  index.ts          → Main exports
  __tests__/
    logger.test.ts
```

### Exports
```typescript
export { log } from "./backend"
export { clientLog } from "./frontend"
export { reconstructFromEvents, formatTimeline } from "./blackbox"
export { shouldLog } from "./levels"
export { getSuggestion } from "./suggestions"
```

### Niveles
`DEBUG`, `INFO`, `WARN`, `ERROR` — configurable via `LOG_LEVEL` env var.

### Black Box
Sistema de replay: almacena eventos con timestamp, permite reconstruir la timeline de lo que paso. Util para debugging post-mortem.

### Nota
`formatPretty` en `backend.ts` es la funcion mas compleja de TODO el proyecto (complejidad ciclomatica 29). Formatea logs con colores ANSI, timestamps, niveles, y contexto.

`@rag-saldivia/db` se importa estaticamente en logger (ADR-005) — import dinamico fallaba silenciosamente en webpack/Next.js.
