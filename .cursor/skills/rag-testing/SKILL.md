---
name: rag-testing
description: Write and run tests for the RAG Saldivia TypeScript monorepo using Bun test. Use when writing new tests, adding test coverage, running the test suite, or when the user says "escribir un test", "agregar tests", "correr los tests", or mentions a specific package to test.
---

# RAG Saldivia — Testing

Reference: `docs/workflows.md` (sección Testing) para el patrón completo de tests.

## Comandos

```bash
bun run test                                    # suite completa via Turborepo
bun test apps/web/src/lib/auth/__tests__/       # auth + RBAC
bun test packages/db/src/__tests__/             # queries de DB
bun test packages/logger/src/__tests__/         # logger + blackbox
bun test packages/config/src/__tests__/         # config loader
```

## Dónde agregar tests

| Qué testear | Directorio |
|-------------|-----------|
| Queries de DB nuevas | `packages/db/src/__tests__/` |
| Lógica de auth / RBAC | `apps/web/src/lib/auth/__tests__/` |
| Config loader | `packages/config/src/__tests__/` |
| Logger / blackbox | `packages/logger/src/__tests__/` |
| Hooks React | `apps/web/src/hooks/__tests__/` |

## Reglas críticas para este proyecto

**1. Siempre DB en memoria**  
Usar `createClient({ url: ":memory:" })` en cada test de DB. Nunca tocar el archivo real de datos.

**2. Solo imports estáticos**  
`await import(...)` dentro de callbacks o `beforeEach` falla silenciosamente en webpack/Next.js.  
Todos los imports deben estar al nivel del módulo.

**3. Tests del logger — formato variable**  
Verificar que el output *contiene* el tipo de evento esperado. No asumir que el formato es JSON ni pretty-print; varía según entorno.

**4. Patrón Arrange → Act → Assert**  
Una sola assertion de comportamiento por test. Usar `beforeEach` para inicializar estado limpio.

## Huecos de cobertura conocidos

Las áreas sin tests unitarios (oportunidades para agregar):
- API routes (`apps/web/src/app/api/`)
- Hooks React (`useRagStream` — candidato prioritario)
- Worker de ingesta (lógica de locking SQLite)
