#!/usr/bin/env bun
/**
 * Worker de ingesta — entry point.
 *
 * F8.30 — Reemplaza el workerLoop manual por BullMQ.
 * Este archivo ahora contiene solo lógica de negocio pura (processJob)
 * y arranca el worker BullMQ al importarse como entry point.
 *
 * BullMQ gestiona: retries con backoff exponencial, concurrencia,
 * graceful shutdown, y locking distribuido.
 *
 * Uso:
 *   bun apps/web/src/workers/ingestion.ts
 */

import { readFile, access } from "fs/promises"
import { getDb, recordIngestionEvent, listActiveReports, updateLastRun, saveResponse, events, getRedisClient } from "@rag-saldivia/db"
import { randomUUID } from "crypto"
import { eq, and, gte, desc } from "drizzle-orm"
import { dispatchEvent } from "@/lib/webhook"
import { log } from "@rag-saldivia/logger/backend"
import { formatDate } from "@/lib/utils"
import {
  ingestionQueue,
  startIngestionWorker,
  scheduleToPattern,
  type IngestionJobData,
  type ScheduledReportJobData,
} from "@/lib/queue"
import type { Job } from "bullmq"

const INGESTOR_URL = process.env["INGESTOR_URL"] ?? "http://localhost:8082"

// ── Lógica de negocio pura ─────────────────────────────────────────────────

export async function processJob(data: IngestionJobData): Promise<void> {
  const { filePath, collection, userId } = data
  const filename = data.filename ?? filePath.split("/").pop() ?? "document.pdf"
  const jobId = `${collection}-${filename}`

  const exists = await access(filePath).then(() => true).catch(() => false)
  if (!exists) {
    throw new Error(`Archivo no encontrado: ${filePath}`)
  }

  const buffer = await readFile(filePath)
  const blob = buffer.buffer.slice(buffer.byteOffset, buffer.byteOffset + buffer.byteLength)

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
    throw new Error(`NV-Ingest respondió ${response.status}: ${body.slice(0, 200)}`)
  }

  log.info("ingestion.completed", { jobId, collection, filename })

  // Dispatch webhook — F2.38
  dispatchEvent("ingestion.completed", { jobId, collection, filename }).catch(() => {})

  // Superficie proactiva — F3.45
  checkProactiveSurface(collection, userId, filename).catch(() => {})

  // Registrar en historial de colecciones — F2.32
  await recordIngestionEvent({ collection, userId, action: "added", filename }).catch(() => {})

  // Notificar al usuario en tiempo real — F8.28
  getRedisClient()
    .publish(`notifications:${userId}`, JSON.stringify({
      id: randomUUID(),
      type: "ingestion.completed",
      ts: Date.now(),
      payload: { collection, filename },
    }))
    .catch(() => {})
}

// ── Informes programados ────────────────────────────────────────────────────

export async function processScheduledReport(data: ScheduledReportJobData): Promise<void> {
  const ragUrl = process.env["RAG_SERVER_URL"] ?? "http://localhost:8081"
  const res = await fetch(`${ragUrl}/v1/chat/completions`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      messages: [{ role: "user", content: data.query }],
      collection_name: data.collection,
      use_knowledge_base: true,
    }),
    signal: AbortSignal.timeout(60000),
  })

  if (!res.ok) throw new Error(`RAG respondió ${res.status}`)

  const body = await res.json() as { choices?: Array<{ message?: { content?: string } }> }
  const content = body.choices?.[0]?.message?.content ?? "(sin respuesta)"

  if (data.destination === "saved") {
    await saveResponse({
      userId: data.userId,
      content: `**Informe programado: ${data.query}**\n\n${content}`,
      sessionTitle: `Informe ${formatDate(new Date())}`,
    })
  } else if (data.destination === "email" && data.email) {
    const smtpHost = process.env["SMTP_HOST"]
    if (!smtpHost) {
      await saveResponse({
        userId: data.userId,
        content: `**Informe (SMTP no configurado): ${data.query}**\n\n${content}`,
        sessionTitle: `Informe ${formatDate(new Date())}`,
      })
    }
    // Si SMTP configurado: implementar envío con nodemailer en el futuro
  }

  await updateLastRun(data.id, data.schedule)
  log.info("ingestion.completed", { reportId: data.id })
}

// ── Superficie proactiva — F3.45 ───────────────────────────────────────────

async function checkProactiveSurface(collection: string, userId: number, filename: string) {
  try {
    const db = getDb()
    const thirtyDaysAgo = Date.now() - 30 * 24 * 60 * 60 * 1000

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

    const docKeywords = filename
      .replace(/\.[^.]+$/, "")
      .toLowerCase()
      .split(/[\s\-_]+/)
      .filter((w) => w.length > 3)

    if (docKeywords.length === 0) return

    const hasMatch = recentQueries.some((q) => {
      const payload = q.payload as Record<string, unknown>
      const queryText = String(payload.query ?? "").toLowerCase()
      return docKeywords.some((kw) => queryText.includes(kw))
    })

    if (!hasMatch) return

    await db.insert(events).values({
      id: randomUUID(),
      ts: Date.now(),
      source: "backend",
      level: "INFO",
      type: "proactive.docs_available",
      userId,
      payload: { collection, filename, matchedQueries: recentQueries.length },
      sequence: await import("@rag-saldivia/db").then((m) => m.getRedisClient().incr("events:seq")),
    })

    // Notificar al usuario
    getRedisClient()
      .publish(`notifications:${userId}`, JSON.stringify({
        id: randomUUID(),
        type: "proactive.docs_available",
        ts: Date.now(),
        payload: { collection, filename },
      }))
      .catch(() => {})
  } catch {
    // No interrumpir el flujo de ingesta
  }
}

// ── Worker BullMQ ───────────────────────────────────────────────────────────

const worker = startIngestionWorker(async (job: Job) => {
  if (job.name === "scheduled-report") {
    await processScheduledReport(job.data as ScheduledReportJobData)
  } else {
    await processJob(job.data as IngestionJobData)
  }
})

worker.on("completed", (job) => {
  log.info("ingestion.completed", { jobId: job.id, jobName: job.name })
})

worker.on("failed", (job, err) => {
  log.error("ingestion.failed", { jobId: job?.id, jobName: job?.name, error: err.message })
  if (job?.data?.userId) {
    getRedisClient()
      .publish(`notifications:${job.data.userId}`, JSON.stringify({
        id: randomUUID(),
        type: "ingestion.error",
        ts: Date.now(),
        payload: { error: err.message, filename: job.data.filename },
      }))
      .catch(() => {})
  }
})

// ── Scheduled reports via BullMQ repeat jobs ─────────────────────────────────

async function syncScheduledReports() {
  try {
    const reports = await listActiveReports()
    for (const report of reports) {
      await ingestionQueue.add("scheduled-report", {
        id: report.id,
        query: report.query,
        collection: report.collection,
        schedule: report.schedule,
        destination: report.destination,
        email: report.email,
        userId: report.userId,
      } as ScheduledReportJobData, {
        repeat: { pattern: scheduleToPattern(report.schedule) },
        jobId: `report-${report.id}`,
      }).catch(() => {})
    }
  } catch {
    // Silencioso
  }
}

// Sincronizar informes programados al arrancar y cada hora
syncScheduledReports().catch(() => {})
setInterval(() => { syncScheduledReports().catch(() => {}) }, 60 * 60 * 1000)

log.info("system.start", { message: `Ingestion worker started via BullMQ (pid: ${process.pid})` })
