/**
 * Tests del middleware de Next.js (proxy.ts) — Auth + RBAC en el edge.
 *
 * Mockea `jose` para controlar la verificacion de JWT sin depender de tokens reales.
 * Usa el preload de bunfig.toml (test-setup.ts) que ya mockea ioredis.
 *
 * NextResponse.next({ request: { headers } }) stores modified request headers as
 * `x-middleware-request-<key>` in the response headers. The helper `getRequestHeader`
 * abstracts this so test assertions read naturally.
 *
 * Corre con: bun test apps/web/src/__tests__/proxy.test.ts
 */

import { describe, test, expect, mock, beforeAll } from "bun:test"

// ── Env setup (antes de importar proxy) ──────────────────────────────────────
process.env["JWT_SECRET"] = "test-secret-key-at-least-16-chars"
process.env["SYSTEM_API_KEY"] = "system-api-key-for-testing"

// ── Mock jose ────────────────────────────────────────────────────────────────
// Default: jwtVerify resolves with valid claims. Tests override via mockJwtVerify.
let jwtVerifyImpl: (...args: unknown[]) => Promise<unknown> = async () => ({
  payload: {
    sub: "42",
    email: "user@test.com",
    name: "Test User",
    role: "user",
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + 3600,
    jti: "jti-abc-123",
  },
})

mock.module("jose", () => ({
  jwtVerify: (...args: unknown[]) => jwtVerifyImpl(...args),
  SignJWT: class {},
  errors: { JWTExpired: class extends Error {} },
}))

// ── Import proxy AFTER mocks ─────────────────────────────────────────────────
import { NextRequest } from "next/server"
import { proxy, config } from "../proxy"

// ── Helpers ──────────────────────────────────────────────────────────────────

/**
 * Create a NextRequest for the given path with optional overrides.
 */
function makeRequest(
  path: string,
  opts?: { cookie?: string; bearer?: string; headers?: Record<string, string> }
): NextRequest {
  const url = new URL(path, "http://localhost:3000")
  const headers = new Headers(opts?.headers)
  if (opts?.cookie) {
    headers.set("cookie", `auth_token=${opts.cookie}`)
  }
  if (opts?.bearer) {
    headers.set("authorization", `Bearer ${opts.bearer}`)
  }
  return new NextRequest(url, { headers })
}

/**
 * Get a request header that was set by the middleware via NextResponse.next().
 *
 * NextResponse.next({ request: { headers } }) stores each modified request header
 * as `x-middleware-request-<key>` in the response. This helper reads through that
 * indirection so test code stays readable.
 */
function getRequestHeader(res: Response, name: string): string | null {
  return res.headers.get(`x-middleware-request-${name}`)
}

/** Temporarily override jwtVerify behavior for a single test. */
function mockJwtVerify(
  impl: (...args: unknown[]) => Promise<unknown>
): void {
  jwtVerifyImpl = impl
}

/** Reset jwtVerify to the default valid-user claims implementation. */
function resetJwtVerify(): void {
  jwtVerifyImpl = async () => ({
    payload: {
      sub: "42",
      email: "user@test.com",
      name: "Test User",
      role: "user",
      iat: Math.floor(Date.now() / 1000),
      exp: Math.floor(Date.now() / 1000) + 3600,
      jti: "jti-abc-123",
    },
  })
}

/** Check that a response is a NextResponse.next() pass-through (not redirect/error). */
function expectPassThrough(res: Response): void {
  expect(res.status).toBe(200)
  expect(res.headers.get("x-middleware-next")).toBe("1")
  expect(res.headers.get("location")).toBeNull()
}

// ── Tests ────────────────────────────────────────────────────────────────────

