# 02 — Arquitectura

## Principio fundamental

Un unico proceso Next.js 16 reemplaza dos servicios del stack viejo:
- Gateway Python FastAPI (puerto 9000) — ahora es `proxy.ts` + API routes
- Frontend SvelteKit (puerto 3000) — ahora es el App Router de Next.js

Esto simplifica el deploy y la operacion. La logica de negocio vive en `packages/` — extraible si algun dia se necesita API separada (ADR-003).

---

## Diagrama de flujo completo

```
                    INTERNET
                       |
                       v
              +--------+--------+
              |   Next.js :3000  |
              |                  |
              |  +------------+  |
              |  | proxy.ts   |  |  <-- Middleware Edge: JWT verify + RBAC
              |  | (Edge)     |  |      Genera x-request-id, propaga x-user-*
              |  +-----+------+  |
              |        |         |
              |  +-----v------+  |
              |  |   Router    |  |
              |  +--+--+--+---+  |
              |     |  |  |      |
              +-----|--|--|------+
                    |  |  |
          +---------+  |  +---------+
          |            |            |
    +-----v----+ +-----v----+ +----v-----+
    |  Pages   | | API      | | Server   |
    |  (SSR)   | | Routes   | | Actions  |
    |  14 rutas| | 18 endp. | | 50+ fns  |
    +-----+----+ +-----+----+ +----+-----+
          |            |            |
          +------+-----+-----+-----+
                 |           |
          +------v---+ +----v--------+
          | packages | | Servicios   |
          |  db      | | externos    |
          |  shared  | |             |
          |  config  | | +--------+  |
          |  logger  | | | Redis  |  |
          +------+---+ | +--------+  |
                 |      | +--------+  |
          +------v---+  | | RAG    |  |
          | SQLite   |  | | :8081  |  |
          | (libsql) |  | +---+----+  |
          +----------+  |     |       |
                        | +---v----+  |
                        | | Milvus |  |
                        | +--------+  |
                        | +--------+  |
                        | | LLM    |  |
                        | | 49B    |  |
                        | +--------+  |
                        +-------------+
```

---

## Flujo de autenticacion

```
1. POST /api/auth/login
   Body: { email, password }
        |
2. Route handler verifica password (bcrypt) contra DB
        |
3. createJwt() — genera JWT con HS256, jti=UUID, exp=24h
        |
4. Response: Set-Cookie: auth_token=<JWT>; HttpOnly; SameSite=Lax
        |
5. Cada request posterior:
   proxy.ts (Edge) extrae JWT de cookie o Authorization header
        |
6. jwtVerify() con jose — verifica firma y expiracion
        |
7. canAccessRoute(claims, pathname) — verifica RBAC
        |
8. Propaga headers: x-user-id, x-user-email, x-user-name, x-user-role, x-user-jti
        |
9. Route handlers: extractClaims(request) lee headers + verifica revocacion en Redis
```

### Revocacion de JWT (logout)

```
1. DELETE /api/auth/logout
        |
2. Lee jti del JWT actual
        |
3. Redis: SET revoked:{jti} 1 EX {ttl_restante}
        |
4. Clear cookie: auth_token=; Max-Age=0
        |
5. Verificacion: extractClaims() consulta Redis antes de retornar claims
   NOTA: proxy.ts (Edge) NO verifica Redis — ioredis no corre en Edge
```

---

## Flujo de una query RAG

```
1. ChatInterface (useChat de AI SDK)
   → POST /api/rag/generate
        |
2. proxy.ts verifica JWT + RBAC
        |
3. extractClaims() — verifica revocacion en Redis
        |
4. ragGenerateStream(body, signal)
   → CRITICO: verifica status HTTP ANTES de streamear
        |
5. fetch("http://localhost:8081/v1/chat/completions", { stream: true })
   → SSE en formato OpenAI: data: {"choices":[{"delta":{"content":"token"}}]}
        |
6. createRagStreamResponse(ragStream)
   → Adapter: parsea SSE de NVIDIA, escribe protocolo AI SDK Data Stream
   → text-start → text-delta (tokens) → data-sources (citations) → text-end
        |
7. Response con ReadableStream → useChat consume en el cliente
        |
MOCK MODE (MOCK_RAG=true):
   → Si OPENROUTER_API_KEY existe: fetch a OpenRouter con modelo configurable
   → Si no: respuesta hardcoded simulada
```

---

