/**
 * Tests for /api/rag/collections route handlers.
 *
 * GET  /api/rag/collections       — list collections (admin sees all, user sees permitted)
 * POST /api/rag/collections       — create collection (admin only)
 * DELETE /api/rag/collections/[name] — delete collection (admin only)
 *
 * Mocks: @rag-saldivia/db, @/lib/rag/client, @/lib/rag/collections-cache, @rag-saldivia/logger
 * Auth: extractClaims reads x-user-* headers from Request (set by middleware/proxy)
 * Runs with: bun test apps/web/src/__tests__/api/collections.test.ts
 */

import { describe, test, expect, mock, beforeEach } from "bun:test"

// ── Mock functions (declared before mock.module) ─────────────────────────

const mockGetCachedRagCollections = mock(() =>
  Promise.resolve(["docs", "contracts", "internal"])
)
const mockInvalidateCollectionsCache = mock(() => Promise.resolve())
const mockGetUserCollections = mock(() =>
  Promise.resolve([
    { name: "docs", permission: "read" },
  ])
)
const mockCanAccessCollection = mock(() => Promise.resolve(true))
const mockRagFetch = mock(() =>
  Promise.resolve(new Response(JSON.stringify({ ok: true }), { status: 200 }))
)

// ── Module mocks ─────────────────────────────────────────────────────────

mock.module("@rag-saldivia/logger/backend", () => ({
  log: { info: () => {}, warn: () => {}, error: () => {} },
}))

mock.module("@rag-saldivia/db", () => ({
  getUserCollections: mockGetUserCollections,
  canAccessCollection: mockCanAccessCollection,
  touchUserPresence: mock(() => Promise.resolve()),
  getRedisClient: mock(() => ({ get: mock(() => Promise.resolve(null)) })),
}))

mock.module("@/lib/rag/client", () => ({
  ragFetch: mockRagFetch,
}))

mock.module("@/lib/rag/collections-cache", () => ({
  getCachedRagCollections: mockGetCachedRagCollections,
  invalidateCollectionsCache: mockInvalidateCollectionsCache,
}))

// ── Import route handlers AFTER mocks ────────────────────────────────────

import { GET, POST } from "@/app/api/rag/collections/route"
import { DELETE } from "@/app/api/rag/collections/[name]/route"

// ── Helpers ──────────────────────────────────────────────────────────────

function makeRequest(
  method: string,
  body?: unknown,
  overrides?: { headers?: Record<string, string> }
): Request {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    "x-user-id": "1",
    "x-user-email": "admin@localhost",
    "x-user-name": "Admin",
    "x-user-role": "admin",
    ...(overrides?.headers ?? {}),
  }

  return new Request("http://localhost:3000/api/rag/collections", {
    method,
    headers,
    ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
  })
}

function makeUserRequest(
  method: string,
  body?: unknown
): Request {
  return makeRequest(method, body, {
    headers: {
      "x-user-id": "10",
      "x-user-email": "user@localhost",
      "x-user-name": "Regular User",
      "x-user-role": "user",
    },
  })
}

function makeUnauthRequest(method: string): Request {
  return new Request("http://localhost:3000/api/rag/collections", {
    method,
    headers: { "Content-Type": "application/json" },
  })
}

// ── Reset mocks between tests ────────────────────────────────────────────

beforeEach(() => {
  mockGetCachedRagCollections.mockClear()
  mockInvalidateCollectionsCache.mockClear()
  mockGetUserCollections.mockClear()
  mockCanAccessCollection.mockClear()
  mockRagFetch.mockClear()

  // Reset to defaults
  mockGetCachedRagCollections.mockImplementation(() =>
    Promise.resolve(["docs", "contracts", "internal"])
  )
  mockGetUserCollections.mockImplementation(() =>
    Promise.resolve([{ name: "docs", permission: "read" }])
  )
  mockRagFetch.mockImplementation(() =>
    Promise.resolve(new Response(JSON.stringify({ ok: true }), { status: 200 }))
  )
})

