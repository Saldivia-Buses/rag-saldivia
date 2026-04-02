/**
 * GET /api/rag/document/[name]
 * Proxy al RAG server para obtener un documento como PDF.
 * F3.40 — preview de doc inline.
 *
 * Si el Blueprint no expone el endpoint de documentos, retorna 404 con mensaje explicativo.
 */

import { extractClaims } from "@/lib/auth/jwt"
import { ragFetch } from "@/lib/rag/client"

export async function GET(
  request: Request,
  { params }: { params: Promise<{ name: string }> }
) {
  const claims = await extractClaims(request)
  if (!claims) return new Response("No autenticado", { status: 401 })

  const { name } = await params
  const docName = decodeURIComponent(name)

  try {
    // Intentar obtener el documento del RAG server
    const res = await ragFetch(`/v1/documents/${encodeURIComponent(docName)}`)

    if ("error" in res) {
      return new Response(
        JSON.stringify({ error: "Documento no disponible", note: "El Blueprint actual puede no exponer este endpoint" }),
        { status: 404, headers: { "Content-Type": "application/json" } }
      )
    }

    if (!res.ok) {
      return new Response("Documento no encontrado", { status: 404 })
    }

    // Reenviar el PDF al cliente
    const blob = await res.blob()
    return new Response(blob, {
      headers: {
        "Content-Type": "application/pdf",
        "Content-Disposition": `inline; filename="${docName}"`,
        "Cache-Control": "private, max-age=300",
      },
    })
  } catch {
    return new Response("Error al obtener el documento", { status: 503 })
  }
}
