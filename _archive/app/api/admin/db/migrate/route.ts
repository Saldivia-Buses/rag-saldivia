/**
 * POST /api/admin/db/migrate — verificar/inicializar DB (CLI)
 *
 * La DB se inicializa automáticamente al arrancar el servidor via getDb().
 * Este endpoint confirma que la conexión está activa.
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb } from "@rag-saldivia/db"

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  try {
    // Verificar que la conexión está activa
    getDb()
    return NextResponse.json({ ok: true, message: "Base de datos inicializada y conectada" })
  } catch (err) {
    return NextResponse.json({ ok: false, error: String(err) }, { status: 500 })
  }
}