## Flujo de mensajeria (Plan 25)

```
1. MessageComposer envia mensaje
   → POST /api/messaging/messages
        |
2. Server action valida con Zod, persiste en msg_messages
        |
3. WebSocket sidecar notifica a miembros del canal
   → Protocolo custom en lib/ws/protocol.ts
        |
4. useMessaging hook actualiza UI reactivamente
        |
5. usePresence hook trackea online/offline via WebSocket
   → lastSeen se actualiza en users.last_seen
        |
6. useTyping hook muestra "usuario escribiendo..."
```

---

## Estructura de paquetes

```
packages/
  db/         → Drizzle ORM + 21 modulos de queries + 4 modulos de schema
                Exporta: getDb(), getRedisClient(), todas las queries, tipos
                Dependencias: drizzle-orm, @libsql/client, ioredis, bcrypt

  shared/     → Zod schemas + tipos compartidos
                Exporta: CoreSchema, RagSchema, MessagingSchema, tipos de roles
                Sin dependencias runtime (solo zod)

  config/     → Config loader desde env vars + YAML
                Exporta: loadConfig(), loadRagParams(), saveRagParams()

  logger/     → Logger estructurado con rotacion de archivos
                Exporta: log(), clientLog(), blackbox replay
                Funcion mas compleja: formatPretty (complejidad ciclomatica 29)
```

---

## ADRs vigentes (decisiones de arquitectura)

