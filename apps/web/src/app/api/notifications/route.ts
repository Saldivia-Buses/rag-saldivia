/**
 * GET /api/notifications
 *
 * Retorna eventos recientes no vistos del tipo:
 * - ingestion.completed
 * - ingestion.error
 * - user.created (solo para admins)
 *
 * "No vistos" se gestiona en el cliente via localStorage["seen_notification_ids"].
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb } from "@rag-saldivia/db"
import { events } from "@rag-saldivia/db"
import { inArray, desc } from "drizzle-orm"

const NOTIFICATION_TYPES = ["ingestion.completed", "ingestion.error", "user.created"]

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) {
    return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  }

  const role = claims.role as string
  const db = getDb()

  const types = role === "admin"
    ? NOTIFICATION_TYPES
    : NOTIFICATION_TYPES.filter((t) => t !== "user.created")

  const rows = await db
    .select({
      id: events.id,
      type: events.type,
      ts: events.ts,
      payload: events.payload,
    })
    .from(events)
    .where(inArray(events.type, types))
    .orderBy(desc(events.ts))
    .limit(20)

  return NextResponse.json({ ok: true, notifications: rows })
}
