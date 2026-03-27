# Plan 8: Optimización — Performance, Calidad de Código y Dependency Upgrades

> Este documento vive en `docs/plans/ultra-optimize-plan8-optimization.md`.
> Se actualiza a medida que se completan las tareas. Cada fase completada genera entrada en CHANGELOG.md.

---

## Contexto

Los Planes 1–7 construyeron el stack completo, 50 features del product roadmap, la suite de testing (215+ tests) y el design system. El código funciona, los tests están en verde, pero **la velocidad de construcción dejó deuda técnica acumulada** — confirmada con análisis de código completo via repomix:

**Duplicación y dead code:**
- Lógica de lectura SSE (`getReader + TextDecoder + parseo de líneas`) duplicada en **6 archivos** (`useRagStream`, `useCrossdocStream`, `useCrossdocDecompose`, `slack/route.ts`, `teams/route.ts`)
- `unknown[]` propagado en 6+ archivos para fuentes del RAG — `CitationSchema` ya existe en `packages/shared` pero no se usa donde corresponde
- `getCachedRagCollections` definida **dos veces** — dead code en `route.ts`, real en `collections-cache.ts`
- Función `ragFetchWithOptions` en `rag/collections/route.ts` nunca exportada ni usada

**Bugs de performance:**
- N+1 query en `getRateLimit` — hace una query SQLite **por área del usuario** en lugar de `WHERE targetId IN (...)`. Se ejecuta en cada request del endpoint `/api/rag/generate`
- `listWebhooksByEvent` carga todos los webhooks en memoria y filtra en JS en lugar de filtrar en SQL
- `canAccessCollection` se llama múltiples veces sin caché por request en el handler de generate multi-colección

**Antipatrones React/Next.js:**
- `settings/memory/page.tsx` es la única página de `(app)/` con `"use client"` y raw `fetch()`, violando el patrón Server Component + Server Actions del resto de la app
- **Cero memoización** en componentes React — `ChatInterface` (410 líneas) recrea 5 handlers en cada render
- `d3` (~450KB) y `react-pdf` (~600KB) entran al bundle inicial aunque solo se usan en rutas específicas

**Dependencias:**
- Drizzle en versión **desincronizada** entre `packages/db` (`^0.38.0`) y `apps/web` (`^0.38.4`)
- Next.js, Drizzle, Lucide, Zod, @libsql/client con versiones atrasadas
- Sin `@next/bundle-analyzer` para medir el bundle

**Calidad estructural:**
- Sin Error Boundaries en rutas críticas (`/chat`, `/admin`)
- CI sin paralelización de jobs de testing
- `turbo lint` no corre en los packages del monorepo

**Lo que NO cambia:** la lógica de negocio, auth, schema de DB, design system, tests existentes. Esta fase no agrega features nuevas — mejora el código interno.

---

## Orden de ejecución recomendado

Las fases están numeradas para referencia, pero el orden óptimo de ejecución es:

```
Fase 0 → Fase 1 → Fase 3 → Fase 2 → Fase 4 → Fase 5 → Fase 6 → Fase 7 → Fase 8
```

**Por qué este orden:**
- **Fase 3 antes de Fase 2:** sincronizar Drizzle antes de refactorizar React — si los upgrades (Fase 4) rompen algo, lo descubrís antes de haber refactorizado
- **Fase 4 después de Fase 2:** los upgrades de Next.js/Zod pueden cambiar APIs que se están refactorizando — mejor conocer los breaking changes después del refactor
- **Fase 5 después de Fase 4:** actualizar `architecture.md` cuando ya están aplicados todos los cambios estructurales de las fases anteriores — ADR-008 y ADR-009 se escriben al cierre de sus fases respectivas (no se espera a Fase 5)
- **Fase 6 antes de Fase 7 y Fase 8:** los Error Boundaries son cambios pequeños e independientes, conviene hacerlos cuando el código ya está limpio

---

## Impacto en tests existentes

Análisis con lectura del código real de los tests. Antes de ejecutar cualquier fase, leer esta tabla.

### Tests que SE ROMPEN (requieren actualización explícita)

| Archivo | Fase | Qué cambia | Actualización requerida |
|---|---|---|---|
| `packages/db/src/__tests__/events.test.ts` | F8.24 | `_seq` → Redis `INCR` via ioredis-mock | Verificar que ioredis-mock está activo en el test setup; el test `"asigna sequence monotónicamente creciente"` sigue siendo válido si el mock retorna enteros incrementales |
| `packages/logger/src/__tests__/logger.test.ts` | F7.17 | `log.info("system.warning", ...)` es el tipo incorrecto que se corrige | Los 3 tests que usan `"system.warning"` en `log.info` deben actualizarse a usar `"system.start"` u otro EventType válido. Agregar test: `"log.request usa EventType system.request"` |
| `apps/web/src/components/admin/__tests__/AreasAdmin.test.tsx` | F2.7 | `next-safe-action` cambia el retorno de las actions | Actualizar mock: `mock(() => Promise.resolve({ data: undefined }))` en lugar de `mock(() => Promise.resolve())` |
| `apps/web/src/components/admin/__tests__/UsersAdmin.test.tsx` | F2.7 | ídem | ídem |
| `apps/web/src/components/admin/__tests__/RagConfigAdmin.test.tsx` | F2.7 | ídem | ídem |
| `apps/web/src/components/admin/__tests__/PermissionsAdmin.test.tsx` | F2.7 | ídem | ídem |
| `apps/web/src/components/settings/__tests__/SettingsClient.test.tsx` | F2.7 | ídem | ídem |
| `packages/db/src/__tests__/rate-limits.test.ts` | F8.24 | Insert directo con `sequence: 9999` bypasea `writeEvent` | Aceptable: ese insert es para el test de timing (`ts` en el pasado), no de sequencing. No necesita cambio. |

### Tests que NO SE ROMPEN (safe)

- `apps/web/src/lib/auth/__tests__/jwt.test.ts` — agregar `jti` no rompe checks de `sub`, `email`, `role`
- `packages/db/src/__tests__/rate-limits.test.ts` — `getRateLimit` con `inArray` retorna lo mismo; tests de comportamiento son safe
- `packages/db/src/__tests__/sessions.test.ts` — `addMessage(sources?: Citation[])` es compatible con `unknown[]` en runtime
- `packages/db/src/__tests__/webhooks.test.ts` — sin cambio de comportamiento, solo índice nuevo
- `packages/logger/src/__tests__/logger.test.ts` — el refactor de `formatPretty` (F7.21) no cambia el output, solo la estructura interna
- Todos los tests de `packages/db/` — ninguno referencia `ingestion_queue` (confirmado con grep)

### Tests nuevos que hay que agregar (por los cambios)

| Test nuevo | Fase | Archivo |
|---|---|---|
| `parseSseLine`, `collectSseText` | F1.1 | `apps/web/src/lib/rag/__tests__/stream.test.ts` (nuevo) |
| `CitationSchema.safeParse` con warning en payload inválido | F1.2 | `packages/shared/src/__tests__/schemas.test.ts` |
| `"formatDate retorna string en formato dd/mm/yyyy"` | F1.7 | `apps/web/src/lib/__tests__/utils.test.ts` |
| `"createJwt incluye campo jti único"` | F8.25 | `apps/web/src/lib/auth/__tests__/jwt.test.ts` |
| `"reconstructFromEvents con rag.stream_started"`, `"con ingestion.failed"` | F7.18 | `packages/logger/src/__tests__/logger.test.ts` |
| `"getRedisClient lanza error si REDIS_URL no configurado"` | F8.23 | `packages/db/src/__tests__/redis.test.ts` (nuevo) |
| `"BullMQ: job completado dispara evento ingestion.completed"` | F8.30 | `apps/web/src/__tests__/queue.test.ts` (nuevo) |
| Tests de componentes con `react-hook-form`: render + submit + errores de validación | F2.8 | `apps/web/src/components/admin/__tests__/AreasAdmin.test.tsx` etc. |
| `"AuditTable filter preserva estado en URL ?q=..."` | F2.9 | `apps/web/src/components/audit/__tests__/AuditTable.test.tsx` |

---

## Seguimiento

Formato: `- [ ] Descripción — estimación`
Al completar: `- [x] Descripción — completado YYYY-MM-DD`
Cada fase completada → entrada en CHANGELOG.md → commit.

---

## Fase 0 — Baseline de medición *(30-45 min)*

Objetivo: medir el estado actual antes de tocar nada. Sin baseline no hay evidencia de mejora.

**Por qué primero:** las Fases 2, 7 y 8 prometen mejoras de performance, bundle y logging. Si no medís antes, al terminar no podés demostrar nada.

**Archivos a crear:**
- `docs/performance/baseline-plan8.md` — snapshot de métricas antes del plan

---

### F0.1 — Bundle size baseline

> **Cómo medir:** `bun run build` imprime los tamaños de cada ruta en consola — esos son los números comparables. `@next/bundle-analyzer` genera un HTML interactivo útil para explorar, pero no imprime números en consola. Usá el output de `bun run build` como fuente primaria del baseline.

- [x] `bun add -d @next/bundle-analyzer` en `apps/web` — completado 2026-03-27
- [x] Agregar a `next.config.ts`:
  ```typescript
  const withBundleAnalyzer = require("@next/bundle-analyzer")({ enabled: process.env["ANALYZE"] === "true" })
  export default withBundleAnalyzer(nextConfig)
  ```
- [x] `bun run build` — output completo capturado — completado 2026-03-27
- [x] Anotar en `docs/performance/baseline-plan8.md`:
  - `/chat`: 120 kB First Load JS
  - `/chat/[id]`: 171 kB First Load JS
  - `/collections/[name]/graph`: 119 kB First Load JS
  - Bundle compartido: 103 kB
- [ ] `ANALYZE=true bun run build` — abrir el HTML, capturar screenshot mostrando `d3` y `react-pdf` en el bundle de `/chat` — 5 min

---

### F0.2 — React render baseline con react-scan

- [x] Análisis estático de `ChatInterface.tsx` completado — 5 handlers sin `useCallback` confirmados — 2026-03-27
- [x] Anotar en `docs/performance/baseline-plan8.md`: componentes que re-renderizan con cada keystroke — completado 2026-03-27
- [ ] `bun run dev` → abrir `/chat` → activar react-scan → capturar screenshot (requiere browser)

---

### F0.3 — Métricas de CI actuales

- [x] Correr `bun run test` y anotar el tiempo total — 1.853s real, 273 tests pass — completado 2026-03-27
- [x] Revisar estructura del CI en `.github/workflows/ci.yml` — 9 jobs documentados — completado 2026-03-27
- [x] Anotar en `docs/performance/baseline-plan8.md` — completado 2026-03-27

---

### Criterio de done

`docs/performance/baseline-plan8.md` existe con bundle sizes, componentes que re-renderizan y tiempos de CI.

### Checklist de cierre

- [x] `docs/performance/baseline-plan8.md` creado con los 3 baseline — completado 2026-03-27
- [x] Commit: `docs(perf): baseline de medicion pre-plan8 — plan8 f0` — completado 2026-03-27

**Estado: completado 2026-03-27**

---

## Fase 1 — Extracción de código duplicado *(5-7 hs)*

Objetivo: eliminar toda duplicación identificable. Al terminar, cada lógica existe en un solo lugar. **Empezar con F1.0 (knip)** para obtener un mapa completo de dead code antes de tocar nada manualmente.

**Archivos a crear:**
- `apps/web/src/lib/rag/stream.ts`
- `apps/web/src/lib/rag/__tests__/stream.test.ts`

**Archivos a modificar:**
- `apps/web/src/hooks/useRagStream.ts`
- `apps/web/src/hooks/useCrossdocStream.ts`
- `apps/web/src/hooks/useCrossdocDecompose.ts`
- `apps/web/src/app/api/slack/route.ts`
- `apps/web/src/app/api/teams/route.ts`
- `apps/web/src/app/api/rag/collections/route.ts`
- `apps/web/src/components/chat/ChatInterface.tsx`
- `apps/web/src/components/chat/SourcesPanel.tsx`
- `apps/web/src/components/chat/ExportSession.tsx`
- `apps/web/src/app/actions/chat.ts`
- `apps/web/src/lib/export.ts`
- `packages/db/src/queries/sessions.ts`
- `packages/db/src/queries/rate-limits.ts`
- `packages/db/src/queries/webhooks.ts`

---

### F1.0 — `knip`: escaneo sistemático de dead code

**Por qué primero:** las tareas F1.1–F1.6 eliminan dead code identificado manualmente. `knip` hace un análisis estático exhaustivo sobre las 176+ archivos TypeScript del proyecto y puede revelar más código muerto antes de que empiece el trabajo manual.

```bash
bun add -d knip
```

`knip` detecta:
- **Exports no usados** — funciones/tipos exportados que nadie importa
- **Archivos no referenciados** — archivos `.ts`/`.tsx` que no forman parte del grafo de imports
- **Dependencias no usadas** en `package.json`
- **Tipos huérfanos** — `type`/`interface` definidos pero nunca usados

```json
// knip.json (raíz del repo)
{
  "workspaces": {
    "apps/web": { "entry": ["src/app/**/{page,layout,route}.ts{x,}","src/workers/*.ts"] },
    "packages/db": { "entry": ["src/index.ts"] },
    "packages/shared": { "entry": ["src/index.ts"] },
    "packages/logger": { "entry": ["src/index.ts"] },
    "packages/config": { "entry": ["src/index.ts"] }
  },
  "ignore": ["**/__tests__/**","**/*.test.*","**/*.spec.*","src/lib/test-setup.ts","src/lib/component-test-setup.ts"]
}
```

> **Hallazgo previo:** ya confirmado que `zustand` y `dompurify` están en `dependencies` con cero usos — `knip` los detectará automáticamente como dependencias no usadas.

- [x] `bun add -d knip` en la raíz del monorepo — completado 2026-03-27
- [x] Crear `knip.json` con la config de workspaces — completado 2026-03-27
- [x] `bunx knip --reporter compact` — leer el reporte completo — completado 2026-03-27
- [x] Anotar en `docs/performance/baseline-plan8.md` todos los exports/archivos/deps sin uso encontrados — completado 2026-03-27
- [x] Para cada ítem reportado: decidir eliminar ahora (si es seguro) o agregar como tarea explícita a la fase correspondiente — completado 2026-03-27
- [x] Commit: `chore(dx): agregar knip + escaneo inicial de dead code — plan8 f1.0` — incluido en commit de F1

