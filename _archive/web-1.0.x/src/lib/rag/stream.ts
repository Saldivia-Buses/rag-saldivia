/**
 * Shared SSE stream reader for the NVIDIA RAG Server.
 *
 * The RAG server (port 8081) streams responses as Server-Sent Events in
 * OpenAI-compatible format: `data: {"choices":[{"delta":{"content":"..."}}]}`
 *
 * This module provides three layers of abstraction:
 *   1. `parseSseLine` — single line -> token string (or null)
 *   2. `readSseTokens` — ReadableStream -> AsyncGenerator of tokens
 *   3. `collectSseText` — full Response -> accumulated string
 *
 * Used by: ai-stream.ts (AI SDK adapter), webhook integrations (Slack/Teams),
 *          and any route that proxies RAG responses.
 * Depends on: nothing (pure streaming utilities, no external deps)
 *
 * Edge case: includes repetition detection to truncate LLM hallucination
 * loops where the model repeats the same ~60-char window multiple times.
 */

/** Number of characters in the sliding window used to detect repeated text. */
const REPETITION_WINDOW = 60

/** How many times the window must repeat to trigger truncation. */
const REPETITION_THRESHOLD = 3

/**
 * Detect hallucination loops by checking if the last REPETITION_WINDOW chars
 * appear REPETITION_THRESHOLD times in the preceding text.
 *
 * @returns Index at which to truncate, or -1 if no repetition detected.
 */
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
 * Parse a single SSE line and extract the content token.
 *
 * Expects OpenAI-compatible format: `data: {"choices":[{"delta":{"content":"token"}}]}`
 * Returns null for `[DONE]` sentinel, non-data lines, malformed JSON,
 * or events without text content (e.g. role-only deltas).
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
 * Async generator that yields individual content tokens from a ReadableStream.
 *
 * Handles the common SSE edge case where a network chunk splits in the middle
 * of a line — incomplete lines are buffered until the next chunk arrives.
 * Always releases the reader lock in the finally block to avoid stream leaks.
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
 * Accumulate the full text from a RAG server response into a single string.
 *
 * Handles both SSE streams (`text/event-stream`) and standard JSON responses
 * (the RAG server may return either depending on configuration).
 *
 * @param response - The fetch Response from the RAG server
 * @param options.maxChars - Hard limit on accumulated text length
 * @param options.detectRepetition - Enable hallucination loop detection;
 *   truncates output when the model repeats the same text block multiple times
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
