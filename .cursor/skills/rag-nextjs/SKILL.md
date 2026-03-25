---
name: rag-nextjs
description: Next.js 15 patterns, file placement, and project-specific conventions for RAG Saldivia. Use when creating new pages, API routes, Server Actions, hooks, or components — or when the user asks "¿dónde pongo X?", "¿cómo agrego una página?", "¿Server Component o Client Component?", "cómo cachear", or mentions SSE, streaming, or Next.js patterns.
---

# RAG Saldivia — Next.js 15 Patterns

Reference: `docs/architecture.md` para el diagrama completo de flujos.

## Dónde poner cada cosa

| Qué | Dónde |
|-----|-------|
| Página nueva | `apps/web/src/app/(app)/[ruta]/page.tsx` (Server Component) |
| Mutación del servidor | `apps/web/src/app/actions/[dominio].ts` (Server Action) |
| Endpoint REST / SSE | `apps/web/src/app/api/[ruta]/route.ts` (Route Handler) |
| UI con estado local | `apps/web/src/components/[nombre].tsx` (Client Component) |
| Lógica de fetch/streaming | `apps/web/src/hooks/use[Nombre].ts` |
| Query de DB nueva | `packages/db/src/queries/[dominio].ts` |
| Tipo compartido web+cli | `packages/shared/src/schemas.ts` |
| Comando CLI nuevo | `apps/cli/src/commands/[nombre].ts` + registrar en `index.ts` |

## Reglas de arquitectura

**Server Components por defecto**  
Agregar `"use client"` solo donde sea imprescindible: chat SSE, modales con estado, sliders, inputs controlados.

**`Bun.*` está prohibido en Next.js**  
El runtime de Next.js es Node.js. Usar `fs/promises`, `crypto` de Node. `Bun.file`, `Bun.write`, etc. rompen el build.

**Timestamps siempre con Temporal**  
`Temporal.Now.instant().epochMilliseconds` — no `Date.now()` ni `new Date()`.

**Validar inputs con Zod**  
Usar los schemas de `@rag-saldivia/shared`. No validar manualmente.

**Loggear puntos críticos**  
Todo endpoint de API y Server Action importante emite un evento al backend logger.

## SSE — trampa crítica

El gateway Python anterior siempre retornaba HTTP 200, incluso en errores, y luego mandaba el error dentro del stream.  
**En este stack: verificar el status HTTP ANTES de comenzar a leer el stream.**  
Si el status no es 200, no intentar hacer `getReader()` — leer el body como JSON de error.

## Cache layers

| Capa | Qué cachea | TTL |
|------|-----------|-----|
| `React.cache()` | queries DB en una request | por request |
| `unstable_cache` | colecciones del RAG Server | 60 s — `tags: ['collections']` |
| Full Route Cache | páginas estáticas | build |
| Router Cache | navegación cliente | sesión |

Invalidar colecciones: `revalidateTag('collections')` al crear o eliminar colecciones.

## Feature workflow

```
1. Leer archivos críticos relevantes
2. Server Component por defecto → agregar "use client" solo si es necesario
3. Datos → Server Action o Route Handler, nunca fetch desde Client Component directo
4. Validar con Zod (shared schemas)
5. Loggear con backend logger
6. Tests unitarios
7. Actualizar CHANGELOG.md
8. Commit convencional
```
