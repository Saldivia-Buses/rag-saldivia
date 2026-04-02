/**
 * Adapter: transforma el SSE del RAG Server NVIDIA al protocolo AI SDK Data Stream.
 *
 * El RAG server emite SSE en formato OpenAI chat completions:
 *   data: {"choices":[{"delta":{"content":"token"}}]}
 *   data: {"choices":[{"delta":{"sources":[...]}}]}
 *   data: [DONE]
 *
 * El AI SDK espera su propio protocolo con text-start, text-delta, data parts, etc.
 * Este adapter lee el primero y escribe el segundo.
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

type RagDataTypes = {
  sources: { citations: Citation[] }
}

export type RagUIMessage = UIMessage<unknown, RagDataTypes>

const TEXT_PART_ID = "msg-text"

/**
 * Crea una Response en formato AI SDK Data Stream a partir de un ReadableStream
 * SSE del RAG Server (NVIDIA o OpenRouter).
 */
export function createRagStreamResponse(ragStream: ReadableStream<Uint8Array>) {
  const stream = createUIMessageStream<UIMessage<unknown, RagDataTypes>>({
    execute: async ({ writer }) => {
      const reader = ragStream.getReader()
      const decoder = new TextDecoder()
      let buffer = ""
      let textStarted = false

      try {
        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split("\n")
          buffer = lines.pop() ?? ""

          for (const line of lines) {
            // Extraer token de texto
            const token = parseSseLine(line)
            if (token) {
              if (!textStarted) {
                writer.write({ type: "text-start", id: TEXT_PART_ID })
                textStarted = true
              }
              writer.write({ type: "text-delta", delta: token, id: TEXT_PART_ID })
            }

            // Extraer citations de delta.sources
            if (line.startsWith("data: ")) {
              const rawData = line.slice(6).trim()
              if (rawData && rawData !== "[DONE]") {
                try {
                  const parsed = JSON.parse(rawData) as NvidiaSseEvent
                  const srcData = parsed.choices?.[0]?.delta?.sources
                  if (srcData) {
                    const result = CitationSchema.array().safeParse(srcData)
                    if (result.success) {
                      writer.write({
                        type: "data-sources",
                        data: { citations: result.data },
                      })
                    }
                  }
                } catch {
                  // Ignorar líneas malformadas
                }
              }
            }
          }
        }

        // Procesar buffer restante
        if (buffer) {
          const token = parseSseLine(buffer)
          if (token) {
            if (!textStarted) {
              writer.write({ type: "text-start", id: TEXT_PART_ID })
              textStarted = true
            }
            writer.write({ type: "text-delta", delta: token, id: TEXT_PART_ID })
          }
        }

        // Cerrar el text part si se abrió
        if (textStarted) {
          writer.write({ type: "text-end", id: TEXT_PART_ID })
        }
      } finally {
        reader.releaseLock()
      }
    },
  })

  return createUIMessageStreamResponse({ stream })
}
