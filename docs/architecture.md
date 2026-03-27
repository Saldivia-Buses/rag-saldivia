# Arquitectura — RAG Saldivia (experimental/ultra-optimize)

> Branch: `experimental/ultra-optimize`
> Última actualización: 2026-03-27 (Plan 8 — Fases 1–4)

---

## Stack anterior (branch main)

```
Usuario → SvelteKit :3000 → gateway.py :9000 → RAG :8081
                                                      ↓
                                              Milvus + NIMs
                                                      ↓
                                         Nemotron-Super-49B
```

## Stack actual (esta branch)

```
Usuario → Next.js :3000 ——————————————————→ RAG :8081
           (UI + auth + proxy)                    ↓
                                          Milvus + NIMs
                                                  ↓
                                     Nemotron-Super-49B
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
│   ├── web/                  → Next.js 16 (servidor único)
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
│   │   │   ├── actions/      → Server Actions (chat, users, areas, settings, auth, projects,
│   │   │   │                    webhooks, reports, external-sources, share, memory)
│   │   │   ├── components/   → React components
│   │   │   ├── hooks/        → useRagStream, useCrossdocStream, useCrossdocDecompose
│   │   │   ├── lib/
│   │   │   │   ├── auth/     → jwt.ts, rbac.ts, current-user.ts
│   │   │   │   ├── rag/      → client.ts (proxy + mock), collections-cache.ts,
│   │   │   │   │                stream.ts (SSE reader compartido — ver ADR-008)
│   │   │   │   ├── safe-action.ts  → cliente next-safe-action con middleware auth
│   │   │   │   └── utils.ts        → formatDate, formatDateTime, cn
│   │   │   ├── workers/      → ingestion.ts (worker de ingesta)
│   │   │   └── middleware.ts → JWT + RBAC en el edge
│   └── cli/                  → CLI TypeScript (rag users/collections/ingest/...)
│       └── src/
│           ├── index.ts      → Commander + REPL interactivo
│           ├── client.ts     → HTTP client al servidor
│           ├── output.ts     → colores, tablas, progreso
│           └── commands/     → status, users, collections, ingest, audit, config
├── packages/
│   ├── shared/               → Zod schemas + TypeScript types
│   ├── db/                   → Drizzle ORM + @libsql/client
│   │   ├── src/schema.ts     → 12 tablas SQLite
│   │   ├── src/connection.ts → singleton WAL
│   │   ├── src/queries/      → users, areas, sessions, events
│   │   └── drizzle.config.ts → Drizzle Kit (schema.ts es la única fuente de verdad)
│   ├── config/               → YAML loader + Zod validation
│   └── logger/               → backend log + frontend log + black box
├── config/                   → YAMLs de configuración (sin cambios)
├── patches/                  → Patches del blueprint NVIDIA (sin cambios)
├── vendor/                   → Submódulo NVIDIA (sin cambios)
├── docs/
│   ├── plans/
│   │   ├── ultra-optimize-plan1-birth.md     → Plan 1: monorepo TS (completado 2026-03-24)
│   │   ├── ultra-optimize-plan2-testing.md   → Plan 2: testing 7 fases (completado 2026-03-25)
│   │   ├── ultra-optimize-plan3-bugfix.md    → Plan 3: bugfix + code quality (completado 2026-03-25)
│   │   └── ultra-optimize-plan8-optimization.md → Plan 8: optimización (en curso)
│   ├── architecture.md       → este archivo
│   ├── workflows.md          → flujos de trabajo del proyecto
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
2. Server verifica contra packages/db (bcrypt-ts)
3. Server crea JWT firmado con JWT_SECRET (jose)
4. Server setea cookie HttpOnly auth_token
5. Todas las requests → middleware.ts verifica el JWT
6. JWT claims (sub, role) se pasan como headers x-user-*
7. Server Components leen getCurrentUser() vía React.cache()
```

### Service-to-service (CLI → API)

```
CLI → Authorization: Bearer <SYSTEM_API_KEY>
   → middleware inyecta headers x-user-id=0, x-user-role=admin
   → extractClaims() lee x-user-* headers (no intenta verificar JWT)
```

## Flujo de RAG query

```
1. Usuario escribe en ChatInterface (Client Component)
2. useRagStream hook: fetch POST /api/rag/generate con historial + collection_name
3. Server verifica permisos de colección (packages/db canAccessCollection)
4. Server hace proxy SSE → RAG Server :8081
5. Server verifica status HTTP ANTES de streamear (fix del bug de gateway.py)
6. Stream se reenvía al browser con ReadableStream
7. ChatInterface acumula los chunks vía parseSseLine() de lib/rag/stream.ts
8. Al terminar, Server Action addMessage() persiste la conversación
```

