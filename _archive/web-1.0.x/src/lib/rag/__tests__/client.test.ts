/**
 * Tests del cliente HTTP del RAG Server.
 * Cubre: ragFetch, ragGenerateStream, mock mode, y createRagError (via exports).
 *
 * detectLanguageHint ya está cubierto en detect-language.test.ts — no duplicar.
 *
 * Corre con: bun test apps/web/src/lib/rag/__tests__/client.test.ts
 */

import { describe, test, expect, afterEach, spyOn } from "bun:test"
import type { RagGenerateRequest, RagError } from "../client"

// Set env BEFORE importing client (module captures at load time)
process.env["RAG_SERVER_URL"] = "http://localhost:8081"
process.env["MOCK_RAG"] = "false"
process.env["RAG_TIMEOUT_MS"] = "5000"

// Remove OpenRouter key to avoid accidental real calls
delete process.env["OPENROUTER_API_KEY"]

import { ragFetch, ragGenerateStream } from "../client"

// ── Helpers ────────────────────────────────────────────────────────────────────

const MINIMAL_BODY: RagGenerateRequest = {
  messages: [{ role: "user", content: "Hola" }],
}

function mockFetchResponse(status: number, body: string, headers?: Record<string, string>) {
  return new Response(body, {
    status,
    headers: { "Content-Type": "application/json", ...headers },
  })
}

function mockFetchStream(status: number, chunks: string[]) {
  const stream = new ReadableStream({
    start(controller) {
      for (const chunk of chunks) {
        controller.enqueue(new TextEncoder().encode(chunk))
      }
      controller.close()
    },
  })

  return new Response(stream, {
    status,
    headers: { "Content-Type": "text/event-stream" },
  })
}

// ── ragFetch ───────────────────────────────────────────────────────────────────

describe("ragFetch", () => {
  let fetchSpy: ReturnType<typeof spyOn>

  afterEach(() => {
    fetchSpy?.mockRestore()
  })

  test("successful response — returns the Response object", async () => {
    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () =>
      mockFetchResponse(200, JSON.stringify({ collections: ["test"] }))
    )

    const result = await ragFetch("/v1/collections")

    expect(result).toBeInstanceOf(Response)
    const response = result as Response
    expect(response.status).toBe(200)

    const json = await response.json()
    expect(json.collections).toEqual(["test"])
  })

  test("passes path and options to fetch correctly", async () => {
    let capturedUrl = ""
    let capturedInit: RequestInit | undefined

    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async (url, init) => {
      capturedUrl = String(url)
      capturedInit = init
      return mockFetchResponse(200, "{}")
    })

    await ragFetch("/v1/collections", {
      method: "POST",
      headers: { "X-Custom": "test" },
      body: JSON.stringify({ name: "col1" }),
    })

    expect(capturedUrl).toBe("http://localhost:8081/v1/collections")
    expect(capturedInit?.method).toBe("POST")
    // signal should be set (AbortSignal.timeout)
    expect(capturedInit?.signal).toBeDefined()
  })

  test("4xx response — returns the Response (caller checks .ok)", async () => {
    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () =>
      mockFetchResponse(404, JSON.stringify({ detail: "not found" }))
    )

    const result = await ragFetch("/v1/collections/nonexistent")

    // ragFetch returns the response as-is for non-network errors
    expect(result).toBeInstanceOf(Response)
    expect((result as Response).status).toBe(404)
  })

  test("5xx response — returns the Response (caller checks .ok)", async () => {
    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () =>
      mockFetchResponse(502, "Bad Gateway")
    )

    const result = await ragFetch("/health")

    expect(result).toBeInstanceOf(Response)
    expect((result as Response).status).toBe(502)
  })

  test("network error (ECONNREFUSED) — returns { error } with code UNAVAILABLE", async () => {
    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () => {
      throw new Error("fetch failed: ECONNREFUSED 127.0.0.1:8081")
    })

    const result = await ragFetch("/health")

    expect(result).not.toBeInstanceOf(Response)
    const { error } = result as { error: RagError }
    expect(error.code).toBe("UNAVAILABLE")
    expect(error.message).toContain("ECONNREFUSED")
    expect(error.suggestion).toBeTruthy()
  })

  test("generic network error — returns { error } with code UNAVAILABLE", async () => {
    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () => {
      throw new Error("DNS resolution failed")
    })

    const result = await ragFetch("/health")

    expect(result).not.toBeInstanceOf(Response)
    const { error } = result as { error: RagError }
    expect(error.code).toBe("UNAVAILABLE")
    expect(error.message).toContain("DNS resolution failed")
  })
})

// ── ragGenerateStream ──────────────────────────────────────────────────────────