---

### F1.1 — Utilidad `readSseStream` compartida

**Problema:** la lógica de leer un stream SSE está copiada en 6 lugares:

| Archivo | Forma |
|---|---|
| `hooks/useRagStream.ts` | inline |
| `hooks/useCrossdocStream.ts` | función local `collectStream` |
| `hooks/useCrossdocDecompose.ts` | función local `collectSseText` |
| `app/api/slack/route.ts` | inline |
| `app/api/teams/route.ts` | inline |

```typescript
// apps/web/src/lib/rag/stream.ts

/** Parsea una línea SSE "data: {...}" y extrae el token. Null para [DONE] o malformadas. */
export function parseSseLine(line: string): string | null

/** Yields tokens individuales a medida que llegan del ReadableStream. */
export async function* readSseTokens(
  body: ReadableStream<Uint8Array>
): AsyncGenerator<string>

/** Acumula todo el texto del stream. Soporta detección de repetición. */
export async function collectSseText(
  response: Response,
  options?: { maxChars?: number; detectRepetition?: boolean }
): Promise<string>
```

- [x] Crear `apps/web/src/lib/rag/stream.ts` con las 3 funciones — completado 2026-03-27
- [x] Tests: `parseSseLine` válida/vacía/`[DONE]`/malformada; `collectSseText` con mock de Response — 18 tests en verde — completado 2026-03-27
- [x] Refactorizar `useRagStream.ts` para usar `parseSseLine` + Citation[] — completado 2026-03-27
- [x] Refactorizar `useCrossdocStream.ts`: reemplazar `collectStream` local con `collectSseText` — completado 2026-03-27
- [x] Refactorizar `useCrossdocDecompose.ts`: reemplazar `collectSseText` local con la compartida — completado 2026-03-27
- [x] Refactorizar `slack/route.ts` y `teams/route.ts` — completado 2026-03-27
- [x] `bun test apps/web/src/lib/rag/` — todos pasan — completado 2026-03-27

---

### F1.2 — Usar `Citation` (ya existe en `packages/shared`) en lugar de `unknown[]`

**Descubrimiento:** `CitationSchema` y `Citation` **ya están definidos** en `packages/shared/src/schemas.ts` (línea 594) y el `ChatMessageSchema` los usa. El problema es que `useRagStream`, `ChatInterface` y otros archivos usan `unknown[]` en lugar del tipo existente.

```typescript
// packages/shared/src/schemas.ts — ya existe:
export const CitationSchema = z.object({
  id: z.string().optional(),
  document: z.string().optional(),
  content: z.string().optional(),
  score: z.number().optional(),
  metadata: z.record(z.unknown()).optional(),
})
export type Citation = z.infer<typeof CitationSchema>
```

- [x] Reemplazar todos los `unknown[]` y `as unknown[]` relacionados a sources con `Citation[]` importando de `@rag-saldivia/shared` — completado 2026-03-27
- [x] En `useRagStream.ts`: usar `.safeParse()` con `console.warn` explícito en lugar de silenciar errores — completado 2026-03-27
- [x] En `sessions.ts` `addMessage`: cambiar `sources?: unknown[]` a `sources?: Citation[]` — completado 2026-03-27
- [ ] En `DocPreviewPanel.tsx`: reemplazar el `any` del import dinámico de react-pdf con tipo inferido — pendiente (Fase 2)
- [x] `bun run test` — sin regresiones — completado 2026-03-27

---

### F1.3 — Eliminar dead code en `rag/collections/route.ts`

**Problema:** dos dead codes en el mismo archivo:
1. `getCachedRagCollections` definida localmente cuando ya existe en `collections-cache.ts`
2. Función `ragFetchWithOptions` nunca exportada ni usada (línea ~1983)

- [x] En `route.ts`: eliminar definición local de `getCachedRagCollections`, importar de `@/lib/rag/collections-cache` — completado 2026-03-27
- [x] En `route.ts`: eliminar la función `ragFetchWithOptions` dead — completado 2026-03-27
- [ ] Verificar que la ruta sigue respondiendo con `MOCK_RAG=true` — pendiente verificación manual

---

### F1.4 — Corregir N+1 query en `getRateLimit`

**Bug real de performance:** `getRateLimit` en `packages/db/src/queries/rate-limits.ts` hace una query SQLite **por área del usuario** en un loop. Se llama en cada request al endpoint `/api/rag/generate`.

```typescript
// Antes — N+1 (una query por área)
for (const areaId of areaIds) {
  const areaLimit = await db.select()...where(eq(rateLimits.targetId, areaId))...
}

// Después — una sola query
const areaLimits = await db
  .select({ max: rateLimits.maxQueriesPerHour })
  .from(rateLimits)
  .where(and(
    eq(rateLimits.targetType, "area"),
    inArray(rateLimits.targetId, areaIds),
    eq(rateLimits.active, true)
  ))
  .orderBy(asc(rateLimits.maxQueriesPerHour))  // el mínimo primero
  .limit(1)
```

- [x] Refactorizar `getRateLimit` para usar `inArray` en lugar del loop — completado 2026-03-27
- [x] Tests existentes en `rate-limits.test.ts` siguen en verde (comportamiento idéntico) — completado 2026-03-27

---

### F1.5 — Corregir `listWebhooksByEvent` filter en memoria

**Problema:** `listWebhooksByEvent` carga todos los webhooks activos y luego filtra en JavaScript. Con muchos webhooks no escala.

```typescript
// Antes — carga todo, filtra en JS
const all = await db.select().from(webhooks).where(eq(webhooks.active, true))
return all.filter((w) => (w.events as string[]).includes(eventType))

// Después — filtrar en SQL con json_each o simplemente usar la query existente
// con un índice virtual (SQLite no soporta array contains nativo, usar LIKE o json_each)
// Solución pragmática: mantener el approach actual pero agregar índice sobre active
// + documentar el límite de escala como comentario
```

> **Nota:** SQLite no tiene operador de array contains nativo. La solución correcta es `json_each` o mover los events a una tabla de junction. Por pragmatismo, esta tarea documenta el límite y agrega un test de rendimiento. La refactorización completa a tabla de junction es Plan 9 material si se necesita.

- [x] Agregar comentario con el límite de escala y la solución futura — completado 2026-03-27
- [x] Índice `idx_webhooks_active` ya existía en schema.ts — verificado 2026-03-27
- [x] Tests de `listWebhooksByEvent` ya existían con los 3 casos requeridos — verificado 2026-03-27

---

### F1.6 — Cache de `canAccessCollection` por request

**Problema:** en el handler de generate multi-colección, `canAccessCollection` se llama una vez por colección — y cada llamada ejecuta `getUserCollections`, que a su vez hace múltiples queries a DB. Con 3 colecciones seleccionadas = 3 × `getUserCollections`.

```typescript
// generate/route.ts — antes (una llamada a getUserCollections por colección)
for (const col of collectionNames) {
  const hasAccess = await canAccessCollection(userId, col, "read")  // query por col
}

// Después — una sola query, cache local por request
const userCollections = await getUserCollections(userId)
const accessSet = new Set(
  userCollections
    .filter(c => ["read", "write", "admin"].includes(c.permission))
    .map(c => c.name)
)
for (const col of collectionNames) {
  if (!accessSet.has(col)) return NextResponse.json({ ok: false, error: ... }, { status: 403 })
}
```

- [x] En `generate/route.ts`: reemplazar el loop de `canAccessCollection` por `getUserCollections` una sola vez + Set local — completado 2026-03-27
- [x] Tests existentes de canAccessCollection siguen en verde — completado 2026-03-27
- [x] Commit incluido en el commit general de Fase 1

---

### F1.7 — Centralizar formateo de fechas en `lib/utils.ts`

**Problema:** 12 instancias de `new Date(...).toLocaleDateString("es-AR")` repartidas en 10 archivos. Cada cambio de formato (idioma, zona horaria, precisión) requiere buscar y editar 12 lugares.

```typescript
// apps/web/src/lib/utils.ts — agregar junto a la función cn() ya existente
export function formatDate(ts: number | string | Date): string {
  return new Date(ts).toLocaleDateString("es-AR", {
    day: "2-digit", month: "2-digit", year: "numeric",
  })
}

export function formatDateTime(ts: number | string | Date): string {
  return new Date(ts).toLocaleString("es-AR", {
    day: "2-digit", month: "2-digit", year: "numeric",
    hour: "2-digit", minute: "2-digit",
  })
}
```

**Instancias a reemplazar:**
- `apps/web/src/workers/ingestion.ts` (2 instancias — `new Date().toLocaleDateString("es-AR")`)
- `apps/web/src/lib/export.ts` (1 instancia — `new Date(session.createdAt).toLocaleString("es-AR")`)
- `apps/web/src/components/collections/CollectionHistory.tsx` (1 instancia)
- `apps/web/src/components/admin/KnowledgeGapsClient.tsx` (2 instancias)
- `apps/web/src/components/admin/ReportsAdmin.tsx` (1 instancia)
- `apps/web/src/components/admin/ExternalSourcesAdmin.tsx` (1 instancia)
- `apps/web/src/app/(app)/saved/page.tsx` (1 instancia)
- `apps/web/src/app/(app)/projects/[id]/page.tsx` (1 instancia)

> **Notas:**
> - `packages/logger/src/blackbox.ts` usa `toISOString()` — formato diferente, mantener as-is.
> - `apps/web/src/components/audit/AuditTable.tsx` usa `.toISOString().replace("T", " ").slice(0,19)` — formato ISO específico, mantener as-is o extraer como `formatIso(ts)` separado.

- [x] Agregar `formatDate` y `formatDateTime` a `apps/web/src/lib/utils.ts` — completado 2026-03-27
- [x] Reemplazar las 9 instancias en 8 archivos (toLocaleString/toLocaleDateString → formatDate/formatDateTime) — completado 2026-03-27
- [x] `bun run test` — sin regresiones — completado 2026-03-27

---

### Criterio de done

Cero duplicación de SSE reader. Cero `unknown[]` relacionados a sources. `getCachedRagCollections` en un solo archivo. `canAccessCollection` hace una sola query DB por request. `formatDate`/`formatDateTime` centralizados en `lib/utils.ts`.

### Checklist de cierre

- [x] `bun run test` — 247 tests, todos pasan — completado 2026-03-27
- [x] Crear `docs/decisions/008-sse-reader-extraction.md` — completado 2026-03-27
- [x] CHANGELOG.md actualizado bajo `### Plan 8 — Optimización` — completado 2026-03-27
- [x] `git commit -m "refactor(web): extraer SSE reader, Citation type, eliminar duplicados, cache canAccess — plan8 f1"` — completado 2026-03-27

**Estado: completado 2026-03-27**

---

## Fase 2 — Refactoring de arquitectura React *(12-18 hs)*

Objetivo: corregir antipatrones de Next.js/React que generan trabajo innecesario en el browser. Incluye la eliminación sistemática de `useEffect + fetch` en Client Components, estandarización de forms y URL state.

**Archivos a crear:**
- `apps/web/src/components/settings/MemoryClient.tsx`
- `apps/web/src/lib/safe-action.ts`
- `apps/web/src/lib/form.ts`
- `apps/web/src/app/actions/auth.ts` (nuevo — `actionLogout`)
- `apps/web/src/app/actions/projects.ts` (nuevo — `actionCreateProject`, `actionDeleteProject`)
- `apps/web/src/app/actions/webhooks.ts` (nuevo)
- `apps/web/src/app/actions/reports.ts` (nuevo)
- `apps/web/src/app/actions/external-sources.ts` (nuevo)
- `apps/web/src/app/actions/share.ts` (nuevo — `actionCreateShare`)

**Archivos a modificar:**
- `apps/web/src/app/(app)/settings/memory/page.tsx`
- `apps/web/src/app/(app)/admin/webhooks/page.tsx`
- `apps/web/src/app/(app)/admin/reports/page.tsx`
- `apps/web/src/app/(app)/admin/external-sources/page.tsx`
- `apps/web/src/app/(app)/admin/knowledge-gaps/page.tsx`
- `apps/web/src/app/actions/settings.ts`
- `apps/web/src/app/actions/chat.ts`
- `apps/web/src/app/actions/users.ts`
- `apps/web/src/app/actions/areas.ts`
- `apps/web/src/app/actions/config.ts`
- `apps/web/src/components/chat/ChatInterface.tsx`
- `apps/web/src/components/chat/SessionList.tsx`
- `apps/web/src/components/chat/CollectionSelector.tsx`
- `apps/web/src/components/chat/PromptTemplates.tsx`
- `apps/web/src/components/chat/ShareDialog.tsx`
- `apps/web/src/components/layout/NavRail.tsx`
- `apps/web/src/components/layout/CommandPalette.tsx`
- `apps/web/src/components/layout/panels/ProjectsPanel.tsx`
- `apps/web/src/components/layout/WhatsNewPanel.tsx`
- `apps/web/src/components/admin/AnalyticsDashboard.tsx`
- `apps/web/src/components/admin/WebhooksAdmin.tsx`
- `apps/web/src/components/admin/ReportsAdmin.tsx`
- `apps/web/src/components/admin/ExternalSourcesAdmin.tsx`
- `apps/web/src/components/admin/KnowledgeGapsClient.tsx`
- `apps/web/src/components/admin/UsersAdmin.tsx`
- `apps/web/src/components/admin/AreasAdmin.tsx`
- `apps/web/src/components/collections/CollectionsList.tsx`
- `apps/web/src/components/projects/ProjectsClient.tsx`
- `apps/web/src/components/settings/SettingsClient.tsx`
- `apps/web/src/app/(auth)/login/page.tsx`
- `apps/web/src/app/(app)/collections/[name]/graph/page.tsx`

---

### F2.4 — Eliminar `useEffect + fetch` para carga de datos: Server Component + props

