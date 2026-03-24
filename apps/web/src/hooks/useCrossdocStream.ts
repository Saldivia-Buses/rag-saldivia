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

const MAX_PARALLEL = 6
const REPETITION_WINDOW = 60
const REPETITION_THRESHOLD = 3
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

function detectRepetition(text: string): number {
  if (text.length <= REPETITION_WINDOW * REPETITION_THRESHOLD) return -1
  const tail = text.slice(-REPETITION_WINDOW)
  const preceding = text.slice(-(REPETITION_WINDOW * (REPETITION_THRESHOLD + 1)), -REPETITION_WINDOW)
  if (preceding.split(tail).length - 1 >= REPETITION_THRESHOLD - 1) {
    const firstIdx = text.indexOf(tail)
    if (firstIdx > 0 && firstIdx < text.length - REPETITION_WINDOW) {
      return firstIdx + REPETITION_WINDOW
    }
  }
  return -1
}

async function collectStream(response: Response): Promise<string> {
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
            if (token && token.length < 500) text += token
          } catch { /* skip */ }
        }
        const truncIdx = detectRepetition(text)
        if (truncIdx > 0) { text = text.slice(0, truncIdx); break }
        if (text.length > MAX_RESPONSE_CHARS) break
      }
    }
  } else {
    const json = await response.json()
    text = json?.choices?.[0]?.message?.content ?? ""
  }
  return text
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
              const text = await collectStream(resp)
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
                const text = await collectStream(resp)
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

        const synthesis = await collectStream(synthResp)
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
