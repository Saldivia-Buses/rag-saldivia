/**
 * Utilidades compartidas para leer streams SSE del RAG Server.
 *
 * Centraliza la lógica de reader + TextDecoder + parseo de líneas que
 * estaba duplicada en useRagStream, useCrossdocStream, useCrossdocDecompose,
 * slack/route.ts y teams/route.ts.
 */

const REPETITION_WINDOW = 60
const REPETITION_THRESHOLD = 3

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

/**
 * Parsea una línea SSE "data: {...}" y extrae el token de contenido.
 * Retorna null para [DONE], líneas malformadas o sin contenido de texto.
 */
export function parseSseLine(line: string): string | null {
  if (!line.startsWith("data: ")) return null
  const data = line.slice(6).trim()
  if (data === "[DONE]") return null
  try {
    const parsed = JSON.parse(data) as {
      choices?: Array<{ delta?: { content?: string } }>
    }
    return parsed.choices?.[0]?.delta?.content ?? null
  } catch {
    return null
  }
}

/**
 * Yields tokens individuales de contenido a medida que llegan del ReadableStream.
 * Incluye buffering de líneas parciales para manejar chunks que cortan en mitad de una línea SSE.
 */
export async function* readSseTokens(
  body: ReadableStream<Uint8Array>
): AsyncGenerator<string> {
  const reader = body.getReader()
  const decoder = new TextDecoder()
  let buffer = ""
  try {
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split("\n")
      buffer = lines.pop() ?? "" // línea incompleta — queda en buffer para el próximo chunk
      for (const line of lines) {
        const token = parseSseLine(line)
        if (token) yield token
      }
    }
    // Procesar cualquier contenido restante en el buffer
    if (buffer) {
      const token = parseSseLine(buffer)
      if (token) yield token
    }
  } finally {
    reader.releaseLock()
  }
}

/**
 * Acumula todo el texto del stream SSE en un string.
 * Maneja tanto respuestas SSE (text/event-stream) como JSON estándar.
 *
 * @param response - La Response del fetch al RAG Server
 * @param options.maxChars - Truncar si el texto supera este límite
 * @param options.detectRepetition - Cortar si se detecta texto repetitivo (útil para modelos con alucinaciones en loop)
 */
export async function collectSseText(
  response: Response,
  options?: { maxChars?: number; detectRepetition?: boolean }
): Promise<string> {
  let text = ""

  if (response.headers.get("content-type")?.includes("text/event-stream") && response.body) {
    for await (const token of readSseTokens(response.body)) {
      text += token

      if (options?.detectRepetition) {
        const truncIdx = detectRepetition(text)
        if (truncIdx > 0) {
          text = text.slice(0, truncIdx)
          break
        }
      }

      if (options?.maxChars && text.length > options.maxChars) break
    }
  } else {
    const json = (await response.json()) as {
      choices?: Array<{ message?: { content?: string } }>
    }
    text = json.choices?.[0]?.message?.content ?? ""
  }

  return text
}