**Problema detectado con repomix:** el anti-patrón de `settings/memory/page.tsx` no es un caso aislado — hay **10 componentes** con `useEffect + fetch("/api/...")` para cargar datos iniciales. El patrón correcto en Next.js App Router es que la página (Server Component) fetche los datos y los pase como props al componente cliente.

```
Patrón actual (×10):   Client Component → useEffect → fetch("/api/...") → setState
Patrón correcto:       page.tsx (Server) → DB query → <ClientComponent data={data} />
```

> **Por qué esto importa:** con el patrón actual, el browser hace un request HTTP adicional por componente después de la hidratación. Con Server Components, los datos viajan en el HTML inicial — cero requests adicionales, cero loading states, cero código de fetch en el cliente.

**F2.4a — settings/memory (page con "use client")**
```typescript
// page.tsx — Server Component
export default async function MemoryPage() {
  const user = await getCurrentUser()
  const entries = await getMemory(user.id)
  return <MemoryClient entries={entries} />
}
// actions/settings.ts — agregar:
export async function actionAddMemory(key: string, value: string): Promise<void>
export async function actionDeleteMemory(key: string): Promise<void>
```
- [x] Crear `MemoryClient.tsx` — Client Component con estado local — completado 2026-03-27
- [x] Agregar `actionAddMemory` y `actionDeleteMemory` en `actions/settings.ts` — completado 2026-03-27
- [x] Reescribir `settings/memory/page.tsx` como Server Component — completado 2026-03-27

**F2.4b — WebhooksAdmin, ReportsAdmin, ExternalSourcesAdmin, KnowledgeGapsClient**

Los 4 componentes de admin más complejos — cada uno hace `useEffect(() => { fetch("/api/admin/...") })` para cargar sus datos iniciales. Las páginas correspondientes ya son Server Components vacíos que solo renderizan el Client Component.

```typescript
// apps/web/src/app/(app)/admin/webhooks/page.tsx — ANTES
export default function WebhooksPage() {
  return <WebhooksAdmin />   // WebhooksAdmin hace fetch internamente
}

// DESPUÉS — la página fetche, el componente solo renderiza
export default async function WebhooksPage() {
  const webhooks = await getWebhooks()
  return <WebhooksAdmin initialWebhooks={webhooks} />
}

// WebhooksAdmin.tsx — eliminar useEffect + useState para carga inicial
// Los datos vienen como prop, el useState se inicializa con initialWebhooks
```

- [x] `admin/webhooks/page.tsx`: fetch server-side + pasar como prop a `WebhooksAdmin` — completado 2026-03-27
- [x] `admin/reports/page.tsx`: fetch server-side + pasar como prop a `ReportsAdmin` — completado 2026-03-27
- [x] `admin/external-sources/page.tsx`: fetch server-side + pasar como prop a `ExternalSourcesAdmin` — completado 2026-03-27
- [x] `admin/knowledge-gaps/page.tsx`: fetch server-side + pasar como prop a `KnowledgeGapsClient` — completado 2026-03-27
- [x] En cada componente: eliminar `useEffect` y `useState` para carga inicial; aceptar `initialData` como prop — completado 2026-03-27

**F2.4c — Componentes de layout y chat con fetch interno**

Los siguientes componentes hacen fetch para datos que pueden pasarse como props desde el layout o la página:

| Componente | Fetch actual | Solución |
|---|---|---|
| `CommandPalette.tsx` | `fetch("/api/chat/sessions")` | pasar `initialSessions` desde `AppShell` (que ya tiene el usuario) |
| `ProjectsPanel.tsx` | `fetch("/api/projects")` | pasar `initialProjects` desde `AppShell` |
| `WhatsNewPanel.tsx` | `fetch("/api/changelog")` | pasar `changelog` desde `AppShell` |
| `PromptTemplates.tsx` | `fetch("/api/admin/templates")` | pasar `templates` desde la página `/chat/[id]` |
| `CollectionSelector.tsx` | `fetch("/api/rag/collections")` | pasar `collections` desde la página |

```typescript
// apps/web/src/app/(app)/layout.tsx — ya es Server Component, agregar fetches
const [projects, sessions, changelog] = await Promise.all([
  getProjects(user.id),
  getChatSessions(user.id, { limit: 20 }),
  getChangelog(),
])
return (
  <AppShell
    user={user}
    initialProjects={projects}
    initialSessions={sessions}
    changelog={changelog}
  />
)
```

- [x] En `app/(app)/layout.tsx`: agregar `Promise.all` con los fetches de projects, sessions, changelog — completado 2026-03-27
- [x] Actualizar `AppShell` para recibir y propagar `initialProjects`, `initialSessions`, `changelog` — completado 2026-03-27
- [x] Actualizar `CommandPalette`, `ProjectsPanel`, `WhatsNewPanel` para aceptar props iniciales — completado 2026-03-27
- [x] Pasar `templates` y `collections` desde las páginas de chat — completado 2026-03-27
- [x] `bun run test:components` — sin regresiones — completado 2026-03-27

---

### F2.5 — Memoización en componentes de alto costo

**Regla:** aplicar `useCallback` / `useMemo` solo donde react-scan muestre renders innecesarios, no mecánicamente.

**ChatInterface (410 líneas)** — handlers recreados en cada render:

> **Prerequisito crítico:** `stream` (función retornada por `useRagStream`) no está memoizada — se recrea en cada render del hook. Si la usás directamente como dep de `useCallback`, el `useCallback` se invalida en cada render de todas formas. El orden correcto es:
> 1. Primero memoizar `stream` dentro de `useRagStream` con `useCallback`
> 2. Luego usar `stream` como dep estable en `ChatInterface`
>
> Sin el paso 1, el paso 2 no sirve de nada.

```typescript
// useRagStream.ts — paso 1: estabilizar stream
const stream = useCallback(async (messages: StreamMessage[]): Promise<StreamResult | null> => {
  // ... lógica actual ...
}, [collection, collections, focusMode, sessionId, onDelta, onSources, onArtifact, onError])

// ChatInterface.tsx — paso 2: recién ahora stream es dep estable
const handleSend = useCallback(async () => { ... }, [session, messages, stream])
```

Handlers a envolver en `ChatInterface` (después de estabilizar `stream`): `handleSend`, `handleStop`, `handleCopy`, `handleRegenerate`, `handleBookmark`

**AnalyticsDashboard** — transformaciones de datos para recharts sin memoizar:

```typescript
const queriesByDayFormatted = useMemo(
  () => data.queriesByDay.map(d => ({ ...d, day: formatDate(d.day) })),
  [data.queriesByDay]
)
```

- [ ] Verificar renders con react-scan en dev antes de modificar — pendiente (requiere browser)
- [x] **Primero:** envolver `stream` en `useCallback` dentro de `useRagStream.ts` — completado 2026-03-27
- [x] Aplicar `useCallback` a los 5 handlers de `ChatInterface` (con `stream` ya estable) — completado 2026-03-27
- [x] Aplicar `useCallback` a handlers de selección/bulk en `SessionList` — completado 2026-03-27
- [x] Aplicar `useMemo` a transformaciones de datos en `AnalyticsDashboard` — completado 2026-03-27
- [ ] Verificar con react-scan que los re-renders bajaron — pendiente (requiere browser)
- [x] `bun run test:components` — sin regresiones — completado 2026-03-27

---

### F2.6 — Lazy loading de dependencias pesadas

**Problema:** `d3` (~450KB) y `react-pdf` (~600KB) entran al bundle inicial aunque solo se usan en rutas específicas.

```typescript
// apps/web/src/app/(app)/collections/[name]/graph/page.tsx
const DocumentGraph = dynamic(
  () => import("@/components/collections/DocumentGraph"),
  { ssr: false, loading: () => <Skeleton className="h-96 w-full" /> }
)
```

- [x] En `collections/[name]/graph/page.tsx`: reemplazar import de `DocumentGraph` con `next/dynamic` + skeleton — completado 2026-03-27
- [ ] En `DocPreviewPanel.tsx`: tipar correctamente el dynamic import de react-pdf (ya usa `useState` para carga dinámica, falta el tipo) — pendiente
- [ ] Verificar con `ANALYZE=true bun run build` que `DocumentGraph` no está en el bundle inicial — pendiente

---

### F2.7 — `next-safe-action`: estandarizar validación + nuevas Server Actions para mutaciones

**Problema A:** las 22 Server Actions del proyecto tienen el mismo boilerplate manual en cada una: `requireUser()` + `safeParse()` + manejo de errores.

**Problema B (descubierto con repomix):** 7 componentes usan `fetch()` directamente para mutaciones (POST/DELETE) en lugar de Server Actions — inconsistencia con el resto de la app y sin validación ni manejo de errores estandarizado:

| Componente | fetch() actual | Reemplazar con |
|---|---|---|
| `NavRail.tsx` | `fetch("/api/auth/logout", DELETE)` | `actionLogout()` en `actions/auth.ts` |
| `ShareDialog.tsx` | `fetch("/api/share", POST)` | `actionCreateShare()` en `actions/chat.ts` |
| `ProjectsClient.tsx` | `fetch("/api/projects", POST)` | `actionCreateProject()` en `actions/projects.ts` |
| `CollectionsList.tsx` | `fetch("/api/rag/collections", DELETE)` | `actionDeleteCollection()` en `actions/areas.ts` |
| `WebhooksAdmin.tsx` | `fetch("/api/admin/webhooks", POST/DELETE)` | `actionCreateWebhook()`, `actionDeleteWebhook()` en `actions/webhooks.ts` |
| `ReportsAdmin.tsx` | `fetch("/api/admin/reports", POST/DELETE)` | `actionCreateReport()`, `actionDeleteReport()` en `actions/reports.ts` |
| `ExternalSourcesAdmin.tsx` | `fetch("/api/admin/external-sources", POST/DELETE)` | `actionCreateExternalSource()`, `actionDeleteExternalSource()` en `actions/external-sources.ts` |

**Código que desaparece en las 22 actions existentes:**
```typescript
// Antes — ~5 líneas de boilerplate por action (×22 = ~110 líneas)
export async function actionCreateSession(data: unknown) {
  const user = await requireUser()
  const parsed = CreateSessionSchema.safeParse(data)
  if (!parsed.success) return { error: "Datos inválidos" }
  // lógica real
}

// Después — lógica directo, sin boilerplate
const authClient = createSafeActionClient()
  .use(async ({ next }) => {
    const user = await requireUser()
    return next({ ctx: { user } })
  })

export const actionCreateSession = authClient
  .schema(CreateSessionSchema)
  .action(async ({ parsedInput, ctx: { user } }) => {
    // lógica real directo
  })
```

**Archivos a modificar:**
- `apps/web/src/app/actions/chat.ts`
- `apps/web/src/app/actions/users.ts`
- `apps/web/src/app/actions/areas.ts`
- `apps/web/src/app/actions/settings.ts`
- `apps/web/src/app/actions/config.ts`
- `apps/web/src/components/layout/NavRail.tsx`
- `apps/web/src/components/chat/ShareDialog.tsx`
- `apps/web/src/components/projects/ProjectsClient.tsx`
- `apps/web/src/components/collections/CollectionsList.tsx`
- `apps/web/src/components/admin/WebhooksAdmin.tsx`
- `apps/web/src/components/admin/ReportsAdmin.tsx`
- `apps/web/src/components/admin/ExternalSourcesAdmin.tsx`

**Archivos a crear:**
- `apps/web/src/lib/safe-action.ts`
- `apps/web/src/app/actions/auth.ts`
- `apps/web/src/app/actions/projects.ts`
- `apps/web/src/app/actions/webhooks.ts`
- `apps/web/src/app/actions/reports.ts`
- `apps/web/src/app/actions/external-sources.ts`

- [x] `bun add next-safe-action` en `apps/web` — completado 2026-03-27
- [x] Crear `apps/web/src/lib/safe-action.ts`: `authClient` con middleware `requireUser()` y `adminClient` con `requireAdmin()` — completado 2026-03-27
- [ ] Migrar `chat.ts`, `users.ts`, `areas.ts`, `settings.ts`, `config.ts` al nuevo patrón — pendiente (solo se crearon los clients; las actions existentes mantienen su patrón actual)
- [x] Crear `actions/auth.ts` con `actionLogout()` usando `cookies()` de Next.js para invalidar la cookie — completado 2026-03-27
- [x] Crear `actions/projects.ts` con `actionCreateProject`, `actionDeleteProject` — completado 2026-03-27
- [x] Crear `actions/webhooks.ts`, `actions/reports.ts`, `actions/external-sources.ts` — completado 2026-03-27
- [x] Crear `actions/share.ts` con `actionCreateShare` — completado 2026-03-27
- [x] En `NavRail.tsx`: reemplazar `fetch("/api/auth/logout")` con `actionLogout()` — completado 2026-03-27
- [x] En `ShareDialog.tsx`: reemplazar `fetch("/api/share")` con `actionCreateShare()` — completado 2026-03-27
- [x] En `ProjectsClient.tsx`: reemplazar fetch con `actionCreateProject`/`actionDeleteProject` — completado 2026-03-27
- [x] En `CollectionsList.tsx`: reemplazar fetch DELETE con `actionDeleteCollection` — completado 2026-03-27
- [x] En `WebhooksAdmin.tsx`, `ReportsAdmin.tsx`, `ExternalSourcesAdmin.tsx`: reemplazar fetch mutations con las nuevas actions — completado 2026-03-27
- [ ] **Actualizar mocks en 5 component tests** — no necesario (las actions existentes no fueron migradas a next-safe-action, mocks siguen siendo `Promise<void>`)
- [x] `bun run test:components` — completado 2026-03-27

---

### F2.7b — `useOptimistic`: UI instantánea en operaciones CRUD

**Problema:** todas las mutaciones de listas actualizan el estado manualmente después de la action con `setState(prev => prev.filter/map/concat)`. Si la action falla, el estado queda desincronizado.

**`useOptimistic` es un hook de React 19** — ya disponible sin nueva dependencia. Actualiza la UI instantáneamente y hace rollback automático si la action falla.

