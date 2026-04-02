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

// Tipo inline para evitar dependencia circular logger → db
export type DbEvent = {
  id: string
  ts: number
  source: string
  level: string
  type: string
  userId: number | null
  sessionId: string | null
  payload: unknown
  sequence: number
}

export type IngestionEventRecord = {
  ts: number
  type: string
  filename?: string
  collection?: string
  sourceId?: string
  error?: string
}

export type ReconstructedState = {
  // Estado de usuarios en ese momento
  users: Map<number, { email: string; role: string; active: boolean; lastAction?: string }>
  // Queries hechas al RAG
  ragQueries: Array<{ ts: number; userId: number; query: string; collection: string; success: boolean }>
  // Eventos de ingesta (started, completed, failed, stalled)
  ingestionEvents: IngestionEventRecord[]
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
    ingestionCount: number
  }
}

// ── Handlers por tipo de evento ────────────────────────────────────────────

function handleAuthLogin(event: DbEvent, payload: Record<string, unknown>, state: ReconstructedState) {
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
    ...(event.userId != null ? { userId: event.userId } : {}),
    summary: `Login: ${payload["email"] ?? "unknown"}`,
  })
}

function handleRagQuery(event: DbEvent, payload: Record<string, unknown>, state: ReconstructedState) {
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
    ...(event.userId != null ? { userId: event.userId } : {}),
    summary: `Query: "${String(payload["query"] ?? "").slice(0, 60)}"`,
  })
}

function handleRagStreamStarted(event: DbEvent, payload: Record<string, unknown>, state: ReconstructedState) {
  state.timeline.push({
    ts: event.ts,
    type: event.type,
    ...(event.userId != null ? { userId: event.userId } : {}),
    summary: `RAG query → col: ${payload["collection"] ?? "?"} | session: ${String(payload["sessionId"] ?? "").slice(0, 8)}`,
  })
}

function handleRagStreamCompleted(event: DbEvent, payload: Record<string, unknown>, state: ReconstructedState) {
  state.timeline.push({
    ts: event.ts,
    type: event.type,
    ...(event.userId != null ? { userId: event.userId } : {}),
    summary: `RAG completado → col: ${payload["collection"] ?? "?"} | ${payload["duration"] != null ? `${payload["duration"]}ms` : ""}`,
  })
}

function handleIngestionStarted(event: DbEvent, payload: Record<string, unknown>, state: ReconstructedState) {
  state.stats.ingestionCount++
  state.ingestionEvents.push({
    ts: event.ts,
    type: event.type,
    ...(payload["filename"] != null ? { filename: String(payload["filename"]) } : {}),
    ...(payload["collection"] != null ? { collection: String(payload["collection"]) } : {}),
    ...(payload["sourceId"] != null ? { sourceId: String(payload["sourceId"]) } : {}),
  })
  state.timeline.push({
    ts: event.ts,
    type: event.type,
    ...(event.userId != null ? { userId: event.userId } : {}),
    summary: `Ingesta iniciada: ${payload["filename"] ?? payload["name"] ?? "?"} → ${payload["collection"] ?? payload["collectionDest"] ?? "?"}`,
  })
}

function handleIngestionCompleted(event: DbEvent, payload: Record<string, unknown>, state: ReconstructedState) {
  state.ingestionEvents.push({
    ts: event.ts,
    type: event.type,
    ...(payload["filename"] != null ? { filename: String(payload["filename"]) } : {}),
    ...(payload["collection"] != null ? { collection: String(payload["collection"]) } : {}),
    ...(payload["sourceId"] != null ? { sourceId: String(payload["sourceId"]) } : {}),
  })
  state.timeline.push({
    ts: event.ts,
    type: event.type,
    ...(event.userId != null ? { userId: event.userId } : {}),
    summary: `Ingesta completada: ${payload["filename"] ?? payload["name"] ?? "?"} → ${payload["collection"] ?? payload["collectionDest"] ?? "?"}`,
  })
}

function handleIngestionFailed(event: DbEvent, payload: Record<string, unknown>, state: ReconstructedState) {
  const message = `Ingesta fallida: ${payload["filename"] ?? payload["name"] ?? "?"} — ${payload["error"] ?? ""}`
  state.ingestionEvents.push({
    ts: event.ts,
    type: event.type,
    ...(payload["filename"] != null ? { filename: String(payload["filename"]) } : {}),
    ...(payload["collection"] != null ? { collection: String(payload["collection"]) } : {}),
    ...(payload["error"] != null ? { error: String(payload["error"]) } : {}),
  })
  state.errors.push({
    ts: event.ts,
    type: event.type,
    message,
  })
  state.timeline.push({
    ts: event.ts,
    type: event.type,
    ...(event.userId != null ? { userId: event.userId } : {}),
    summary: message.slice(0, 80),
  })
}

