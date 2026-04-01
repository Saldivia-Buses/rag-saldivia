# 16 — Rendimiento, Caching y Lazy Loading

> Analisis medido sobre el codigo real — no teoria.
> Herramientas: grep exhaustivo + repomix + Explore agent
> Fecha: 2026-03-31

---

## Score por categoria (actualizado post Plan 26)

| Categoria | Implementado | Score | Gap principal |
|-----------|-------------|-------|--------------|
| Caching | 90% | ⭐⭐⭐⭐⭐ | Cache-Control en collections, Redis cache admin stats (Plan 26) |
| Lazy loading | 90% | ⭐⭐⭐⭐⭐ | Loading skeletons en TODAS las rutas (Plan 26). recharts no aplica (CSS puro). |
| Bundle | 85% | ⭐⭐⭐⭐ | Standalone output + compression + optimizePackageImports (Plan 26) |
| Data fetching | 80% | ⭐⭐⭐⭐ | Server Components + Promise.all sistematico |
| Memoizacion | 85% | ⭐⭐⭐⭐⭐ | React.memo en MessageList, ChannelList + existentes (Plan 26) |
| Streaming | 85% | ⭐⭐⭐⭐⭐ | HTTP status check pre-stream, AI SDK adapter limpio |
| Base de datos | 85% | ⭐⭐⭐⭐⭐ | SQLite PRAGMAs WAL + busy_timeout (Plan 26). Redis cache admin stats. |

---

## 1. CACHING — Lo que hay

### Redis cache de colecciones (produccion-ready)

**Archivo:** `lib/rag/collections-cache.ts`

```
getCachedRagCollections() → Redis GET con TTL 60s
                         → fallback a fetch directo si Redis caido
invalidateCollectionsCache() → Redis DEL en POST/DELETE coleccion
```

- Migrado de `unstable_cache` (per-process) a Redis (compartido entre instancias)
- Invalidacion correcta: `actionCreateCollection` y `actionDeleteCollection` llaman `invalidateCollectionsCache()` + `revalidatePath("/collections")`
- Usado en: `/admin/permissions/page.tsx`, `/admin/collections/page.tsx`, `/collections/page.tsx`

### revalidatePath sistematico

**46 llamadas** a `revalidatePath()` en 7 archivos de actions:

| Action file | Paths revalidados |
|------------|------------------|
| `chat.ts` | `/chat`, `/chat/{id}` |
| `collections.ts` | `/collections` |
| `admin.ts` | `/admin/users` |
| `areas.ts` | `/admin/areas`, `/admin/permissions` |
| `roles.ts` | `/admin` |
| `settings.ts` | `/settings`, `/`, layout |
| `messaging.ts` | `/messaging`, `/messaging/{channelId}` |

### HTTP Cache-Control (1 sola ruta)

**Unico endpoint con cache HTTP:** `/api/rag/document/[name]` → `Cache-Control: private, max-age=300`

Todo el resto de API routes NO tiene headers de cache.

### Lo que se implemento (Plan 26)

- `Cache-Control: private, max-age=60` en `/api/rag/collections` GET
- Cache Redis de analytics queries para admin dashboard
- React.memo en MessageList y ChannelList

### Lo que FALTA en caching

| Oportunidad | Esfuerzo | Impacto |
|------------|----------|---------|
| Cache Redis de `getUserCollections()` (TTL 300s) | 2 horas | Alto — llamado en CADA page load protegido |
| `Cache-Control` en `/api/health` | 5 min | Bajo — reduce polling |
| `revalidateTag()` en vez de `revalidatePath()` | 4 horas | Medio — invalidacion mas granular (Next.js 16 compatible) |

---

## 2. LAZY LOADING — Lo que hay

### next/dynamic encontrado

- `ReactScanProvider.tsx` → `dynamic(() => import("./ReactScan"), { ssr: false })` (solo dev)
- (DocumentGraph fue archivado en Plan 13 — no hay mas usos activos de dynamic aparte de ReactScan)
- Mocked en tests: `component-test-setup.ts` y `test-setup.ts` mockean `next/dynamic`

### Suspense boundaries (1 sola)

- `/app/(auth)/login/page.tsx` → `<Suspense>` wrapping `useSearchParams()`
- **Ningun otro Suspense** en el proyecto

### loading.tsx existentes (3 rutas)

| Ruta | Tiene loading.tsx |
|------|------------------|
| `/chat` | Si — Skeleton UI |
| `/collections` | Si — SkeletonTable |
| `/messaging` | Si — Skeleton |

### Skeleton components (maduros)

`Skeleton`, `SkeletonText`, `SkeletonAvatar`, `SkeletonCard`, `SkeletonTable` — todos con tests y stories en Storybook.

### Lo que se implemento (Plan 26)

