/**
 * DELETE /api/rag/collections/[name] — eliminar colección (solo admin)
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { ragFetch } from "@/lib/rag/client"

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ name: string }> }
) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Solo admins" }, { status: 403 })

  const { name } = await params

  try {
    await ragFetch(`/v1/collections/${encodeURIComponent(name)}`, { method: "DELETE" } as Parameters<typeof ragFetch>[1])
  } catch {
    // En modo mock: simular éxito
  }

  return NextResponse.json({ ok: true })
}