## Base de datos única (SQLite)

Todos los datos de la aplicación en un solo archivo SQLite (`data/app.db`).

| Tabla | Contenido |
|---|---|
| `users` | Usuarios con roles y preferencias |
| `areas` | Áreas de la organización |
| `user_areas` | Many-to-many usuarios ↔ áreas |
| `area_collections` | Permisos área ↔ colección (nivel read/write) |
| `chat_sessions` | Sesiones de chat |
| `chat_messages` | Mensajes de cada sesión |
| `message_feedback` | Feedback thumbs up/down |
| `ingestion_jobs` | Jobs de ingesta con estado |
| `ingestion_alerts` | Alertas de jobs fallidos |
| `ingestion_queue` | Cola de ingesta con locking optimista |
| `audit_log` | Acciones de usuarios (legacy, mantenida por compatibilidad) |
| `events` | **Black Box** — todos los eventos del sistema |

**Nota:** La DB usa `@libsql/client` (JS puro, sin compilación nativa), no `better-sqlite3`.
Conexión singleton con WAL mode habilitado. Timestamps en epoch ms via Temporal API.

El schema en `packages/db/src/schema.ts` es la **única fuente de verdad** — `drizzle-kit push` sincroniza la DB sin SQL manual (ver ADR-003 en Drizzle Kit, Fase 3 del Plan 8).

### Cola de ingesta con locking optimista (SQLite)

La tabla `ingestion_queue` implementa locking optimista para coordinar workers:

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

> **Nota (Fase 8):** La tabla `ingestion_queue` y el locking SQLite serán reemplazados por **BullMQ sobre Redis** en la Fase 8 del Plan 8. Ver ADR-010 y la sección "Redis (Fase 8)" más abajo.

---

## Utilidades de stream SSE

Toda la lógica de lectura de streams SSE vive en un único lugar:

**`apps/web/src/lib/rag/stream.ts`** — tres funciones públicas:

| Función | Descripción |
|---|---|
| `parseSseLine(line)` | Extrae el token de contenido de una línea `data: {...}`. Null para `[DONE]`, líneas sin prefijo, JSON malformado o delta sin content |
| `readSseTokens(body)` | AsyncGenerator que yields tokens individuales. Incluye buffering de líneas parciales para manejar chunks que cortan en mitad de una línea SSE |
| `collectSseText(response, options?)` | Acumula todo el texto del stream. Soporta `maxChars` y `detectRepetition`. Maneja SSE y respuestas JSON estándar |

**Consumers:** `useRagStream`, `useCrossdocStream`, `useCrossdocDecompose`, `api/slack/route.ts`, `api/teams/route.ts`.

Antes de la extracción (Plan 8 — Fase 1), esta lógica estaba copiada en 5 lugares con variantes sutiles. Ver ADR-008 para la justificación completa.

---

## Black Box System

Todos los eventos del sistema (frontend + backend) se almacenan en la tabla `events`:

```typescript
{
  id: string        // UUID
  ts: number        // epoch ms (Temporal.Now.instant().epochMilliseconds)
  source: 'frontend' | 'backend'
  level: LogLevel   // TRACE..FATAL
  type: EventType   // 'auth.login' | 'rag.query' | 'user.created' | 'system.request' | ...
  userId: number | null
  sessionId: string | null
  payload: JSON     // contexto del evento
  sequence: number  // monotónico, para replay ordenado
}
```

Para reconstruir el estado después de un crash:
```bash
rag audit replay --from 2026-03-25
```

Internamente usa `packages/logger/blackbox.ts:reconstructFromEvents()`.

---

## Caching

Next.js 16 tiene 4 capas de caché, todas en uso:

| Capa | Qué cachea | TTL |
|---|---|---|
| `React.cache()` | queries DB dentro de una request | request |
| `unstable_cache` | colecciones del RAG | 60s |
| Full Route Cache | páginas estáticas (admin config) | build |
| Router Cache | navegación del cliente | session |

Invalidación: `revalidateTag('collections')` al crear/eliminar colecciones.

---

## Patrones React/Next.js (Plan 8 — Fase 2)

### Server Components por defecto

Todo componente que carga datos usa el patrón Server Component → props → Client Component:

```
page.tsx (Server Component)
  → query DB / RAG
  → <ClientComponent data={data} />
```

