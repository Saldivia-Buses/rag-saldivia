/**
 * POST /api/admin/db/seed — aplicar seed de desarrollo (CLI)
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb, users, areas, userAreas, areaCollections, createUser } from "@rag-saldivia/db"

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  const db = getDb()
  const now = Date.now()

  // Upsert área General
  const existingAreas = await db.select().from(areas).limit(1)
  let areaId = existingAreas[0]?.id
  if (!areaId) {
    const inserted = await db.insert(areas).values({
      name: "General",
      description: "Área general de acceso",
      createdAt: now,
    }).returning()
    const area = inserted[0]
    if (!area) throw new Error("No se pudo crear el área General")
    areaId = area.id

    await db.insert(areaCollections).values({
      areaId,
      collectionName: "tecpia",
      permission: "admin",
    }).onConflictDoNothing()
  }

  // Upsert admin solo si no hay usuarios
  const existingUsers = await db.select().from(users).limit(1)
  if (existingUsers.length === 0) {
    const admin = await createUser({
      email: "admin@localhost",
      name: "Admin (dev)",
      password: "changeme",
      role: "admin",
    })
    await db.insert(userAreas).values({ userId: admin.id, areaId }).onConflictDoNothing()
  }

  return NextResponse.json({ ok: true, message: "Seed aplicado" })
}
