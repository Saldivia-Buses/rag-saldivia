/**
 * DELETE /api/rag/collections/[name] — delete collection (admin only)
 *
 * Invalidates Redis cache after deletion so the UI reflects the change immediately.
 */

import { NextResponse } from "next/server"
import { ragFetch } from "@/lib/rag/client"
import { invalidateCollectionsCache } from "@/lib/rag/collections-cache"
import { requireAdmin, apiError, apiServerError } from "@/lib/api-utils"
import { CollectionNameSchema } from "@rag-saldivia/shared"

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ name: string }> }
) {
  const claims = await requireAdmin(request)
  if (claims instanceof NextResponse) return claims

  const { name } = await params
  const parsed = CollectionNameSchema.safeParse(name)
  if (!parsed.success) {
    return apiError(parsed.error.issues[0]?.message ?? "Nombre de colección inválido")
  }

  try {
    const res = await ragFetch(`/v1/collections/${encodeURIComponent(name)}`, {
      method: "DELETE",
    } as Parameters<typeof ragFetch>[1])
    if ("error" in res) {
      return apiError(`No se pudo eliminar la colección: ${res.error.message}`, 502)
    }
  } catch (error) {
    return apiServerError(error, `DELETE /api/rag/collections/${name}`, Number(claims.sub))
  }

  await invalidateCollectionsCache().catch(() => {})
  return NextResponse.json({ ok: true })
}
