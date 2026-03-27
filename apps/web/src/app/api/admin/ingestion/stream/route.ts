/**
 * GET /api/admin/ingestion/stream
 *
 * SSE endpoint que emite eventos de ingesta en tiempo real via BullMQ QueueEvents.
 * F8.30 — Reemplaza el polling SQLite cada 3s por eventos de BullMQ sobre Redis.
 *
 * Solo accesible por admins.
 */

import { extractClaims } from "@/lib/auth/jwt"
import { ingestionQueue, createQueueEvents, type IngestionJobData } from "@/lib/queue"
import { Job } from "bullmq"

export const runtime = "nodejs"

function jobToDto(job: Job<IngestionJobData>) {
  return {
    id: job.id,
    name: job.name,
    data: job.data,
    state: job.returnvalue !== undefined ? "completed" : "waiting",
    progress: job.progress,
    failedReason: job.failedReason,
    timestamp: job.timestamp,
    finishedOn: job.finishedOn,
  }
}

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims || claims.role !== "admin") {
    return new Response("No autorizado", { status: 401 })
  }

  const stream = new ReadableStream({
    async start(controller) {
      const encoder = new TextEncoder()
      const queueEvents = createQueueEvents()

      function emit(data: unknown) {
        try {
          controller.enqueue(encoder.encode(`data: ${JSON.stringify(data)}\n\n`))
        } catch {
          // cliente desconectado
        }
      }

      // Estado inicial — jobs activos, esperando y recientes
      try {
        const jobs = await ingestionQueue.getJobs(["active", "waiting", "completed", "failed"])
        emit({ type: "init", jobs: jobs.map(jobToDto) })
      } catch {
        emit({ type: "init", jobs: [] })
      }

      // Suscribirse a eventos BullMQ en tiempo real
      queueEvents.on("completed", async ({ jobId }) => {
        const job = await Job.fromId(ingestionQueue, jobId)
        if (job) emit({ type: "completed", job: jobToDto(job) })
      })

      queueEvents.on("failed", async ({ jobId, failedReason }) => {
        const job = await Job.fromId(ingestionQueue, jobId)
        emit({ type: "failed", job: job ? jobToDto(job) : { id: jobId }, error: failedReason })
      })

      queueEvents.on("progress", async ({ jobId, data }) => {
        const job = await Job.fromId(ingestionQueue, jobId)
        if (job) emit({ type: "progress", job: jobToDto(job), progress: data })
      })

      queueEvents.on("active", async ({ jobId }) => {
        const job = await Job.fromId(ingestionQueue, jobId)
        if (job) emit({ type: "active", job: jobToDto(job) })
      })

      queueEvents.on("waiting", async ({ jobId }) => {
        emit({ type: "waiting", jobId })
      })

      // Limpiar cuando el cliente desconecta
      request.signal.addEventListener("abort", () => {
        queueEvents.close().catch(() => {})
        try { controller.close() } catch { /* ya cerrado */ }
      })
    },
  })

  return new Response(stream, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      "Connection": "keep-alive",
      "X-Accel-Buffering": "no",
    },
  })
}
