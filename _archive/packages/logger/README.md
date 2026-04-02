# `@rag-saldivia/logger`

Logger estructurado para el servidor, integración con la tabla **`events`** (`@rag-saldivia/db`), **blackbox replay** y **rotación de archivos** bajo `logs/`.

## Niveles

`TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL` — controlados por `LOG_LEVEL` (default `INFO`).

## Uso

```typescript
import { log } from "@rag-saldivia/logger/backend"

log.info("rag.query", { collection: "docs" }, { userId: 1, requestId: "…" })
```

El tipo del primer argumento (`EventType`) debe ser un valor válido del `EventTypeSchema` en `@rag-saldivia/shared`.

## Event types

Definidos en `EventTypeSchema` en `packages/shared/src/schemas.ts`, por ejemplo:

`auth.login`, `auth.logout`, `rag.query`, `rag.stream_started`, `ingestion.started`, `ingestion.completed`, `system.request`, `client.error`, …

Lista completa: ver el enum en el archivo citado.

## Black box

Los eventos persistidos permiten reconstruir el estado aproximado del sistema con `reconstructFromEvents()` en `src/blackbox.ts` (usado por `rag audit replay`).

**No** usar replay sobre tablas enormes sin paginar; está pensado para auditoría y diagnóstico.

## Rotación de archivos

- Archivos: `logs/backend.log`, `logs/errors.log`, `logs/frontend.log`.
- Rotación al superar **10 MB**; se conservan hasta **5** archivos históricos (.1, .2, …).
- Los tamaños se persisten en Redis (`log:sizes`) para coordinar entre instancias.

## Retención de eventos en SQLite

La limpieza opcional de filas viejas en `events` respeta `LOG_RETENTION_DAYS` (default **90**) vía `deleteOldEvents` en `packages/db/src/queries/events-cleanup.ts`.

## Tests

```bash
bun test packages/logger/
```