describe("ragGenerateStream", () => {
  let fetchSpy: ReturnType<typeof spyOn>

  afterEach(() => {
    fetchSpy?.mockRestore()
  })

  test("successful stream — returns { stream, contentType }", async () => {
    const sseChunks = [
      `data: ${JSON.stringify({ choices: [{ delta: { content: "Hello" } }] })}\n\n`,
      `data: [DONE]\n\n`,
    ]

    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () =>
      mockFetchStream(200, sseChunks)
    )

    const result = await ragGenerateStream(MINIMAL_BODY)

    expect("stream" in result).toBe(true)
    const { stream, contentType } = result as { stream: ReadableStream; contentType: string }
    expect(contentType).toBe("text/event-stream")

    // Verify we can read the stream
    const reader = stream.getReader()
    const { value, done } = await reader.read()
    expect(done).toBe(false)
    expect(new TextDecoder().decode(value)).toContain("Hello")
    reader.releaseLock()
  })

  test("sends correct request to RAG server endpoint", async () => {
    let capturedUrl = ""
    let capturedBody: Record<string, unknown> = {}
    let capturedHeaders: Record<string, string> = {}

    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async (url, init) => {
      capturedUrl = String(url)
      capturedBody = JSON.parse(init?.body as string)
      capturedHeaders = Object.fromEntries(new Headers(init?.headers as HeadersInit).entries())
      return mockFetchStream(200, ["data: [DONE]\n\n"])
    })

    const body: RagGenerateRequest = {
      messages: [{ role: "user", content: "test" }],
      collection_name: "docs",
      temperature: 0.7,
    }

    await ragGenerateStream(body)

    expect(capturedUrl).toBe("http://localhost:8081/v1/chat/completions")
    expect(capturedBody.stream).toBe(true) // must inject stream: true
    expect(capturedBody.messages).toEqual(body.messages)
    expect(capturedBody.collection_name).toBe("docs")
    expect(capturedBody.temperature).toBe(0.7)
    expect(capturedHeaders["content-type"]).toBe("application/json")
    expect(capturedHeaders["accept"]).toBe("text/event-stream")
  })

  test("HTTP 4xx — returns { error } with code FORBIDDEN", async () => {
    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () =>
      mockFetchResponse(403, "Access denied")
    )

    const result = await ragGenerateStream(MINIMAL_BODY)

    expect("error" in result).toBe(true)
    const { error } = result as { error: RagError }
    expect(error.code).toBe("FORBIDDEN")
    expect(error.message).toContain("403")
    expect(error.message).toContain("Access denied")
  })

  test("HTTP 429 — returns { error } with code FORBIDDEN (< 500)", async () => {
    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () =>
      mockFetchResponse(429, "Rate limit exceeded")
    )

    const result = await ragGenerateStream(MINIMAL_BODY)

    expect("error" in result).toBe(true)
    const { error } = result as { error: RagError }
    expect(error.code).toBe("FORBIDDEN")
  })

  test("HTTP 5xx — returns { error } with code UPSTREAM_ERROR", async () => {
    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () =>
      mockFetchResponse(500, "Internal server error")
    )

    const result = await ragGenerateStream(MINIMAL_BODY)

    expect("error" in result).toBe(true)
    const { error } = result as { error: RagError }
    expect(error.code).toBe("UPSTREAM_ERROR")
    expect(error.message).toContain("500")
  })

  test("HTTP 502 — returns { error } with code UPSTREAM_ERROR", async () => {
    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () =>
      mockFetchResponse(502, "Bad Gateway")
    )

    const result = await ragGenerateStream(MINIMAL_BODY)

    expect("error" in result).toBe(true)
    const { error } = result as { error: RagError }
    expect(error.code).toBe("UPSTREAM_ERROR")
    expect(error.message).toContain("502")
  })

  test("response body is null — returns { error } UPSTREAM_ERROR", async () => {
    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () => {
      // Create a response with ok=true but no body
      // Response constructor always creates a body, so we override it
      const resp = new Response(null, { status: 200 })
      Object.defineProperty(resp, "body", { value: null })
      return resp
    })

    const result = await ragGenerateStream(MINIMAL_BODY)

    expect("error" in result).toBe(true)
    const { error } = result as { error: RagError }
    expect(error.code).toBe("UPSTREAM_ERROR")
    expect(error.message).toContain("no retornó stream")
  })

  test("ECONNREFUSED — returns { error } with code UNAVAILABLE", async () => {
    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () => {
      throw new Error("fetch failed: ECONNREFUSED 127.0.0.1:8081")
    })

    const result = await ragGenerateStream(MINIMAL_BODY)

    expect("error" in result).toBe(true)
    const { error } = result as { error: RagError }
    expect(error.code).toBe("UNAVAILABLE")
    expect(error.message).toContain("ECONNREFUSED")
    // Should get a helpful suggestion from logger/suggestions
    expect(error.suggestion).toBeTruthy()
  })

  test("AbortError (timeout) — returns { error } with code TIMEOUT", async () => {
    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async (_url, _init) => {
      // Simulate an AbortError as if the timeout fired
      const err = new DOMException("The operation was aborted", "AbortError")
      throw err
    })

    const result = await ragGenerateStream(MINIMAL_BODY)

    expect("error" in result).toBe(true)
    const { error } = result as { error: RagError }
    expect(error.code).toBe("TIMEOUT")
    expect(error.message).toContain("Timeout")
  })

  test("user abort signal — respects external abort", async () => {
    const userAbort = new AbortController()
    // Abort immediately
    userAbort.abort()

    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async (_url, init) => {
      // The signal from ragGenerateStream should be aborted because user aborted
      if (init?.signal?.aborted) {
        throw new DOMException("The operation was aborted", "AbortError")
      }
      return mockFetchStream(200, ["data: [DONE]\n\n"])
    })

    const result = await ragGenerateStream(MINIMAL_BODY, userAbort.signal)

    expect("error" in result).toBe(true)
    const { error } = result as { error: RagError }
    expect(error.code).toBe("TIMEOUT")
  })

  test("generic non-ECONNREFUSED error — returns UPSTREAM_ERROR", async () => {
    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () => {
      throw new Error("TLS handshake failed")
    })

    const result = await ragGenerateStream(MINIMAL_BODY)

    expect("error" in result).toBe(true)
    const { error } = result as { error: RagError }
    expect(error.code).toBe("UPSTREAM_ERROR")
    expect(error.message).toContain("TLS handshake failed")
  })

  test("error body is truncated to 200 chars in message", async () => {
    const longBody = "x".repeat(500)

    fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () =>
      mockFetchResponse(500, longBody)
    )

    const result = await ragGenerateStream(MINIMAL_BODY)

    expect("error" in result).toBe(true)
    const { error } = result as { error: RagError }
    // The error body is sliced to 200 chars
    expect(error.message.length).toBeLessThan(300) // status prefix + 200 chars of body
    expect(error.message).not.toContain("x".repeat(500))
  })
})

