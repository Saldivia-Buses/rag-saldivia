/**
 * Tests for POST /api/rag/generate route handler.
 *
 * Covers: auth, validation, rate limiting, collection access, streaming, error paths.
 * Runs with: bun test apps/web/src/__tests__/api/rag-generate.test.ts
 */

import { describe, test, expect, mock, beforeEach } from "bun:test"

// ── Mock functions (declared before mock.module) ─────────────────────────

const mockRagGenerateStream = mock(() =>
  Promise.resolve({
    stream: new ReadableStream({
      start(controller) {
        controller.enqueue(new TextEncoder().encode("data: test\n\n"))
        controller.close()
      },
    }),
    contentType: "text/event-stream",
  })
)

const mockCreateRagStreamResponse = mock(
  (stream: ReadableStream) => new Response(stream, { headers: { "Content-Type": "text/event-stream" } })
)

const mockCanAccessCollection = mock(() => Promise.resolve(true))
const mockGetUserCollections = mock(() => Promise.resolve([]))
const mockGetRateLimit = mock(() => Promise.resolve(null))
const mockCountQueriesLastHour = mock(() => Promise.resolve(0))
const mockGetProjectBySession = mock(() => Promise.resolve(null))
const mockGetMemoryAsContext = mock(() => Promise.resolve(null))
const mockDispatchEvent = mock(() => Promise.resolve())

// ── Module mocks ─────────────────────────────────────────────────────────

mock.module("@rag-saldivia/logger/backend", () => ({
  log: { info: () => {}, warn: () => {}, error: () => {} },
}))

mock.module("@rag-saldivia/logger/suggestions", () => ({
  getSuggestion: () => null,
}))

mock.module("@rag-saldivia/db", () => ({
  canAccessCollection: mockCanAccessCollection,
  getUserCollections: mockGetUserCollections,
  getRateLimit: mockGetRateLimit,
  countQueriesLastHour: mockCountQueriesLastHour,
  getProjectBySession: mockGetProjectBySession,
  getMemoryAsContext: mockGetMemoryAsContext,
  touchUserPresence: mock(() => Promise.resolve()),
  getRedisClient: mock(() => ({ get: mock(() => Promise.resolve(null)) })),
}))

mock.module("@/lib/rag/client", () => ({
  ragGenerateStream: mockRagGenerateStream,
  detectLanguageHint: () => "",
}))

mock.module("@/lib/rag/ai-stream", () => ({
  createRagStreamResponse: mockCreateRagStreamResponse,
}))

mock.module("@/lib/webhook", () => ({
  dispatchEvent: mockDispatchEvent,
}))

mock.module("@rag-saldivia/shared", () => ({
  FOCUS_MODES: [
    { id: "precise", label: "Preciso", systemPrompt: "Be precise and factual." },
  ],
}))

// ── Import route handler AFTER mocks ─────────────────────────────────────

import { POST } from "@/app/api/rag/generate/route"

// ── Helpers ──────────────────────────────────────────────────────────────

/** Build a valid authenticated request */
function makeRequest(body: unknown, overrides?: { headers?: Record<string, string> }): Request {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    "x-user-id": "1",
    "x-user-email": "admin@localhost",
    "x-user-name": "Admin",
    "x-user-role": "admin",
    ...(overrides?.headers ?? {}),
  }

  return new Request("http://localhost:3000/api/rag/generate", {
    method: "POST",
    headers,
    body: JSON.stringify(body),
  })
}

/** Build a request without auth headers */
function makeUnauthRequest(body: unknown): Request {
  return new Request("http://localhost:3000/api/rag/generate", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  })
}

const VALID_BODY = {
  messages: [{ role: "user", content: "What is RAG?" }],
  collection_name: "default",
}

// ── Reset ────────────────────────────────────────────────────────────────