- `loading.tsx` para TODAS las rutas que faltaban: `/chat/[id]`, `/admin`, `/admin/users`, `/collections/[name]`, `/settings`
- AdminDashboard no usa recharts — usa barras CSS puras (no aplica lazy load)

### Lo que FALTA en lazy loading

| Oportunidad | Esfuerzo | Impacto |
|------------|----------|---------|
| Suspense boundaries en messaging (message lists) | 2 horas | Medio — streaming de listas |
| Message list virtualization (react-window/tanstack-virtual) | 4 horas | Alto para canales con muchos mensajes |

---

## 3. BUNDLE OPTIMIZATION — Lo que hay

### React Compiler activo

`next.config.ts`: `reactCompiler: true` — auto-memoizacion habilitada. Los `useMemo`/`useCallback` manuales coexisten.

### Bundle analyzer integrado

`ANALYZE=true` activa `withBundleAnalyzer`. Baseline capturado: `/chat` 120 kB, `/chat/[id]` 171 kB.

### Lucide React 1.x

Upgrade de 0.475.0 → 1.7.0 — mejor tree-shaking en v1.

### Barrel exports (riesgo controlado)

`packages/db/src/index.ts` re-exporta 21 modulos. `packages/shared/src/schemas/index.ts` re-exporta 3 schemas. Estandar para monorepo — tree-shaking de webpack deberia eliminar lo no usado, pero no verificado en build.

### Lo que FALTA

| Oportunidad | Esfuerzo | Impacto |
|------------|----------|---------|
| Verificar tree-shaking de barrel exports con `ANALYZE=true` | 30 min | Diagnostico |
| `optimizePackageImports` en next.config para lucide-react | 5 min | Bajo-medio |

---

## 4. MEMOIZACION — Lo que hay (extensiva)

### React.memo (9 componentes)

| Componente | Archivo | Razon |
|-----------|---------|-------|
| `HighlightedCode` | ArtifactPanel.tsx:49 | Evita re-highlight en cada render |
| `CodeView` | ArtifactPanel.tsx:104 | Sub-componente de artifact |
| `MermaidPreview` | ArtifactPanel.tsx:140 | Rendering de diagramas |
| `HtmlPreview` | ArtifactPanel.tsx:215 | Iframe preview |
| `SvgPreview` | ArtifactPanel.tsx:228 | SVG sanitizado |
| `ArtifactCard` | MarkdownMessage.tsx:77 | Card de artifact en mensajes |
| `MarkdownOnly` | MarkdownMessage.tsx:172 | Markdown sin artifacts |
| `MarkdownWithCodeBlockArtifacts` | MarkdownMessage.tsx:208 | Markdown con code blocks |
| `MarkdownMessage` | MarkdownMessage.tsx:267 | Componente exportado |

### useMemo (20+ usos)

Patrones encontrados:
- `ChatInterface.tsx`: `allArtifacts`, `streamingArtifact`, `panelArtifacts` (lineas 204, 220, 228)
- `SessionList.tsx`: `filtered` — filtrado client-side sin server round-trip (linea 36)
- `MarkdownMessage.tsx`: `extractArtifacts(content)`, `components` de markdown (lineas 173, 215, 275)
- `ChannelList.tsx`: `grouped` — agrupacion de canales (linea 60)
- `MessageItem.tsx`: `author` lookup (linea 79)

### useCallback (30+ usos)

Todos los event handlers en componentes interactivos:
- `ChatInterface`: scrollToBottom, handleCopy, handleFeedback, handleTitleSave, handleRetry (lineas 280-331)
- `SessionList`: confirmDelete (linea 58)
- `AdminUsers`: refreshUsers, confirmDeleteUser (lineas 98, 239)
- `AdminRoles`: confirmDelete (linea 233)
- `AdminAreas`: confirmDelete (linea 119)
- `AdminCollections`: confirmDelete (linea 67)
- `CollectionsList`: confirmDelete (linea 70)
- `ChannelView`: handleOpenThread, handleReply, handleOptimisticMessage (lineas 45-53)
- Hooks: useTyping (startTyping, stopTyping), useMessaging (subscribe, unsubscribe, clearMessages), useCopyToClipboard (copy), useLocalStorage (set), usePresence (setStatus)

### Lo que FALTA

| Oportunidad | Esfuerzo | Impacto |
|------------|----------|---------|
| React.memo en `MessageList` (re-render en cada mensaje nuevo) | 30 min | Alto en canales activos |
| React.memo en `ChannelList` (re-render en cada cambio de canal) | 20 min | Medio |
| Verificar renders con ReactScan en dev (tool existe, no hay metricas) | 1 hora | Diagnostico |

---

## 5. DATA FETCHING — Lo que hay (maduro)

### Server Components como default

Todas las pages son Server Components que fetchean datos server-side:

