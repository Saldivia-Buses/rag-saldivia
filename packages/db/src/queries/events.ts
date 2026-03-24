/**
 * Queries de eventos (black box).
 * Todo el sistema escribe aquí — es el registro inmutable de lo que pasó.
 */

import { eq, gte, lte, and, desc, asc } from "drizzle-orm"
import { getDb } from "../connection"
import { events } from "../schema"
import type { EventSource, LogLevel, EventType } from "@rag-saldivia/shared"

const db = getDb()

// Secuencia monotónica en memoria para esta instancia del proceso.
// Si el proceso reinicia, continúa desde el máximo en DB.
let _seq: number | null = null

async function nextSequence(): Promise<number> {
  if (_seq === null) {
    const last = await db.query.events.findFirst({
      orderBy: (e, { desc }) => [desc(e.sequence)],
    })
    _seq = (last?.sequence ?? 0) + 1
  } else {
    _seq++
  }
  return _seq
}

// ── Escritura ──────────────────────────────────────────────────────────────

export async function writeEvent(data: {
  source: EventSource
  level: LogLevel
  type: EventType
  userId?: number | null
  sessionId?: string | null
  payload?: Record<string, unknown>
}) {
  const sequence = await nextSequence()
  const [event] = await db
    .insert(events)
    .values({
      id: crypto.randomUUID(),
      ts: Date.now(),
      source: data.source,
      level: data.level,
      type: data.type,
      userId: data.userId ?? null,
      sessionId: data.sessionId ?? null,
      payload: data.payload ?? {},
      sequence,
    })
    .returning()
  return event
}

// ── Lectura / Audit ────────────────────────────────────────────────────────

export async function queryEvents(filters: {
  fromTs?: number
  toTs?: number
  source?: EventSource
  level?: LogLevel
  type?: EventType
  userId?: number
  limit?: number
  offset?: number
  order?: "asc" | "desc"
}) {
  const conditions = []

  if (filters.fromTs) conditions.push(gte(events.ts, filters.fromTs))
  if (filters.toTs) conditions.push(lte(events.ts, filters.toTs))
  if (filters.source) conditions.push(eq(events.source, filters.source))
  if (filters.level) conditions.push(eq(events.level, filters.level))
  if (filters.type) conditions.push(eq(events.type, filters.type))
  if (filters.userId) conditions.push(eq(events.userId, filters.userId))

  return db.query.events.findMany({
    where: conditions.length > 0 ? and(...conditions) : undefined,
    orderBy: filters.order === "asc"
      ? (e, { asc }) => [asc(e.sequence)]
      : (e, { desc }) => [desc(e.sequence)],
    limit: filters.limit ?? 100,
    offset: filters.offset ?? 0,
  })
}

// ── Black Box: Reconstruct ─────────────────────────────────────────────────

export async function getEventsForReplay(fromTs: number, toTs?: number) {
  const conditions = [gte(events.ts, fromTs)]
  if (toTs) conditions.push(lte(events.ts, toTs))

  return db.query.events.findMany({
    where: and(...conditions),
    orderBy: (e, { asc }) => [asc(e.sequence)],
  })
}