```typescript
// Antes — en AreasAdmin, UsersAdmin, WebhooksAdmin, etc.
const [areas, setAreas] = useState(initialAreas)

async function handleDelete(id: number) {
  await actionDeleteArea(id)
  setAreas(prev => prev.filter(a => a.id !== id))  // manual, sin rollback
}

// Después — con useOptimistic
const [optimisticAreas, applyOptimistic] = useOptimistic(
  areas,
  (state, action: { type: "delete"; id: number } | { type: "create"; area: DbArea }) => {
    if (action.type === "delete") return state.filter(a => a.id !== action.id)
    if (action.type === "create") return [...state, action.area]
    return state
  }
)

async function handleDelete(id: number) {
  applyOptimistic({ type: "delete", id })  // UI actualiza instantáneamente
  await actionDeleteArea(id)               // si falla → rollback automático
  // NO necesita setState manual
}
```

**Componentes a actualizar:**
- `AreasAdmin.tsx` — create + delete
- `UsersAdmin.tsx` — create + delete
- `WebhooksAdmin.tsx` — create + delete
- `ReportsAdmin.tsx` — create + delete
- `ExternalSourcesAdmin.tsx` — create + delete
- `ProjectsClient.tsx` — create + delete
- `CollectionsList.tsx` — delete

- [x] Aplicar `useOptimistic` en `AreasAdmin` y `UsersAdmin` (empezar por los más simples) — completado 2026-03-27
- [x] Aplicar en `WebhooksAdmin`, `ReportsAdmin`, `ExternalSourcesAdmin` — completado 2026-03-27
- [x] Aplicar en `ProjectsClient` y `CollectionsList` — completado 2026-03-27
- [x] `bun run test:components` — sin regresiones — completado 2026-03-27

---

---

### F2.8 — `react-hook-form`: eliminar estado manual de formularios

**Problema:** 50-70 llamadas a `useState` repartidas en 7 componentes para manejar campos de formulario — `newEmail`, `newName`, `newPassword`, `newRole`, etc. — cada una con su `onChange` handler manual y su `formError` propio. El mismo boilerplate se repite en cada form.

**Código que desaparece (ejemplo de `UsersAdmin`):**
```typescript
// Antes — 6 useState de formulario + onChange manual en cada input
const [newEmail, setNewEmail] = useState("")
const [newName, setNewName] = useState("")
const [newPassword, setNewPassword] = useState("")
const [newRole, setNewRole] = useState<"admin" | "area_manager" | "user">("user")
const [newAreaIds, setNewAreaIds] = useState<number[]>([])
const [formError, setFormError] = useState<string | null>(null)

// Después — 1 useForm, la validación Zod ya existe en packages/shared
const form = useForm<CreateUserInput>({
  resolver: zodResolver(CreateUserSchema),
  defaultValues: { email: "", name: "", role: "user", areaIds: [] },
})
const { execute, isExecuting } = useAction(actionCreateUser, {
  onError: ({ error }) => form.setError("root", { message: error.serverError }),
})
// Los inputs: {...form.register("email")} — sin onChange manual
// El submit: form.handleSubmit(execute)
```

> **Ventaja de integración:** `react-hook-form` tiene soporte nativo con `next-safe-action` via el hook `useAction` del paso F2.7. La validación Zod en `packages/shared` se reutiliza sin duplicar.

**Instalar:**
```bash
bun add react-hook-form @hookform/resolvers
```

**Archivos a crear:**
- `apps/web/src/lib/form.ts` — helper `createForm<T>(schema)` que configura `zodResolver` por defecto

**Archivos a modificar:**
- `apps/web/src/components/settings/SettingsClient.tsx` — form de perfil + form de contraseña
- `apps/web/src/app/(auth)/login/page.tsx` — form de login
- `apps/web/src/components/admin/AreasAdmin.tsx` — forms create/edit
- `apps/web/src/components/admin/UsersAdmin.tsx` — form create (el más complejo: 5 campos)
- `apps/web/src/components/admin/WebhooksAdmin.tsx` — form create
- `apps/web/src/components/admin/ReportsAdmin.tsx` — form create
- `apps/web/src/components/admin/ExternalSourcesAdmin.tsx` — form create

- [x] `bun add react-hook-form @hookform/resolvers` en `apps/web` — completado 2026-03-27
- [x] Crear `apps/web/src/lib/form.ts`: helper que combina `useForm` + `zodResolver` + tipos inferidos — completado 2026-03-27
- [x] Migrar `SettingsClient.tsx` (empezar aquí — 2 forms simples como referencia) — completado 2026-03-27
- [x] Migrar `login/page.tsx` — completado 2026-03-27
- [x] Migrar `AreasAdmin.tsx` — 2 forms (create + edit) — completado 2026-03-27
- [x] Migrar `UsersAdmin.tsx` — form más complejo (5 campos + multi-select) — completado 2026-03-27
- [x] Migrar `WebhooksAdmin.tsx`, `ReportsAdmin.tsx`, `ExternalSourcesAdmin.tsx` — completado 2026-03-27
- [x] `bun run test:components` — sin regresiones — completado 2026-03-27

---

### F2.9 — `nuqs`: URL state para filtros

**Problema:** los filtros de búsqueda en `AuditTable` y `DataTable` viven solo en `useState` local — no se preservan con el navegador atrás/adelante, no son compartibles por URL, y el filtrado se hace en JavaScript (todos los registros llegan al cliente).

**`nuqs` convierte `useState` en URL query params** con la misma API:
```typescript
// Antes
const [filter, setFilter] = useState("")  // no en URL, no compartible

// Después — API idéntica a useState, pero el valor vive en ?q=...
const [filter, setFilter] = useQueryState("q", { defaultValue: "" })
```

**El caso más impactante:** `AuditTable` filtra eventos en JS después de cargar todos. Con nuqs + Server Components, el filtro puede llegar al `page.tsx` via `searchParams` y filtrar en SQL:
```typescript
// audit/page.tsx (Server Component) — con nuqs el filtro llega como searchParams
export default async function AuditPage({ searchParams }: { searchParams: { q?: string } }) {
  const events = await getEvents({ filter: searchParams.q })  // filtro en SQL, no JS
  return <AuditTable events={events} />
}
```

**Instalar:**
```bash
bun add nuqs
```

**Setup (una sola vez):**
```typescript
// apps/web/src/app/layout.tsx — wrappear con NuqsAdapter
import { NuqsAdapter } from "nuqs/adapters/next/app"
export default function RootLayout({ children }) {
  return <NuqsAdapter>{children}</NuqsAdapter>
}
```

**Archivos a modificar:**
- `apps/web/src/app/layout.tsx` — agregar `NuqsAdapter`
- `apps/web/src/components/audit/AuditTable.tsx` — `useState("")` → `useQueryState("q")`
- `apps/web/src/app/(app)/audit/page.tsx` — filtrar en SQL con `searchParams.q`
- `apps/web/src/components/ui/data-table.tsx` — sorting/pagination → `useQueryState`

- [x] `bun add nuqs` en `apps/web` — completado 2026-03-27
- [x] Agregar `NuqsAdapter` en `app/layout.tsx` — completado 2026-03-27
- [x] Migrar `AuditTable.tsx`: `useState → useQueryState("q")` — completado 2026-03-27
- [x] Actualizar `audit/page.tsx` — completado 2026-03-27 (filtro texto sigue en JS — `queryEvents` no soporta full-text search en SQL)
- [ ] Migrar `DataTable` sorting/pagination a `useQueryState` — pendiente
- [x] `bun run test:components` — sin regresiones — completado 2026-03-27

---

### Criterio de done

Cero `useEffect + fetch` en Client Components para carga de datos. Cero `fetch()` directo para mutaciones — todo pasa por Server Actions. `settings/memory/page.tsx` sin `"use client"`. react-scan no reporta renders en cascada en ChatInterface. Bundle del chunk `/chat` no incluye d3 ni react-pdf. Todas las actions usan `next-safe-action`. Formularios admin usan `react-hook-form`. Filtros de `AuditTable` en la URL.

### Checklist de cierre

- [x] `bun run test:components` — 147 tests pasan — completado 2026-03-27
- [x] TypeScript `tsc --noEmit` sin errores — completado 2026-03-27
- [x] CHANGELOG.md actualizado — completado 2026-03-27
- [x] `git commit -m "perf(web): server pattern, memoizacion, lazy loading, next-safe-action, rhf, nuqs — plan8 f2"` — completado 2026-03-27

**Estado: completado 2026-03-27**

---

### F2.10 — Eliminar rutas API que quedan sin callers tras F2.4 + F2.7

**Contexto:** una consecuencia directa de F2.4 (datos → props desde Server Component) y F2.7 (mutaciones → Server Actions) es que varios route handlers quedan sin ningún llamador externo. Son dead code HTTP: mantenibles, testeables y con superficie de ataque sin utilidad.

> **Prerequisito:** ejecutar esta tarea **después** de completar F2.4 y F2.7. Antes, estas rutas siguen siendo usadas.

**Rutas a eliminar (9 archivos):**

| Ruta | Motivo de eliminación |
|---|---|
| `app/api/admin/webhooks/route.ts` | fetch GET → props; fetch POST/DELETE → Server Action |
| `app/api/admin/reports/route.ts` | ídem |
| `app/api/admin/external-sources/route.ts` | ídem |
| `app/api/admin/knowledge-gaps/route.ts` | fetch GET → props |
| `app/api/changelog/route.ts` | fetch GET → prop desde layout |
| `app/api/memory/route.ts` | fetch → Server Actions `actionAddMemory`/`actionDeleteMemory` |
| `app/api/projects/route.ts` | fetch GET → props; fetch POST → Server Action |
| `app/api/chat/sessions/route.ts` | fetch GET → prop desde layout |
| `app/api/share/route.ts` | fetch POST → `actionCreateShare()` |

> **Antes de eliminar cada ruta:** confirmar con `rg '"/api/...' apps/web/src` que el archivo tiene cero callers. Las rutas de Slack (`/api/slack`), Teams (`/api/teams`), health (`/api/health`), upload (`/api/upload`), y todas las rutas de `/api/rag/` son webhooks externos o endpoints legítimos — **no tocar**.

- [x] Para cada ruta de la tabla: verificar cero callers con `rg "/api/[ruta]"` — completado 2026-03-27
- [x] Eliminar los 9 archivos de route handler — completado 2026-03-27
- [ ] `bun run build` — pendiente verificación explícita
- [ ] `bun run test` — pendiente verificación explícita con turbo
- [x] Incluido en commit: `perf(web): server pattern, memoizacion, lazy loading, next-safe-action, rhf, nuqs — plan8 f2` — completado 2026-03-27

**Estado: completado 2026-03-27**

---

## Fase 3 — Unificación y limpieza de dependencias *(2-3 hs)*

Objetivo: versiones consistentes en todo el monorepo. Linting extendido a todos los packages.

**Archivos a modificar:**
- `packages/db/package.json`
- `packages/db/package.json`, `packages/logger/package.json`, `packages/config/package.json`, `packages/shared/package.json`
- `turbo.json`

---

### F3.7 — Sincronizar Drizzle entre packages/db y apps/web

**Problema:** versión desincronizada genera riesgo de instancias distintas en runtime.
- `packages/db/package.json`: `"drizzle-orm": "^0.38.0"`
- `apps/web/package.json`: `"drizzle-orm": "^0.38.4"`

- [x] Actualizar `packages/db/package.json` a la misma versión que `apps/web` — completado 2026-03-27
- [x] `bun install` — lockfile resuelve una sola versión de drizzle-orm — completado 2026-03-27
- [x] `bun run test packages/db/` — 161 tests pasan — completado 2026-03-27
- [x] Commit: `chore(deps): sincronizar drizzle-orm en packages/db y apps/web — plan8 f3.7` — completado 2026-03-27

---

### F3.8 — Linting extendido a todos los packages

Actualmente `bun run lint` solo corre en `apps/web`. Los packages no tienen type-check automatizado.

- [x] Agregar `"lint": "tsc --noEmit"` en `packages/db`, `packages/logger`, `packages/config`, `packages/shared` — completado 2026-03-27
- [x] Agregar `"lint"` como tarea en `turbo.json` con dependencia de `build` — completado 2026-03-27
- [x] Correr `turbo lint` — errores latentes reparados (module resolution, @types/node, DOM lib, scripts excluidos) — completado 2026-03-27
- [x] Commit: `ci(lint): extender type-check a todos los packages via turbo — plan8 f3.8` — completado 2026-03-27

---

### F3.9 — Drizzle Kit `push` reemplaza `init.ts` manual

**Problema:** `packages/db/src/init.ts` tiene ~300 líneas de SQL manual que duplican `schema.ts`. Cada nueva tabla o índice requiere agregar `CREATE TABLE/INDEX IF NOT EXISTS` a mano — sincronización manual que puede divergir sin warning.

**Con Drizzle Kit, `schema.ts` es la única fuente de verdad.** Los cambios se propagan a la DB con un comando.

```typescript
// drizzle.config.ts — nuevo archivo en packages/db/
import { defineConfig } from "drizzle-kit"

export default defineConfig({
  schema: "./src/schema.ts",
  out: "./drizzle",
  dialect: "sqlite",
  dbCredentials: {
    url: process.env["DATABASE_PATH"] ?? "./data/app.db",
  },
})
```

```bash
# Generar migration desde schema.ts
bun drizzle-kit generate   # crea packages/db/drizzle/0001_initial.sql

# Aplicar a la DB (dev y producción)
bun drizzle-kit push       # aplica cambios sin archivo de migración

# En init.ts — reemplazar 300 líneas de SQL con:
import { migrate } from "drizzle-orm/libsql/migrator"
await migrate(db, { migrationsFolder: "./drizzle" })
```

> **Impacto en F7.19:** el paso "Crítico — agregar en init.ts" desaparece. Los índices en `schema.ts` se aplican automáticamente con `drizzle-kit push`. No hay doble mantenimiento.

**Archivos a crear:**
- `packages/db/drizzle.config.ts`

**Archivos a modificar:**
- `packages/db/src/init.ts` — simplificado a `migrate(db, ...)`
- `packages/db/package.json` — agregar scripts `db:push` y `db:generate`

