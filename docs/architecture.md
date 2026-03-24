# Arquitectura — RAG Saldivia (experimental/ultra-optimize)

> Branch: `experimental/ultra-optimize`
> Última actualización: 2026-03-24

---

## Stack actual (main branch)

```
Usuario → SvelteKit :3000 → gateway.py :9000 → RAG :8081
                                                      ↓
                                              Milvus + NIMs
                                                      ↓
                                         Nemotron-3-Super-120B
```

## Stack objetivo (esta branch)

```
Usuario → Next.js :3000 ——————————————————→ RAG :8081
           (UI + auth + proxy)                    ↓
                                          Milvus + NIMs
                                                  ↓
                                     Nemotron-3-Super-120B
```

### Beneficios del servidor único

- Un solo proceso para deployar y monitorear
- Elimina un salto de red (SvelteKit → gateway.py → RAG)
- Server Components: cero JS al browser por defecto
- Server Actions: mutaciones sin capa de API separada

---

## Estructura del monorepo

```
rag-saldivia/
├── apps/
│   ├── web/                  → Next.js 15 (servidor único)
│   │   ├── src/
│   │   │   ├── app/          → App Router
│   │   │   │   ├── (auth)/login/
│   │   │   │   ├── (app)/chat/
│   │   │   │   ├── (app)/admin/
│   │   │   │   ├── (app)/settings/
│   │   │   │   ├── (app)/audit/
│   │   │   │   ├── api/auth/        → login, logout, refresh
│   │   │   │   ├── api/rag/         → generate (SSE proxy), collections
│   │   │   │   ├── api/audit/       → events, replay, export
│   │   │   │   ├── api/log/         → recibe eventos del browser
│   │   │   │   └── api/health/
│   │   │   ├── actions/      → Server Actions (chat, users, areas, settings)
│   │   │   ├── components/   → React components
│   │   │   ├── lib/
│   │   │   │   ├── auth/     → jwt.ts, rbac.ts, current-user.ts
│   │   │   │   └── rag/      → client.ts (proxy + mock)
│   │   │   └── middleware.ts → JWT + RBAC en el edge
│   └── cli/                  → CLI TypeScript (rag users/collections/ingest/...)
│       └── src/
│           ├── index.ts      → Commander + REPL interactivo
│           ├── client.ts     → HTTP client al servidor
│           ├── output.ts     → colores, tablas, progreso
│           └── commands/     → status, users, collections, ingest, audit, config
├── packages/
│   ├── shared/               → Zod schemas + TypeScript types
│   ├── db/                   → Drizzle ORM + better-sqlite3
│   │   ├── src/schema.ts     → 14 tablas SQLite
│   │   ├── src/connection.ts → singleton + WAL + pragmas
│   │   └── src/queries/      → users, areas, sessions, events
│   ├── config/               → YAML loader + Zod validation
│   └── logger/               → backend log + frontend log + black box
├── config/                   → YAMLs de configuración (sin cambios)
├── patches/                  → Patches del blueprint NVIDIA (sin cambios)
├── vendor/                   → Submódulo NVIDIA (sin cambios)
├── docs/
│   ├── plans/ultra-optimize.md
│   ├── architecture.md       → este archivo
│   ├── cli.md
│   ├── blackbox.md
│   └── onboarding.md
├── scripts/
│   └── setup.ts              → script de onboarding
├── CHANGELOG.md
├── turbo.json
└── package.json              → Bun workspaces root
```

---

## Flujo de autenticación

```
1. Usuario → POST /api/auth/login (email + password)
2. Server verifica contra packages/db (bcrypt)
3. Server crea JWT firmado con JWT_SECRET
4. Server setea cookie HttpOnly auth_token
5. Todas las requests → middleware.ts verifica el JWT
6. JWT claims (sub, role) se pasan como headers x-user-*
7. Server Components leen getCurrentUser() vía React.cache()
```

## Flujo de RAG query

```
1. Usuario escribe en ChatInterface (Client Component)
2. fetch POST /api/rag/generate con el historial + collection_name
3. Server verifica permisos de colección (packages/db canAccessCollection)
4. Server hace proxy SSE → RAG Server :8081
5. Server verifica status HTTP ANTES de streamear (fix del bug de gateway.py)
6. Stream se reenvía al browser con ReadableStream
7. ChatInterface acumula los chunks y actualiza el estado
8. Al terminar, Server Action addMessage() persiste la conversación
```

## Base de datos única (SQLite)

Todos los datos de la aplicación en un solo archivo SQLite (`data/app.db`).

| Tabla | Contenido |
|---|---|
| `users` | Usuarios con roles y preferencias |
| `areas` | Áreas de la organización |
| `user_areas` | Many-to-many usuarios ↔ áreas |
| `area_collections` | Permisos área ↔ colección |
| `chat_sessions` | Sesiones de chat |
| `chat_messages` | Mensajes de cada sesión |
| `message_feedback` | Feedback thumbs up/down |
| `ingestion_jobs` | Jobs de ingesta con estado |
| `ingestion_alerts` | Alertas de jobs fallidos |
| `ingestion_queue` | Cola de ingesta (reemplaza Redis) |
| `audit_log` | Acciones de usuarios (legacy) |
| `events` | **Black Box** — todos los eventos del sistema |

### Reemplazo de Redis

La cola de ingesta que antes usaba Redis se reemplaza con la tabla `ingestion_queue` + locking optimista:

```sql
-- Worker toma un job (transacción atómica, SQLite serializa writes)
BEGIN;
SELECT id FROM ingestion_queue
WHERE status = 'pending' AND locked_at IS NULL
ORDER BY priority DESC, created_at ASC
LIMIT 1;
UPDATE ingestion_queue SET status = 'locked', locked_at = ?, locked_by = ?
WHERE id = ?;
COMMIT;
```

## Black Box System

Todos los eventos del sistema (frontend + backend) se almacenan en la tabla `events`:

```typescript
{
  id: string        // UUID
  ts: number        // epoch ms (Temporal.Now.instant().epochMilliseconds)
  source: 'frontend' | 'backend'
  level: LogLevel   // TRACE..FATAL
  type: EventType   // 'auth.login' | 'rag.query' | 'user.created' | ...
  userId: number | null
  sessionId: string | null
  payload: JSON     // contexto del evento
  sequence: number  // monotónico, para replay ordenado
}
```

Para reconstruir el estado después de un crash:
```bash
rag audit replay --from 2026-03-24
```

Internamente usa `packages/logger/blackbox.ts:reconstructFromEvents()`.

---

## Caching

Next.js 15 tiene 4 capas de caché, todas en uso:

| Capa | Qué cachea | TTL |
|---|---|---|
| `React.cache()` | queries DB dentro de una request | request |
| `unstable_cache` | colecciones del RAG | 60s |
| Full Route Cache | páginas estáticas (admin config) | build |
| Router Cache | navegación del cliente | session |

Invalidación: `revalidateTag('collections')` al crear/eliminar colecciones.
