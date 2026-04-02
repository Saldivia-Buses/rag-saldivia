/**
 * Tests for auth route handlers: login, refresh, logout.
 * Plan 26: access (15m) + refresh (7d) token rotation.
 *
 * Run: cd apps/web && bun test ./src/__tests__/auth-routes.test.ts
 */

import { describe, test, expect, mock } from "bun:test"

// ── Env setup (BEFORE any imports that read env) ────────────────────────────
process.env["JWT_SECRET"] = "test-secret-key-at-least-16-chars"
process.env["NODE_ENV"] = "test"
process.env["REDIS_URL"] = "redis://localhost:6379"

// ── Mock data ───────────────────────────────────────────────────────────────

const MOCK_USER = {
  id: 42,
  email: "user@test.com",
  name: "Test User",
  role: "user",
  active: true,
  passwordHash: "hashed",
  preferences: { defaultCollection: "default" },
  createdAt: Date.now(),
  lastLogin: Date.now(),
  userAreas: [],
}

const MOCK_INACTIVE_USER = {
  ...MOCK_USER,
  id: 99,
  email: "inactive@test.com",
  active: false,
}

// ── Mock @rag-saldivia/db ───────────────────────────────────────────────────

const mockRedisStore = new Map<string, string>()

const mockRedisClient = {
  get: mock(async (key: string) => mockRedisStore.get(key) ?? null),
  set: mock(async (key: string, value: string, _ex?: string, _ttl?: number) => {
    mockRedisStore.set(key, value)
    return "OK"
  }),
}

mock.module("@rag-saldivia/db", () => ({
  verifyPassword: mock(async (email: string, password: string) => {
    if (email === MOCK_USER.email && password === "correct-password") {
      return MOCK_USER
    }
    return null
  }),
  getUserByEmail: mock(async (email: string) => {
    if (email === MOCK_USER.email) return MOCK_USER
    if (email === MOCK_INACTIVE_USER.email) return MOCK_INACTIVE_USER
    return null
  }),
  getUserById: mock(async (id: number) => {
    if (id === MOCK_USER.id) return MOCK_USER
    if (id === MOCK_INACTIVE_USER.id) return MOCK_INACTIVE_USER
    return null
  }),
  getRedisClient: () => mockRedisClient,
}))

// ── Mock @rag-saldivia/logger/backend ───────────────────────────────────────

mock.module("@rag-saldivia/logger/backend", () => ({
  log: {
    info: mock(() => {}),
    warn: mock(() => {}),
    error: mock(() => {}),
    debug: mock(() => {}),
    fatal: mock(() => {}),
  },
}))

// ── Dynamic imports AFTER mocks ─────────────────────────────────────────────

const { POST: loginHandler } = await import("../app/api/auth/login/route.js")
const { POST: refreshHandler } = await import("../app/api/auth/refresh/route.js")
const { DELETE: logoutHandler } = await import("../app/api/auth/logout/route.js")
const { createAccessToken, createRefreshToken, verifyJwt } =
  await import("../lib/auth/jwt.js")

// ── Helpers ─────────────────────────────────────────────────────────────────

function getAllSetCookieHeaders(response: Response): string[] {
  // getSetCookie() is the standard way to get multiple Set-Cookie headers
  if ("getSetCookie" in response.headers && typeof response.headers.getSetCookie === "function") {
    return response.headers.getSetCookie()
  }
  // Fallback: iterate
  const cookies: string[] = []
  response.headers.forEach((value, key) => {
    if (key.toLowerCase() === "set-cookie") {
      cookies.push(value)
    }
  })
  return cookies
}

// ═════════════════════════════════════════════════════════════════════════════
// LOGIN (POST /api/auth/login)
// ═════════════════════════════════════════════════════════════════════════════