| Pagina | Fetching pattern |
|--------|-----------------|
| `/chat` | `Promise.all([listSessions(), getUserById()])` |
| `/chat/[id]` | `Promise.all([getSession(), listSessions(), getTemplates(), getUserById(), getUserCollections()])` |
| `/admin/users` | `Promise.all([listUsers(), listAreas()])` |
| `/admin/permissions` | `Promise.all([listAreas(), getCachedRagCollections()])` |
| `/collections` | `getCachedRagCollections()` |

### Promise.all sistematico

31 usos de `Promise.all` / `Promise.allSettled` en el proyecto:
- Pages: queries paralelas en lugar de secuenciales
- Webhooks: `Promise.allSettled(hooks.map(...))` — previene cascada de fallos
- Analytics: 4 queries SQL en paralelo

### Sin N+1 queries

- `getRateLimit()` usa `inArray` para batch lookup
- `canAccessCollection()` usa single `getUserCollections()`
- Admin pages: fetch areas + collections una vez, no per-item

### Lo que FALTA

| Oportunidad | Esfuerzo | Impacto |
|------------|----------|---------|
| Verificar que no quedan `useEffect + fetch` en Client Components | 2 horas | Medio |

---

## 6. STREAMING — Lo que hay (solido)

### Patron critico verificado

`ragGenerateStream()` en `lib/rag/client.ts`: **verifica HTTP status ANTES de leer el stream.** Si el response no es 200, retorna error sin consumir el stream. Esto previene el bug del gateway Python original.

### Pipeline completo

```
NVIDIA RAG :8081 (SSE formato OpenAI)
  → ReadableStream + TextDecoder (chunked reading)
    → parseSseLine() extrae tokens
      → createUIMessageStream() escribe AI SDK protocol
        → text-start → text-delta → data-sources → text-end
          → useChat() consume en browser
```

### SSE headers correctos

Endpoints de streaming usan: `Cache-Control: "no-cache"`, `Connection: "keep-alive"`, `X-Accel-Buffering: "no"`.

### WebSocket para real-time

Messaging usa WS sidecar: presencia (heartbeat 15s), typing indicators, notificaciones.

### Lo que FALTA

| Oportunidad | Esfuerzo | Impacto |
|------------|----------|---------|
| Compresion gzip en SSE responses | 1 hora | Bajo — tokens son chicos |
| Streaming de message lists (cursor + SSE) | 8 horas | Medio para canales grandes |
| Backpressure handling en WebSocket | 4 horas | Bajo — edge case |

---

## 7. BASE DE DATOS — Lo que hay

### Paginacion cursor-based (messaging)

`WHERE createdAt < ? ORDER BY createdAt DESC LIMIT {limit}` con cursor compuesto `(createdAt, id)` para evitar duplicados.

### Indices

- `idx_events_query` compuesto en `(type, userId, ts)` para analytics O(log n)
- `idx_users_api_key` en users.apiKeyHash
- `idx_audit_user`, `idx_audit_timestamp` en audit_log
- `idx_chat_sessions_user_updated` compuesto para listado de sesiones
- `idx_msg_channel_created` compuesto para mensajes por canal
- FTS5 disponible para busqueda full-text (`fts5.ts`)

### Cleanup automatico

`deleteOldEvents()` ejecutado en worker diario — evita crecimiento infinito de la tabla events.

### Lo que FALTA

| Oportunidad | Esfuerzo | Impacto |
|------------|----------|---------|
| Cache Redis de analytics (queries costosas, TTL 60s) | 2 horas | Alto — admin dashboard es lento |
| Cache Redis de user collections (TTL 300s, invalidar en permisos) | 2 horas | Alto — se consulta en cada page |
| Verificar que FTS5 cubre sessions, messages, annotations | 1 hora | Medio — search podria ser lento sin FTS |
| Audit de indices con EXPLAIN QUERY PLAN | 2 horas | Diagnostico |

---

## Plan de accion priorizado

### Quick wins (< 1 hora cada uno)

1. `loading.tsx` para `/chat/[id]`, `/collections/[name]`, `/admin/analytics` — 3 archivos, 15 min cada uno
2. `Cache-Control: private, max-age=60` en `/api/rag/collections` GET — 10 min
3. `optimizePackageImports: ["lucide-react"]` en next.config — 5 min
4. `React.memo` en MessageList y ChannelList — 30 min

### Medio esfuerzo (2-4 horas)

5. Cache Redis de `getUserCollections()` con TTL 300s e invalidacion en area/permission changes — 2 horas
6. Cache Redis de analytics queries con TTL 60s — 2 horas
7. Lazy load de recharts en AdminDashboard con `next/dynamic` — 30 min
8. Verificar tree-shaking con `ANALYZE=true` y fix barrel exports si necesario — 2 horas

### Mayor esfuerzo (4+ horas)

9. Message virtualization con tanstack-virtual para canales grandes — 4 horas
10. Audit de indices con EXPLAIN QUERY PLAN en queries lentas — 2 horas
11. Streaming de message lists con cursor pagination — 8 horas
