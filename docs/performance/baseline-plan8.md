# Baseline de performance — Pre Plan 8

> Snapshot tomado el 2026-03-27 antes de aplicar cualquier cambio del Plan 8.
> Estas métricas son el punto de comparación para todas las fases de optimización.

---

## F0.1 — Bundle size baseline

**Build:** `bun run build` en `apps/web`
**Entorno:** Next.js 15.5.14, Bun 1.3.11, producción
**Tiempo de build:** 16.4 s (compile: 6.6 s)

### Rutas de UI — tamaños comparables

| Ruta | Page size | First Load JS |
|---|---|---|
| `/chat` | 2.01 kB | **120 kB** |
| `/chat/[id]` | 30.2 kB | **171 kB** |
| `/collections` | 5.04 kB | 116 kB |
| `/collections/[name]/graph` | 4.05 kB | **119 kB** |
| `/extract` | 5.63 kB | 117 kB |
| `/login` | 8.87 kB | 120 kB |
| `/audit` | 2.41 kB | 114 kB |
| `/admin/analytics` | 128 kB | **239 kB** |
| `/admin/users` | 5.52 kB | 117 kB |
| `/settings` | 3.71 kB | 115 kB |
| `/settings/memory` | 4.6 kB | 116 kB |
| `/upload` | 2.57 kB | 105 kB |

### Bundle compartido (First Load JS shared by all)

| Chunk | Tamaño |
|---|---|
| `chunks/380029e9-fed582fa60b790f9.js` | 54.2 kB |
| `chunks/7106-de56942f64025513.js` | 46 kB |
| other shared chunks | 2.64 kB |
| **Total shared** | **103 kB** |

**Middleware:** 39.9 kB

### Notas sobre librerías pesadas identificadas

- `d3` (~450 KB minificado) — se usa únicamente en `/collections/[name]/graph` pero puede estar entrando en el bundle compartido o en `/admin/analytics`. La ruta `/admin/analytics` tiene 239 kB First Load, significativamente mayor que el resto.
- `react-pdf` (~600 KB) — verificar si está en el bundle de `/chat/[id]` (171 kB, el mayor entre rutas de chat).
- `recharts` — usado en `/admin/analytics`.

**Observación:** `/chat/[id]` es 51 kB más pesado que el shared bundle base — esto incluye `ChatInterface` y sus dependencias directas.

---

## F0.2 — React render baseline (análisis estático + react-scan)

### Método

Análisis estático de `ChatInterface.tsx` (410 líneas). El plan requiere correr react-scan en dev con keystroke activo para medir renders reales; este baseline documenta lo que el análisis de código confirma.

### Handlers sin `useCallback` en `ChatInterface`

Los siguientes 5 handlers se recrean en cada render del componente:

| Línea | Handler | Acción |
|---|---|---|
| 91 | `handleSend()` | Envía mensaje al RAG via streaming |
| 149 | `handleRegenerate()` | Regenera la última respuesta |
| 156 | `handleCopy(messageId, content)` | Copia al clipboard |
| 162 | `handleToggleSaved(messageId, content)` | Guarda/quita de guardados |
| 173 | `handleFeedback(messageId, rating)` | Envía feedback up/down |

### Componentes que re-renderizan con cada keystroke (estimado)

Cadena de renders esperada por cada tecla en el input del chat:
1. `ChatInterface` — state `query` cambia → re-render completo
2. Todos los child components que no están memoizados y reciben props de `ChatInterface`
3. `SourcesPanel`, `ArtifactsPanel` — reciben props derivadas → re-render innecesario

**Sin ningún `useCallback` ni `useMemo` en `ChatInterface`**, cada keystroke recrea los 5 handlers y potencialmente causa re-renders en cadena en los componentes hijos que los reciben como props.

### Estado actual de memoización

- `useCallback`: **0 usos** en `ChatInterface`
- `useMemo`: **0 usos** en `ChatInterface`
- `React.memo`: no aplicado a los hijos directos de `ChatInterface`

---

## F0.3 — Métricas de CI

### Tests locales (`bun run test`)