// ── GET /api/rag/collections ─────────────────────────────────────────────

describe("GET /api/rag/collections", () => {
  test("returns all collections for admin user", async () => {
    const request = makeRequest("GET")
    const response = await GET(request)

    expect(response.status).toBe(200)
    const json = await response.json()
    expect(json.ok).toBe(true)
    expect(json.data).toEqual(["docs", "contracts", "internal"])
    expect(mockGetCachedRagCollections).toHaveBeenCalledTimes(1)
    // Admin should NOT hit getUserCollections
    expect(mockGetUserCollections).not.toHaveBeenCalled()
  })

  test("returns filtered collections for non-admin user", async () => {
    const request = makeUserRequest("GET")
    const response = await GET(request)

    expect(response.status).toBe(200)
    const json = await response.json()
    expect(json.ok).toBe(true)
    // User only has "docs" permission, so only "docs" should appear
    expect(json.data).toEqual(["docs"])
    expect(mockGetUserCollections).toHaveBeenCalledTimes(1)
  })

  test("returns empty array when non-admin user has no permissions", async () => {
    mockGetUserCollections.mockImplementation(() => Promise.resolve([]))

    const request = makeUserRequest("GET")
    const response = await GET(request)

    expect(response.status).toBe(200)
    const json = await response.json()
    expect(json.data).toEqual([])
  })

  test("returns 401 when not authenticated", async () => {
    const request = makeUnauthRequest("GET")
    const response = await GET(request)

    expect(response.status).toBe(401)
    const json = await response.json()
    expect(json.ok).toBe(false)
  })
})

// ── POST /api/rag/collections ────────────────────────────────────────────

describe("POST /api/rag/collections", () => {
  test("creates collection with valid name", async () => {
    const request = makeRequest("POST", { name: "new-collection" })
    const response = await POST(request)

    expect(response.status).toBe(200)
    const json = await response.json()
    expect(json.ok).toBe(true)
    expect(mockRagFetch).toHaveBeenCalledTimes(1)
    // Verify ragFetch was called with correct path and body
    const [path, opts] = mockRagFetch.mock.calls[0] as [string, RequestInit]
    expect(path).toBe("/v1/collections")
    expect(opts.method).toBe("POST")
    const body = JSON.parse(opts.body as string)
    expect(body.collection_name).toBe("new-collection")
  })

  test("invalidates cache after successful creation", async () => {
    const request = makeRequest("POST", { name: "new-collection" })
    await POST(request)

    expect(mockInvalidateCollectionsCache).toHaveBeenCalledTimes(1)
  })

  test("returns 400 for invalid collection name (uppercase)", async () => {
    const request = makeRequest("POST", { name: "Invalid Name!" })
    const response = await POST(request)

    expect(response.status).toBe(400)
    const json = await response.json()
    expect(json.ok).toBe(false)
  })

  test("returns 400 for empty collection name", async () => {
    const request = makeRequest("POST", { name: "" })
    const response = await POST(request)

    expect(response.status).toBe(400)
    const json = await response.json()
    expect(json.ok).toBe(false)
  })

  test("returns 400 for collection name with special chars", async () => {
    const request = makeRequest("POST", { name: "col@#$%" })
    const response = await POST(request)

    expect(response.status).toBe(400)
  })

  test("returns 403 when non-admin tries to create", async () => {
    // Need to add body for POST — re-create with body
    const userRequest = new Request("http://localhost:3000/api/rag/collections", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-user-id": "10",
        "x-user-email": "user@localhost",
        "x-user-name": "Regular User",
        "x-user-role": "user",
      },
      body: JSON.stringify({ name: "new-collection" }),
    })
    const response = await POST(userRequest)

    expect(response.status).toBe(403)
    const json = await response.json()
    expect(json.ok).toBe(false)
  })

  test("returns 401 when not authenticated", async () => {
    const request = new Request("http://localhost:3000/api/rag/collections", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: "test" }),
    })
    const response = await POST(request)

    expect(response.status).toBe(401)
  })

  test("returns 502 when RAG server returns error", async () => {
    mockRagFetch.mockImplementation(() =>
      Promise.resolve({ error: { code: "UPSTREAM_ERROR", message: "RAG down", suggestion: "" } })
    )

    const request = makeRequest("POST", { name: "new-collection" })
    const response = await POST(request)

    expect(response.status).toBe(502)
    const json = await response.json()
    expect(json.ok).toBe(false)
  })
})

