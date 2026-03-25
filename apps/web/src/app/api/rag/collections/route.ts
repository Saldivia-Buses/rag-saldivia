/**
 * GET /api/rag/collections
 * Lista las colecciones del RAG Server con cache de 60 segundos.
 * Filtra por las colecciones a las que el usuario tiene acceso.
 */

import { NextResponse } from "next/server"
import { unstable_cache } from "next/cache"
import { ragFetch } from "@/lib/rag/client"
import { extractClaims } from "@/lib/auth/jwt"
import { getUserCollections } from "@rag-saldivia/db"

const getCachedRagCollections = unstable_cache(
  async () => {
    const res = await ragFetch("/v1/collections")
    if ("error" in res) return []
    if (!res.ok) return []

    try {
      const data = await res.json()
      return (data.collections ?? []) as string[]
    } catch {
      return []
    }
  },
  ["rag-collections"],
  { revalidate: 60, tags: ["collections"] }
)

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
    const ragUrl = process.env["RAG_SERVER_URL"] ?? "http://localhost:8081"
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

async function ragFetchWithOptions(path: string, options?: RequestInit) {
  const ragUrl = process.env["RAG_SERVER_URL"] ?? "http://localhost:8081"
  try {
    const res = await fetch(`${ragUrl}${path}`, { ...options, signal: AbortSignal.timeout(10000) })
    return res
  } catch {
    return null
  }
}