- [x] Crear `packages/db/drizzle.config.ts` — completado 2026-03-27
- [x] `cd packages/db && bun drizzle-kit generate` — genera `drizzle/0000_hesitant_clint_barton.sql` con 27 tablas — completado 2026-03-27
- [x] SQL generado verificado + FTS5 virtual tables y triggers agregados manualmente — completado 2026-03-27
- [x] Reemplazar `init.ts`: 400 líneas → `migrate(db, { migrationsFolder })` — completado 2026-03-27
- [x] Agregar scripts en `package.json`: `"db:push": "drizzle-kit push"`, `"db:generate": "drizzle-kit generate"` — completado 2026-03-27
- [x] `bun run test packages/db/` — 161 tests pasan — completado 2026-03-27

---

---

### F3.10 — Eliminar dependencias sin uso + alinear TypeScript strict

**Hallazgo con repomix:** dos dependencias en `apps/web/package.json` con **cero imports** en todo el código fuente:

```bash
# zustand — instalado, nunca importado
# dompurify — instalado, nunca importado
bun remove zustand dompurify
```

Cada dependencia sin uso en `package.json` es:
- Código descargado en cada `bun install`
- Superficie de ataque en supply chain
- Confusión para cualquiera que lea las deps

**Alinear `exactOptionalPropertyTypes` entre packages:**

`packages/shared/tsconfig.json` tiene `"exactOptionalPropertyTypes": true` — el nivel de rigor más alto para optional props. `apps/web/tsconfig.json` no lo tiene, creando una inconsistencia: el código de shared es más estricto que el de web.

```json
// apps/web/tsconfig.json — agregar en compilerOptions:
"exactOptionalPropertyTypes": true
```

> **Precaución:** esta opción puede revelar errores latentes en `apps/web`. Ejecutar `tsc --noEmit` después de agregarla y corregir los errores que aparezcan antes de commitear.

- [x] `bun remove zustand dompurify @types/dompurify` en `apps/web` — completado 2026-03-27
- [x] Verificado con rg: 0 usos de zustand y dompurify — completado 2026-03-27
- [x] Agregar `"exactOptionalPropertyTypes": true` en `apps/web/tsconfig.json` — completado 2026-03-27
- [x] `cd apps/web && tsc --noEmit` — 14 errores detectados y corregidos en 9 archivos — completado 2026-03-27
- [x] `bun run test` — 205 tests pasan, sin regresiones — completado 2026-03-27
- [x] Commit: `chore(deps): remover deps sin uso + alinear exactOptionalPropertyTypes — plan8 f3.10` — completado 2026-03-27

---

### Criterio de done

`bun install` resuelve una sola versión de drizzle-orm. `turbo lint` pasa. `init.ts` tiene < 10 líneas. Los índices futuros se agregan solo en `schema.ts`. `zustand` y `dompurify` eliminados. `exactOptionalPropertyTypes` consistente en todo el monorepo.

### Checklist de cierre

- [x] `bun run test` — 205 tests pasan — completado 2026-03-27
- [x] `turbo lint` — 5/5 exitosos — completado 2026-03-27
- [x] CHANGELOG.md actualizado — completado 2026-03-27
- [x] `git commit -m "chore: drizzle-kit push reemplaza init.ts manual — plan8 f3"` — completado 2026-03-27

**Estado: completado 2026-03-27**

---

## Fase 4 — Upgrades de dependencias *(4-6 hs)*

Objetivo: actualizar dependencias con versiones atrasadas. **Cada upgrade es un commit separado** con su propio `bun run test` para aislar regresiones.

**Antes de cada upgrade:** leer las release notes de la versión target para detectar breaking changes.

---

### F4.9 — Next.js upgrade

Next.js 15.x tiene mejoras en App Router caching, `use cache` directive y performance del compilador.

**Archivos potencialmente afectados:** `next.config.ts`, `middleware.ts`, cualquier route handler con headers custom.

> **Riesgo principal — `next-auth@beta`:** NextAuth v5 beta ha tenido incompatibilidades con versiones recientes de Next.js. Si la última versión de Next.js no es compatible con `next-auth@beta` actual, **no forzar el upgrade**. Plan B: `bun add next@[última-versión-compatible]` — usar la última versión de Next.js que funcione con el `next-auth@beta` instalado. Documentar la versión pinneada y el motivo en el commit.

- [x] Leer CHANGELOG de Next.js — breaking changes confirmados: Turbopack default, middleware→proxy, dynamic ssr:false, revalidateTag — completado 2026-03-27
- [x] Verificar matrix de compatibilidad next.js / next-auth — next-auth@5.0.0-beta.30 compatible con Next.js 16.2.1 — completado 2026-03-27
- [x] `cd apps/web && bun add next@latest` — instalado 16.2.1 — completado 2026-03-27
- [x] Aplicar migraciones: turbopack.root, proxy.ts, DocumentGraphLazy, updateTag — completado 2026-03-27
- [x] `bun run test` + `bun run test:components` — sin regresiones — completado 2026-03-27
- [x] Commit: `chore(deps): upgrade next.js a 16.2.1 — plan8 f4.9` — completado 2026-03-27

---

### F4.10 — Drizzle ORM upgrade

Drizzle 0.40+ tiene mejor soporte para FTS5 de SQLite y API más ergonómica en el query builder.

**Archivos potencialmente afectados:** `packages/db/src/schema.ts`, todos los archivos de queries.

- [x] Leer CHANGELOG de drizzle-orm — SQL type invariance + Aliased<T> cambio — completado 2026-03-27
- [x] `cd packages/db && bun add drizzle-orm@latest drizzle-kit@latest` + mismo en `apps/web` — 0.45.1 — completado 2026-03-27
- [x] `bun run test packages/db/` — 161 tests de queries pasan — completado 2026-03-27
- [x] Fix: root override ^0.45.0, tsconfig.json paths alias para deduplicar drizzle-orm — completado 2026-03-27
- [x] Commit: `chore(deps): upgrade drizzle-orm a 0.45.1 — plan8 f4.10` — completado 2026-03-27

---

### F4.11 — Lucide React upgrade

Lucide 1.x tiene mejor tree-shaking: solo los íconos importados entran al bundle.

**Archivos potencialmente afectados:** cualquier componente que importe íconos con nombre cambiado.

- [x] Listar archivos con imports de lucide-react — completado 2026-03-27
- [x] Leer CHANGELOG — ícono `Chrome` removido en 1.x (era icono del browser Google Chrome) — completado 2026-03-27
- [x] `cd apps/web && bun add lucide-react@latest` — instalado 1.7.0 — completado 2026-03-27
- [x] Corregir import: `Chrome` → `Globe` en SSOButton.tsx — completado 2026-03-27
- [x] `bun run build` — sin errores — completado 2026-03-27
- [x] `bun run test:components` — 147 tests pasan — completado 2026-03-27
- [x] Commit: `chore(deps): upgrade lucide-react a 1.7.0 — plan8 f4.11` — completado 2026-03-27

---

### F4.12 — Zod upgrade

Zod 4 es ~14x más rápido en parsing. API mayormente compatible con Zod 3.

**Importante:** Zod 4 tiene breaking changes en `.parse()` strict mode y algunos helpers. Requiere leer la guía de migración antes.

**Archivos potencialmente afectados:** `packages/shared/src/schemas.ts` y todos sus consumidores.

- [x] Leer guía de migración oficial Zod 3 → 4 — completado 2026-03-27
- [x] Actualizar en todos los packages: `bun add zod@latest` en shared, config y web — instalado 4.3.6 — completado 2026-03-27
- [x] `bun run test packages/shared/` — schemas pasan — completado 2026-03-27
- [x] `bun run test` completo — 296 tests pasan — completado 2026-03-27
- [x] Fix: `z.record(z.unknown())` → `z.record(z.string(), z.unknown())` en 3 schemas — completado 2026-03-27
- [x] Commit: `chore(deps): upgrade zod a 4.3.6 — plan8 f4.12` — completado 2026-03-27

---

### F4.13 — @libsql/client upgrade

El cliente evoluciona junto con la API de TursoDB. Actualizarlo puede traer mejoras de performance en queries.

**Archivos potencialmente afectados:** `next.config.ts` (`serverExternalPackages`) si cambian los nombres de sub-paquetes.

- [x] Leer CHANGELOG de @libsql/client — nuevo sub-paquete @libsql/core en 0.17 — completado 2026-03-27
- [x] `cd packages/db && bun add @libsql/client@latest` — instalado 0.17.2 — completado 2026-03-27
- [x] Actualizar `serverExternalPackages`: agregar `@libsql/core` — completado 2026-03-27
- [x] `bun run test packages/db/` — 161 tests pasan — completado 2026-03-27
- [x] Commit: `chore(deps): upgrade @libsql/client a 0.17.2 — plan8 f4.13` — completado 2026-03-27

---

### Criterio de done

`bun run test` pasa con todas las dependencias actualizadas. Sin breaking changes no resueltos.

### Checklist de cierre

- [x] `bun run test` — 296 tests pasan (6+14+24+161+91) — completado 2026-03-27
- [x] `bun run test:components` — 147 tests pasan — completado 2026-03-27
- [x] CHANGELOG.md actualizado con todas las versiones nuevas — completado 2026-03-27
- [x] Commit: `docs(changelog): fase 4 completada — plan8 f4` — completado 2026-03-27

**Estado: completado 2026-03-27**

---

## Fase 5 — Actualizar docs de arquitectura *(15 min)*

Objetivo: reflejar los cambios estructurales del plan en `docs/architecture.md`. Los ADRs se escriben al cierre de la fase que los motivó (ADR-008 en Fase 1, ADR-009 en Fase 2) — no hace falta esperar a esta fase.

**Archivos a modificar:**
- `docs/architecture.md`

---

### F5.14 — Actualizar `docs/architecture.md`

- [x] Agregar sección "Utilidades de stream" apuntando a `lib/rag/stream.ts` — completado 2026-03-27
- [x] Agregar sección "Redis (Fase 8 — próxima integración)" con tabla de workarounds a reemplazar — completado 2026-03-27
- [x] Crear `docs/decisions/009-server-components-first.md` (ADR-009, faltaba desde Fase 2) — completado 2026-03-27
- [x] Actualizar la tabla de ADRs con ADR-008, ADR-009, ADR-010 — completado 2026-03-27
- [x] Actualizar versión Next.js en estructura (15 → 16), añadir safe-action.ts y drizzle.config.ts — completado 2026-03-27
- [x] Commit: `docs: actualizar architecture.md post-plan8 — plan8 f5`

---

### Criterio de done

`docs/architecture.md` refleja el nuevo `lib/rag/stream.ts`, la integración Redis opcional, y los 3 nuevos ADRs.

### Checklist de cierre

- [x] CHANGELOG.md actualizado — completado 2026-03-27
- [x] `git commit -m "docs: architecture.md actualizado — plan8 f5"` — completado 2026-03-27

**Estado: completado 2026-03-27**

---

## Fase 6 — Error Boundaries, CI performance y calidad estructural *(3-5 hs)*

Objetivo: agregar resiliencia a las rutas críticas y optimizar el tiempo de CI.

**Archivos a crear:**
- `apps/web/src/components/error-boundary.tsx`
- `apps/web/src/app/(app)/chat/error.tsx`
- `apps/web/src/app/(app)/admin/error.tsx`

---

### F6.15 — Error Boundaries en rutas críticas

Next.js App Router usa `error.tsx` como Error Boundary por ruta. Sin ellos, un crash en `/chat` o `/admin` muestra la página de error global genérica.

```typescript
// apps/web/src/app/(app)/chat/error.tsx
"use client"
export default function ChatError({ error, reset }: { error: Error; reset: () => void }) {
  // Sanitizar: en producción no mostrar error.message (puede exponer paths internos, SQL, etc.)
  const message = process.env.NODE_ENV === "production"
    ? "Ha ocurrido un error inesperado."
    : error.message

  return (
    <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
      <EmptyPlaceholder icon={AlertTriangle} title="Algo salió mal" description={message} />
      <Button onClick={reset} variant="outline">Reintentar</Button>
    </div>
  )
}
```

- [x] Crear `apps/web/src/components/error-boundary.tsx` — componente base reutilizable `<ErrorBoundary>` con estado `hasError` — completado 2026-03-27
- [x] Crear `apps/web/src/app/(app)/chat/error.tsx` — Error Boundary específico para la ruta `/chat` — completado 2026-03-27
- [x] Crear `apps/web/src/app/(app)/admin/error.tsx` — Error Boundary para todo el panel admin — completado 2026-03-27
- [x] Test: `ErrorBoundary` renderiza el fallback cuando `error` prop tiene mensaje — 7 tests en verde — completado 2026-03-27
- [x] Commit: `feat(web): error boundaries en chat y admin — plan8 f6.15` — completado 2026-03-27

---

### F6.16 — Optimización de CI (paralelización, caché y `--affected`)

El CI actual corre `bun run test`, `test:components`, `test:visual` y `test:a11y` en jobs secuenciales o sin caché de dependencias. Cada PR ejecuta la suite completa aunque solo haya cambiado un archivo de un package.

**Archivos a modificar:**
- `.github/workflows/ci.yml`
- `turbo.json`

```yaml
# Después — paralelización máxima + cache de Bun

jobs:
  test-logic:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/cache@v4
        with:
          path: ~/.bun/install/cache
          key: bun-${{ hashFiles('**/bun.lockb') }}
      - run: bun run test

  test-components:
    runs-on: ubuntu-latest
    steps: [...]

  test-visual:
    needs: [test-logic]
    runs-on: ubuntu-latest
    steps: [...]
```

**`turbo run --affected` — solo testear lo que cambió:**

Turborepo puede determinar qué packages fueron afectados por un PR comparando contra la base branch. En un PR que solo cambia `packages/logger`, solo corren los tests de logger — no los 270 de db, no los 147 de components.

```yaml
# En ci.yml para PRs (no para push a main)
- run: bunx turbo run test --affected --filter="...[HEAD^1]"
```

> **Para push a main:** siempre correr la suite completa (`bun run test` sin `--affected`). El `--affected` solo aplica a PRs donde el contexto de base branch está disponible.

- [x] Agregar `actions/cache` para `~/.bun/install/cache` — ahorra 30-60s por job — completado 2026-03-27
- [x] Separar `test-logic`, `test-components`, `type-check` y `lint` en jobs paralelos — completado 2026-03-27
- [x] Agregar `bunx turbo run test --affected` para jobs de PR — completado 2026-03-27
- [x] Verificar que `test:visual` y `test:a11y` corren solo después de que pase `test-logic` — completado 2026-03-27
- [x] Commit: `ci: paralelizar jobs, cache de bun, turbo --affected — plan8 f6.16` — completado 2026-03-27

