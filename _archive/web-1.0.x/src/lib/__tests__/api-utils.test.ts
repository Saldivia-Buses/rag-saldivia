/**
 * Tests for api-utils.ts — shared helpers for API route handlers.
 * Corre con: bun test apps/web/src/lib/__tests__/api-utils.test.ts
 */

import { describe, test, expect, mock, beforeEach } from "bun:test"
import type { JwtClaims } from "@rag-saldivia/shared"

// ── Mocks ────────────────────────────────────────────────────────────────────

// Track calls to log.error for apiServerError tests
const mockLogError = mock(() => {})

mock.module("@rag-saldivia/logger/backend", () => ({
  log: {
    info: mock(() => {}),
    warn: mock(() => {}),
    error: mockLogError,
  },
}))

// Mock extractClaims — controlled per-test via mockExtractClaims
let mockExtractClaimsResult: JwtClaims | null = null

mock.module("@/lib/auth/jwt", () => ({
  extractClaims: async () => mockExtractClaimsResult,
}))

// Import AFTER mocks are set up
import { requireAuth, requireAdmin, apiOk, apiError, apiServerError } from "../api-utils"

// ── Helpers ──────────────────────────────────────────────────────────────────

function makeRequest(path = "/"): Request {
  return new Request(`http://localhost${path}`)
}

const validUserClaims: JwtClaims = {
  sub: "42",
  email: "user@test.com",
  name: "Test User",
  role: "user",
  iat: Math.floor(Date.now() / 1000),
  exp: Math.floor(Date.now() / 1000) + 3600,
  jti: "test-jti-123",
}

const validAdminClaims: JwtClaims = {
  sub: "1",
  email: "admin@test.com",
  name: "Admin User",
  role: "admin",
  iat: Math.floor(Date.now() / 1000),
  exp: Math.floor(Date.now() / 1000) + 3600,
  jti: "admin-jti-456",
}

// ── Tests ────────────────────────────────────────────────────────────────────

describe("apiOk", () => {
  test("returns JSON with { ok: true, data } when data is provided", async () => {
    const response = apiOk({ users: [{ id: 1 }] })
    expect(response.status).toBe(200)
    const body = await response.json()
    expect(body.ok).toBe(true)
    expect(body.data).toEqual({ users: [{ id: 1 }] })
  })

  test("returns JSON with { ok: true } without data field when called with no args", async () => {
    const response = apiOk()
    expect(response.status).toBe(200)
    const body = await response.json()
    expect(body.ok).toBe(true)
    expect(body.data).toBeUndefined()
  })

  test("uses custom status code when provided", async () => {
    const response = apiOk({ created: true }, 201)
    expect(response.status).toBe(201)
    const body = await response.json()
    expect(body.ok).toBe(true)
    expect(body.data).toEqual({ created: true })
  })

  test("returns content-type application/json", () => {
    const response = apiOk({ test: true })
    expect(response.headers.get("content-type")).toContain("application/json")
  })
})

describe("apiError", () => {
  test("returns JSON with { ok: false, error } for default 400", async () => {
    const response = apiError("Bad input")
    expect(response.status).toBe(400)
    const body = await response.json()
    expect(body.ok).toBe(false)
    expect(body.error).toBe("Bad input")
  })

  test("uses custom status code", async () => {
    const response = apiError("Not found", 404)
    expect(response.status).toBe(404)
    const body = await response.json()
    expect(body.error).toBe("Not found")
  })

  test("includes details when provided", async () => {
    const response = apiError("Validation failed", 422, { field: "email", issue: "invalid" })
    const body = await response.json()
    expect(body.ok).toBe(false)
    expect(body.error).toBe("Validation failed")
    expect(body.details).toEqual({ field: "email", issue: "invalid" })
  })

  test("omits details field when not provided", async () => {
    const response = apiError("Forbidden", 403)
    const body = await response.json()
    expect(body.details).toBeUndefined()
  })
})

describe("apiServerError", () => {
  beforeEach(() => {
    mockLogError.mockClear()
  })

  test("returns 500 with generic error message", async () => {
    const response = apiServerError(new Error("DB crashed"), "/api/users")
    expect(response.status).toBe(500)
    const body = await response.json()
    expect(body.ok).toBe(false)
    expect(body.error).toBe("Error interno del servidor")
  })

  test("calls log.error with endpoint context", () => {
    apiServerError(new Error("timeout"), "/api/rag/generate")
    expect(mockLogError).toHaveBeenCalledTimes(1)
    const [event, data] = mockLogError.mock.calls[0] as [string, Record<string, unknown>]
    expect(event).toBe("system.error")
    expect(data.endpoint).toBe("/api/rag/generate")
    expect(data.error).toContain("timeout")
  })

  test("passes userId context when provided", () => {
    apiServerError("something broke", "/api/chat", 42)
    expect(mockLogError).toHaveBeenCalledTimes(1)
    const args = mockLogError.mock.calls[0] as unknown[]
    // Third arg is the user context
    expect(args[2]).toEqual({ userId: 42 })
  })

  test("omits userId context when not provided", () => {
    apiServerError("something broke", "/api/chat")
    expect(mockLogError).toHaveBeenCalledTimes(1)
    const args = mockLogError.mock.calls[0] as unknown[]
    // Third arg should be undefined
    expect(args[2]).toBeUndefined()
  })

  test("handles string errors", async () => {
    const response = apiServerError("plain string error", "/api/test")
    expect(response.status).toBe(500)
    const [, data] = mockLogError.mock.calls[0] as [string, Record<string, unknown>]
    expect(data.error).toBe("plain string error")
  })
})

describe("requireAuth", () => {
  test("returns claims when user is authenticated", async () => {
    mockExtractClaimsResult = validUserClaims
    const result = await requireAuth(makeRequest())
    // Should not be a Response — should be claims
    expect(result).toHaveProperty("sub", "42")
    expect(result).toHaveProperty("email", "user@test.com")
    expect(result).toHaveProperty("role", "user")
  })

  test("returns 401 NextResponse when no auth", async () => {
    mockExtractClaimsResult = null
    const result = await requireAuth(makeRequest())
    // Should be a NextResponse with 401
    expect(result).toHaveProperty("status", 401)
    const body = await (result as Response).json()
    expect(body.ok).toBe(false)
    expect(body.error).toBe("No autenticado")
  })
})

describe("requireAdmin", () => {
  test("returns claims when user is admin", async () => {
    mockExtractClaimsResult = validAdminClaims
    const result = await requireAdmin(makeRequest())
    expect(result).toHaveProperty("sub", "1")
    expect(result).toHaveProperty("role", "admin")
  })

  test("returns 401 when user is not authenticated", async () => {
    mockExtractClaimsResult = null
    const result = await requireAdmin(makeRequest())
    expect(result).toHaveProperty("status", 401)
    const body = await (result as Response).json()
    expect(body.ok).toBe(false)
    expect(body.error).toBe("No autenticado")
  })

  test("returns 403 when authenticated user is not admin", async () => {
    mockExtractClaimsResult = validUserClaims
    const result = await requireAdmin(makeRequest())
    expect(result).toHaveProperty("status", 403)
    const body = await (result as Response).json()
    expect(body.ok).toBe(false)
    expect(body.error).toBe("Solo administradores")
  })

  test("returns 403 for area_manager role", async () => {
    mockExtractClaimsResult = { ...validUserClaims, role: "area_manager" }
    const result = await requireAdmin(makeRequest())
    expect(result).toHaveProperty("status", 403)
  })
})
