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

  // Admins ven todas las colecciones del RAG
  // Otros usuarios ven solo las que tienen permisos
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
