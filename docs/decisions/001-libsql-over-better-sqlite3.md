# ADR-001: @libsql/client en lugar de better-sqlite3

**Fecha:** 2026-03-24
**Estado:** Aceptado

---

## Contexto

El stack necesita una base de datos SQLite embebida para auth, sesiones, ingestion queue y audit log.
El entorno de desarrollo es WSL2 sobre Windows; el entorno de producción es Ubuntu 24.04 nativo.
El runtime principal es Bun, pero el código también corre en el contexto de Next.js (webpack, Node.js).

## Opciones consideradas

- **better-sqlite3:** cliente SQLite síncrono y muy maduro. Contras: requiere compilación nativa con `node-gyp` — falla en Bun porque Bun no expone el ABI de Node.js de la misma manera; el proceso de build se rompe en WSL2 sin tools de compilación instaladas.
- **@libsql/client (libSQL):** fork de SQLite mantenido por Turso, escrito en JS puro (no hay addon nativo). Pros: sin compilación, compatible con Bun y Node.js; API async moderna. Contras: ligeramente menos maduro; la integración con Drizzle ORM es `drizzle-orm/libsql` en lugar de `drizzle-orm/better-sqlite3`.
- **SQLite via Bun.sqlite:** API nativa de Bun, cero dependencias. Contras: no disponible en el contexto webpack/Next.js (en `packages/*` importados desde `apps/web`); ataría toda la DB al runtime Bun.

## Decisión

Elegimos **@libsql/client** porque elimina la compilación nativa sin sacrificar funcionalidad.

El driver se importa en `packages/db/src/connection.ts` via `drizzle-orm/libsql`.
En `apps/web/next.config.ts` se agrega `@libsql/client` y `libsql` a `serverExternalPackages` para excluirlos del bundling de webpack (que no sabe manejar addons nativos aunque no los haya).

Un workaround adicional para WSL2: `scripts/link-libsql.sh` crea symlinks manuales de `@libsql` en `apps/web/node_modules/` porque Bun workspaces en filesystem Windows no crea symlinks correctamente. **Este script no es necesario en Ubuntu nativo (producción).**

## Consecuencias

**Positivas:**
- `bun install` funciona sin node-gyp ni herramientas de compilación C/C++.
- El mismo paquete `@rag-saldivia/db` corre en Bun (CLI, worker de ingesta) y en Next.js (Server Components, API routes) sin cambios.
- Compatible con `bun:test` — los tests de DB usan SQLite en memoria con `:memory:`.

**Negativas / trade-offs:**
- `@libsql/client` es async-only; no existe API síncrona. Las queries de DB siempre requieren `await`.
- En WSL2, hay que correr `scripts/link-libsql.sh` después de `bun install` la primera vez.
- Si Bun implementa un ABI de Node.js completo en el futuro, `better-sqlite3` podría ser una opción más performante.

## Referencias

- `packages/db/src/connection.ts`
- `apps/web/next.config.ts` — sección `serverExternalPackages` y `webpack.externals`
- `scripts/link-libsql.sh`
- CHANGELOG.md línea ~392: "migrado de better-sqlite3 a @libsql/client"
