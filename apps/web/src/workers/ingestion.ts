#!/usr/bin/env bun
/**
 * Worker de ingesta — reemplaza saldivia/ingestion_worker.py
 *
 * Sondea la tabla ingestion_queue en SQLite, toma jobs con locking optimista,
 * los envía al NV-Ingest en :8082, y actualiza el estado.
 *
 * SQLite serializa writes → no hay race conditions entre workers.
 *
 * Uso:
 *   bun src/workers/ingestion.ts
 *   (arranca automáticamente con el servidor Next.js en producción)
 */

import { eq, and, isNull } from "drizzle-orm"
import { getDb, ingestionQueue } from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"

const INGESTOR_URL = process.env["INGESTOR_URL"] ?? "http://localhost:8082"
const POLL_INTERVAL_MS = 2000
const MAX_RETRIES = 3
const RETRY_DELAYS_MS = [10_000, 30_000, 60_000]
const WORKER_ID = `worker-${process.pid}-${Date.now()}`

let _shutdown = false

// ── Graceful shutdown ──────────────────────────────────────────────────────

process.on("SIGTERM", () => {
  log.info("system.warning", { message: "SIGTERM recibido — finalizando job actual y apagando" })
  _shutdown = true
})

process.on("SIGINT", () => {
  _shutdown = true
})

// ── Procesar un job ────────────────────────────────────────────────────────

async function processJob(job: typeof ingestionQueue.$inferSelect): Promise<boolean> {
  const { id, filePath, collection } = job

  try {
    const file = Bun.file(filePath)
    const exists = await file.exists()
    if (!exists) {
      log.error("ingestion.failed", { jobId: id, reason: "file_not_found", filePath })
      return false
    }

    const blob = await file.arrayBuffer()
    const filename = filePath.split("/").pop() ?? filePath.split("\\").pop() ?? "document.pdf"

    const formData = new FormData()
    formData.append("documents", new Blob([blob], { type: "application/pdf" }), filename)
    formData.append("data", JSON.stringify({ collection_name: collection }))

    const response = await fetch(`${INGESTOR_URL}/v1/documents`, {
      method: "POST",
      body: formData,
      signal: AbortSignal.timeout(600_000), // 10 minutos máximo
    })

    if (!response.ok) {
      const body = await response.text().catch(() => "")
      log.error("ingestion.failed", {
        jobId: id,
        status: response.status,
        body: body.slice(0, 200),
      })
      return false
    }

    log.info("ingestion.completed", { jobId: id, collection, filename })
    return true
  } catch (error) {
    log.error("ingestion.failed", { jobId: id, error: String(error) })
    return false
  }
}

async function processWithRetry(job: typeof ingestionQueue.$inferSelect): Promise<boolean> {
  const db = getDb()

  for (let attempt = 0; attempt < MAX_RETRIES; attempt++) {
    if (_shutdown) return false

    const success = await processJob(job)
    if (success) return true

    if (attempt < MAX_RETRIES - 1) {
      const delay = RETRY_DELAYS_MS[attempt] ?? 60_000
      log.info("system.warning", {
        message: `Job ${job.id} falló, reintento ${attempt + 1}/${MAX_RETRIES - 1} en ${delay / 1000}s`,
      })

      // Actualizar retry_count en DB
      await db
        .update(ingestionQueue)
        .set({ retryCount: (job.retryCount ?? 0) + 1 })
        .where(eq(ingestionQueue.id, job.id))

      await new Promise((r) => setTimeout(r, delay))
    }
  }

  log.error("ingestion.failed", { jobId: job.id, reason: `failed after ${MAX_RETRIES} attempts` })
  return false
}

// ── Loop principal ─────────────────────────────────────────────────────────

async function workerLoop() {
  const db = getDb()
  log.info("system.start", { workerId: WORKER_ID, ingestorUrl: INGESTOR_URL })

  while (!_shutdown) {
    try {
      // Tomar el próximo job disponible con locking optimista
      const now = Date.now()

      // Liberar jobs bloqueados por más de 15 minutos (worker muerto)
      await db
        .update(ingestionQueue)
        .set({ lockedAt: null, lockedBy: null, status: "pending" })
        .where(
          and(
            eq(ingestionQueue.status, "locked"),
            // locked_at < now - 15min
          )
        )

      // SELECT + lock en una transacción implícita (SQLite serializa)
      const [job] = await db
        .select()
        .from(ingestionQueue)
        .where(and(eq(ingestionQueue.status, "pending"), isNull(ingestionQueue.lockedAt)))
        .orderBy(ingestionQueue.priority, ingestionQueue.createdAt)
        .limit(1)

      if (!job) {
        // Sin jobs — esperar antes de reintentar
        await new Promise((r) => setTimeout(r, POLL_INTERVAL_MS))
        continue
      }

      // Bloquear el job
      await db
        .update(ingestionQueue)
        .set({ status: "locked", lockedAt: now, lockedBy: WORKER_ID, startedAt: now })
        .where(and(eq(ingestionQueue.id, job.id), isNull(ingestionQueue.lockedAt)))

      // Procesar
      const success = await processWithRetry(job)

      // Actualizar resultado
      await db
        .update(ingestionQueue)
        .set({
          status: success ? "done" : "error",
          completedAt: Date.now(),
          lockedAt: null,
          lockedBy: null,
          error: success ? null : "Falló después de todos los reintentos",
        })
        .where(eq(ingestionQueue.id, job.id))
    } catch (error) {
      log.error("system.error", { error: String(error), context: "ingestion_worker_loop" })
      await new Promise((r) => setTimeout(r, POLL_INTERVAL_MS * 2))
    }
  }

  log.info("system.warning", { message: `Worker ${WORKER_ID} apagado limpiamente` })
}

// ── Main ───────────────────────────────────────────────────────────────────
workerLoop().catch((err) => {
  log.fatal("system.error", { error: String(err), context: "ingestion_worker" })
  process.exit(1)
})
