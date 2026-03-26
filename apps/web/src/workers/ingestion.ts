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
import { readFile, access } from "fs/promises"
import { getDb, ingestionQueue, recordIngestionEvent, listActiveReports, updateLastRun, saveResponse, events } from "@rag-saldivia/db"
import { eq, and, gte, desc } from "drizzle-orm"
import { randomUUID } from "crypto"
import { dispatchEvent } from "@/lib/webhook"
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
  const { id, filePath, collection, userId } = job

  try {
    const exists = await access(filePath).then(() => true).catch(() => false)
    if (!exists) {
      log.error("ingestion.failed", { jobId: id, reason: "file_not_found", filePath })
      return false
    }

    const buffer = await readFile(filePath)
    const blob = buffer.buffer.slice(buffer.byteOffset, buffer.byteOffset + buffer.byteLength)
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

    // Dispatch webhook — F2.38
    dispatchEvent("ingestion.completed", { jobId: id, collection, filename }).catch(() => {})

    // Superficie proactiva — F3.45
    checkProactiveSurface(collection, userId, filename ?? "").catch(() => {})

    // Registrar en historial de colecciones — F2.32
    try {
      await recordIngestionEvent({ collection, userId, action: "added", filename })
    } catch { /* no bloquear si falla */ }

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

// ── Superficie proactiva (F3.45) ────────────────────────────────────────────
/**
 * Cruza el nuevo documento con queries recientes del usuario.
 * Si hay solapamiento de keywords, genera una notificación proactiva.
 */
async function checkProactiveSurface(collection: string, userId: number, filename: string) {
  try {
    const db = getDb()
    const thirtyDaysAgo = Date.now() - 30 * 24 * 60 * 60 * 1000

    // Obtener queries recientes del usuario en esta colección
    const recentQueries = await db
      .select({ payload: events.payload })
      .from(events)
      .where(
        and(
          eq(events.type, "rag.stream_started"),
          eq(events.userId, userId),
          gte(events.ts, thirtyDaysAgo)
        )
      )
      .orderBy(desc(events.ts))
      .limit(20)

    if (recentQueries.length === 0) return

    // Extraer keywords del filename (simple: palabras de > 3 chars sin extensión)
    const docKeywords = filename
      .replace(/\.[^.]+$/, "")
      .toLowerCase()
      .split(/[\s\-_]+/)
      .filter((w) => w.length > 3)

    if (docKeywords.length === 0) return

    // Verificar solapamiento con alguno de los queries recientes
    const hasMatch = recentQueries.some((q) => {
      const payload = q.payload as Record<string, unknown>
      const queryText = String(payload.query ?? "").toLowerCase()
      return docKeywords.some((kw) => queryText.includes(kw))
    })

    if (!hasMatch) return

    // Insertar evento proactivo
    await db.insert(events).values({
      id: randomUUID(),
      ts: Date.now(),
      source: "backend",
      level: "INFO",
      type: "proactive.docs_available",
      userId,
      payload: { collection, filename, matchedQueries: recentQueries.length },
      sequence: Date.now(),
    })

    log.info("system.warning", { message: `Proactive surface triggered for user ${userId}: ${filename}` })
  } catch (err) {
    // No interrumpir el flujo de ingesta
    log.info("system.warning", { message: `Proactive surface check failed: ${String(err).slice(0, 100)}` })
  }
}

// ── Scheduled reports processor ─────────────────────────────────────────────
async function processScheduledReports() {
  try {
    const reports = await listActiveReports()
    for (const report of reports) {
      try {
        const ragUrl = process.env["RAG_SERVER_URL"] ?? "http://localhost:8081"
        const res = await fetch(`${ragUrl}/v1/chat/completions`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            messages: [{ role: "user", content: report.query }],
            collection_name: report.collection,
            use_knowledge_base: true,
          }),
          signal: AbortSignal.timeout(60000),
        })

        if (res.ok) {
          const data = await res.json() as { choices?: Array<{ message?: { content?: string } }> }
          const content = data.choices?.[0]?.message?.content ?? "(sin respuesta)"

          if (report.destination === "saved") {
            await saveResponse({
              userId: report.userId,
              content: `**Informe programado: ${report.query}**\n\n${content}`,
              sessionTitle: `Informe ${new Date().toLocaleDateString("es-AR")}`,
            })
          } else if (report.destination === "email" && report.email) {
            const smtpHost = process.env["SMTP_HOST"]
            if (!smtpHost) {
              log.warn("system.warning", { reportId: report.id })
              // Fallback: guardar igualmente
              await saveResponse({
                userId: report.userId,
                content: `**Informe (SMTP no configurado): ${report.query}**\n\n${content}`,
                sessionTitle: `Informe ${new Date().toLocaleDateString("es-AR")}`,
              })
            }
            // Si SMTP configurado: implementar envío de email con nodemailer en el futuro
          }

          await updateLastRun(report.id, report.schedule)
          log.info("ingestion.completed", { reportId: report.id })
        }
      } catch (err) {
        log.error("system.error", { reportId: report.id, error: String(err) })
      }
    }
  } catch (err) {
    log.error("system.error", { error: String(err) })
  }
}

// ── Main ───────────────────────────────────────────────────────────────────
// Procesar informes programados cada 5 minutos
setInterval(() => { processScheduledReports().catch(() => {}) }, 5 * 60 * 1000)

workerLoop().catch((err) => {
  log.fatal("system.error", { error: String(err), context: "ingestion_worker" })
  process.exit(1)
})