// ── RagError shape ─────────────────────────────────────────────────────────────

describe("RagError shape", () => {
  test("error always includes code, message, and suggestion", async () => {
    const fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () => {
      throw new Error("something broke")
    })

    const result = await ragFetch("/health")

    expect(result).not.toBeInstanceOf(Response)
    const { error } = result as { error: RagError }
    expect(error).toHaveProperty("code")
    expect(error).toHaveProperty("message")
    expect(error).toHaveProperty("suggestion")
    expect(typeof error.code).toBe("string")
    expect(typeof error.message).toBe("string")
    expect(typeof error.suggestion).toBe("string")
    // Suggestion should never be empty — fallback exists
    expect(error.suggestion.length).toBeGreaterThan(0)

    fetchSpy.mockRestore()
  })

  test("known ECONNREFUSED pattern triggers specific suggestion from logger", async () => {
    const fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () => {
      throw new Error("ECONNREFUSED 127.0.0.1:8081")
    })

    const result = await ragFetch("/health")
    const { error } = result as { error: RagError }

    // The logger/suggestions module has a specific pattern for ECONNREFUSED.*8081
    expect(error.suggestion).toContain("RAG Server")

    fetchSpy.mockRestore()
  })
})

// ── Mock mode ──────────────────────────────────────────────────────────────────

describe("mockRagStream (MOCK_RAG behavior)", () => {
  // NOTE: MOCK_RAG is captured at module load time as `false` in this test file.
  // We cannot test the mock path directly without re-importing the module.
  // Instead, we verify that with MOCK_RAG=false, real fetch is always called.

  test("with MOCK_RAG=false, ragGenerateStream always calls fetch", async () => {
    let fetchCalled = false
    const fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () => {
      fetchCalled = true
      return mockFetchStream(200, ["data: [DONE]\n\n"])
    })

    await ragGenerateStream(MINIMAL_BODY)
    expect(fetchCalled).toBe(true)

    fetchSpy.mockRestore()
  })

  test("with MOCK_RAG=false, ragFetch always calls fetch", async () => {
    let fetchCalled = false
    const fetchSpy = spyOn(globalThis, "fetch").mockImplementationOnce(async () => {
      fetchCalled = true
      return mockFetchResponse(200, "{}")
    })

    await ragFetch("/health")
    expect(fetchCalled).toBe(true)

    fetchSpy.mockRestore()
  })
})