function handleIngestionStalled(event: DbEvent, payload: Record<string, unknown>, state: ReconstructedState) {
  const message = `Ingesta estancada: ${payload["filename"] ?? "?"} (${payload["duration"] != null ? `${payload["duration"]}ms` : "sin duración"})`
  state.ingestionEvents.push({
    ts: event.ts,
    type: event.type,
    ...(payload["filename"] != null ? { filename: String(payload["filename"]) } : {}),
    ...(payload["collection"] != null ? { collection: String(payload["collection"]) } : {}),
  })
  state.errors.push({
    ts: event.ts,
    type: event.type,
    message,
  })
  state.timeline.push({
    ts: event.ts,
    type: event.type,
    ...(event.userId != null ? { userId: event.userId } : {}),
    summary: message.slice(0, 80),
  })
}

function handleError(event: DbEvent, payload: Record<string, unknown>, state: ReconstructedState) {
  state.errors.push({
    ts: event.ts,
    type: event.type,
    message: String(payload["error"] ?? payload["message"] ?? ""),
    suggestion: String(payload["suggestion"] ?? ""),
  })
  state.timeline.push({
    ts: event.ts,
    type: event.type,
    ...(event.userId != null ? { userId: event.userId } : {}),
    summary: `Error: ${String(payload["error"] ?? payload["message"] ?? "").slice(0, 80)}`,
  })
}

function handleUserCreatedOrUpdated(event: DbEvent, payload: Record<string, unknown>, state: ReconstructedState) {
  if (payload["userId"]) {
    const uid = Number(payload["userId"])
    state.users.set(uid, {
      email: String(payload["email"] ?? ""),
      role: String(payload["role"] ?? "user"),
      active: payload["active"] !== false,
    })
  }
}

function handleUserDeleted(event: DbEvent, payload: Record<string, unknown>, state: ReconstructedState) {
  if (payload["userId"]) {
    const uid = Number(payload["userId"])
    const u = state.users.get(uid)
    if (u) state.users.set(uid, { ...u, active: false })
  }
}

function handleDefault(event: DbEvent, payload: Record<string, unknown>, state: ReconstructedState) {
  state.timeline.push({
    ts: event.ts,
    type: event.type,
    ...(event.userId != null ? { userId: event.userId } : {}),
    summary: JSON.stringify(payload).slice(0, 80),
  })
}

type EventHandler = (event: DbEvent, payload: Record<string, unknown>, state: ReconstructedState) => void

const EVENT_HANDLERS: Record<string, EventHandler> = {
  "auth.login": handleAuthLogin,
  "rag.query": handleRagQuery,
  "rag.query_crossdoc": handleRagQuery,
  "rag.stream_started": handleRagStreamStarted,
  "rag.stream_completed": handleRagStreamCompleted,
  "rag.error": handleError,
  "system.error": handleError,
  "client.error": handleError,
  "ingestion.started": handleIngestionStarted,
  "ingestion.completed": handleIngestionCompleted,
  "ingestion.failed": handleIngestionFailed,
  "ingestion.stalled": handleIngestionStalled,
  "user.created": handleUserCreatedOrUpdated,
  "user.updated": handleUserCreatedOrUpdated,
  "user.deleted": handleUserDeleted,
}

// ── Función principal ───────────────────────────────────────────────────────

/**
 * Reconstruye el estado del sistema a partir de eventos (blackbox replay). Carga todos los eventos
 * en memoria — no usar en producción con cientos de miles de filas sin paginar. El estado refleja
 * el pasado observado, no predice el futuro.
 */
export function reconstructFromEvents(events: DbEvent[]): ReconstructedState {
  const state: ReconstructedState = {
    users: new Map(),
    ragQueries: [],
    ingestionEvents: [],
    errors: [],
    timeline: [],
    stats: {
      totalEvents: events.length,
      errorCount: 0,
      warnCount: 0,
      uniqueUsers: 0,
      ragQueryCount: 0,
      ingestionCount: 0,
    },
  }

  for (const event of events) {
    const payload = event.payload as Record<string, unknown>

    if (event.level === "ERROR" || event.level === "FATAL") state.stats.errorCount++
    if (event.level === "WARN") state.stats.warnCount++
    if (event.userId && !state.users.has(event.userId)) state.stats.uniqueUsers++

    const handler = EVENT_HANDLERS[event.type] ?? handleDefault
    handler(event, payload, state)
  }

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
    `Ingestas: ${state.stats.ingestionCount}`,
    "",
    "=== Timeline (más reciente primero) ===",
  ]

  for (const event of state.timeline.slice(0, 50)) {
    const ts = new Date(event.ts).toISOString().replace("T", " ").slice(0, 19)
    const user = event.userId ? `[user=${event.userId}]` : ""
    lines.push(`  ${ts}  ${event.type.padEnd(25)} ${user}  ${event.summary}`)
  }

  if (state.ingestionEvents.length > 0) {
    lines.push("", "=== Ingestas ===")
    for (const ing of state.ingestionEvents.slice(-20)) {
      const ts = new Date(ing.ts).toISOString().replace("T", " ").slice(0, 19)
      const file = ing.filename ?? "?"
      const col = ing.collection ?? "?"
      const errPart = ing.error ? ` — ERROR: ${ing.error.slice(0, 60)}` : ""
      lines.push(`  ${ts}  ${ing.type.padEnd(25)}  ${file} → ${col}${errPart}`)
    }
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
