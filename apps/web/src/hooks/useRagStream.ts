"use client"

import { useState, useRef } from "react"

export type ChatPhase = "idle" | "streaming" | "done" | "error"

type StreamMessage = { role: "user" | "assistant"; content: string }

type StreamResult = {
  fullContent: string
  sources: unknown[]
}

type UseRagStreamOptions = {
  sessionId: string
  collection: string
  focusMode?: string
  onDelta: (fullContent: string) => void
  onSources: (sources: unknown[]) => void
  onError: (message: string) => void
}

/**
 * Encapsula fetch + lectura del stream SSE + abort controller.
 * ChatInterface solo maneja estado de mensajes; este hook maneja el transporte.
 */
export function useRagStream({
  sessionId,
  collection,
  focusMode,
  onDelta,
  onSources,
  onError,
}: UseRagStreamOptions) {
  const [phase, setPhase] = useState<ChatPhase>("idle")
  const abortRef = useRef<AbortController | null>(null)

  function abort() {
    abortRef.current?.abort()
  }

  async function stream(messages: StreamMessage[]): Promise<StreamResult | null> {
    setPhase("streaming")
    abortRef.current = new AbortController()

    let fullContent = ""
    let sources: unknown[] = []

    try {
      const res = await fetch("/api/rag/generate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          messages: messages.map((m) => ({ role: m.role, content: m.content })),
          collection_name: collection,
          session_id: sessionId,
          use_knowledge_base: true,
          focus_mode: focusMode ?? "detallado",
        }),
        signal: abortRef.current.signal,
      })

      if (!res.ok) {
        const data = await res.json().catch(() => ({ error: "Error desconocido" }))
        throw new Error(data.error ?? `Error ${res.status}`)
      }

      if (!res.body) throw new Error("No stream")

      const reader = res.body.getReader()
      const decoder = new TextDecoder()

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        for (const line of chunk.split("\n")) {
          if (!line.startsWith("data: ")) continue
          const data = line.slice(6).trim()
          if (data === "[DONE]") continue

          try {
            const parsed = JSON.parse(data)
            const delta = parsed.choices?.[0]?.delta?.content ?? ""
            if (delta) {
              fullContent += delta
              onDelta(fullContent)
            }
            const srcData = parsed.choices?.[0]?.delta?.sources
            if (srcData) {
              sources = srcData
              onSources(sources)
            }
          } catch {
            // ignorar líneas malformadas del stream
          }
        }
      }

      setPhase("done")
      return { fullContent, sources }
    } catch (err) {
      if (err instanceof Error && err.name === "AbortError") {
        setPhase("idle")
        return null
      }

      const message = err instanceof Error ? err.message : String(err)
      onError(message)
      setPhase("error")
      return null
    }
  }

  return { phase, stream, abort }
}