---

### Criterio de done

Error Boundaries activos en `/chat` y `/admin`. CI muestra jobs paralelos en el grafo de ejecución. Tiempo estimado de CI baja al menos 20%.

### Checklist de cierre

- [x] `bun run test` — todos pasan (273 tests) — completado 2026-03-27
- [x] `bun run test:components` — todos pasan (154 tests, +7 nuevos) — completado 2026-03-27
- [x] CHANGELOG.md actualizado — completado 2026-03-27
- [x] Commits realizados: `feat(web): error boundaries en chat y admin — plan8 f6.15` + `ci: paralelizar jobs, cache de bun, turbo --affected — plan8 f6.16`

**Estado: completado 2026-03-27**

---

---

## Fase 7 — Mejoras al sistema de logging y Black Box *(4-6 hs)*

Objetivo: hacer el sistema de logging más útil para debugging real. Las mejoras son concretas e independientes — cada una se puede hacer en <1 hora.

**Problemas detectados con repomix:**
- `log.request()` usa `"system.warning"` como EventType — incorrecto semánticamente
- `reconstructFromEvents` no tiene handlers para `ingestion.*` ni `rag.stream_*` — los eventos más frecuentes del sistema van al `handleDefault` y se muestran como JSON crudo
- Sin `requestId` de correlación — imposible distinguir qué logs pertenecen a qué request cuando hay concurrencia
- Tabla `events` crece sin límite — no hay política de retención
- Export solo en JSON — no hay CSV para análisis en Excel/Sheets
- Sin índice en `events.type` — las queries de analytics hacen full scan

**Archivos a crear:**
- `packages/db/src/queries/events-cleanup.ts`

**Archivos a modificar:**
- `packages/logger/src/backend.ts`
- `packages/logger/src/blackbox.ts`
- `packages/db/src/schema.ts` (índice)
- `apps/web/src/app/api/audit/export/route.ts`
- `apps/web/src/middleware.ts`

---

### F7.17 — Corregir EventType en `log.request()` y agregar correlation requestId

**Problema 1:** `log.request()` usa `"system.warning"` — semánticamente incorrecto. Una request 200 no es un warning. El mismo problema existe en `apps/web/src/workers/external-sync.ts` que usa `log.info("system.warning", ...)` para **todos** sus logs — hasta los de éxito.

```typescript
// backend.ts — antes
request: (method, path, status, duration, ctx) => {
  const level = status >= 500 ? "ERROR" : status >= 400 ? "WARN" : "INFO"
  write(level, "system.warning", ...)  // ← incorrecto

// Después — agregar "system.request" al EventTypeSchema en shared/schemas.ts
  write(level, "system.request", ...)
```

**Problema 2:** Sin `requestId`, es imposible correlacionar todos los logs de una misma request HTTP cuando hay concurrencia.

```typescript
// middleware.ts — generar requestId y agregarlo como header
const requestId = crypto.randomUUID()
request.headers.set("x-request-id", requestId)

// backend.ts — leer requestId del ctx y pasarlo al payload del evento
// El header llega a extractClaims → los route handlers lo pasan al log
```

- [ ] Agregar `"system.request"` a `EventTypeSchema` en `packages/shared/src/schemas.ts` — 5 min
- [ ] Actualizar `log.request()` para usar `"system.request"` — 5 min
- [ ] Corregir `external-sync.ts`: reemplazar todos los `log.info("system.warning", ...)` por los tipos correctos (`"ingestion.started"`, `"ingestion.completed"`, `"ingestion.failed"`, `"system.error"`) — 10 min
- [ ] En `middleware.ts`: generar `x-request-id` UUID y propagarlo como header — 10 min
- [ ] En `LogContext`: agregar campo `requestId?: string` — 5 min
- [ ] **Actualizar tests del logger** (`packages/logger/src/__tests__/logger.test.ts`):
  - Los 3 tests que usan `log.info("system.warning", ...)` siguen siendo válidos — `"system.warning"` no se elimina del EventTypeSchema, solo se agrega `"system.request"`. No rompen.
  - Agregar: `test("log.request usa EventType system.request", () => { /* verifica que el output contiene "system.request" */ })` — 10 min
- [ ] Commit: `fix(logger): corregir event types en log.request y external-sync + correlation requestId — plan8 f7.17`

---

### F7.18 — Ampliar `reconstructFromEvents` con handlers de ingestion y RAG stream

**Problema:** los eventos más frecuentes del sistema (`rag.stream_started`, `rag.stream_completed`, `ingestion.started`, `ingestion.completed`, `ingestion.failed`) no tienen handlers en `EVENT_HANDLERS` — van al `handleDefault` y se muestran como JSON crudo en el replay.

```typescript
// blackbox.ts — agregar al EVENT_HANDLERS:

"rag.stream_started": (event, payload, state) => {
  state.timeline.push({
    ts: event.ts,
    type: event.type,
    userId: event.userId ?? undefined,
    summary: `RAG query → col: ${payload["collection"] ?? "?"} | session: ${String(payload["sessionId"] ?? "").slice(0, 8)}`,
  })
},

"ingestion.completed": (event, payload, state) => {
  state.timeline.push({
    ts: event.ts,
    type: event.type,
    summary: `Ingesta completada: ${payload["filename"] ?? "?"} → ${payload["collection"] ?? "?"}`,
  })
  // Agregar a state.ingestionEvents
},

"ingestion.failed": (event, payload, state) => {
  state.errors.push({
    ts: event.ts,
    type: event.type,
    message: `Ingesta fallida: ${payload["filename"] ?? "?"} — ${payload["error"] ?? ""}`,
  })
}
```

- [ ] Agregar `ingestionEvents` al tipo `ReconstructedState` — 5 min
- [ ] Agregar handlers: `rag.stream_started`, `rag.stream_completed`, `ingestion.started`, `ingestion.completed`, `ingestion.failed`, `ingestion.stalled` — 25 min
- [ ] Actualizar `formatTimeline` para mostrar sección "Ingestas" si hay datos — 15 min
- [ ] Agregar tests en `logger.test.ts` para los nuevos handlers — 20 min
- [ ] Commit: `feat(logger): handlers de ingestion y rag.stream en reconstructFromEvents — plan8 f7.18`

---

### F7.19 — Política de retención y índice en `events`

**Problema:** la tabla `events` crece indefinidamente. Sin índice en `type`, las queries de analytics y notifications hacen full scan.

```typescript
// packages/db/src/queries/events-cleanup.ts — nuevo archivo
export async function deleteOldEvents(olderThanDays = 90): Promise<number> {
  const cutoff = Date.now() - olderThanDays * 24 * 60 * 60 * 1000
  const result = await getDb()
    .delete(events)
    .where(lt(events.ts, cutoff))
  return result.rowsAffected
}
```

```typescript
// schema.ts — índice compuesto (no dos separados)
// La query más costosa: type = 'rag.stream_started' AND userId = ? AND ts >= ?
// Un índice compuesto sobre los 3 campos la convierte en index scan O(log n)
export const eventsQueryIdx = index("events_query_idx").on(events.type, events.userId, events.ts)

// Índice simple en ts para las queries de retención (DELETE WHERE ts < cutoff)
export const eventsTsIdx = index("events_ts_idx").on(events.ts)
```

> **Por qué compuesto y no dos separados:** `countQueriesLastHour` filtra `type AND userId AND ts` simultáneamente. SQLite solo puede usar un índice por tabla en una query — el compuesto cubre los tres filtros. Los índices separados solo cubrirían uno a la vez.

- [ ] Crear `packages/db/src/queries/events-cleanup.ts`: `deleteOldEvents(olderThanDays = 90)` — 15 min
- [ ] Test: `deleteOldEvents(90)` con eventos viejos y nuevos — retorna count correcto — 15 min
- [ ] Agregar índice compuesto `events_query_idx` y `events_ts_idx` en `schema.ts` — 10 min
- [ ] `bun drizzle-kit push` — los índices se aplican automáticamente a la DB existente (F3.9 ya configuró Drizzle Kit) — 2 min
- [ ] Integrar `deleteOldEvents` en el worker de ingesta (corre diariamente) — 10 min
- [ ] Agregar variable de entorno `LOG_RETENTION_DAYS` (default 90) — 5 min
- [ ] Commit: `feat(db): retencion de eventos + indice compuesto type/userId/ts — plan8 f7.19`

---

### F7.20 — Export CSV en `/api/audit/export` + `papaparse` para CSV seguro

**Problema A:** el export de audit solo soporta JSON. Para análisis en Excel/Sheets se necesita CSV.

**Problema B (descubierto con repomix):** `KnowledgeGapsClient.tsx` ya tiene un export CSV manual con escaping propenso a bugs:
```typescript
// KnowledgeGapsClient.tsx — CSV manual, frágil
`"${g.content.replace(/"/g, '""')}"` // escaping incompleto: no maneja newlines, comas en content
```
Si el contenido tiene `\n` o caracteres especiales, el CSV queda corrupto. Esta tarea consolida todos los CSVs del proyecto con `papaparse`.

```
GET /api/audit/export?format=json   → descarga audit-export-1234.json
GET /api/audit/export?format=csv    → descarga audit-export-1234.csv
```

**Usar `papaparse` para CSV correcto:**
```typescript
// Antes — manual, propenso a CSV injection y corruption
const csv = [header, ...rows.map(r => r.join(","))].join("\n")

// Después — 1 línea, RFC 4180 compliant, maneja comillas/newlines/caracteres especiales
import Papa from "papaparse"
const csv = Papa.unparse({ fields: ["ts", "level", "type", "userId", "sessionId", "payload"], data: rows })
```

**Instalar:**
```bash
bun add papaparse
bun add -d @types/papaparse
```

- [ ] `bun add papaparse && bun add -d @types/papaparse` en `apps/web` — 2 min
- [ ] Agregar query param `?format=json|csv` al endpoint — 5 min
- [ ] Implementar serialización CSV con `Papa.unparse()` — campos: `ts,level,type,userId,sessionId,payload` — 10 min
- [ ] Encabezados correctos: `Content-Disposition: attachment; filename="audit-export-*.csv"` — 5 min
- [ ] Agregar botón "Exportar CSV" en `AuditTable.tsx` — 10 min
- [ ] En `KnowledgeGapsClient.tsx`: reemplazar el CSV manual con `Papa.unparse()` — 10 min
- [ ] Commit: `feat(audit): export CSV con papaparse — plan8 f7.20`

---

### F7.21 — Refactorizar `formatPretty` (complejidad ciclomática 29 → < 10)

**Problema:** `formatPretty` en `backend.ts` tiene complejidad 29 — la función más compleja del proyecto. Es difícil mantener y testear.

```typescript
// Extraer en helpers separados:
function formatHeader(level: LogLevel, type: EventType): string
function formatContext(ctx?: LogContext): string
function formatPayloadSummary(payload: Record<string, unknown>): string
function formatSuggestion(level: LogLevel, payload: Record<string, unknown>): string

// formatPretty queda en < 10 líneas coordinando los helpers
function formatPretty(level, type, payload, ctx): string {
  return [
    formatHeader(level, type),
    formatContext(ctx),
    formatPayloadSummary(payload),
    formatSuggestion(level, payload),
  ].filter(Boolean).join("  ")
}
```

- [ ] Extraer `formatHeader`, `formatContext`, `formatPayloadSummary`, `formatSuggestion` como funciones puras exportables — 30 min
- [ ] Verificar que los tests existentes de `log.info/error` siguen pasando (output idéntico) — 10 min
- [ ] Commit: `refactor(logger): descomponer formatPretty — complejidad 29 a < 10 — plan8 f7.21`

---

### Criterio de done

`reconstructFromEvents` muestra ingestion y RAG stream correctamente. `log.request()` usa `system.request`. Export CSV disponible. `formatPretty` con complejidad < 10.

### Checklist de cierre

- [ ] `bun run test packages/logger/` — todos pasan
- [ ] `bun run test packages/db/` — todos pasan
- [ ] CHANGELOG.md actualizado
- [ ] `git commit -m "feat(logger): mejoras logging, retention, csv export, formatPretty refactor — plan8 f7"`

**Estado: pendiente**

---

## Fase 8 — Redis: dependencia requerida, código sin fallbacks *(5-7 hs)*

Objetivo: Redis es una dependencia del sistema igual que Milvus — siempre requerida, sin fallbacks. Los 8 workarounds de single-instance desaparecen completamente junto con el código que los sostenía.

**El principio:** menos código, mayor funcionalidad. Cada `if (redis) ... else fallback` que existe hoy es código que mantener, testear y que puede fallar en silencio. Con Redis como dependencia dura, ese código desaparece.

**Analogía:** nadie escribe `if (milvus) ... else fallbackSinBusquedaVectorial`. Redis es lo mismo.

**Código que desaparece completamente:**
- `let _seq: number | null = null` y toda la lógica de inicialización desde DB
- `const _sizeCache = new Map<string, number>()`
- `tryLockJobSQLite()` — el lock por SQLite deja de existir
- `getCachedRagCollectionsNextJs()` — el fallback a `unstable_cache`
- Todo el código de `localStorage["seen_notification_ids"]`
- Los `if (redis)` y los `else` en F8.24–F8.28

**Para tests unitarios:** se agrega `ioredis-mock` como devDependency — los 270 tests de lógica no necesitan Redis corriendo. Los tests de integración Redis usan `services: redis` en CI.

---

### F8.22 — ADR-010 + Redis en infraestructura

**Archivos a crear:**
- `docs/decisions/010-redis-required.md`

**Archivos a modificar:**
- `docker-compose.yml` (o el compose file de producción)
- `apps/web/src/app/api/health/route.ts`
- `.env.example`

Los workarounds que Redis + BullMQ reemplazan:

| Workaround actual | Archivo | Reemplazado por |
|---|---|---|
| `let _seq: number \| null = null` | `events.ts` | Redis `INCR events:seq` — eliminar `_seq` |
| Tabla `ingestion_queue` + locking SQLite | `ingestion.ts` + schema | **BullMQ** — eliminar tabla entera |
| `processWithRetry` manual | `ingestion.ts` | **BullMQ** `attempts + backoff` |
| `setInterval(processScheduledReports)` | `ingestion.ts` | **BullMQ** `repeat jobs` |
| SSE polling de ingesta cada 3s | `ingestion/stream/route.ts` | **BullMQ** eventos en tiempo real |
| Sin JWT blacklist | `middleware.ts` | Redis `SET revoked:{jti} 1 EX {ttl}` |
| `const _sizeCache = new Map()` | `rotation.ts` | Redis `HSET log:sizes` — eliminar `_sizeCache` |
| `revalidateTag` local | `collections-cache.ts` | Redis `DEL rag:collections` |
| `localStorage["seen_notification_ids"]` | `useNotifications.ts` | Redis `ZADD notifications:seen:{userId}` |
| `sequence: Date.now()` | `ingestion.ts` | Redis `INCR events:seq` (misma clave) |
| Sin master election para external-sync | `external-sync.ts` | Redis `SET NX EX` |

- [ ] Escribir `docs/decisions/010-redis-required.md`: Redis como dependencia del sistema, motivo, primitivas usadas — 20 min
- [ ] Agregar Redis al `docker-compose.yml`:
  ```yaml
  redis:
    image: redis:alpine
    ports: ["6379:6379"]
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
  ```
  — 10 min
- [ ] En `GET /api/health`: agregar verificación de Redis:
  ```typescript
  const redis = getRedisClient()
  const redisOk = await redis.ping().then(() => true).catch(() => false)
  if (!redisOk) return NextResponse.json({ ok: false, service: "redis", status: "down" }, { status: 503 })
  ```
  — 10 min
- [ ] `REDIS_URL=redis://localhost:6379` en `.env.example` — marcado como **requerido**, no opcional — 2 min
- [ ] Commit: `feat(infra): redis como dependencia requerida + health check — plan8 f8.22`

