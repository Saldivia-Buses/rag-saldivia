/**
 * POST /api/admin/db/reset — borrar y recrear la DB (CLI) — DESTRUCTIVO
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import {
  getDb, users, areas, userAreas, areaCollections,
  chatSessions, chatMessages, messageFeedback,
  ingestionJobs, ingestionAlerts, ingestionQueue, events,
} from "@rag-saldivia/db"

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  if (process.env["NODE_ENV"] === "production") {
    return NextResponse.json({ ok: false, error: "No se puede resetear en producción" }, { status: 403 })
  }

  try {
    const db = getDb()

    // Truncar todas las tablas en orden seguro
    await db.delete(events)
    await db.delete(messageFeedback)
    await db.delete(chatMessages)
    await db.delete(chatSessions)
    await db.delete(ingestionAlerts)
    await db.delete(ingestionQueue)
    await db.delete(ingestionJobs)
    await db.delete(areaCollections)
    await db.delete(userAreas)
    await db.delete(users)
    await db.delete(areas)

    return NextResponse.json({ ok: true, message: "DB reseteada" })
  } catch (err) {
    return NextResponse.json({ ok: false, error: String(err) }, { status: 500 })
  }
}