beforeEach(() => {
  mockRagGenerateStream.mockClear()
  mockCreateRagStreamResponse.mockClear()
  mockCanAccessCollection.mockClear()
  mockGetUserCollections.mockClear()
  mockGetRateLimit.mockClear()
  mockCountQueriesLastHour.mockClear()
  mockGetProjectBySession.mockClear()
  mockGetMemoryAsContext.mockClear()
  mockDispatchEvent.mockClear()

  // Reset to defaults
  mockRagGenerateStream.mockImplementation(() =>
    Promise.resolve({
      stream: new ReadableStream({
        start(controller) {
          controller.enqueue(new TextEncoder().encode("data: test\n\n"))
          controller.close()
        },
      }),
      contentType: "text/event-stream",
    })
  )
  mockCanAccessCollection.mockImplementation(() => Promise.resolve(true))
  mockGetRateLimit.mockImplementation(() => Promise.resolve(null))
  mockCountQueriesLastHour.mockImplementation(() => Promise.resolve(0))
  mockGetMemoryAsContext.mockImplementation(() => Promise.resolve(null))
  mockGetProjectBySession.mockImplementation(() => Promise.resolve(null))
})

// ── Tests ────────────────────────────────────────────────────────────────

describe("POST /api/rag/generate — auth", () => {
  test("returns 401 when no auth headers present", async () => {
    const request = makeUnauthRequest(VALID_BODY)
    const response = await POST(request)

    expect(response.status).toBe(401)
    const json = await response.json()
    expect(json.ok).toBe(false)
    expect(json.error).toContain("autenticado")
  })

  test("returns 401 with partial auth headers (missing x-user-role)", async () => {
    const request = new Request("http://localhost:3000/api/rag/generate", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-user-id": "1",
        // missing x-user-role
      },
      body: JSON.stringify(VALID_BODY),
    })

    const response = await POST(request)
    expect(response.status).toBe(401)
  })
})

describe("POST /api/rag/generate — validation", () => {
  test("returns 400 when body is empty", async () => {
    const request = makeRequest({})
    const response = await POST(request)

    expect(response.status).toBe(400)
    const json = await response.json()
    expect(json.ok).toBe(false)
  })

  test("returns 400 when messages array is empty", async () => {
    const request = makeRequest({ messages: [] })
    const response = await POST(request)

    expect(response.status).toBe(400)
    const json = await response.json()
    expect(json.ok).toBe(false)
    expect(json.error).toContain("messages")
  })

  test("returns 400 when body is invalid JSON", async () => {
    const request = new Request("http://localhost:3000/api/rag/generate", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-user-id": "1",
        "x-user-email": "admin@localhost",
        "x-user-name": "Admin",
        "x-user-role": "admin",
      },
      body: "not json",
    })

    const response = await POST(request)
    expect(response.status).toBe(400)
  })
})

describe("POST /api/rag/generate — rate limiting", () => {
  test("returns 429 when user exceeds rate limit", async () => {
    mockGetRateLimit.mockImplementation(() => Promise.resolve(10))
    mockCountQueriesLastHour.mockImplementation(() => Promise.resolve(10))

    const request = makeRequest(VALID_BODY)
    const response = await POST(request)

    expect(response.status).toBe(429)
    const json = await response.json()
    expect(json.error).toContain("10")
  })

  test("allows request when under rate limit", async () => {
    mockGetRateLimit.mockImplementation(() => Promise.resolve(10))
    mockCountQueriesLastHour.mockImplementation(() => Promise.resolve(5))

    const request = makeRequest(VALID_BODY)
    const response = await POST(request)

    // Should proceed past rate limiting (200 from stream)
    expect(response.status).toBe(200)
  })

  test("allows request when no rate limit is set (null)", async () => {
    mockGetRateLimit.mockImplementation(() => Promise.resolve(null))

    const request = makeRequest(VALID_BODY)
    const response = await POST(request)

    expect(response.status).toBe(200)
  })
})