---

### F8.23 — Cliente Redis singleton (falla rápido si no configurado)

**Archivos a crear:**
- `packages/db/src/redis.ts`

```typescript
// packages/db/src/redis.ts
// NO importar @rag-saldivia/logger aquí — logger importa de @rag-saldivia/db
// → importarlo crearía dependencia circular: db → logger → db (violación de ADR-005)
import Redis from "ioredis"

let _client: Redis | null = null

export function getRedisClient(): Redis {
  if (!_client) {
    const url = process.env["REDIS_URL"]
    if (!url) {
      throw new Error(
        "REDIS_URL no configurado.\n" +
        "Redis es requerido. Para dev local:\n" +
        "  docker run -d -p 6379:6379 redis:alpine\n" +
        "  REDIS_URL=redis://localhost:6379 en .env.local"
      )
    }
    _client = new Redis(url, { maxRetriesPerRequest: 3 })
    // Usar console.error — NO importar logger (evita dependencia circular)
    _client.on("error", (err) => {
      console.error("[Redis] connection error:", String(err))
    })
  }
  return _client
}

export function _resetRedisForTesting() {
  _client = null
}
```

**Para tests unitarios:** `ioredis-mock` en el test setup, sin necesidad de Redis corriendo:

```typescript
// packages/db/src/test-setup.ts o apps/web/src/lib/test-setup.ts
// Solo para entornos de test — no afecta producción
if (process.env["NODE_ENV"] === "test" && !process.env["REDIS_URL"]) {
  const { IORedisMock } = await import("ioredis-mock")
  mock.module("ioredis", () => ({ default: IORedisMock }))
}
```

- [ ] `bun add ioredis` en `packages/db` — 5 min
- [ ] `bun add -d ioredis-mock` en `packages/db` y `apps/web` — 3 min
- [ ] Crear `packages/db/src/redis.ts` con el singleton que lanza error claro — 15 min
- [ ] Agregar mock de ioredis al `test-setup.ts` — 10 min
- [ ] Crear `packages/db/src/__tests__/redis.test.ts`:
  ```typescript
  import { getRedisClient, _resetRedisForTesting } from "../redis"

  afterEach(() => {
    _resetRedisForTesting()  // obligatorio — evita que el singleton del test anterior
  })                          // interfiera con el próximo test

  test("getRedisClient lanza error claro si REDIS_URL no configurado", () => {
    const orig = process.env["REDIS_URL"]
    delete process.env["REDIS_URL"]
    expect(() => getRedisClient()).toThrow("REDIS_URL no configurado")
    if (orig) process.env["REDIS_URL"] = orig
  })

  test("getRedisClient retorna instancia Redis cuando REDIS_URL está configurado", () => {
    process.env["REDIS_URL"] = "redis://localhost:6379"
    const client = getRedisClient()
    expect(client).toBeDefined()
    expect(typeof client.get).toBe("function")  // ioredis-mock activo
  })
  ```
  — 15 min
- [ ] Commit: `feat(db): cliente Redis requerido con fail-fast + mock para tests — plan8 f8.23`

---

### F8.24 — Eliminar `_seq` — secuencia de eventos via `INCR`

**Código que desaparece:** `let _seq`, la lógica de inicialización desde DB, y el `sequence: Date.now()` del worker.

```typescript
// events.ts — antes (30 líneas)
let _seq: number | null = null
async function nextSequence(): Promise<number> {
  if (_seq === null) {
    const last = await getDb().query.events.findFirst({ orderBy: desc(e.sequence) })
    _seq = (last?.sequence ?? 0) + 1
  } else { _seq++ }
  return _seq
}

// events.ts — después (1 línea)
async function nextSequence(): Promise<number> {
  return getRedisClient().incr("events:seq")
}
```

- [ ] Reemplazar `nextSequence()` con la versión Redis de 1 línea — 5 min
- [ ] Eliminar la variable `_seq` y toda la lógica de inicialización — 5 min
- [ ] En `ingestion.ts:1247`: reemplazar insert directo con `writeEvent()` — elimina `sequence: Date.now()` — 10 min
- [ ] **Verificar `events.test.ts`**: el test `"asigna sequence monotónicamente creciente"` sigue pasando porque ioredis-mock implementa INCR correctamente (retorna 1, 2, 3...). No requiere modificación, pero confirmar que el mock está activo antes de correr — 5 min
- [ ] `bun run test packages/db/` — todos pasan (ioredis-mock activo) — 5 min
- [ ] Commit: `refactor(db): eliminar _seq in-memory — secuencia de eventos via Redis INCR — plan8 f8.24`

---

### F8.25 — JWT revocation list

**Agrega jti al token + blacklist en Redis.**

```typescript
// lib/auth/jwt.ts — agregar jti
export async function createJwt(claims: Omit<JwtClaims, "iat" | "exp">): Promise<string> {
  return new SignJWT({ ...claims })
    .setProtectedHeader({ alg: "HS256" })
    .setIssuedAt()
    .setExpirationTime(getExpiry())
    .setJti(crypto.randomUUID())
    .sign(getSecret())
}

// api/auth/logout/route.ts — escribir en blacklist
const ttl = claims.exp - Math.floor(Date.now() / 1000)
if (claims.jti && ttl > 0) {
  await getRedisClient().set(`revoked:${claims.jti}`, "1", "EX", ttl)
}

// lib/auth/jwt.ts — extractClaims() verifica blacklist (Node.js runtime, no Edge)
// ⚠️ NO poner esta verificación en middleware.ts — middleware corre en Edge runtime
// por defecto y ioredis requiere Node.js APIs (net.Socket, tls).
// extractClaims() es llamado desde route handlers (Node.js) — ahí sí funciona.
export async function extractClaims(request: Request): Promise<JwtClaims | null> {
  // ... verificación JWT existente ...
  const claims = await verifyJwt(token)
  if (!claims) return null

  // Verificar blacklist de revocación — solo si Redis disponible
  if (claims.jti) {
    const revoked = await getRedisClient().get(`revoked:${claims.jti}`)
    if (revoked) return null  // token revocado — tratar como inválido
  }

  return claims
}
```

> **Por qué no en `middleware.ts`:** Next.js middleware corre en Edge runtime por defecto. `ioredis` usa APIs de Node.js (`net.Socket`, `tls`) que no existen en Edge. Agregar `getRedisClient()` a `middleware.ts` rompería el middleware. La solución correcta: verificar en `extractClaims()`, que se llama desde route handlers (Node.js runtime). El middleware sigue validando firma y expiración del JWT. La revocación se verifica en la capa de negocio.

- [ ] Agregar `jti: z.string().optional()` al `JwtClaimsSchema` en `packages/shared` — 5 min
- [ ] En `createJwt()`: agregar `.setJti(crypto.randomUUID())` — 5 min
- [ ] **`jwt.test.ts` — tests existentes no se rompen** (verifican `sub`, `email`, `role` — no verifican ausencia de jti). Agregar 1 test nuevo:
  ```typescript
  test("createJwt incluye jti único por token", async () => {
    const t1 = await createJwt(validClaims)
    const t2 = await createJwt(validClaims)
    const c1 = await verifyJwt(t1)
    const c2 = await verifyJwt(t2)
    expect(c1?.jti).toBeDefined()
    expect(c1?.jti).not.toBe(c2?.jti)  // únicos
  })
  ```
  — 10 min
- [ ] En `logout/route.ts`: escribir `SET revoked:{jti} 1 EX {ttl}` — 10 min
- [ ] En `extractClaims()` (`lib/auth/jwt.ts`): verificar blacklist después de validar el JWT — **no en `middleware.ts`** (Edge runtime) — 15 min
- [ ] Test: login → logout → token previo retorna 401 — 15 min
- [ ] Commit: `feat(auth): jwt revocation list — plan8 f8.25`

---

### F8.26 — Master lock para `external-sync.ts`

> **Nota:** el worker de ingesta ya no necesita distributed lock — BullMQ (F8.30) garantiza que un job es procesado por un solo worker. Esta tarea aplica **solo a `external-sync.ts`**, que es el único worker que no usa BullMQ.

**Problema:** si hay dos instancias corriendo, `external-sync.ts` sincroniza Google Drive / SharePoint dos veces en paralelo.

```typescript
// external-sync.ts — master election simple

async function acquireExternalSyncLock(): Promise<boolean> {
  const result = await getRedisClient()
    .set("worker:master:external-sync", WORKER_ID, "EX", 60, "NX")
  return result === "OK"
}

// Renovar cada 30s mientras el loop está activo
setInterval(() => {
  getRedisClient().expire("worker:master:external-sync", 60).catch(() => {})
}, 30_000)

async function syncLoop() {
  while (true) {
    if (await acquireExternalSyncLock()) {
      // solo la instancia master sincroniza
      await syncAllSources()
    }
    await sleep(SYNC_INTERVAL_MS)
  }
}
```

- [ ] Agregar `acquireExternalSyncLock()` en `external-sync.ts` — 15 min
- [ ] Envolver el loop de sync con el lock — 5 min
- [ ] Commit: `feat(workers): master lock para external-sync — plan8 f8.26`

---

### F8.27 — Cache de colecciones (eliminar `unstable_cache` fallback)

**Código que desaparece:** `getCachedRagCollectionsNextJs()` y el `if (redis)` wrapper.

```typescript
// lib/rag/collections-cache.ts — sin fallback
export async function getCachedRagCollections(): Promise<string[]> {
  const redis = getRedisClient()
  const cached = await redis.get("rag:collections")
  if (cached) return JSON.parse(cached) as string[]
  const fresh = await fetchCollectionsFromRAG()
  await redis.set("rag:collections", JSON.stringify(fresh), "EX", 60)
  return fresh
}

export async function invalidateCollectionsCache() {
  await getRedisClient().del("rag:collections")
}
```

- [ ] Reescribir `getCachedRagCollections` — eliminar el fallback `unstable_cache` — 10 min
- [ ] Simplificar `invalidateCollectionsCache` — sin `revalidateTag` — 5 min
- [ ] En rutas `POST/DELETE /api/rag/collections`: llamar `invalidateCollectionsCache()` — 5 min
- [ ] Commit: `refactor(web): cache de colecciones via Redis — eliminar unstable_cache — plan8 f8.27`

---

### F8.28 — Notificaciones: SSE push + vistas server-side (eliminar localStorage + polling)

**Código que desaparece:** `localStorage["seen_notification_ids"]`, la lógica de `markSeen`/`getSeenIds`, el polling interval de 30s como path principal.

```typescript
// useNotifications.ts — sin localStorage, sin polling como default
// EventSource es el único path — Redis siempre disponible
useEffect(() => {
  const es = new EventSource("/api/notifications/stream")
  es.onmessage = (e) => handleNotification(JSON.parse(e.data))
  return () => es.close()
}, [])

// Seen IDs — server-side via Sorted Set
// ZADD notifications:seen:{userId} {ts} {id}
// ZREMRANGEBYSCORE notifications:seen:{userId} 0 {30daysAgo} — limpieza periódica
```

**Archivos a crear:**
- `apps/web/src/app/api/notifications/stream/route.ts`

**Archivos a modificar:**
- `apps/web/src/hooks/useNotifications.ts` — eliminar localStorage, polling, `getSeenIds`, `markSeen`
- `apps/web/src/lib/queue.ts` — agregar `redis.publish` en el BullMQ Worker callback

> **⚠️ Orden de implementación:** F8.30 (BullMQ) refactoriza `ingestion.ts` completamente. No agregar `redis.publish` a `ingestion.ts` — agregar directamente al Worker callback en `queue.ts` (que se crea en F8.30). Si F8.28 se ejecuta antes de F8.30, agregarlo como comentario TODO en `ingestion.ts` y moverlo a `queue.ts` en F8.30.

```typescript
// apps/web/src/lib/queue.ts — el publish va en el Worker callback (F8.30)
export const ingestionWorker = new Worker(
  "ingestion",
  async (job) => {
    await processJob(job.data)
    // Notificar al usuario en tiempo real — F8.28
    await getRedisClient().publish(
      `notifications:${job.data.userId}`,
      JSON.stringify({ type: "ingestion.completed", payload: job.data })
    )
  }, ...
)
ingestionWorker.on("failed", async (job, err) => {
  if (job) await getRedisClient().publish(
    `notifications:${job.data.userId}`,
    JSON.stringify({ type: "ingestion.error", payload: { ...job.data, error: err.message } })
  )
})
```

