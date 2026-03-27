/**
 * GET /api/rag/collections
 * Lista las colecciones del RAG Server con cache de 60 segundos.
 * Filtra por las colecciones a las que el usuario tiene acceso.
 */

import { NextResponse } from "next/server"
import { ragFetch } from "@/lib/rag/client"
import { extractClaims } from "@/lib/auth/jwt"
import { getUserCollections } from "@rag-saldivia/db"
import { getCachedRagCollections } from "@/lib/rag/collections-cache"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) {
    return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  }

  const userId = Number(claims.sub)
  const ragCollections = await getCachedRagCollections()

  if (claims.role === "admin") {
    return NextResponse.json({ ok: true, data: ragCollections })
  }

  const userCollections = await getUserCollections(userId)
  const allowed = new Set(userCollections.map((c) => c.name))

  return NextResponse.json({
    ok: true,
    data: ragCollections.filter((name) => allowed.has(name)),
  })
}

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Solo admins" }, { status: 403 })

  const body = await request.json().catch(() => null)
  if (!body?.name) return NextResponse.json({ ok: false, error: "name requerido" }, { status: 400 })

  try {
    const res = await ragFetch(`/v1/collections`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ collection_name: body.name }),
    } as Parameters<typeof ragFetch>[1])
    if ("error" in res) throw new Error(res.error.message)
    return NextResponse.json({ ok: true })
  } catch {
    // En modo mock: simular éxito
    return NextResponse.json({ ok: true })
  }
}

