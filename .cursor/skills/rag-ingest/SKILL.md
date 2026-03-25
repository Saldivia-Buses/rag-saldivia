---
name: rag-ingest
description: Ingest documents into RAG Saldivia collections — understand tiers, job states, the queue, and monitoring. Use when ingesting documents, debugging ingestion failures, monitoring job progress, understanding the ingestion queue, or when the user says "ingestar", "subir documentos", "está fallando la ingesta", "job stalled", or asks about ingestion tiers or states.
---

# RAG Saldivia — Ingesta de Documentos

## Cómo funciona

1. La UI o CLI encola un job en la tabla `ingestion_queue`
2. El worker (`apps/web/src/workers/ingestion.ts`) toma jobs con locking optimista
3. El worker sube el archivo al RAG Server y hace polling del progreso
4. El estado del job se refleja en `ingestion_jobs`
5. Los errores se registran en `ingestion_alerts`

**No hay Redis** — la cola y el locking son puramente SQLite.

## Tiers de ingesta

Los tiers determinan el tiempo estimado de procesamiento según tamaño del documento:

| Tier | Descripción |
|------|-------------|
| `tiny` | Documentos pequeños (< 10 páginas) |
| `small` | 10–50 páginas |
| `medium` | 50–200 páginas |
| `large` | Documentos grandes (> 200 páginas) |

## Estados de un job

```
pending → running → done
                 ↘ stalled (timeout sin actualización)
                 ↘ error   (fallo del RAG Server)
                 ↘ cancelled (cancelado por usuario)
```

## Locking de la cola — diseño no obvio

El worker hace `SELECT + UPDATE locked_at` en una sola transacción.  
SQLite serializa los writes — sin race condition, sin Redis.  
Un job está "tomado" cuando `lockedAt != null` y `lockedBy = <worker_id>`.  
Jobs con `lockedAt` antiguo (stalled) pueden ser re-tomados por otro worker.

## Monitorear ingesta

```bash
rag ingest status          # tabla de jobs con progreso
rag ingest cancel <jobId>  # cancelar un job en curso

# Eventos en el black box
rag audit log --type ingestion.started
rag audit log --type ingestion.completed
rag audit log --type ingestion.failed
```

## Iniciar una ingesta

```bash
rag ingest start                          # wizard interactivo
rag ingest start -c <colección> -p /ruta  # con flags directos
```

## Alertas

Los jobs en estado `error` generan una entrada en `ingestion_alerts`.  
Ver y resolver desde `/admin` en la UI, o desde la API.

## Schema relevante

`packages/db/src/schema.ts` — tablas `ingestionJobs`, `ingestionQueue`, `ingestionAlerts`.