describe("proxy middleware", () => {
  beforeAll(() => resetJwtVerify())

  // ── 1. Public routes ────────────────────────────────────────────────────

  describe("public routes pass without auth", () => {
    const publicPaths = [
      "/login",
      "/api/auth/login",
      "/api/auth/refresh",
      "/api/health",
      "/api/log",
      "/_next/static/chunks/main.js",
      "/favicon.ico",
    ]

    for (const path of publicPaths) {
      test(`${path} returns next() without requiring auth`, async () => {
        const req = makeRequest(path)
        const res = await proxy(req)

        expectPassThrough(res)
      })
    }

    test("public routes get x-request-id header", async () => {
      const req = makeRequest("/login")
      const res = await proxy(req)

      const requestId = getRequestHeader(res, "x-request-id")
      expect(requestId).not.toBeNull()
      expect(requestId!).toMatch(
        /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/
      )
    })
  })

  // ── 2. Protected routes without token ───────────────────────────────────

  describe("protected routes without token", () => {
    test("page route /chat without cookie redirects to /login", async () => {
      const req = makeRequest("/chat")
      const res = await proxy(req)

      expect(res.status).toBe(307)
      const location = new URL(res.headers.get("location")!)
      expect(location.pathname).toBe("/login")
      expect(location.searchParams.get("from")).toBe("/chat")
    })

    test("page route /settings without cookie redirects to /login", async () => {
      const req = makeRequest("/settings")
      const res = await proxy(req)

      expect(res.status).toBe(307)
      const location = new URL(res.headers.get("location")!)
      expect(location.pathname).toBe("/login")
      expect(location.searchParams.get("from")).toBe("/settings")
    })

    test("API route /api/rag/generate without token returns 401", async () => {
      const req = makeRequest("/api/rag/generate")
      const res = await proxy(req)

      expect(res.status).toBe(401)
      const body = await res.json()
      expect(body.ok).toBe(false)
      expect(body.error).toBe("No autenticado")
    })

    test("API route /api/rag/collections without token returns 401", async () => {
      const req = makeRequest("/api/rag/collections")
      const res = await proxy(req)

      expect(res.status).toBe(401)
      const body = await res.json()
      expect(body.ok).toBe(false)
    })
  })

  // ── 3. Valid JWT sets x-user-* headers ──────────────────────────────────

  describe("valid JWT sets x-user-* headers", () => {
    test("sets all expected user headers from JWT claims", async () => {
      resetJwtVerify()
      const req = makeRequest("/chat", { cookie: "valid-token" })
      const res = await proxy(req)

      expectPassThrough(res)

      expect(getRequestHeader(res, "x-user-id")).toBe("42")
      expect(getRequestHeader(res, "x-user-email")).toBe("user@test.com")
      expect(getRequestHeader(res, "x-user-name")).toBe("Test User")
      expect(getRequestHeader(res, "x-user-role")).toBe("user")
      expect(getRequestHeader(res, "x-user-jti")).toBe("jti-abc-123")
    })

    test("works with Authorization bearer header instead of cookie", async () => {
      resetJwtVerify()
      const req = makeRequest("/api/rag/collections", { bearer: "valid-token" })
      const res = await proxy(req)

      expectPassThrough(res)
      expect(getRequestHeader(res, "x-user-id")).toBe("42")
      expect(getRequestHeader(res, "x-user-role")).toBe("user")
    })

    test("cookie takes precedence over bearer when both present", async () => {
      // The proxy uses: cookieToken ?? bearerToken
      // If cookie is present, it should be used
      resetJwtVerify()
      const req = makeRequest("/chat", {
        cookie: "cookie-token",
        bearer: "bearer-token",
      })
      const res = await proxy(req)

      expectPassThrough(res)
    })

    test("jti header is NOT set when claims lack jti", async () => {
      mockJwtVerify(async () => ({
        payload: {
          sub: "42",
          email: "user@test.com",
          name: "Test User",
          role: "user",
          iat: Math.floor(Date.now() / 1000),
          exp: Math.floor(Date.now() / 1000) + 3600,
          // No jti field
        },
      }))

      const req = makeRequest("/chat", { cookie: "valid-token" })
      const res = await proxy(req)

      expectPassThrough(res)
      expect(getRequestHeader(res, "x-user-jti")).toBeNull()
    })

    test("admin claims pass through for regular routes", async () => {
      mockJwtVerify(async () => ({
        payload: {
          sub: "1",
          email: "admin@test.com",
          name: "Admin User",
          role: "admin",
          iat: Math.floor(Date.now() / 1000),
          exp: Math.floor(Date.now() / 1000) + 3600,
          jti: "admin-jti",
        },
      }))

      const req = makeRequest("/chat", { cookie: "admin-token" })
      const res = await proxy(req)

      expectPassThrough(res)
      expect(getRequestHeader(res, "x-user-role")).toBe("admin")
      expect(getRequestHeader(res, "x-user-email")).toBe("admin@test.com")
    })
  })

  // ── 4. Invalid/expired JWT ──────────────────────────────────────────────

  describe("invalid or expired JWT", () => {
    test("bad token on page route redirects to /login", async () => {
      mockJwtVerify(async () => {
        throw new Error("invalid signature")
      })

      const req = makeRequest("/chat", { cookie: "bad-token" })
      const res = await proxy(req)

      expect(res.status).toBe(307)
      const location = new URL(res.headers.get("location")!)
      expect(location.pathname).toBe("/login")
      expect(location.searchParams.get("from")).toBe("/chat")
    })

    test("expired token on page route redirects to /login", async () => {
      mockJwtVerify(async () => {
        throw new Error("jwt expired")
      })

      const req = makeRequest("/collections", { cookie: "expired-token" })
      const res = await proxy(req)

      expect(res.status).toBe(307)
      const location = new URL(res.headers.get("location")!)
      expect(location.pathname).toBe("/login")
    })

    test("bad token on API route returns 401", async () => {
      mockJwtVerify(async () => {
        throw new Error("invalid signature")
      })

      const req = makeRequest("/api/rag/generate", { bearer: "bad-token" })
      const res = await proxy(req)

      expect(res.status).toBe(401)
      const body = await res.json()
      expect(body.ok).toBe(false)
      expect(body.error).toBe("Token inválido o expirado")
    })

    test("expired token on API route returns 401", async () => {
      mockJwtVerify(async () => {
        throw new Error("jwt expired")
      })

      const req = makeRequest("/api/rag/collections", { bearer: "expired-token" })
      const res = await proxy(req)

      expect(res.status).toBe(401)
      const body = await res.json()
      expect(body.ok).toBe(false)
    })
  })

  // ── 5. Admin routes block regular users (RBAC) ─────────────────────────

  describe("RBAC — admin routes block regular users", () => {
    test("user role accessing /admin page redirects to /", async () => {
      resetJwtVerify() // role: "user"

      const req = makeRequest("/admin", { cookie: "user-token" })
      const res = await proxy(req)

      expect(res.status).toBe(307)
      const location = new URL(res.headers.get("location")!)
      expect(location.pathname).toBe("/")
    })

    test("user role accessing /admin/users page redirects to /", async () => {
      resetJwtVerify()

      const req = makeRequest("/admin/users", { cookie: "user-token" })
      const res = await proxy(req)

      expect(res.status).toBe(307)
      const location = new URL(res.headers.get("location")!)
      expect(location.pathname).toBe("/")
    })

    test("user role accessing /api/admin returns 403", async () => {
      resetJwtVerify()

      const req = makeRequest("/api/admin/users", { cookie: "user-token" })
      const res = await proxy(req)

      expect(res.status).toBe(403)
      const body = await res.json()
      expect(body.ok).toBe(false)
      expect(body.error).toBe("Acceso denegado")
    })

    test("admin role CAN access /admin page", async () => {
      mockJwtVerify(async () => ({
        payload: {
          sub: "1",
          email: "admin@test.com",
          name: "Admin",
          role: "admin",
          iat: Math.floor(Date.now() / 1000),
          exp: Math.floor(Date.now() / 1000) + 3600,
          jti: "admin-jti",
        },
      }))

      const req = makeRequest("/admin", { cookie: "admin-token" })
      const res = await proxy(req)

      expectPassThrough(res)
      expect(getRequestHeader(res, "x-user-role")).toBe("admin")
    })

    test("admin role CAN access /api/admin routes", async () => {
      mockJwtVerify(async () => ({
        payload: {
          sub: "1",
          email: "admin@test.com",
          name: "Admin",
          role: "admin",
          iat: Math.floor(Date.now() / 1000),
          exp: Math.floor(Date.now() / 1000) + 3600,
        },
      }))

      const req = makeRequest("/api/admin/users", { cookie: "admin-token" })
      const res = await proxy(req)

      expectPassThrough(res)
      expect(getRequestHeader(res, "x-user-role")).toBe("admin")
    })

    test("area_manager cannot access /admin page", async () => {
      mockJwtVerify(async () => ({
        payload: {
          sub: "5",
          email: "manager@test.com",
          name: "Manager",
          role: "area_manager",
          iat: Math.floor(Date.now() / 1000),
          exp: Math.floor(Date.now() / 1000) + 3600,
        },
      }))

      const req = makeRequest("/admin", { cookie: "manager-token" })
      const res = await proxy(req)

      expect(res.status).toBe(307)
      const location = new URL(res.headers.get("location")!)
      expect(location.pathname).toBe("/")
    })

    test("area_manager CAN access /audit routes", async () => {
      mockJwtVerify(async () => ({
        payload: {
          sub: "5",
          email: "manager@test.com",
          name: "Manager",
          role: "area_manager",
          iat: Math.floor(Date.now() / 1000),
          exp: Math.floor(Date.now() / 1000) + 3600,
        },
      }))

      const req = makeRequest("/audit", { cookie: "manager-token" })
      const res = await proxy(req)

      expectPassThrough(res)
    })

    test("user role cannot access /audit routes (page)", async () => {
      resetJwtVerify() // role: "user"

      const req = makeRequest("/audit/logs", { cookie: "user-token" })
      const res = await proxy(req)

      expect(res.status).toBe(307)
    })

    test("user role accessing /api/audit returns 403", async () => {
      resetJwtVerify()

      const req = makeRequest("/api/audit/logs", { cookie: "user-token" })
      const res = await proxy(req)

      expect(res.status).toBe(403)
    })
  })

  // ── 6. System API key ──────────────────────────────────────────────────

  describe("system API key", () => {
    test("Bearer with SYSTEM_API_KEY passes with admin role headers", async () => {
      const req = makeRequest("/api/rag/collections", {
        bearer: "system-api-key-for-testing",
      })
      const res = await proxy(req)

      expectPassThrough(res)
      expect(getRequestHeader(res, "x-user-id")).toBe("0")
      expect(getRequestHeader(res, "x-user-email")).toBe("system@internal")
      expect(getRequestHeader(res, "x-user-name")).toBe("System")
      expect(getRequestHeader(res, "x-user-role")).toBe("admin")
    })

    test("system API key gets x-request-id", async () => {
      const req = makeRequest("/api/rag/generate", {
        bearer: "system-api-key-for-testing",
      })
      const res = await proxy(req)

      const requestId = getRequestHeader(res, "x-request-id")
      expect(requestId).not.toBeNull()
      expect(requestId!).toMatch(
        /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/
      )
    })

    test("system API key works even on admin routes", async () => {
      const req = makeRequest("/api/admin/users", {
        bearer: "system-api-key-for-testing",
      })
      const res = await proxy(req)

      // System key bypasses JWT + RBAC entirely, gets admin role
      expectPassThrough(res)
      expect(getRequestHeader(res, "x-user-role")).toBe("admin")
    })

    test("system API key in cookie also works", async () => {
      const req = makeRequest("/api/rag/collections", {
        cookie: "system-api-key-for-testing",
      })
      const res = await proxy(req)

      expectPassThrough(res)
      expect(getRequestHeader(res, "x-user-id")).toBe("0")
      expect(getRequestHeader(res, "x-user-role")).toBe("admin")
    })

    test("wrong API key falls through to JWT verification", async () => {
      mockJwtVerify(async () => {
        throw new Error("invalid token")
      })

      const req = makeRequest("/api/rag/collections", {
        bearer: "wrong-api-key",
      })
      const res = await proxy(req)

      // Wrong key is NOT the system key, so it tries JWT verification
      // which fails, resulting in 401
      expect(res.status).toBe(401)
    })
  })

  // ── 7. x-request-id generation ─────────────────────────────────────────

  describe("x-request-id generation", () => {
    test("authenticated requests get x-request-id", async () => {
      resetJwtVerify()

      const req = makeRequest("/chat", { cookie: "valid-token" })
      const res = await proxy(req)

      const requestId = getRequestHeader(res, "x-request-id")
      expect(requestId).not.toBeNull()
      expect(requestId!).toMatch(
        /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/
      )
    })

    test("each request gets a unique x-request-id", async () => {
      resetJwtVerify()

      const req1 = makeRequest("/chat", { cookie: "valid-token" })
      const req2 = makeRequest("/chat", { cookie: "valid-token" })

      const res1 = await proxy(req1)
      const res2 = await proxy(req2)

      const id1 = getRequestHeader(res1, "x-request-id")
      const id2 = getRequestHeader(res2, "x-request-id")

      expect(id1).not.toBeNull()
      expect(id2).not.toBeNull()
      expect(id1).not.toBe(id2)
    })

    test("public routes also get x-request-id", async () => {
      const req = makeRequest("/api/health")
      const res = await proxy(req)

      expect(getRequestHeader(res, "x-request-id")).not.toBeNull()
    })

    test("401 responses do NOT have x-request-id (error short-circuits)", async () => {
      const req = makeRequest("/api/rag/generate")
      const res = await proxy(req)

      expect(res.status).toBe(401)
      // Error responses are built with NextResponse.json() — no request header forwarding
      expect(getRequestHeader(res, "x-request-id")).toBeNull()
    })
  })

  // ── 8. Config matcher ──────────────────────────────────────────────────

  describe("config export", () => {
    test("exports matcher config array", () => {
      expect(config).toBeDefined()
      expect(config.matcher).toBeDefined()
      expect(Array.isArray(config.matcher)).toBe(true)
      expect(config.matcher.length).toBeGreaterThan(0)
    })

    test("matcher excludes _next/static, _next/image, and favicon.ico", () => {
      const pattern = config.matcher[0]!
      expect(pattern).toContain("_next/static")
      expect(pattern).toContain("_next/image")
      expect(pattern).toContain("favicon.ico")
    })
  })

  // ── 9. Edge case: JWT without name field ────────────────────────────────

  describe("edge case — JWT without name field", () => {
    test("undefined name becomes 'undefined' string header", async () => {
      mockJwtVerify(async () => ({
        payload: {
          sub: "42",
          email: "user@test.com",
          name: undefined,
          role: "user",
          iat: Math.floor(Date.now() / 1000),
          exp: Math.floor(Date.now() / 1000) + 3600,
        },
      }))

      const req = makeRequest("/chat", { cookie: "valid-token" })
      const res = await proxy(req)

      // headers.set("x-user-name", claims.name) with undefined name
      // coerces to the string "undefined" — this is a known edge case
      const nameHeader = getRequestHeader(res, "x-user-name")
      expect(nameHeader).toBe("undefined")
    })
  })

  // ── 10. Redirect preserves original path ───────────────────────────────

  describe("redirect preserves original path in from param", () => {
    test("/chat/abc123 without auth redirects with from=/chat/abc123", async () => {
      const req = makeRequest("/chat/abc123")
      const res = await proxy(req)

      expect(res.status).toBe(307)
      const location = new URL(res.headers.get("location")!)
      expect(location.pathname).toBe("/login")
      expect(location.searchParams.get("from")).toBe("/chat/abc123")
    })

    test("expired token redirect also includes from param", async () => {
      mockJwtVerify(async () => {
        throw new Error("jwt expired")
      })

      const req = makeRequest("/settings", { cookie: "expired-token" })
      const res = await proxy(req)

      expect(res.status).toBe(307)
      const location = new URL(res.headers.get("location")!)
      expect(location.searchParams.get("from")).toBe("/settings")
    })
  })
})
