---
name: rag-db
description: Work with the RAG Saldivia database — Drizzle ORM, SQLite schema, queries, migrations, and the ingestion queue. Use when modifying the schema, writing DB queries, adding a new table, asking about data structure, or when the user mentions "Drizzle", "schema", "tabla", "query", "migración", or a specific table name.
---

# RAG Saldivia — Base de datos

## Fuentes autoritativas

- **Schema completo** (12 tablas): `packages/db/src/schema.ts`
- **Queries por dominio**: `packages/db/src/queries/`
- **Conexión singleton**: `packages/db/src/connection.ts`

Leer `packages/db/src/schema.ts` antes de modificar cualquier tabla o query.

## Decisiones de diseño no obvias

**`@libsql/client`, no `better-sqlite3`**  
El proyecto usa `@libsql/client` (JS puro, sin compilación nativa). No instalar ni usar `better-sqlite3`. Esto es crítico para compatibilidad en distintos entornos.

**Timestamps = INTEGER epoch ms vía Temporal API**  
Todas las columnas de fecha son `INTEGER` (epoch ms). Usar `Temporal.Now.instant().epochMilliseconds` — no `Date.now()`, no `new Date()`. Esto elimina el bug de timezone de SQLite que afectaba al stack anterior.

**Import estático obligatorio**  
`@rag-saldivia/db` debe importarse de forma estática (no `await import(...)`). Los imports dinámicos fallan silenciosamente en webpack/Next.js.

**WAL mode**  
La conexión usa WAL mode. Permite lecturas concurrentes pero un solo writer. Tener esto en cuenta al diseñar operaciones pesadas de escritura.

**Cola sin Redis**  
La tabla `ingestion_queue` reemplaza Redis con locking optimista. SQLite serializa los writes — no hay race condition. El worker hace `SELECT + UPDATE locked_at` en una sola transacción.

**Relaciones para `.query` con `with`**  
Para usar la API `db.query.tabla.findFirst({ with: { ... } })` las relaciones deben declararse en la sección `Relations` al final del schema. Sin eso, los `with` no funcionan.

## Cómo agregar una entidad nueva

1. Agregar la tabla en `packages/db/src/schema.ts` con `sqliteTable(...)`
2. Declarar los tipos inferidos al final: `export type DbX = typeof x.$inferSelect`
3. Agregar las relaciones si aplica
4. Crear el archivo `packages/db/src/queries/[dominio].ts` con las funciones CRUD
5. Re-exportar desde `packages/db/src/index.ts`
6. Escribir tests en `packages/db/src/__tests__/[dominio].test.ts` con DB en memoria

## Tablas principales

| Tabla | Contenido |
|-------|-----------|
| `users` | Usuarios — roles: `admin \| area_manager \| user` |
| `areas` | Áreas/departamentos |
| `user_areas` | Many-to-many usuarios ↔ áreas |
| `area_collections` | Permisos área ↔ colección (`read \| write \| admin`) |
| `chat_sessions` | Sesiones de chat |
| `chat_messages` | Mensajes (`user \| assistant \| system`) |
| `ingestion_jobs` | Jobs con tiers (`tiny \| small \| medium \| large`) y estados |
| `ingestion_queue` | Cola de ingesta con locking (reemplaza Redis) |
| `events` | Black Box — log inmutable del sistema |