describe("POST /api/auth/login", () => {
  test("valid credentials -> 200 with both Set-Cookie headers", async () => {
    const request = new Request("http://localhost:3000/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: MOCK_USER.email, password: "correct-password" }),
    })

    const response = await loginHandler(request)
    const body = await response.json()

    expect(response.status).toBe(200)
    expect(body.ok).toBe(true)
    expect(body.data.user.id).toBe(42)
    expect(body.data.user.email).toBe("user@test.com")
    expect(body.data.user.name).toBe("Test User")

    // Both cookies must be set
    const cookies = getAllSetCookieHeaders(response)
    const authCookie = cookies.find((c) => c.startsWith("auth_token="))
    const refreshCookie = cookies.find((c) => c.startsWith("refresh_token="))

    expect(authCookie).toBeDefined()
    expect(refreshCookie).toBeDefined()
  })

  test("auth_token cookie has SameSite=Lax", async () => {
    const request = new Request("http://localhost:3000/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: MOCK_USER.email, password: "correct-password" }),
    })

    const response = await loginHandler(request)
    const cookies = getAllSetCookieHeaders(response)
    const authCookie = cookies.find((c) => c.startsWith("auth_token="))

    expect(authCookie).toContain("SameSite=Lax")
    expect(authCookie).toContain("HttpOnly")
    expect(authCookie).toContain("Path=/")
  })

  test("refresh_token cookie has SameSite=Strict and Path=/api/auth/refresh", async () => {
    const request = new Request("http://localhost:3000/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: MOCK_USER.email, password: "correct-password" }),
    })

    const response = await loginHandler(request)
    const cookies = getAllSetCookieHeaders(response)
    const refreshCookie = cookies.find((c) => c.startsWith("refresh_token="))

    expect(refreshCookie).toContain("SameSite=Strict")
    expect(refreshCookie).toContain("Path=/api/auth/refresh")
    expect(refreshCookie).toContain("HttpOnly")
  })

  test("invalid email -> 401", async () => {
    const request = new Request("http://localhost:3000/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: "nonexistent@test.com", password: "anything" }),
    })

    const response = await loginHandler(request)
    const body = await response.json()

    expect(response.status).toBe(401)
    expect(body.ok).toBe(false)
    expect(body.error).toContain("incorrectos")
  })

  test("invalid password -> 401", async () => {
    const request = new Request("http://localhost:3000/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: MOCK_USER.email, password: "wrong-password" }),
    })

    const response = await loginHandler(request)
    const body = await response.json()

    expect(response.status).toBe(401)
    expect(body.ok).toBe(false)
  })

  test("inactive account -> 403", async () => {
    const request = new Request("http://localhost:3000/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: MOCK_INACTIVE_USER.email, password: "anything" }),
    })

    const response = await loginHandler(request)
    const body = await response.json()

    expect(response.status).toBe(403)
    expect(body.ok).toBe(false)
    expect(body.error).toContain("desactivada")
  })

  test("invalid body (missing fields) -> 400", async () => {
    const request = new Request("http://localhost:3000/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: "no-password" }),
    })

    const response = await loginHandler(request)
    const body = await response.json()

    expect(response.status).toBe(400)
    expect(body.ok).toBe(false)
    expect(body.error).toContain("inv\u00e1lidos")
  })

  test("invalid body (empty object) -> 400", async () => {
    const request = new Request("http://localhost:3000/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({}),
    })

    const response = await loginHandler(request)
    expect(response.status).toBe(400)
  })
})

// ═════════════════════════════════════════════════════════════════════════════
// REFRESH (POST /api/auth/refresh)
// ═════════════════════════════════════════════════════════════════════════════

describe("POST /api/auth/refresh", () => {
  test("valid refresh token in cookie -> 200 with new token pair", async () => {
    const refreshToken = await createRefreshToken(String(MOCK_USER.id))

    const request = new Request("http://localhost:3000/api/auth/refresh", {
      method: "POST",
      headers: {
        cookie: `refresh_token=${encodeURIComponent(refreshToken)}`,
      },
    })

    const response = await refreshHandler(request)
    const body = await response.json()

    expect(response.status).toBe(200)
    expect(body.ok).toBe(true)

    const cookies = getAllSetCookieHeaders(response)
    const authCookie = cookies.find((c) => c.startsWith("auth_token="))
    const newRefreshCookie = cookies.find((c) => c.startsWith("refresh_token="))

    expect(authCookie).toBeDefined()
    expect(newRefreshCookie).toBeDefined()
  })

  test("no refresh cookie -> 401", async () => {
    const request = new Request("http://localhost:3000/api/auth/refresh", {
      method: "POST",
    })

    const response = await refreshHandler(request)
    const body = await response.json()

    expect(response.status).toBe(401)
    expect(body.ok).toBe(false)
    expect(body.error).toContain("No refresh token")
  })

  test("invalid/expired refresh token -> 401", async () => {
    const request = new Request("http://localhost:3000/api/auth/refresh", {
      method: "POST",
      headers: {
        cookie: "refresh_token=invalid.token.garbage",
      },
    })

    const response = await refreshHandler(request)
    const body = await response.json()

    expect(response.status).toBe(401)
    expect(body.ok).toBe(false)
  })

  test("access token sent as refresh (type != 'refresh') -> 401", async () => {
    // Create an access token (type: "access") and try to use it as a refresh token
    const accessToken = await createAccessToken({
      sub: String(MOCK_USER.id),
      email: MOCK_USER.email,
      name: MOCK_USER.name,
      role: MOCK_USER.role as "admin" | "area_manager" | "user",
    })

    const request = new Request("http://localhost:3000/api/auth/refresh", {
      method: "POST",
      headers: {
        cookie: `refresh_token=${encodeURIComponent(accessToken)}`,
      },
    })

    const response = await refreshHandler(request)
    const body = await response.json()

    expect(response.status).toBe(401)
    expect(body.ok).toBe(false)
    expect(body.error).toContain("inv\u00e1lido")
  })

  test("refresh rotates tokens: old refresh JTI gets revoked in Redis", async () => {
    mockRedisStore.clear()
    mockRedisClient.set.mockClear()

    const refreshToken = await createRefreshToken(String(MOCK_USER.id))
    const refreshClaims = await verifyJwt(refreshToken)

    const request = new Request("http://localhost:3000/api/auth/refresh", {
      method: "POST",
      headers: {
        cookie: `refresh_token=${encodeURIComponent(refreshToken)}`,
      },
    })

    await refreshHandler(request)

    // The old refresh token's JTI should have been revoked in Redis
    expect(mockRedisClient.set).toHaveBeenCalled()
    const setCall = mockRedisClient.set.mock.calls.find(
      (call: unknown[]) => (call[0] as string) === `revoked:${refreshClaims!.jti}`
    )
    expect(setCall).toBeDefined()
  })
})

