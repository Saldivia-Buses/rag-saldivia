/**
 * POST /api/extract
 * Extrae campos estructurados de una colección.
 * F3.50 — extracción estructurada a tabla.
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { canAccessCollection } from "@rag-saldivia/db"
import { ragFetch } from "@/lib/rag/client"

type Field = { name: string; description: string }
type ExtractionRow = Record<string, string>

const MOCK_RAG = process.env["MOCK_RAG"] === "true"

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })

  const userId = Number(claims.sub)
  const body = await request.json().catch(() => null) as {
    collection?: string
    fields?: Field[]
  } | null

  if (!body?.collection || !body.fields?.length) {
    return NextResponse.json({ ok: false, error: "collection y fields son requeridos" }, { status: 400 })
  }

  const hasAccess = await canAccessCollection(userId, body.collection, "read")
  if (!hasAccess) {
    return NextResponse.json({ ok: false, error: `Sin acceso a la colección '${body.collection}'` }, { status: 403 })
  }

  if (MOCK_RAG) {
    // Datos simulados en modo mock
    const mockData: ExtractionRow[] = [
      { documento: "contrato-2024.pdf", ...Object.fromEntries(body.fields.map((f) => [f.name, `Valor de ${f.name}`])) },
      { documento: "acuerdo-marco.pdf", ...Object.fromEntries(body.fields.map((f) => [f.name, `Otro valor de ${f.name}`])) },
    ]
    return NextResponse.json({ ok: true, data: mockData, fields: body.fields.map((f) => f.name) })
  }

  try {
    // Obtener lista de documentos de la colección
    const docsRes = await ragFetch(`/v1/collections/${encodeURIComponent(body.collection)}/documents`)
    let docs: string[] = []
    if (!("error" in docsRes) && docsRes.ok) {
      const data = await docsRes.json() as { documents?: string[] }
      docs = data.documents ?? []
    }

    if (docs.length === 0) {
      return NextResponse.json({ ok: true, data: [], fields: body.fields.map((f) => f.name), note: "No se encontraron documentos en la colección" })
    }

    const results: ExtractionRow[] = []
    const fieldsList = body.fields.map((f) => `- ${f.name}: ${f.description}`).join("\n")
    const prompt = `Extract ONLY the following fields from the document as JSON. Return ONLY valid JSON with these exact keys. If a field is not found, use null.\nFields:\n${fieldsList}`

    for (const doc of docs.slice(0, 20)) { // límite de 20 docs
      try {
        const ragRes = await ragFetch("/v1/chat/completions", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            messages: [
              { role: "system", content: `You are extracting structured data from document: ${doc}` },
              { role: "user", content: prompt },
            ],
            collection_name: body.collection,
            use_knowledge_base: true,
            max_tokens: 500,
          }),
        } as Parameters<typeof ragFetch>[1])

        if (!("error" in ragRes) && ragRes.ok) {
          const data = await ragRes.json() as { choices?: Array<{ message?: { content?: string } }> }
          const content = data.choices?.[0]?.message?.content ?? "{}"
          const jsonMatch = content.match(/\{[\s\S]*\}/)
          if (jsonMatch) {
            const extracted = JSON.parse(jsonMatch[0]) as Record<string, string>
            results.push({ documento: doc, ...extracted })
          }
        }
      } catch { /* ignorar errores por doc */ }
    }

    return NextResponse.json({ ok: true, data: results, fields: ["documento", ...body.fields.map((f) => f.name)] })
  } catch (err) {
    return NextResponse.json({ ok: false, error: String(err) }, { status: 500 })
  }
}
