# ADR-010: Redis como dependencia requerida del sistema

**Fecha:** 2026-03-27  
**Estado:** Aceptada  
**Contexto:** Plan 8 — Fase 8 (F8.22–F8.30)

---

## Contexto

El stack de RAG Saldivia acumuló 11 workarounds de single-instance a medida que se construyeron features que lógicamente requieren almacenamiento distribuido:

| Workaround eliminado | Archivo original | Reemplazado por |
|---|---|---|
| `let _seq: number \| null = null` | `packages/db/src/queries/events.ts` | Redis `INCR events:seq` |
| `const _sizeCache = new Map()` | `packages/logger/src/rotation.ts` | Redis `HSET log:sizes` |
| `ingestion_queue` SQLite + locking optimista | `packages/db/src/schema.ts` | BullMQ sobre Redis |
| `processWithRetry` manual | `apps/web/src/workers/ingestion.ts` | BullMQ `attempts + backoff` |
| `setInterval(processScheduledReports)` | `apps/web/src/workers/ingestion.ts` | BullMQ repeat jobs |
| SSE polling cada 3s sobre SQLite | `api/admin/ingestion/stream/route.ts` | BullMQ `QueueEvents` |
| Sin JWT blacklist | `apps/web/src/middleware.ts` | Redis `SET revoked:{jti} EX` |
| `unstable_cache` local para colecciones | `lib/rag/collections-cache.ts` | Redis `GET/SET rag:collections` |
| `localStorage["seen_notification_ids"]` | `hooks/useNotifications.ts` | Redis Sorted Set `ZADD` |
| Sin master election para external-sync | `apps/web/src/workers/external-sync.ts` | Redis `SET NX EX` |
| `sequence: Date.now()` en ingestion worker | `apps/web/src/workers/ingestion.ts` | Redis `INCR events:seq` |

## Decisión

**Redis es una dependencia del sistema igual que Milvus.** No existe código de fallback para un stack sin Redis.

La función `getRedisClient()` en `packages/db/src/redis.ts` lanza un error claro con instrucciones si `REDIS_URL` no está configurado. No retorna null, no tiene modo degradado.

## Primitivas Redis utilizadas

| Primitiva | Uso | Clave |
|---|---|---|
| `INCR` | Secuencia monotónica de eventos | `events:seq` |
| `HSET/HGET` | Cache de tamaños de archivos de log | `log:sizes` |
| `SET EX` | JWT blacklist (revocación) | `revoked:{jti}` |
| `GET/SET EX` | Cache de colecciones RAG | `rag:collections` |
| `ZADD/ZSCORE` | Notificaciones vistas por usuario | `notifications:seen:{userId}` |
| `PUBLISH/SUBSCRIBE` | Notificaciones en tiempo real | `notifications:{userId}` |
| `SET NX EX` | Master election para external-sync | `worker:master:external-sync` |
| BullMQ (sobre Redis) | Cola de ingesta distribuida | prefijo `bull:ingestion:` |

## Consecuencias

**Positivas:**
- ~350 líneas de código eliminadas (workarounds + tabla `ingestion_queue`)
- Reintentos de ingesta con backoff exponencial sin código manual
- JWT revocación inmediata al logout (antes era imposible)
- Notificaciones en tiempo real sub-segundo (antes: polling cada 30s)
- Cache de colecciones compartido entre instancias (antes: por proceso)
- Worker de ingesta sin locking SQLite manual

**Negativas:**
- Dependencia adicional en el stack de infraestructura
- Los tests unitarios requieren `ioredis-mock` como devDependency

## Para tests unitarios

`ioredis-mock` reemplaza `ioredis` en el entorno de tests via `mock.module()` en el preload de bun:test. Los 270+ tests de lógica no requieren Redis corriendo.

## Para CI

El workflow de CI incluye `services: redis` para que los tests de integración tengan Redis disponible.

## Analogía

> Nadie escribe `if (milvus) ... else fallbackSinBusquedaVectorial`.  
> Redis es lo mismo.
