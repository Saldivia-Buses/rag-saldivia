/**
 * Black Box — Reconstrucción del estado del sistema.
 *
 * Lee los eventos en orden de secuencia desde la tabla events
 * y reconstruye lo que pasó: usuarios, queries, errores, en qué orden.
 *
 * Uso: rag audit replay --from 2026-03-24
 *
 * Implementación completa en Fase 5. Esta versión define la API y el
 * algoritmo de reconstrucción base.
 */

import type { DbEvent } from "@rag-saldivia/db"

export type ReconstructedState = {
  // Estado de usuarios en ese momento
  users: Map<number, { email: string; role: string; active: boolean; lastAction?: string }>
  // Queries hechas al RAG
  ragQueries: Array<{ ts: number; userId: number; query: string; collection: string; success: boolean }>
  // Errores registrados
  errors: Array<{ ts: number; type: string; message: string; suggestion?: string }>
  // Timeline de eventos (los más recientes primero en la UI)
  timeline: Array<{ ts: number; type: string; userId?: number; summary: string }>
  // Estadísticas
  stats: {
    totalEvents: number
    errorCount: number
    warnCount: number
    uniqueUsers: number
    ragQueryCount: number
  }
}

export function reconstructFromEvents(events: DbEvent[]): ReconstructedState {
  const state: ReconstructedState = {
    users: new Map(),
    ragQueries: [],
    errors: [],
    timeline: [],
    stats: {
      totalEvents: events.length,
      errorCount: 0,
      warnCount: 0,
      uniqueUsers: 0,
      ragQueryCount: 0,
    },
  }

  for (const event of events) {
    const payload = event.payload as Record<string, unknown>

    // Actualizar stats
    if (event.level === "ERROR" || event.level === "FATAL") state.stats.errorCount++
    if (event.level === "WARN") state.stats.warnCount++
    if (event.userId) {
      if (!state.users.has(event.userId)) state.stats.uniqueUsers++
    }

    // Procesar por tipo de evento
    switch (event.type) {
      case "auth.login":
        if (event.userId) {
          state.users.set(event.userId, {
            email: String(payload["email"] ?? ""),
            role: String(payload["role"] ?? "user"),
            active: true,
            lastAction: `Login a las ${new Date(event.ts).toISOString()}`,
          })
        }
        state.timeline.push({
          ts: event.ts,
          type: event.type,
          userId: event.userId ?? undefined,
          summary: `Login: ${payload["email"] ?? "unknown"}`,
        })
        break

      case "rag.query":
      case "rag.query_crossdoc":
        state.stats.ragQueryCount++
        if (event.userId) {
          state.ragQueries.push({
            ts: event.ts,
            userId: event.userId,
            query: String(payload["query"] ?? "").slice(0, 100),
            collection: String(payload["collection"] ?? ""),
            success: !payload["error"],
          })
        }
        state.timeline.push({
          ts: event.ts,
          type: event.type,
          userId: event.userId ?? undefined,
          summary: `Query: "${String(payload["query"] ?? "").slice(0, 60)}"`,
        })
        break

      case "rag.error":
      case "system.error":
      case "client.error":
        state.errors.push({
          ts: event.ts,
          type: event.type,
          message: String(payload["error"] ?? payload["message"] ?? ""),
          suggestion: String(payload["suggestion"] ?? ""),
        })
        state.timeline.push({
          ts: event.ts,
          type: event.type,
          userId: event.userId ?? undefined,
          summary: `Error: ${String(payload["error"] ?? payload["message"] ?? "").slice(0, 80)}`,
        })
        break

      case "user.created":
      case "user.updated":
        if (payload["userId"]) {
          const uid = Number(payload["userId"])
          state.users.set(uid, {
            email: String(payload["email"] ?? ""),
            role: String(payload["role"] ?? "user"),
            active: payload["active"] !== false,
          })
        }
        break

      case "user.deleted":
        if (payload["userId"]) {
          const uid = Number(payload["userId"])
          const u = state.users.get(uid)
          if (u) state.users.set(uid, { ...u, active: false })
        }
        break

      default:
        state.timeline.push({
          ts: event.ts,
          type: event.type,
          userId: event.userId ?? undefined,
          summary: JSON.stringify(payload).slice(0, 80),
        })
    }
  }

  // Ordenar timeline por ts descendente (más reciente primero)
  state.timeline.sort((a, b) => b.ts - a.ts)

  return state
}

export function formatTimeline(state: ReconstructedState): string {
  const lines: string[] = [
    `=== Black Box Replay ===`,
    `Total eventos: ${state.stats.totalEvents}`,
    `Errores: ${state.stats.errorCount}`,
    `Warnings: ${state.stats.warnCount}`,
    `Usuarios únicos: ${state.stats.uniqueUsers}`,
    `Queries RAG: ${state.stats.ragQueryCount}`,
    "",
    "=== Timeline (más reciente primero) ===",
  ]

  for (const event of state.timeline.slice(0, 50)) {
    const ts = new Date(event.ts).toISOString().replace("T", " ").slice(0, 19)
    const user = event.userId ? `[user=${event.userId}]` : ""
    lines.push(`  ${ts}  ${event.type.padEnd(25)} ${user}  ${event.summary}`)
  }

  if (state.errors.length > 0) {
    lines.push("", "=== Errores registrados ===")
    for (const err of state.errors.slice(-10)) {
      const ts = new Date(err.ts).toISOString().replace("T", " ").slice(0, 19)
      lines.push(`  ${ts}  ${err.type}: ${err.message}`)
      if (err.suggestion) lines.push(`            → ${err.suggestion}`)
    }
  }

  return lines.join("\n")
}