Cero `useEffect + fetch` en Client Components para carga de datos.

### Server Actions para mutaciones

Todas las mutaciones van por Server Actions en `apps/web/src/app/actions/`. Tipadas con `next-safe-action` y `zodResolver`. Ver ADR-009 para la justificación completa.

### Lazy loading de dependencias pesadas

`d3` (~450 KB) y `react-pdf` (~600 KB) se cargan con `next/dynamic` solo en las rutas que los necesitan:

```typescript
const DocumentGraph = dynamic(() => import("@/components/collections/DocumentGraph"), { ssr: false })
```

---

## Redis (Fase 8 — próxima integración)

Redis se integrará como **dependencia requerida del sistema** en la Fase 8 del Plan 8, al mismo nivel que Milvus — sin fallback SQLite.

**Por qué requerido (no opcional):** los workarounds actuales en SQLite/memoria son correctos pero tienen limitaciones en escenarios de alta concurrencia y multi-proceso. Redis provee primitivas que SQLite no puede replicar eficientemente:

| Workaround actual | Reemplazado por |
|---|---|
| Tabla `ingestion_queue` + locking SQLite | BullMQ sobre Redis |
| `sequence` monotónico en memoria (`_seq`) | Redis `INCR events:seq` |
| Sin JWT blacklist | Redis `SET revoked:{jti} 1 EX {ttl}` |
| Cache de tamaños de logs en memoria | Redis `HSET log:sizes` |
| `localStorage["seen_notification_ids"]` | Redis `ZADD notifications:seen:{userId}` |
| Sin master election para external-sync | Redis `SET NX EX` |

Ver ADR-010 (`docs/decisions/010-redis-required.md` — se crea en Fase 8) para la decisión completa.

**Variable de entorno:** `REDIS_URL=redis://localhost:6379` (requerido en Fase 8).

---

## Design System (Plan 7)

El proyecto tiene un design system completo "Warm Intelligence" aplicado a las 24 páginas:

- **Tokens CSS** en `apps/web/src/app/globals.css` — crema cálido + navy profundo + dark mode cálido
- **Tailwind v4** con `@tailwindcss/postcss` — requiere `postcss.config.js` explícito
- **Storybook 8** en `apps/web/stories/` — catálogo de componentes con addon-a11y
- **Densidad adaptiva** — `data-density="compact|spacious"` en los layouts

Ver `docs/design-system.md` para la documentación completa.

---

## Suite de testing (Planes 5, 6 + Plan 8)

| Capa | Herramienta | Archivos |
|---|---|---|
| Lógica pura | bun:test | `packages/*/src/__tests__/`, `apps/web/src/lib/__tests__/` |
| Componentes React | @testing-library/react + happy-dom | `apps/web/src/components/**/__tests__/` |
| Visual regression | Playwright screenshots | `apps/web/tests/visual/` |
| A11y WCAG AA | axe-playwright | `apps/web/tests/a11y/` |
| E2E flujos | Maestro YAML | `apps/web/tests/e2e/` |
| Performance | react-scan | activo en dev, overlay visual |

Ver `docs/testing.md` para la guía completa.

---

## Workers de background

```
apps/web/src/workers/
  ingestion.ts       → procesa cola de ingesta (polling SQLite, locking optimista)
  external-sync.ts   → sincroniza fuentes externas según schedule
```

Los workers se invocan como endpoints serverless o via cron jobs. Usan el mismo `packages/db` que el servidor Next.js.

---

## Architecture Decision Records

Decisiones de arquitectura documentadas en `docs/decisions/`:

| ADR | Impacto |
|---|---|
| 001 — libsql | Por qué usamos libsql en lugar de better-sqlite3 |
| 002 — CJS | Por qué packages/* no tienen "type": "module" |
| 003 — proceso único | Por qué Next.js unifica gateway + frontend |
| 004 — Temporal API | Por qué usamos Date.now() y no utilidades de SQLite |
| 005 — import estático logger | Por qué logger importa db estáticamente |
| 006 — testing strategy | Metas de cobertura y qué testear por tipo de código |
| 007 — funciones reales en tests | Por qué los tests usan las funciones reales de queries |
| 008 — SSE reader extraction | Por qué `lib/rag/stream.ts` centraliza toda la lógica de streams |
| 009 — Server Components first | Por qué Server Components + Server Actions reemplazan useEffect+fetch |
| 010 — Redis requerido *(Fase 8)* | Por qué Redis es dependencia del sistema sin fallback SQLite |