// ═════════════════════════════════════════════════════════════════════════════
// LOGOUT (DELETE /api/auth/logout)
// ═════════════════════════════════════════════════════════════════════════════

describe("DELETE /api/auth/logout", () => {
  test("with valid tokens -> 200 with both cookies cleared (Max-Age=0)", async () => {
    const accessToken = await createAccessToken({
      sub: String(MOCK_USER.id),
      email: MOCK_USER.email,
      name: MOCK_USER.name,
      role: MOCK_USER.role as "admin" | "area_manager" | "user",
    })
    const refreshToken = await createRefreshToken(String(MOCK_USER.id))

    const request = new Request("http://localhost:3000/api/auth/logout", {
      method: "DELETE",
      headers: {
        cookie: `auth_token=${encodeURIComponent(accessToken)}; refresh_token=${encodeURIComponent(refreshToken)}`,
      },
    })

    const response = await logoutHandler(request)
    const body = await response.json()

    expect(response.status).toBe(200)
    expect(body.ok).toBe(true)

    const cookies = getAllSetCookieHeaders(response)
    const clearedAuth = cookies.find((c) => c.startsWith("auth_token="))
    const clearedRefresh = cookies.find((c) => c.startsWith("refresh_token="))

    expect(clearedAuth).toBeDefined()
    expect(clearedAuth).toContain("Max-Age=0")
    expect(clearedRefresh).toBeDefined()
    expect(clearedRefresh).toContain("Max-Age=0")
  })

  test("without tokens -> 200 (idempotent)", async () => {
    const request = new Request("http://localhost:3000/api/auth/logout", {
      method: "DELETE",
    })

    const response = await logoutHandler(request)
    const body = await response.json()

    expect(response.status).toBe(200)
    expect(body.ok).toBe(true)

    // Cookies should still be cleared even if no tokens were present
    const cookies = getAllSetCookieHeaders(response)
    const clearedAuth = cookies.find((c) => c.startsWith("auth_token="))
    const clearedRefresh = cookies.find((c) => c.startsWith("refresh_token="))

    expect(clearedAuth).toContain("Max-Age=0")
    expect(clearedRefresh).toContain("Max-Age=0")
  })

  test("logout revokes both access and refresh JTIs in Redis", async () => {
    mockRedisStore.clear()
    mockRedisClient.set.mockClear()

    const accessToken = await createAccessToken({
      sub: String(MOCK_USER.id),
      email: MOCK_USER.email,
      name: MOCK_USER.name,
      role: MOCK_USER.role as "admin" | "area_manager" | "user",
    })
    const refreshToken = await createRefreshToken(String(MOCK_USER.id))

    const accessClaims = await verifyJwt(accessToken)
    const refreshClaims = await verifyJwt(refreshToken)

    const request = new Request("http://localhost:3000/api/auth/logout", {
      method: "DELETE",
      headers: {
        cookie: `auth_token=${encodeURIComponent(accessToken)}; refresh_token=${encodeURIComponent(refreshToken)}`,
      },
    })

    await logoutHandler(request)

    // Both JTIs should be revoked
    const setCalls = mockRedisClient.set.mock.calls.map((c: unknown[]) => c[0] as string)
    expect(setCalls).toContain(`revoked:${accessClaims!.jti}`)
    expect(setCalls).toContain(`revoked:${refreshClaims!.jti}`)
  })
})
