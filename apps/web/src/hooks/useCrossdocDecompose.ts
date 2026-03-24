/**
 * useCrossdocDecompose — portado de patches/frontend/new/useCrossdocDecompose.ts
 *
 * Cambios respecto al original:
 * - URL: /api/v1/generate → /api/rag/generate
 * - Eliminada dependencia de useSettingsStore del blueprint (usa parámetros directos)
 * - Mantiene toda la lógica de Jaccard dedup y follow-up generation
 */

import { useCallback } from "react"

const JACCARD_THRESHOLD = 0.65
const RAG_URL = "/api/rag/generate"

function jaccard(a: string, b: string): number {
  const setA = new Set(a.toLowerCase().split(/\s+/))
  const setB = new Set(b.toLowerCase().split(/\s+/))
  const intersection = new Set([...setA].filter((x) => setB.has(x)))
  const union = new Set([...setA, ...setB])
  return union.size === 0 ? 0 : intersection.size / union.size
}

function dedup(queries: string[]): string[] {
  const result: string[] = []
  for (const q of queries) {
    if (!result.some((existing) => jaccard(existing, q) >= JACCARD_THRESHOLD)) {
      result.push(q)
    }
  }
  return result
}

async function collectSseText(response: Response): Promise<string> {
  let text = ""
  if (response.headers.get("content-type")?.includes("text/event-stream")) {
    const reader = response.body?.getReader()
    const decoder = new TextDecoder()
    if (reader) {
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        const chunk = decoder.decode(value, { stream: true })
        for (const line of chunk.split("\n")) {
          if (!line.startsWith("data: ")) continue
          try {
            const data = JSON.parse(line.slice(6))
            const token = data?.choices?.[0]?.delta?.content
            if (token) text += token
          } catch { /* skip */ }
        }
      }
    }
  } else {
    const json = await response.json()
    text = json?.choices?.[0]?.message?.content ?? ""
  }
  return text
}

export type CrossdocDecomposeOptions = {
  maxSubQueries?: number
  model?: string
}

export function useCrossdocDecompose() {
  const decompose = useCallback(
    async (question: string, opts: CrossdocDecomposeOptions = {}): Promise<string[]> => {
      const maxSub = opts.maxSubQueries ?? 0

      const prompt = `You are a search query decomposer for a technical document retrieval system.

Given the user's question, generate multiple retrieval-focused sub-queries. Each sub-query should:
- Target a SPECIFIC product, component, or technical specification
- Use generic catalog/manual terminology (not user-specific context)
- Be at most 15 words
- Be independent — each should retrieve different documents

Return ONLY the sub-queries, one per line. No numbering, no explanations.

User question: ${question}`

      const response = await fetch(RAG_URL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          messages: [{ role: "user", content: prompt }],
          use_knowledge_base: false,
          max_tokens: 2048,
        }),
      })

      if (!response.ok) throw new Error(`Decompose failed: ${response.status}`)

      const text = await collectSseText(response)

      let subQueries = text
        .split("\n")
        .map((line) => line.replace(/^\d+[\.\)]\s*/, "").trim())
        .filter((line) => line.length > 5 && line.length < 200)

      subQueries = dedup(subQueries)

      if (maxSub > 0 && subQueries.length > maxSub) {
        subQueries = subQueries.slice(0, maxSub)
      }

      return subQueries
    },
    []
  )

  const generateFollowUps = useCallback(
    async (failedQueries: string[]): Promise<string[]> => {
      const prompt = `These search queries returned no useful results:
${failedQueries.map((q) => `- ${q}`).join("\n")}

Generate alternative queries using synonyms, broader terms, or different technical vocabulary.
One query per line, no numbering.`

      const response = await fetch(RAG_URL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          messages: [{ role: "user", content: prompt }],
          use_knowledge_base: false,
          max_tokens: 1024,
        }),
      })

      if (!response.ok) return []

      const text = await collectSseText(response)
      return text
        .split("\n")
        .map((line) => line.replace(/^\d+[\.\)]\s*/, "").trim())
        .filter((line) => line.length > 5 && line.length < 200)
    },
    []
  )

  return { decompose, generateFollowUps }
}