describe("POST /api/rag/generate — collection access", () => {
  test("returns 403 when user cannot access collection", async () => {
    mockCanAccessCollection.mockImplementation(() => Promise.resolve(false))

    const request = makeRequest({
      messages: [{ role: "user", content: "test" }],
      collection_name: "secret-collection",
    })
    const response = await POST(request)

    expect(response.status).toBe(403)
    const json = await response.json()
    expect(json.error).toContain("Sin acceso")
  })

  test("allows request when user has collection access", async () => {
    mockCanAccessCollection.mockImplementation(() => Promise.resolve(true))

    const request = makeRequest(VALID_BODY)
    const response = await POST(request)

    expect(response.status).toBe(200)
  })

  test("returns 403 for multi-collection when user lacks access to one", async () => {
    mockGetUserCollections.mockImplementation(() =>
      Promise.resolve([
        { name: "allowed", permission: "read" },
      ])
    )

    const request = makeRequest({
      messages: [{ role: "user", content: "test" }],
      collection_names: ["allowed", "forbidden"],
    })
    const response = await POST(request)

    expect(response.status).toBe(403)
    const json = await response.json()
    expect(json.error).toContain("Sin acceso")
  })
})

describe("POST /api/rag/generate — streaming success", () => {
  test("returns streaming response on valid request", async () => {
    const request = makeRequest(VALID_BODY)
    const response = await POST(request)

    expect(response.status).toBe(200)
    expect(mockRagGenerateStream).toHaveBeenCalledTimes(1)
    expect(mockCreateRagStreamResponse).toHaveBeenCalledTimes(1)
  })

  test("dispatches webhook event after stream", async () => {
    const request = makeRequest(VALID_BODY)
    await POST(request)

    expect(mockDispatchEvent).toHaveBeenCalledWith(
      "query.completed",
      expect.objectContaining({ userId: 1 })
    )
  })
})

describe("POST /api/rag/generate — RAG errors (SSE error hidden in 200 prevention)", () => {
  test("returns 502 when RAG returns upstream error", async () => {
    mockRagGenerateStream.mockImplementation(() =>
      Promise.resolve({
        error: {
          code: "UPSTREAM_ERROR",
          message: "RAG Server respondio 500",
          suggestion: "Check the server",
        },
      })
    )

    const request = makeRequest(VALID_BODY)
    const response = await POST(request)

    expect(response.status).toBe(502)
    const json = await response.json()
    expect(json.ok).toBe(false)
  })

  test("returns 504 on timeout", async () => {
    mockRagGenerateStream.mockImplementation(() =>
      Promise.resolve({
        error: {
          code: "TIMEOUT",
          message: "Timeout after 120000ms",
          suggestion: "Try again",
        },
      })
    )

    const request = makeRequest(VALID_BODY)
    const response = await POST(request)

    expect(response.status).toBe(504)
  })

  test("returns 503 when RAG is unavailable", async () => {
    mockRagGenerateStream.mockImplementation(() =>
      Promise.resolve({
        error: {
          code: "UNAVAILABLE",
          message: "ECONNREFUSED",
          suggestion: "Start RAG server",
        },
      })
    )

    const request = makeRequest(VALID_BODY)
    const response = await POST(request)

    expect(response.status).toBe(503)
  })
})

describe("POST /api/rag/generate — message normalization", () => {
  test("handles AI SDK parts format (normalizes to content)", async () => {
    const request = makeRequest({
      messages: [
        {
          role: "user",
          parts: [
            { type: "text", text: "Hello " },
            { type: "text", text: "world" },
          ],
        },
      ],
      collection_name: "default",
    })

    const response = await POST(request)

    expect(response.status).toBe(200)
    // Verify ragGenerateStream was called with normalized messages
    const callArgs = mockRagGenerateStream.mock.calls[0]?.[0] as Record<string, unknown>
    const messages = callArgs?.["messages"] as Array<{ role: string; content: string }>
    // The last user message should be the normalized one (system messages may be prepended)
    const userMsg = messages.find((m) => m.role === "user")
    expect(userMsg?.content).toBe("Hello world")
  })
})
