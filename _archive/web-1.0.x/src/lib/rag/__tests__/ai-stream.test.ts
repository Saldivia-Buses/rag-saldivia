/**
 * Tests for the NVIDIA SSE -> AI SDK Data Stream adapter.
 *
 * `createRagStreamResponse` takes a ReadableStream of SSE bytes (from the
 * NVIDIA RAG server) and returns a Response whose body is an AI SDK
 * UI message stream.  These tests verify the full pipeline: SSE input ->
 * adapter -> Response body -> parsed AI SDK chunks.
 *
 * Run with: bun test apps/web/src/lib/rag/__tests__/ai-stream.test.ts
 */

import { describe, test, expect } from "bun:test"
import { createRagStreamResponse } from "../ai-stream"

// ── Helpers ────────────────────────────────────────────────────────────────

/** Build a ReadableStream<Uint8Array> from an array of string chunks. */
function createSseStream(chunks: string[]): ReadableStream<Uint8Array> {
  const encoder = new TextEncoder()
  return new ReadableStream({
    start(controller) {
      for (const chunk of chunks) {
        controller.enqueue(encoder.encode(chunk))
      }
      controller.close()
    },
  })
}

/** Build a single NVIDIA SSE data line with a text token. */
function tokenLine(content: string): string {
  return `data: ${JSON.stringify({ choices: [{ delta: { content } }] })}\n`
}

/** Build a single NVIDIA SSE data line with sources (citations). */
function sourcesLine(sources: unknown[]): string {
  return `data: ${JSON.stringify({ choices: [{ delta: { sources } }] })}\n`
}

/** The DONE sentinel that ends a NVIDIA SSE stream. */
const DONE_LINE = "data: [DONE]\n"

/**
 * Read an AI SDK SSE Response body and return the parsed chunk objects.
 *
 * The AI SDK serializes chunks as `data: <JSON>\n\n`.  This helper reads
 * all of them and returns an array of parsed objects.
 */
async function parseAiSdkResponse(
  response: Response
): Promise<Array<Record<string, unknown>>> {
  const text = await response.text()
  const chunks: Array<Record<string, unknown>> = []

  for (const line of text.split("\n")) {
    const trimmed = line.trim()
    if (trimmed.startsWith("data: ")) {
      try {
        chunks.push(JSON.parse(trimmed.slice(6)) as Record<string, unknown>)
      } catch {
        // skip malformed (shouldn't happen from AI SDK)
      }
    }
  }

  return chunks
}

/** Filter chunks by type. */
function chunksOfType(chunks: Array<Record<string, unknown>>, type: string) {
  return chunks.filter((c) => c.type === type)
}

// ── createRagStreamResponse ────────────────────────────────────────────────