Ejecutado con Turborepo (ejecución paralela):

| Package | Tests | Archivos | Tiempo interno |
|---|---|---|---|
| `@rag-saldivia/shared` | 6 pass | 1 archivo | 19 ms |
| `@rag-saldivia/config` | 14 pass | 1 archivo | 86 ms |
| `@rag-saldivia/logger` | 24 pass | 1 archivo | 138 ms |
| `@rag-saldivia/db` | 161 pass | 17 archivos | 1383 ms |
| `@rag-saldivia/web` (lib) | 68 pass | 6 archivos | 1750 ms |
| **Total** | **273 tests** | **26 archivos** | — |

**Tiempo real wall-clock:** `0m1.853s`
**Turborepo time (paralelo):** `1.819s`
**Resultado:** 273 pass, 0 fail

### Tests de componentes (`bun run test:components`)

> Pendiente de medir en ejecución separada. Suite actual: 147 tests (20 componentes).

### Estructura del CI (`.github/workflows/ci.yml`)

Jobs actuales en CI (todos secuenciales, sin paralelización explícita entre jobs del mismo workflow):

| Job | Trigger | Tiempo estimado |
|---|---|---|
| `commitlint` | PRs | ~1-2 min |
| `changelog` | PRs | ~30 s |
| `type-check` | PRs + push dev | ~2-3 min |
| `test` | PRs + push dev | ~1-2 min |
| `coverage` | PRs | ~2-3 min |
| `lint` | PRs + push dev | ~2-3 min |
| `component-tests` | PRs + push dev | ~3-5 min |
| `visual-regression` | PRs + push dev | ~8-12 min (build Storybook + Playwright) |
| `accessibility` | PRs + push dev | ~5-8 min (dev server + Playwright) |

**Problema identificado:** Los jobs `test`, `coverage`, `component-tests`, `visual-regression` y `accessibility` podrían paralelizarse (actualmente corren secuencialmente dentro del mismo workflow, aunque GitHub Actions los ejecuta en runners separados por ser jobs distintos). Sin embargo, el job `test` y `coverage` realizan el mismo setup (`bun install` + `bun run test`) por separado — duplicación de trabajo.

---

## Versiones de dependencias (pre-upgrade)

| Dependencia | Version en `apps/web` | Version en `packages/db` | Última disponible |
|---|---|---|---|
| `next` | `^15.0.0` (instalado: 15.5.14) | — | 15.5.14 ✓ |
| `react` | `^19.2.4` | — | ~19.x |
| `drizzle-orm` | `^0.38.4` | `^0.38.0` | ~0.45.x |
| `zod` | `^3.25.0` | — | ~3.25.x / Zod 4 en RC |
| `@libsql/client` | — | `^0.14.0` | ~0.14.x |
| `lucide-react` | `^0.475.0` | — | ~0.475.x |
| `typescript` | `^6.0.2` | `^6.0.0` | 6.x |
| `@next/bundle-analyzer` | `^16.2.1` (recién instalado) | — | 16.2.1 ✓ |

**Desincronización Drizzle:** `apps/web` usa `^0.38.4` y `packages/db` usa `^0.38.0`. Ambas resuelven a la misma versión de patch pero el constraint es diferente — se unificará en Fase 3.

---

## Resumen ejecutivo

| Métrica | Valor baseline |
|---|---|
| Bundle `/chat` First Load JS | 120 kB |
| Bundle `/chat/[id]` First Load JS | **171 kB** |
| Bundle `/collections/[name]/graph` First Load JS | 119 kB |
| Bundle compartido | 103 kB |
| Ruta más pesada (`/admin/analytics`) | **239 kB** |
| Handlers sin `useCallback` en `ChatInterface` | **5** |
| Componentes sin memoización | `ChatInterface` + hijos |
| Tests totales | **273 pass, 0 fail** |
| Tiempo `bun run test` (wall-clock) | **1.853 s** |
| Tiempo de build | **16.4 s** |

---

*Actualizar este documento al completar cada fase relevante con los valores post-optimización.*
