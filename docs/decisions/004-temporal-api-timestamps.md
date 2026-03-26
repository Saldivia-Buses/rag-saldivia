# ADR-004: Temporal API para timestamps (en lugar de Date o Date.now())

**Fecha:** 2026-03-24
**Estado:** Aceptado

---

## Contexto

El stack necesita timestamps consistentes en múltiples lugares: eventos de audit log, columnas `created_at` de la DB, locking de la ingestion queue (`locked_at`), expiración de tokens JWT.

El stack original (Python) tenía un bug conocido donde `_ts()` generaba timestamps inconsistentes según el contexto en que se llamaba. Al reescribir en TypeScript se quería evitar la misma clase de problema.

## Opciones consideradas

- **`Date.now()`:** retorna epoch en milisegundos. Simple, universal. Contras: `new Date()` y `Date.now()` se pueden mezclar accidentalmente; la API `Date` tiene comportamientos heredados confusos (meses 0-indexed, mutabilidad, etc.).
- **`Date.now()` con convención estricta:** usar solo `Date.now()` en todo el codebase, nunca `new Date()`. Pros: simple. Contras: requiere disciplina manual; fácil de violar sin que TypeScript lo detecte.
- **Temporal API (TC39 Stage 3):** `Temporal.Now.instant().epochMilliseconds`. API moderna, inmutable, sin ambigüedad de timezone. Disponible en Bun y Node.js 22+ sin polyfill. Cons: API más verbosa; todavía Stage 3 (aunque ya en todos los runtimes modernos).
- **Librería externa (date-fns, dayjs):** agrega dependencia para algo que el runtime ya provee.

## Decisión

Elegimos **Temporal API** para todos los timestamps del sistema: `Temporal.Now.instant().epochMilliseconds`.

Esta elección:
1. Elimina la ambigüedad entre `Date.now()` (ms) y `new Date().getTime()` (ms) vs `new Date()` (objeto mutable).
2. La verbosidad es intencional — `Temporal.Now.instant().epochMilliseconds` es auto-documentado.
3. Disponible nativamente en Bun (runtime de desarrollo y producción).

Los valores se almacenan en SQLite como `INTEGER` (epoch ms). No se usa `TEXT` con ISO strings para evitar conversiones en queries.

## Consecuencias

**Positivas:**
- Un solo patrón para todos los timestamps en el codebase, imposible de confundir con `new Date()`.
- Inmutabilidad: `Temporal.Instant` no tiene métodos que muten el objeto.
- Cero dependencias externas.

**Negativas / trade-offs:**
- Temporal API es Stage 3; si Bun o Node.js lo remueven (improbable dado el avance), habría que agregar el polyfill `@js-temporal/polyfill`.
- Los timestamps en la DB son epoch ms (integers), no ISO strings. Las herramientas de visualización de SQLite muestran números en lugar de fechas legibles — hay que convertir manualmente al inspeccionar la DB.

## Referencias

- `CLAUDE.md` — "Temporal API para todos los timestamps → elimina el bug `_ts()` de SQLite"
- `packages/db/src/schema.ts` — columnas `created_at`, `locked_at` como `integer`
- `apps/web/src/workers/ingestion.ts` — uso de `Temporal.Now.instant().epochMilliseconds` en locking
