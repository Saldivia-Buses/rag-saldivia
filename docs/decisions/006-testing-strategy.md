# ADR-006: Estrategia de testing — cobertura por capas y enforcement en CI

**Fecha:** 2026-03-26
**Estado:** Aceptado

---

## Contexto

El Plan 2 estableció 79 tests unitarios para las capas más críticas (auth, DB core, logger, config).
El Plan 4 agregó ~50 features sin tests nuevos — las query files de `packages/db` pasaron de 3 a 17
archivos, y solo 3 tienen cobertura.

Sin metas de cobertura explícitas ni enforcement automatizado, la deuda de tests crece silenciosamente.
El Plan 5 establece las reglas como contrato.

## El principio central

> **Si el código no tiene test, no está terminado.**

Esto no es una aspiración — es un criterio de done que el CI enforcea.

## Opciones consideradas

- **100% de cobertura en todo el repo:** imposible de sostener con Next.js (edge middleware,
  React components, SSE). Los tests artificiales que mockean todo no aportan valor real.
- **Sin threshold — solo "escribir tests cuando se pueda":** el estado actual. No funciona.
- **Cobertura diferenciada por capa con threshold enforced:** cada capa tiene una meta
  acorde a su testeabilidad real. El CI enforcea las metas. Las capas no testeables unitariamente
  tienen requisitos alternativos (integración manual, Playwright futuro).

## Decisión

Elegimos **cobertura diferenciada por capa** con las siguientes metas:

| Capa | Target | Herramienta | Cuándo se enforcea |
|------|--------|-------------|-------------------|
| `packages/*` | **95%** | `bun test --coverage` | CI en cada PR |
| `apps/web/src/lib/` | **95%** | `bun test --coverage` | CI en cada PR |
| `apps/web/src/hooks/` | **80%** | `bun test --coverage` | CI en cada PR |
| API routes | cobertura funcional | test manual documentado en PR | revisión humana |
| React components | no requerido ahora | Playwright (futuro Plan 6) | — |
| Edge middleware | no requerido | no testeable unitariamente | — |

## Matriz "tipo de código → test requerido"

Esta tabla es el contrato. Aplica a todo código nuevo desde el momento en que este ADR fue aceptado.

| Tipo de código | Test requerido | Dónde va el test | En qué PR |
|----------------|----------------|------------------|-----------|
| Query nueva en `packages/db/src/queries/` | Test unitario contra SQLite en memoria | `packages/db/src/__tests__/[dominio].test.ts` | **mismo PR** |
| Función pura en `apps/web/src/lib/` | Test unitario | `apps/web/src/lib/__tests__/[nombre].test.ts` | **mismo PR** |
| Lógica pura extraída de un hook | Test unitario | `apps/web/src/lib/[dominio]/__tests__/` | **mismo PR** |
| Server Action nueva | Test de la query subyacente + test manual | `packages/db/src/__tests__/` + PR comment | **mismo PR** |
| API route nueva | Test de integración manual documentado | comentario en PR con curl/resultado | **mismo PR** |
| Hook React nuevo | Extraer lógica pura → test unitario de esa lógica | `apps/web/src/lib/` | **mismo PR** |
| React component nuevo | No requerido ahora | — | — |
| Schema Zod nuevo en `packages/shared` | Test de validación (casos válidos + inválidos) | `packages/shared/src/__tests__/` | **mismo PR** |
| Config loader modificado | Test del caso nuevo | `packages/config/src/__tests__/` | **mismo PR** |

## Enforcement

### CI (`.github/workflows/ci.yml`)
- En cada PR: `bun run test:coverage` — falla si alguna capa está por debajo del threshold
- En pushes directos a `dev`: `bun run test` — solo verifica que los tests pasan (sin coverage)

### Pre-push hook (`.husky/pre-push`)
- `bun run type-check` — igual que antes, no se agrega coverage al pre-push
- Razón: el pre-push debe ser rápido. La cobertura se verifica en CI.

### Threshold en `bunfig.toml`
```toml
[test]
coverageThreshold = 0.95
coverageSkipTestFiles = true
```

## Consecuencias

**Positivas:**
- Los huecos de cobertura se vuelven visibles y bloqueantes en PRs.
- El agente tiene una tabla explícita para saber qué testear sin consultar a nadie.
- La deuda de tests no puede crecer silenciosamente.
- Los tests de `packages/*` son rápidos (SQLite en memoria) — no penalizan el CI.

**Negativas / trade-offs:**
- Los API routes no tienen cobertura automatizada hasta que se implemente Playwright (Plan 6).
  Esto es un trade-off consciente — los tests de routes de Next.js son frágiles y costosos de mantener.
- El threshold 0.95 puede ser bloqueante si se agrega código legacy sin tests en un PR de urgencia.
  Workaround: abrir un issue de deuda técnica y mergearlo rápido en el PR siguiente.

## Referencias

- `docs/plans/ultra-optimize-plan5-testing-foundation.md`
- `bunfig.toml` — threshold configuration
- `.github/workflows/ci.yml` — job `test` con coverage
- `.cursor/skills/rag-testing/SKILL.md` — referencia operativa para el agente
