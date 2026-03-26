---
name: rag-testing
description: Write and run tests for the RAG Saldivia TypeScript monorepo using Bun test. Use when writing new tests, adding test coverage, running the test suite, or when the user says "escribir un test", "agregar tests", "correr los tests", or mentions a specific package to test.
---

# RAG Saldivia — Testing

Reference: `docs/workflows.md` (sección Testing), `docs/decisions/006-testing-strategy.md`.

## Regla de oro

> **Si el código no tiene test, no está terminado.**

Esto no es una sugerencia. Es un criterio de done que el CI enforcea con `coverageThreshold = 0.95`.

## Comandos

```bash
# Suite completa via Turborepo
bun run test

# Suite completa con cobertura (lo que corre en CI en PRs)
bun run test:coverage

# Por paquete — más rápido durante desarrollo
bun test packages/db/src/__tests__/             # queries de DB
bun test packages/logger/src/__tests__/         # logger + blackbox
bun test packages/config/src/__tests__/         # config loader
bun test packages/shared/src/__tests__/         # schemas Zod
bun test apps/web/src/lib/auth/__tests__/       # auth + RBAC
bun test apps/web/src/lib/__tests__/            # utilidades web
bun test apps/web/src/lib/rag/__tests__/        # cliente RAG

# Con cobertura por paquete
bun test packages/db/src/__tests__/ --coverage
```

## Metas de cobertura por capa

| Capa | Target | Enforced en CI |
|------|--------|----------------|
| `packages/*` | **95%** | ✅ sí |
| `apps/web/src/lib/` | **95%** | ✅ sí |
| `apps/web/src/hooks/` | **80%** | ✅ sí |
| API routes | cobertura funcional (test manual documentado) | revisión humana |
| React components | no requerido ahora | — |

## Matriz "tipo de código → test requerido"

Usá esta tabla para saber qué testear antes de hacer commit.

| Tipo de código | Test requerido | Dónde | En qué PR |
|----------------|----------------|-------|-----------|
| Query nueva en `packages/db/src/queries/` | Test unitario SQLite en memoria | `packages/db/src/__tests__/[dominio].test.ts` | **mismo PR** |
| Función pura en `apps/web/src/lib/` | Test unitario | `apps/web/src/lib/__tests__/[nombre].test.ts` | **mismo PR** |
| Lógica pura de un hook | Extraer a `lib/` → test allí | `apps/web/src/lib/[dominio]/__tests__/` | **mismo PR** |
| Schema Zod nuevo en `packages/shared` | Test validación (válido + inválido) | `packages/shared/src/__tests__/` | **mismo PR** |
| Server Action nueva | Test de la query subyacente | `packages/db/src/__tests__/` | **mismo PR** |
| API route nueva | Test manual documentado en PR | comentario en PR | **mismo PR** |
| Config loader modificado | Test del caso nuevo | `packages/config/src/__tests__/` | **mismo PR** |

## Dónde agregar tests

| Qué testear | Directorio |
|-------------|-----------|
| Queries de DB | `packages/db/src/__tests__/` |
| Lógica de auth / RBAC | `apps/web/src/lib/auth/__tests__/` |
| Config loader | `packages/config/src/__tests__/` |
| Logger / blackbox | `packages/logger/src/__tests__/` |
| Schemas Zod | `packages/shared/src/__tests__/` |
| Utilidades web (`lib/`) | `apps/web/src/lib/__tests__/` |
| Cliente RAG / detección idioma | `apps/web/src/lib/rag/__tests__/` |
| Hooks (lógica pura extraída) | `apps/web/src/lib/[dominio]/__tests__/` |

## Patrón de test para packages/db

```typescript
import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import * as schema from "../schema"

// CRÍTICO: setear antes de cualquier import que use getDb()
process.env["DATABASE_PATH"] = ":memory:"

const client = createClient({ url: ":memory:" })
const testDb = drizzle(client, { schema })

beforeAll(async () => {
  // Solo las tablas que necesita este archivo
  await client.executeMultiple(`
    CREATE TABLE IF NOT EXISTS users (...);
    CREATE TABLE IF NOT EXISTS mi_tabla (...);
  `)
})

afterEach(async () => {
  // Limpiar entre tests para aislar estado
  await client.executeMultiple(`
    DELETE FROM mi_tabla;
    DELETE FROM users;
  `)
})

describe("nombreDelModulo", () => {
  test("describe el comportamiento esperado", async () => {
    // arrange → act → assert
  })
})
```

## Reglas críticas

**1. DB en memoria siempre**
`createClient({ url: ":memory:" })` en cada test de DB. Nunca tocar el archivo real.

**2. `process.env["DATABASE_PATH"] = ":memory:"` ANTES de los imports**
Las funciones de query llaman `getDb()` internamente. Si `sessions.ts` lo llama a nivel de módulo
(línea 9), el env var tiene que estar seteado antes del import del módulo de test.

**3. Solo imports estáticos al nivel del módulo**
`await import(...)` dentro de callbacks `describe` o `beforeEach` falla silenciosamente en webpack.
Todos los imports al tope del archivo.

**4. Arrange → Act → Assert**
Una sola assertion de comportamiento por test. `beforeEach` para estado limpio.

**5. Tests del logger — verificar por contenido**
No asumir formato JSON ni pretty-print. Verificar que el output *contiene* el tipo de evento.

**6. Mock de fetch para tests de webhook**
```typescript
import { spyOn } from "bun:test"
const mockFetch = spyOn(globalThis, "fetch").mockResolvedValue(
  new Response(null, { status: 200 })
)
```

## Estado de cobertura

| Paquete / módulo | Tests | Estado |
|------------------|-------|--------|
| `packages/db` — users, areas, saved | ✅ ~37 tests | cubierto |
| `packages/logger` | ✅ 24 tests | cubierto |
| `packages/config` | ✅ 14 tests | cubierto |
| `packages/shared` | ✅ 6 tests | cubierto |
| `apps/web` auth + RBAC | ✅ 17 tests | cubierto |
| `apps/web` lib/export, lib/changelog, lib/rag/detect-language | ✅ 27 tests | cubierto |
| `packages/db` — sessions, events, memory, annotations, tags, shares, templates, rate-limits, webhooks, reports, collection-history, projects, search, external-sources | ⏳ Plan 5 F3 | en progreso |
| `apps/web` lib/webhook, lib/rag/detect-artifact | ⏳ Plan 5 F4 | en progreso |
