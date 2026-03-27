/**
 * useCrossdocStream — portado de patches/frontend/new/useCrossdocStream.ts
 *
 * Cambios respecto al original:
 * - URL: /api/v1/generate → /api/rag/generate
 * - Eliminada dependencia de useSettingsStore y useCollectionsStore del blueprint
 * - Usa parámetros directos (collection, settings)
 * - Mantiene toda la lógica de pipeline: decompose → parallel queries → follow-ups → synthesis
 */

import { useCallback, useRef } from "react"
import { useCrossdocDecompose } from "./useCrossdocDecompose.js"
import { collectSseText } from "@/lib/rag/stream"

const MAX_PARALLEL = 6
const MAX_RESPONSE_CHARS = 15000
const RAG_URL = "/api/rag/generate"

export type CrossdocPhase =
  | "decomposing"
  | "querying"
  | "following-up"
  | "synthesizing"
  | "done"
  | "error"

export type CrossdocProgress = {
  phase: CrossdocPhase
  subQueries?: string[]
  completed?: number
  total?: number
  synthesis?: string
  error?: string
}

export type CrossdocSettings = {
  collection: string
  maxSubQueries?: number
  enableFollowUps?: boolean
  vdbTopK?: number
  rerankerTopK?: number
}


function hasUsefulData(text: string): boolean {
  const trimmed = text.trim()
  if (trimmed.length < 3) return false
  const emptyPatterns = [
    /^(no|sin)\s+(information|data|results|context)/i,
    /^out of context$/i,
    /^i (cannot|can't|don't)/i,
    /^$/,
  ]
  return !emptyPatterns.some((p) => p.test(trimmed))
}

export function useCrossdocStream() {
  const { decompose, generateFollowUps } = useCrossdocDecompose()
  const abortRef = useRef<AbortController | null>(null)

  const execute = useCallback(
    async (
      question: string,
      settings: CrossdocSettings,
      onProgress: (progress: CrossdocProgress) => void
    ): Promise<string> => {
      abortRef.current = new AbortController()
      const signal = abortRef.current.signal

      try {
        // Phase 1: Decompose
        onProgress({ phase: "decomposing" })
        const subQueries = await decompose(question, {
          maxSubQueries: settings.maxSubQueries ?? 0,
        })
        onProgress({ phase: "querying", subQueries, completed: 0, total: subQueries.length })

        // Phase 2: Parallel RAG queries
        type Result = { query: string; response: string; success: boolean }
        const results: Result[] = []
        const batches: string[][] = []
        for (let i = 0; i < subQueries.length; i += MAX_PARALLEL) {
          batches.push(subQueries.slice(i, i + MAX_PARALLEL))
        }

        let completed = 0
        for (const batch of batches) {
          if (signal.aborted) break

          const batchResults = await Promise.allSettled(
            batch.map(async (query) => {
              const resp = await fetch(RAG_URL, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                signal,
                body: JSON.stringify({
                  messages: [{ role: "user", content: query }],
                  use_knowledge_base: true,
                  collection_name: settings.collection,
                  vdb_top_k: settings.vdbTopK ?? 10,
                  reranker_top_k: settings.rerankerTopK ?? 5,
                  max_tokens: 2048,
                }),
              })
              const text = await collectSseText(resp, { maxChars: MAX_RESPONSE_CHARS, detectRepetition: true })
              return { query, response: text, success: hasUsefulData(text) }
            })
          )

          for (const r of batchResults) {
            if (r.status === "fulfilled") results.push(r.value)
            else results.push({ query: "", response: "", success: false })
            completed++
          }
          onProgress({ phase: "querying", subQueries, completed, total: subQueries.length })
        }

        // Phase 3: Follow-ups
        if (settings.enableFollowUps !== false) {
          const failed = results.filter((r) => !r.success).map((r) => r.query).filter(Boolean)
          if (failed.length > 0) {
            onProgress({ phase: "following-up", subQueries, completed, total: subQueries.length })
            const alternatives = await generateFollowUps(failed)
            for (const alt of alternatives.slice(0, 10)) {
              if (signal.aborted) break
              try {
                const resp = await fetch(RAG_URL, {
                  method: "POST",
                  headers: { "Content-Type": "application/json" },
                  signal,
                  body: JSON.stringify({
                    messages: [{ role: "user", content: alt }],
                    use_knowledge_base: true,
                    collection_name: settings.collection,
                    max_tokens: 2048,
                  }),
                })
                const text = await collectSseText(resp, { maxChars: MAX_RESPONSE_CHARS, detectRepetition: true })
                if (hasUsefulData(text)) {
                  results.push({ query: alt, response: text, success: true })
                }
              } catch { /* skip */ }
            }
          }
        }

        // Phase 4: Synthesize
        onProgress({ phase: "synthesizing" })
        const successResults = results.filter((r) => r.success)
        const context = successResults
          .map((r, i) => `[Sub-query ${i + 1}: "${r.query}"]\n${r.response}`)
          .join("\n\n---\n\n")

        const synthesisPrompt = `You are a senior engineer writing a comprehensive technical answer.

Based on the following retrieval results from multiple sub-queries, write a single unified answer to the user's original question.

Rules:
- Cite sources when possible (mention which sub-query or document the info came from)
- Include specific numbers, measurements, and technical specifications
- Be thorough but concise — cover all relevant information
- Use professional technical language

Original question: ${question}

Retrieval results:
${context}`

        const synthResp = await fetch(RAG_URL, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            messages: [{ role: "user", content: synthesisPrompt }],
            use_knowledge_base: false,
            max_tokens: 4096,
          }),
        })

        const synthesis = await collectSseText(synthResp)
        onProgress({ phase: "done", synthesis })
        return synthesis
      } catch (error) {
        if ((error as Error).name === "AbortError") return ""
        onProgress({ phase: "error", error: (error as Error).message ?? "Crossdoc failed" })
        return ""
      }
    },
    [decompose, generateFollowUps]
  )

  const abort = useCallback(() => {
    abortRef.current?.abort()
  }, [])

  return { execute, abort }
}
