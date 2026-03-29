/**
 * GET /api/collections/[name]/embeddings
 * Retorna la estructura de grafo de similitud entre documentos de una colección.
 * F3.46 — grafo de documentos.
 *
 * Si el RAG server no expone embeddings, retorna datos simulados para MVP.
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { ragFetch } from "@/lib/rag/client"

type GraphNode = { id: string; name: string; group?: number }
type GraphEdge = { source: string; target: string; weight: number }

export async function GET(
  request: Request,
  { params }: { params: Promise<{ name: string }> }
) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })

  const { name } = await params
  const collection = decodeURIComponent(name)

  try {
    // Intentar obtener documentos de la colección del RAG server
    const res = await ragFetch(`/v1/collections/${encodeURIComponent(collection)}/documents`)

    let docs: string[] = []
    if (!("error" in res) && res.ok) {
      try {
        const data = await res.json() as { documents?: string[] }
        docs = data.documents ?? []
      } catch { /* ignorar */ }
    }

    if (docs.length === 0) {
      // Datos simulados para MVP
      docs = ["doc-a.pdf", "doc-b.pdf", "doc-c.pdf", "doc-d.pdf", "doc-e.pdf"]
    }

    // Construir grafo con similitud simulada (en producción vendría de embeddings reales)
    const nodes: GraphNode[] = docs.map((d, i) => ({ id: d, name: d, group: Math.floor(i / 2) }))
    const edges: GraphEdge[] = []

    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        // Similitud simulada: mayor entre docs del mismo grupo
        const weight = nodes[i]?.group === nodes[j]?.group ? 0.7 + Math.random() * 0.25 : 0.2 + Math.random() * 0.3
        if (weight > 0.4) {
          edges.push({ source: nodes[i]!.id, target: nodes[j]!.id, weight })
        }
      }
    }

    return NextResponse.json({ ok: true, nodes, edges, collection })
  } catch {
    return NextResponse.json({ ok: false, error: "Error al obtener el grafo" }, { status: 500 })
  }
}
