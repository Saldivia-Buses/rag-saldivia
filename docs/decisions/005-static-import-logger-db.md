# ADR-005: Import estático de @rag-saldivia/db en el logger (no dinámico)

**Fecha:** 2026-03-25
**Estado:** Aceptado

---

## Contexto

El paquete `packages/logger` necesita persistir eventos en la base de datos (tabla `events`).
Esto crea una dependencia `logger → db`. Durante el desarrollo del Plan 2 (testing), se descubrió que los eventos del logger no se estaban persistiendo en absoluto — silenciosamente.

## El bug

`packages/logger/src/backend.ts` usaba un import dinámico para evitar dependencia circular en tiempo de tipos:

```typescript
const db = await import("@rag-saldivia/db" as any);
```

En el contexto de webpack (Next.js), este import dinámico fallaba silenciosamente — webpack no podía resolver el paquete workspace en el contexto async dentro del módulo, y la función `persistEvent` completaba sin error pero sin escribir nada a la DB.

## Opciones consideradas

- **Import dinámico (`await import(...)`):** intento original. Contras: falla silenciosamente en webpack; difícil de debuggear porque no lanza error.
- **Import estático (`import { db } from "@rag-saldivia/db"`):** el módulo se resuelve en tiempo de carga. Pros: webpack lo maneja correctamente; cualquier error de resolución falla en startup, no silenciosamente en runtime. Contras: crea una dependencia en tiempo de módulo entre `logger` y `db` — si `db` falla al inicializar, el logger también falla.
- **Separar la responsabilidad:** que el logger solo emita eventos y que otro módulo se encargue de persistirlos. Pros: desacoplamiento puro. Contras: requiere refactor significativo del sistema de eventos; over-engineering para el tamaño actual del proyecto.

## Decisión

Elegimos **import estático** porque el fallo silencioso del import dinámico es peor que un fallo en startup.

Si `@rag-saldivia/db` no puede inicializar (DB corrupta, path incorrecto), es preferible que el servidor falle inmediatamente con un mensaje claro a que funcione aparentemente bien pero no persista ningún evento de audit.

```typescript
// backend.ts — CORRECTO
import { db } from "@rag-saldivia/db";

// INCORRECTO — falla silenciosamente en webpack
const db = await import("@rag-saldivia/db" as any);
```

## Consecuencias

**Positivas:**
- Los eventos se persisten correctamente. El audit log y el black box replay funcionan.
- Fallo en startup si la DB no está disponible — falla fast, falla visiblemente.

**Negativas / trade-offs:**
- Si se necesitara usar `packages/logger` en un contexto sin DB (tests unitarios del logger), hay que mockear `@rag-saldivia/db` o inicializar una instancia en memoria. Los tests actuales del logger hacen esto correctamente.
- La dependencia `logger → db` es unidireccional y aceptada. `db` no importa de `logger`.

## Referencias

- `packages/logger/src/backend.ts` — import estático de `@rag-saldivia/db`
- `packages/logger/package.json` — `@rag-saldivia/db` como dependencia explícita
- CHANGELOG.md: "reemplazar lazy-load dinámico por import estático — en webpack/Next.js el dynamic import fallaba silenciosamente"
- `CLAUDE.md` — "Logger + DB: import estático"