describe("createRagStreamResponse", () => {
  test("returns a valid Response with correct content-type", async () => {
    const stream = createSseStream([DONE_LINE])
    const response = createRagStreamResponse(stream)

    expect(response).toBeInstanceOf(Response)
    const ct = response.headers.get("content-type")
    expect(ct).toContain("text/event-stream")
  })

  test("simple tokens produce text-start, text-delta(s), text-end", async () => {
    const stream = createSseStream([
      tokenLine("Hello"),
      tokenLine(" world"),
      DONE_LINE,
    ])

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    const starts = chunksOfType(chunks, "text-start")
    expect(starts).toHaveLength(1)
    expect(starts[0]!.id).toBe("msg-text")

    const deltas = chunksOfType(chunks, "text-delta")
    expect(deltas).toHaveLength(2)
    expect(deltas[0]!.delta).toBe("Hello")
    expect(deltas[1]!.delta).toBe(" world")

    const ends = chunksOfType(chunks, "text-end")
    expect(ends).toHaveLength(1)
    expect(ends[0]!.id).toBe("msg-text")
  })

  test("text-start appears exactly once regardless of token count", async () => {
    const stream = createSseStream([
      tokenLine("a"),
      tokenLine("b"),
      tokenLine("c"),
      tokenLine("d"),
      DONE_LINE,
    ])

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    expect(chunksOfType(chunks, "text-start")).toHaveLength(1)
    expect(chunksOfType(chunks, "text-delta")).toHaveLength(4)
    expect(chunksOfType(chunks, "text-end")).toHaveLength(1)
  })

  test("citations in delta.sources emit data-sources chunk", async () => {
    const citations = [
      { document: "report.pdf", content: "Revenue grew 12%", score: 0.95 },
      { document: "summary.pdf", content: "Q4 results", score: 0.88 },
    ]

    const stream = createSseStream([
      tokenLine("Based on the report, "),
      sourcesLine(citations),
      tokenLine("revenue grew."),
      DONE_LINE,
    ])

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    const sourceChunks = chunksOfType(chunks, "data-sources")
    expect(sourceChunks).toHaveLength(1)

    const data = sourceChunks[0]!.data as { citations: unknown[] }
    expect(data.citations).toHaveLength(2)
    expect(data.citations[0]).toMatchObject({
      document: "report.pdf",
      content: "Revenue grew 12%",
      score: 0.95,
    })
  })

  test("citations with only optional fields pass validation", async () => {
    // CitationSchema has all fields optional — a minimal object should work
    const stream = createSseStream([
      tokenLine("text"),
      sourcesLine([{ document: "doc.pdf" }]),
      DONE_LINE,
    ])

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    const sourceChunks = chunksOfType(chunks, "data-sources")
    expect(sourceChunks).toHaveLength(1)
  })

  test("invalid citations (fail Zod) are silently skipped", async () => {
    // sources with wrong types should fail CitationSchema.array().safeParse()
    const stream = createSseStream([
      tokenLine("text"),
      sourcesLine(["not-an-object", 42]),
      DONE_LINE,
    ])

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    // No data-sources chunk should appear
    const sourceChunks = chunksOfType(chunks, "data-sources")
    expect(sourceChunks).toHaveLength(0)

    // But text tokens should still work
    const deltas = chunksOfType(chunks, "text-delta")
    expect(deltas).toHaveLength(1)
  })

  test("[DONE] sentinel does not produce any token or crash", async () => {
    const stream = createSseStream([
      tokenLine("answer"),
      DONE_LINE,
    ])

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    const deltas = chunksOfType(chunks, "text-delta")
    expect(deltas).toHaveLength(1)
    expect(deltas[0]!.delta).toBe("answer")

    // No chunk should have [DONE] as content
    for (const chunk of chunks) {
      expect(chunk.delta).not.toBe("[DONE]")
    }
  })

  test("empty stream returns valid response with no text parts", async () => {
    const stream = createSseStream([])
    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    // No text-start/end because no tokens were written
    expect(chunksOfType(chunks, "text-start")).toHaveLength(0)
    expect(chunksOfType(chunks, "text-delta")).toHaveLength(0)
    expect(chunksOfType(chunks, "text-end")).toHaveLength(0)
  })

  test("stream with only DONE produces no text parts", async () => {
    const stream = createSseStream([DONE_LINE])
    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    expect(chunksOfType(chunks, "text-start")).toHaveLength(0)
    expect(chunksOfType(chunks, "text-delta")).toHaveLength(0)
    expect(chunksOfType(chunks, "text-end")).toHaveLength(0)
  })

  test("malformed JSON lines are silently ignored", async () => {
    const stream = createSseStream([
      tokenLine("before"),
      "data: {broken json!!!}\n",
      tokenLine("after"),
      DONE_LINE,
    ])

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    const deltas = chunksOfType(chunks, "text-delta")
    expect(deltas).toHaveLength(2)
    expect(deltas[0]!.delta).toBe("before")
    expect(deltas[1]!.delta).toBe("after")
  })

  test("non-data lines (event:, id:, comments) are ignored", async () => {
    const stream = createSseStream([
      ": this is a comment\n",
      "event: message\n",
      tokenLine("token"),
      "id: 42\n",
      DONE_LINE,
    ])

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    const deltas = chunksOfType(chunks, "text-delta")
    expect(deltas).toHaveLength(1)
    expect(deltas[0]!.delta).toBe("token")
  })

  test("partial buffer across chunks: line split mid-token", async () => {
    // Simulate a network chunk that splits an SSE line in the middle
    const fullLine = tokenLine("split-token")
    const midpoint = Math.floor(fullLine.length / 2)

    const stream = createSseStream([
      fullLine.slice(0, midpoint),
      fullLine.slice(midpoint),
      DONE_LINE,
    ])

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    const deltas = chunksOfType(chunks, "text-delta")
    expect(deltas).toHaveLength(1)
    expect(deltas[0]!.delta).toBe("split-token")
  })

  test("multiple lines in a single chunk are all processed", async () => {
    // All SSE lines arrive in one network chunk
    const stream = createSseStream([
      tokenLine("one") + tokenLine("two") + tokenLine("three") + DONE_LINE,
    ])

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    const deltas = chunksOfType(chunks, "text-delta")
    expect(deltas).toHaveLength(3)
    expect(deltas.map((d) => d.delta)).toEqual(["one", "two", "three"])
  })

  test("tokens with special characters are preserved", async () => {
    const specialContent = 'El costo es $1.500, con IVA "incluido" & más\n'
    const stream = createSseStream([
      tokenLine(specialContent),
      DONE_LINE,
    ])

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    const deltas = chunksOfType(chunks, "text-delta")
    expect(deltas).toHaveLength(1)
    expect(deltas[0]!.delta).toBe(specialContent)
  })

  test("tokens with unicode characters are preserved", async () => {
    const stream = createSseStream([
      tokenLine("Hola"),
      tokenLine(" "),
      tokenLine("cafe\u0301"),
      DONE_LINE,
    ])

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    const deltas = chunksOfType(chunks, "text-delta")
    expect(deltas).toHaveLength(3)
    expect(deltas[2]!.delta).toBe("cafe\u0301")
  })

  test("interleaved tokens and sources maintain correct order", async () => {
    const stream = createSseStream([
      tokenLine("According to "),
      sourcesLine([{ document: "policy.pdf", content: "Section 5" }]),
      tokenLine("the policy states "),
      sourcesLine([{ document: "manual.pdf", content: "Chapter 3" }]),
      tokenLine("that access is restricted."),
      DONE_LINE,
    ])

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    // Verify we got all parts
    const deltas = chunksOfType(chunks, "text-delta")
    expect(deltas).toHaveLength(3)

    const sourceChunks = chunksOfType(chunks, "data-sources")
    expect(sourceChunks).toHaveLength(2)

    // Verify ordering: text-start should come first, then interleaved content
    const typeSequence = chunks.map((c) => c.type)
    const textStartIdx = typeSequence.indexOf("text-start")
    const textEndIdx = typeSequence.indexOf("text-end")
    expect(textStartIdx).toBeLessThan(textEndIdx)

    // All text-delta and data-sources should be between start and end
    for (let i = 0; i < chunks.length; i++) {
      const t = chunks[i]!.type
      if (t === "text-delta" || t === "data-sources") {
        expect(i).toBeGreaterThan(textStartIdx)
        expect(i).toBeLessThan(textEndIdx)
      }
    }
  })

  test("delta without content but with sources only emits data-sources", async () => {
    // A line that has sources but NO content token
    const stream = createSseStream([
      sourcesLine([{ document: "doc.pdf" }]),
      DONE_LINE,
    ])

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    // No text parts — sources-only lines don't trigger text-start
    expect(chunksOfType(chunks, "text-start")).toHaveLength(0)
    expect(chunksOfType(chunks, "text-delta")).toHaveLength(0)
    expect(chunksOfType(chunks, "text-end")).toHaveLength(0)

    // But data-sources should still appear
    expect(chunksOfType(chunks, "data-sources")).toHaveLength(1)
  })

  test("trailing buffer without newline is still processed", async () => {
    // Simulate a stream that ends without a trailing newline
    const line = `data: ${JSON.stringify({ choices: [{ delta: { content: "trailing" } }] })}`
    const stream = createSseStream([line]) // no \n at end

    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    const deltas = chunksOfType(chunks, "text-delta")
    expect(deltas).toHaveLength(1)
    expect(deltas[0]!.delta).toBe("trailing")
  })

  test("large number of tokens produces correct count", async () => {
    const tokenCount = 100
    const lines: string[] = []
    for (let i = 0; i < tokenCount; i++) {
      lines.push(tokenLine(`t${i}`))
    }
    lines.push(DONE_LINE)

    const stream = createSseStream(lines)
    const response = createRagStreamResponse(stream)
    const chunks = await parseAiSdkResponse(response)

    const deltas = chunksOfType(chunks, "text-delta")
    expect(deltas).toHaveLength(tokenCount)
    expect(deltas[0]!.delta).toBe("t0")
    expect(deltas[tokenCount - 1]!.delta).toBe(`t${tokenCount - 1}`)
  })
})
