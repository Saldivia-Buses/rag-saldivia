/**
 * BullMQ — definición central de queues y workers de ingesta.
 *
 * F8.30 — Reemplaza:
 *   - tabla ingestion_queue (SQLite) + locking optimista
 *   - workerLoop() manual con processWithRetry()
 *   - setInterval(processScheduledReports) → BullMQ repeat job
 *   - SSE polling de SQLite cada 3s → BullMQ QueueEvents en tiempo real
 *
 * ADR-010: Redis es requerido. BullMQ está construido sobre Redis.
 *
 * IMPORTANTE:
 * - getBullMQConnection() crea conexiones separadas para Queue y Worker
 *   porque BullMQ requiere maxRetriesPerRequest: null en ioredis (v5+)
 * - NO pasar la instancia singleton de getRedisClient() a BullMQ directamente
 *   — BullMQ gestiona sus propias conexiones internas (subscriber + publisher)
 */

import { Queue, Worker, QueueEvents, Job } from "bullmq"
import IORedis from "ioredis"

function getBullMQConnection(): IORedis {
  const url = process.env["REDIS_URL"]
  if (!url) throw new Error("REDIS_URL no configurado (BullMQ)")
  return new IORedis(url, { maxRetriesPerRequest: null })
}

export const ingestionQueue = new Queue("ingestion", {
  connection: getBullMQConnection(),
  defaultJobOptions: {
    attempts: 3,
    backoff: { type: "exponential", delay: 10_000 },
    removeOnComplete: 100,
    removeOnFail: 200,
  },
})

/**
 * Factory para QueueEvents — usado por la ruta SSE del kanban.
 * Crea una nueva instancia dedicada por llamada.
 */
export function createQueueEvents(): QueueEvents {
  return new QueueEvents("ingestion", { connection: getBullMQConnection() })
}

/**
 * Factory para el Worker — solo llamar desde el worker entry point (ingestion.ts).
 * Acepta la función de procesamiento como parámetro para evitar dependencia circular.
 */
/**
 * Inicia el worker de BullMQ. No debe llamarse desde route handlers: va en el entry point del
 * worker (`workers/ingestion.ts`). Llamar dos veces puede producir advertencias de BullMQ.
 */
export function startIngestionWorker(
  processJobFn: (job: Job) => Promise<void>
): Worker {
  return new Worker("ingestion", processJobFn, {
    connection: getBullMQConnection(),
    concurrency: 1,
  })
}

export function scheduleToPattern(schedule: "daily" | "weekly" | "monthly"): string {
  switch (schedule) {
    case "daily": return "0 0 * * *"
    case "weekly": return "0 0 * * 0"
    case "monthly": return "0 0 1 * *"
  }
}

export type IngestionJobData = {
  filePath: string
  collection: string
  userId: number
  filename?: string
}

export type ScheduledReportJobData = {
  id: string
  query: string
  collection: string
  schedule: "daily" | "weekly" | "monthly"
  destination: "saved" | "email"
  email?: string | null
  userId: number
}

export { Job }