| ADR | Decision | Impacto |
|-----|----------|---------|
| 001 | libsql sobre better-sqlite3 | packages/db/ — libsql permite future-proof a Turso |
| 002 | CJS sobre ESM en packages | packages/*/tsconfig.json — ESM rompe webpack/Next.js |
| 003 | Next.js como proceso unico | Toda la arquitectura — 1 proceso reemplaza 2 |
| 004 | Temporal API para timestamps | Toda query con fechas — Date.now(), nunca _ts() de SQLite |
| 005 | Import estatico de db en logger | packages/logger/src/backend.ts — import dinamico fallaba |
| 006 | Estrategia de testing | Toda la suite — bun:test + happy-dom + Playwright |
| 007 | Funciones reales sobre helpers en tests | packages/db/src/__tests__/ — no mockear queries |
| 008 | Lector SSE compartido | SUPERADA — AI SDK adoptado en Plan 14 |
| 009 | Server Components primero | Todo el frontend — "use client" solo donde necesario |
| 010 | Redis como dependencia requerida | getRedisClient() nunca retorna null, lanza error |
| 011 | UI strategy | SUPERADA por Plan Maestro 1.0.x |
| 012 | Stack definitivo | Decision formal del stack completo |

---

## Grafo de dependencias (medido por imports)

### Nodos centrales (mayor blast radius)

| Modulo | Importado por N archivos | Un cambio rompe... |
|--------|-------------------------|-------------------|
| `@rag-saldivia/db` | 45 | Queries, actions, routes, components |
| `@/lib/auth/current-user` | 43 | Auth en SSR + todas las actions |
| `@/lib/utils` | 39 | Styling de todos los componentes (cn()) |
| `@rag-saldivia/shared` | 35 | Tipos en web + packages |
| `@/lib/auth/jwt` | 33 | Login, logout, refresh, proxy |
| `@/lib/safe-action` | 13 | TODAS las server actions |
| `@/lib/api-utils` | 13 | Todos los API routes |
| `@rag-saldivia/logger` | 11 | Logging server-side |

### Dependencias inter-packages

```
shared ← (ninguna dep interna)         config ← (ninguna dep interna)
   ↑
   |-- db ← shared (tipos)
   |        ↑
   |        └── logger ← db (import estatico, ADR-005)
   |
   └── apps/web importa todo
```

**Sin dependencias circulares.** La unica relacion delicada es logger → db (import estatico, ADR-005).

### Riesgo de cambio por modulo

| Si cambia... | Afecta | Riesgo |
|-------------|--------|--------|
| `db/schema/*` | 21 queries + actions + routes (migracion DB) | CRITICO |
| `lib/auth/jwt.ts` | 33 archivos (proxy + routes + actions) | CRITICO |
| `lib/safe-action.ts` | 10 archivos de actions + componentes que las llaman | CRITICO |
| `@rag-saldivia/shared` | 35 archivos en web + packages | ALTO |
| `lib/utils.ts` | 39 archivos (cn() en todos los componentes) | MEDIO |
| `lib/rag/client.ts` | 3 archivos (generate route, mock mode) | MEDIO |
| `packages/logger/*` | 11 archivos (invisible, no rompe features) | BAJO |

### Modulos aislados (bajo riesgo de cambio)

`lib/webhook.ts`, `lib/changelog.ts`, `lib/export.ts`, cada hook individual, `workers/messaging-notifications.ts`, cada componente de `messaging/` — todos tienen 1-2 dependientes.

---

## State management — como fluye el estado

El proyecto NO usa ninguna libreria de state management (Zustand, Jotai, Redux). Esto es una decision implicita, no accidental.

### Patron: Server-first con islands de client state

```
Server Components (pages, layouts)
  |
  | props (datos de DB via queries directas)
  |
  v
Client Components ("use client")
  |
  | useState/useTransition/useOptimistic — estado LOCAL
  | useChat (AI SDK) — estado de streaming
  | useCallback/useMemo — memoizacion
  |
  v
Server Actions (mutaciones)
  |
  | revalidatePath() — invalida cache de Next.js
  |
  v
Server Components re-render con datos frescos
```

**178 usos de useState/useContext/useOptimistic/useTransition** en 39 archivos. El estado vive local en cada componente — no hay store global compartido.

**Componentes con mas estado local:**
- `AdminAreas.tsx` — 14 usos (CRUD + optimistic updates)
- `AdminUsers.tsx` — 13 usos
- `AdminRoles.tsx` — 12 usos
- `ChatInterface.tsx` — 11 usos (streaming + UI)
- `AdminCollections.tsx` — 8 usos
- `CollectionsList.tsx` — 8 usos

**Unico contexto global:** `ChatLayout.tsx` tiene `createContext` para el toggle de sidebar. No hay otro contexto compartido.

**Por que funciona sin store global:** Cada pagina es un Server Component que fetchea sus datos y los pasa como props. Las mutaciones van via server actions que llaman `revalidatePath()`. No hay estado que necesite compartirse entre paginas — cada navegacion es un re-fetch server-side.

**Riesgo:** Si alguna feature futura necesita estado compartido entre rutas (ej: notificaciones en tiempo real visibles en toda la app), habria que agregar un provider o una libreria. Hoy no es necesario.

---

## Error handling — como se manejan los errores

### Capa por capa

**118 usos de catch/onError/toast en 53 archivos.** El error handling existe pero no esta formalizado en un patron unico.

**Middleware edge (proxy.ts):**
- Token invalido → 401 JSON o redirect a `/login`
- Sin permisos → 403 JSON o redirect a `/`
- Error cripto → token tratado como invalido

**API routes:**
- Patron comun: `try { ... } catch { return NextResponse.json({ error }, { status: 500 }) }`
- RAG client: errores tipados (`UNAVAILABLE`, `TIMEOUT`, `FORBIDDEN`, `UPSTREAM_ERROR`) con sugerencia de fix
- Validacion Zod falla → 400 con mensaje del schema

**Server actions (next-safe-action):**
- `handleServerError: (e) => e.message` — el error se propaga como string al cliente
- El caller accede a `result?.serverError` o `result?.data`
- No hay toast automatico — cada componente decide como mostrar el error

**Componentes:**
- `ErrorBoundary` global en el arbol React (catch de renders fallidos)
- `error.tsx` en `/chat` y `/messaging` — paginas de error por segmento
- `sonner` (toast) disponible pero NO usado sistematicamente para errores de actions
- Patron comun en admin: `try { await action() } catch { /* silenciado o console.error */ }`

**Redis caido:**
- `isRevoked()` retorna `false` (fail-open) — el usuario sigue autenticado
- `getCachedRagCollections()` fallback a fetch directo — funciona sin cache
- BullMQ se reconecta automaticamente

**RAG server caido:**
- `ragGenerateStream()` retorna `{ error: { code: "UNAVAILABLE", suggestion: "..." } }`
- El route handler convierte a JSON 503
- `MOCK_RAG=true` es el workaround para desarrollo

### Lo que falta

| Gap | Impacto |
|-----|---------|
| Toast automatico en server action errors | Los admin components silencian errores |
| Retry automatico en fetch failures | El usuario tiene que recargar manualmente |
| Error reporting centralizado (Sentry, etc.) | Errores en produccion no se trackean |
| Rate limit feedback al usuario | Si se excede el rate limit, no hay UI feedback |
