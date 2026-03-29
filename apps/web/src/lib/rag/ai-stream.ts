/**
 * Adapter: transforma el SSE del RAG Server NVIDIA al protocolo AI SDK Data Stream.
 *
 * El RAG server emite SSE en formato OpenAI chat completions:
 *   data: {"choices":[{"delta":{"content":"token"}}]}
 *   data: {"choices":[{"delta":{"sources":[...]}}]}
 *   data: [DONE]
 *
 * El AI SDK espera su propio protocolo con text-delta, data parts, etc.
 * Este adapter lee el primero y escribe el segundo.
 *
 * Citations se pasan como `data-sources` parts (custom data type).
 */

import {
  createUIMessageStream,
  createUIMessageStreamResponse,
  type UIMessage,
} from "ai"
import { parseSseLine } from "./stream"
import { CitationSchema, type Citation } from "@rag-saldivia/shared"

type NvidiaSseEvent = {
  choices?: Array<{
    delta?: {
      content?: string
      sources?: unknown
    }
  }>
}

// Custom data types para nuestro stream — citations del RAG
type RagDataTypes = {
  sources: { citations: Citation[] }
}

export type RagUIMessage = UIMessage<unknown, RagDataTypes>

type RagWriter = Parameters<
  Parameters<typeof createUIMessageStream<UIMessage<unknown, RagDataTypes>>>[0]["execute"]
>[0]["writer"]

/**
 * Crea una Response en formato AI SDK Data Stream a partir de un ReadableStream
 * SSE del RAG Server NVIDIA.
 */
export function createRagStreamResponse(ragStream: ReadableStream<Uint8Array>) {
  const stream = createUIMessageStream<UIMessage<unknown, RagDataTypes>>({
    execute: async ({ writer }) => {
      const reader = ragStream.getReader()
      const decoder = new TextDecoder()
      let buffer = ""

      try {
        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split("\n")
          buffer = lines.pop() ?? ""

          for (const line of lines) {
            processLine(line, writer)
          }
        }

        if (buffer) {
          processLine(buffer, writer)
        }
      } finally {
        reader.releaseLock()
      }
    },
  })

  return createUIMessageStreamResponse({ stream })
}

let partIdCounter = 0

function processLine(line: string, writer: RagWriter) {
  // Extraer token de texto
  const token = parseSseLine(line)
  if (token) {
    writer.write({ type: "text-delta", delta: token, id: `t-${partIdCounter++}` })
  }

  // Extraer citations de delta.sources
  if (!line.startsWith("data: ")) return
  const rawData = line.slice(6).trim()
  if (!rawData || rawData === "[DONE]") return

  try {
    const parsed = JSON.parse(rawData) as NvidiaSseEvent
    const srcData = parsed.choices?.[0]?.delta?.sources
    if (!srcData) return

    const result = CitationSchema.array().safeParse(srcData)
    if (!result.success) return

    writer.write({
      type: "data-sources",
      data: { citations: result.data },
    })
  } catch {
    // Ignorar líneas malformadas
  }
}