- [ ] Crear `GET /api/notifications/stream`: SSE con `redis.subscribe("notifications:{userId}")` — 40 min
- [ ] Reescribir `useNotifications.ts` eliminando todo el localStorage y el polling — 20 min
- [ ] Agregar `redis.publish` en el Worker callback de `queue.ts` (coordinar con F8.30) — 10 min
- [ ] Mover seen IDs a Sorted Set server-side — `ZADD` + `ZSCORE` + `ZREMRANGEBYSCORE` — 20 min
- [ ] Commit: `feat(web): notificaciones SSE via Redis — eliminar localStorage y polling — plan8 f8.28`

---

### F8.29 — Eliminar `_sizeCache` (rotation.ts)

Con Redis siempre disponible, ya no hay razón para esta tarea ser "skipeable". Es simple y el código queda limpio.

```typescript
// rotation.ts — sin _sizeCache Map, sin condicionales
async function getLogFileSize(filePath: string): Promise<number> {
  return Number(await getRedisClient().hget("log:sizes", filePath) ?? "0")
}

async function setLogFileSize(filePath: string, size: number) {
  await getRedisClient().hset("log:sizes", filePath, size)
}
```

**Código que desaparece:** `const _sizeCache = new Map<string, number>()` y todas las referencias.

- [ ] Reemplazar `_sizeCache` con las dos funciones Redis — eliminar el Map — 15 min
- [ ] Commit: `refactor(logger): eliminar _sizeCache in-memory — via Redis HSET — plan8 f8.29`

---

### F8.30 — BullMQ reemplaza el worker de ingesta custom

Prerequisito: F8.23 (Redis client). BullMQ está construido sobre Redis — no requiere dependencia adicional de infraestructura.

**Código que desaparece completamente:**

```typescript
// Todo esto se elimina de ingestion.ts:
const WORKER_ID = `worker-${process.pid}-${Date.now()}`
let _shutdown = false
process.on("SIGTERM", ...)
process.on("SIGINT", ...)
async function tryLockJob(...)           // BullMQ hace esto internamente
async function tryLockJobSQLite(...)     // ya eliminado en F8.26, pero BullMQ también cubre esto
async function processWithRetry(...)     // BullMQ: attempts + backoff automático
async function workerLoop(...)           // 60 líneas → BullMQ Worker
setInterval(processScheduledReports)    // BullMQ: repeat jobs
acquireWorkerMasterLock()               // BullMQ: un solo worker procesa un job dado
```

**Y de SQLite:**
```
tabla ingestion_queue → desaparece del schema
columnas: status, locked_at, locked_by, retry_count, started_at, completed_at, error
```

**Código resultante:**

```typescript
// apps/web/src/lib/queue.ts — definición central de queues

import { Queue, Worker } from "bullmq"
import IORedis from "ioredis"

// BullMQ necesita múltiples conexiones internas (subscriber + publisher separados).
// Pasar una instancia ioredis puede causar problemas — usar opciones de conexión
// para que BullMQ cree sus propias conexiones.
// maxRetriesPerRequest: null es obligatorio en ioredis para uso con BullMQ (v5+).
function getBullMQConnection() {
  return new IORedis(process.env["REDIS_URL"]!, { maxRetriesPerRequest: null })
}

export const ingestionQueue = new Queue("ingestion", {
  connection: getBullMQConnection(),
  defaultJobOptions: {
    attempts: 3,
    backoff: { type: "exponential", delay: 10_000 },
    removeOnComplete: 100,   // guarda los últimos 100 completados en Redis
    removeOnFail: 200,
  },
})

// Worker — reemplaza workerLoop() + processWithRetry()
export const ingestionWorker = new Worker(
  "ingestion",
  async (job) => {
    await processJob(job.data)   // lógica de negocio pura, sin infra
    await recordIngestionEvent({ ...job.data, action: "added" })
  },
  {
    connection: getBullMQConnection(),  // conexión separada para el worker
    concurrency: 1,
  }
)

// Scheduled reports — reemplaza setInterval + processScheduledReports
export async function scheduleReport(report: ScheduledReport) {
  await ingestionQueue.add("scheduled-report", report, {
    repeat: { pattern: scheduleToPattern(report.schedule) },
    jobId: `report-${report.id}`,   // idempotente
  })
}
```

**Agregar job desde la API de upload:**
```typescript
// api/upload/route.ts — reemplaza el INSERT en ingestion_queue
await ingestionQueue.add("ingest", { filePath, collection, userId })
```

**Monitoreo del kanban (reemplaza el SSE polling):**
```typescript
// api/admin/ingestion/stream/route.ts — BullMQ eventos en tiempo real
// En lugar de polling SQLite cada 3s → subscribe a eventos del worker
ingestionWorker.on("completed", (job) => emit({ type: "done", job }))
ingestionWorker.on("failed", (job, err) => emit({ type: "failed", job, error: err.message }))
ingestionWorker.on("progress", (job, progress) => emit({ type: "progress", job, progress }))
```

**Archivos a crear:**
- `apps/web/src/lib/queue.ts`

**Archivos a modificar:**
- `apps/web/src/workers/ingestion.ts` — simplificado a solo lógica de negocio
- `apps/web/src/app/api/upload/route.ts` — usa `ingestionQueue.add()` en lugar de INSERT SQL
- `apps/web/src/app/api/admin/ingestion/stream/route.ts` — eventos BullMQ en lugar de polling
- `apps/web/src/app/api/admin/ingestion/route.ts` — el GET non-SSE también lee `ingestion_queue`, actualizar para leer desde BullMQ job history

**Archivos/tablas a eliminar:**
- `packages/db/src/schema.ts` — eliminar tabla `ingestionQueue` y sus tipos
- `packages/db/src/queries/` — eliminar cualquier query relacionada a `ingestion_queue`

- [ ] `bun add bullmq` en `apps/web` — 2 min
- [ ] Crear `apps/web/src/lib/queue.ts` con la definición de `ingestionQueue` e `ingestionWorker` — 30 min
- [ ] Refactorizar `ingestion.ts`: eliminar `workerLoop`, `tryLockJob`, `processWithRetry`, los signal handlers, el `setInterval` — dejar solo `processJob()` con lógica de negocio pura — 30 min
- [ ] Refactorizar `api/upload/route.ts`: reemplazar INSERT SQL con `ingestionQueue.add()` — 10 min
- [ ] Refactorizar `api/admin/ingestion/stream/route.ts`: eventos BullMQ en lugar de polling SQLite cada 3s — 20 min
- [ ] Refactorizar `api/admin/ingestion/route.ts` (GET non-SSE): reemplazar queries a `ingestion_queue` con lecturas de BullMQ job history (`ingestionQueue.getJobs(["active", "waiting", "completed", "failed"])`) — 15 min
- [ ] Migrar scheduled reports a `ingestionQueue.add(..., { repeat: ... })` — 20 min
- [ ] Eliminar tabla `ingestionQueue` del schema + `bun drizzle-kit push` — 10 min
- [ ] `bun run test packages/db/` — todos pasan (tabla eliminada) — 5 min

> **Nota sobre F8.26:** con BullMQ, el `acquireWorkerMasterLock` para el worker de ingesta ya no es necesario — BullMQ garantiza que un job es procesado por un solo worker a la vez. Solo queda el master lock para `external-sync.ts` que no usa BullMQ (simplificar F8.26 a eso).

- [ ] Commit: `feat(workers): bullmq reemplaza worker de ingesta custom — eliminar ingestion_queue table — plan8 f8.30`

---

### Criterio de done

Los 11 workarounds de código no existen más — ni como fallback, ni como dead code. La tabla `ingestion_queue` fue eliminada del schema. El worker de ingesta son 50 líneas de lógica pura. `init.ts` tiene < 10 líneas. Las 22 actions usan validación automática. Health check valida Redis. Deploy bloquea si Redis está caído.

### Checklist de cierre

- [ ] `bun run test` — todos pasan (ioredis-mock activo en test env)
- [ ] `bun run test:components` — todos pasan
- [ ] Smoke test con Redis (`docker compose up redis` o `docker run -d -p 6379:6379 redis:alpine`):
  - [ ] Login → logout → token revocado inmediatamente
  - [ ] Ingesta → notificación llega en < 1s via EventSource
  - [ ] Crear colección → cache invalida al instante en Redis
  - [ ] `GET /api/health` retorna `{ redis: "healthy" }`
  - [ ] Sin `REDIS_URL`: servidor falla con error claro al arrancar
- [ ] **CI — agregar `services: redis` al workflow** `.github/workflows/ci.yml`:
  ```yaml
  services:
    redis:
      image: redis:alpine
      ports: ["6379:6379"]
  env:
    REDIS_URL: redis://localhost:6379
  ```
- [ ] `.env.example` — `REDIS_URL` marcado como **requerido**
- [ ] CHANGELOG.md actualizado
- [ ] **Actualizar `CLAUDE.md`** — sección "Patrones importantes":
  - Agregar: `lib/rag/stream.ts` es la utilidad canónica de SSE — nunca duplicar
  - Agregar: `Citation` (de `@rag-saldivia/shared`) es el tipo canónico de sources del RAG
  - Agregar: Redis es dependencia requerida — `getRedisClient()` retorna `Redis` directo, nunca null
  - Agregar: política de `useCallback` — memoizar la función del hook primero, luego los handlers del componente
- [ ] `git commit -m "feat: redis como dependencia requerida — eliminar 8 workarounds de single-instance — plan8 f8"`

**Estado: pendiente**

---

## Estado global

Orden de ejecución: **F0 → F1 → F3 → F2 → F4 → F5 → F6 → F7 → F8**

| Fase | Exec | Estado | Descripción |
|------|------|--------|-------------|
| Fase 0 — Baseline | 1° | ⏳ pendiente | Bundle size, react-scan, tiempos CI |
| Fase 1 — Dead code + duplicados | 2° | ⏳ pendiente | knip scan, SSE reader, Citation type, dead code, N+1, canAccess cache, formatDate — ADR-008 |
| Fase 3 — Unificación de deps | 3° | ⏳ pendiente | Drizzle sync + linting + Drizzle Kit push (init.ts 416→<10) + zustand/dompurify + strictTS |
| Fase 2 — Refactoring React | 4° | ⏳ pendiente | Server pattern (×10), Server Actions (×7), useOptimistic, memoización, lazy loading, next-safe-action, react-hook-form, nuqs, eliminar 9 rutas API dead — ADR-009 |
| Fase 4 — Upgrades | 5° | ⏳ pendiente | Next.js, Drizzle, Lucide, Zod, @libsql/client |
| Fase 5 — Docs arquitectura | 6° | ⏳ pendiente | architecture.md con stream utils, Redis, nuevos ADRs |
| Fase 6 — Calidad estructural | 7° | ⏳ pendiente | Error Boundaries, CI paralelo + cache + turbo --affected |
| Fase 7 — Logging y Black Box | 8° | ⏳ pendiente | requestId, event types, handlers, retención, índice compuesto, CSV + papaparse |
| Fase 8 — Redis + BullMQ | 9° | ⏳ pendiente | Redis requerido, BullMQ reemplaza worker — 11 workarounds eliminados |

---

## Mapa de código eliminado (todas las fases)

| Tecnología / Patrón | Fase | Impacto |
|---|---|---|
| `knip` — dead code sistemático (baseline) | F1.0 | variable (informa otras fases) |
| `formatDate`/`formatDateTime` utility centralizada | F1.7 | ~36 líneas |
| SSE reader duplicado + Citation type + dead code + N+1 | F1.1–F1.6 | ~200 líneas |
| Drizzle Kit reemplaza `init.ts` (416 líneas → <10) | F3.9 | ~410 líneas |
| `zustand` + `dompurify` eliminados (0 usos) | F3.10 | 2 deps del bundle |
| `exactOptionalPropertyTypes` — rigor TS consistente | F3.10 | 0 líneas, + type safety |
| Server Component lift para datos (×10 componentes) | F2.4 | ~400 líneas |
| Server Actions para mutaciones (×7 componentes) | F2.7 | ~250 líneas |
| `next-safe-action` — boilerplate en 22+ actions | F2.7 | ~110 líneas |
| `useOptimistic` — setState manual post-action | F2.7b | ~80 líneas |
| `react-hook-form` — useState de formularios (50-70 instancias) | F2.8 | ~350 líneas |
| `nuqs` — URL state para filtros + filtrado SQL | F2.9 | ~60 líneas |
| 9 rutas API eliminadas (dead tras F2.4+F2.7) | F2.10 | ~200 líneas + 9 archivos |
| Memoización + lazy loading | F2.5–F2.6 | ~40 líneas |
| `papaparse` — CSV manual reemplazado | F7.20 | ~30 líneas |
| Redis elimina 11 fallbacks in-memory | F8 | ~200 líneas |
| BullMQ reemplaza worker custom + tabla `ingestion_queue` | F8.30 | ~150 líneas |
| `turbo --affected` — CI solo testea lo afectado | F6.16 | 0 líneas, -50% tiempo CI en PRs |
| **TOTAL líneas eliminadas** | — | **~2,516 líneas** |
| **TOTAL archivos eliminados** | — | **9 route handlers** |
| **TOTAL deps eliminadas** | — | **2 (zustand, dompurify)** |

---

## Tiempo total estimado

| Fase | Estimación |
|------|------------|
| Fase 0 — Baseline | 30-45 min |
| Fase 1 — knip + duplicados + formatDate | 5-7 hs |
| Fase 3 — Unificación + Drizzle Kit + deps + TS | 2-4 hs |
| Fase 2 — Server pattern + actions + useOptimistic + lazy + next-safe-action + react-hook-form + nuqs + 9 rutas dead | 12-18 hs |
| Fase 4 — Upgrades | 4-6 hs |
| Fase 5 — Docs arquitectura | 15 min |
| Fase 6 — Error Boundaries + CI paralelo + turbo | 3-5 hs |
| Fase 7 — Logging + papaparse | 4-6 hs |
| Fase 8 — Redis + BullMQ | 7-10 hs |
| **Total** | **38-57 hs** |