// ── DELETE /api/rag/collections/[name] ───────────────────────────────────

describe("DELETE /api/rag/collections/[name]", () => {
  test("deletes collection with valid name", async () => {
    const request = new Request("http://localhost:3000/api/rag/collections/my-collection", {
      method: "DELETE",
      headers: {
        "x-user-id": "1",
        "x-user-email": "admin@localhost",
        "x-user-name": "Admin",
        "x-user-role": "admin",
      },
    })

    const response = await DELETE(request, {
      params: Promise.resolve({ name: "my-collection" }),
    })

    expect(response.status).toBe(200)
    const json = await response.json()
    expect(json.ok).toBe(true)
    expect(mockRagFetch).toHaveBeenCalledTimes(1)
    const [path] = mockRagFetch.mock.calls[0] as [string]
    expect(path).toContain("my-collection")
  })

  test("invalidates cache after successful deletion", async () => {
    const request = new Request("http://localhost:3000/api/rag/collections/docs", {
      method: "DELETE",
      headers: {
        "x-user-id": "1",
        "x-user-email": "admin@localhost",
        "x-user-name": "Admin",
        "x-user-role": "admin",
      },
    })

    await DELETE(request, { params: Promise.resolve({ name: "docs" }) })

    expect(mockInvalidateCollectionsCache).toHaveBeenCalledTimes(1)
  })

  test("returns 400 for invalid collection name", async () => {
    const request = new Request("http://localhost:3000/api/rag/collections/INVALID%20NAME!", {
      method: "DELETE",
      headers: {
        "x-user-id": "1",
        "x-user-email": "admin@localhost",
        "x-user-name": "Admin",
        "x-user-role": "admin",
      },
    })

    const response = await DELETE(request, {
      params: Promise.resolve({ name: "INVALID NAME!" }),
    })

    expect(response.status).toBe(400)
    const json = await response.json()
    expect(json.ok).toBe(false)
  })

  test("returns 403 when non-admin tries to delete", async () => {
    const request = new Request("http://localhost:3000/api/rag/collections/docs", {
      method: "DELETE",
      headers: {
        "x-user-id": "10",
        "x-user-email": "user@localhost",
        "x-user-name": "Regular User",
        "x-user-role": "user",
      },
    })

    const response = await DELETE(request, {
      params: Promise.resolve({ name: "docs" }),
    })

    expect(response.status).toBe(403)
  })

  test("returns 401 when not authenticated", async () => {
    const request = new Request("http://localhost:3000/api/rag/collections/docs", {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
    })

    const response = await DELETE(request, {
      params: Promise.resolve({ name: "docs" }),
    })

    expect(response.status).toBe(401)
  })

  test("returns 502 when RAG server returns error on delete", async () => {
    mockRagFetch.mockImplementation(() =>
      Promise.resolve({ error: { code: "UNAVAILABLE", message: "Connection refused", suggestion: "" } })
    )

    const request = new Request("http://localhost:3000/api/rag/collections/docs", {
      method: "DELETE",
      headers: {
        "x-user-id": "1",
        "x-user-email": "admin@localhost",
        "x-user-name": "Admin",
        "x-user-role": "admin",
      },
    })

    const response = await DELETE(request, {
      params: Promise.resolve({ name: "docs" }),
    })

    expect(response.status).toBe(502)
    const json = await response.json()
    expect(json.ok).toBe(false)
  })
})
