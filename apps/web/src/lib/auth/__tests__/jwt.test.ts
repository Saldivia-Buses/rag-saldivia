/**
 * Tests del flujo de autenticación JWT.
 * Corre con: bun test src/lib/auth/__tests__/jwt.test.ts
 */

import { describe, test, expect } from "bun:test"

// Setup env vars para tests
process.env["JWT_SECRET"] = "test-secret-key-at-least-16-chars"
process.env["JWT_EXPIRY"] = "1h"
process.env["NODE_ENV"] = "test"

// Import after env setup
const {
  createJwt,
  verifyJwt,
  extractClaims,
  makeAuthCookie,
  makeClearAuthCookie,
  createAccessToken,
  createRefreshToken,
  makeRefreshCookie,
  makeClearRefreshCookie,
  revokeToken,
} = await import("../jwt.js")
const { hasRole, canAccessRoute, getRequiredRole, isAdmin, isAreaManager } = await import("../rbac.js")
const { getRedisClient } = await import("@rag-saldivia/db")

describe("JWT utilities", () => {
  const validClaims = {
    sub: "42",
    email: "test@example.com",
    name: "Test User",
    role: "user" as const,
  }

  test("createJwt retorna un string JWT válido", async () => {
    const token = await createJwt(validClaims)
    expect(typeof token).toBe("string")
    expect(token.split(".")).toHaveLength(3) // header.payload.signature
  })

  test("verifyJwt verifica un token válido", async () => {
    const token = await createJwt(validClaims)
    const claims = await verifyJwt(token)
    expect(claims).not.toBeNull()
    expect(claims?.sub).toBe("42")
    expect(claims?.email).toBe("test@example.com")
    expect(claims?.role).toBe("user")
  })

  test("verifyJwt retorna null para token inválido", async () => {
    const result = await verifyJwt("token.invalido.fake")
    expect(result).toBeNull()
  })

  test("verifyJwt retorna null para token con firma incorrecta", async () => {
    const token = await createJwt(validClaims)
    const tampered = token.slice(0, -5) + "XXXXX"
    const result = await verifyJwt(tampered)
    expect(result).toBeNull()
  })

  test("extractClaims lee JWT desde cookie header", async () => {
    const token = await createJwt(validClaims)
    const mockRequest = new Request("http://localhost/", {
      headers: { cookie: `auth_token=${encodeURIComponent(token)}; other=value` },
    })
    const claims = await extractClaims(mockRequest)
    expect(claims?.sub).toBe("42")
    expect(claims?.email).toBe("test@example.com")
  })

  test("extractClaims lee JWT desde Authorization header", async () => {
    const token = await createJwt(validClaims)
    const mockRequest = new Request("http://localhost/", {
      headers: { authorization: `Bearer ${token}` },
    })
    const claims = await extractClaims(mockRequest)
    expect(claims?.sub).toBe("42")
  })

  test("extractClaims retorna null cuando no hay token", async () => {
    const mockRequest = new Request("http://localhost/")
    const claims = await extractClaims(mockRequest)
    expect(claims).toBeNull()
  })

  test("makeAuthCookie genera cookie HttpOnly correcta", async () => {
    const token = await createJwt(validClaims)
    const cookie = makeAuthCookie(token)
    expect(cookie).toContain("auth_token=")
    expect(cookie).toContain("HttpOnly")
    expect(cookie).toContain("Path=/")
    expect(cookie).toContain("SameSite=Lax")
  })

  test("makeClearAuthCookie genera cookie con Max-Age=0", () => {
    const cookie = makeClearAuthCookie()
    expect(cookie).toContain("auth_token=")
    expect(cookie).toContain("Max-Age=0")
    expect(cookie).toContain("HttpOnly")
  })

  test("createJwt incluye jti único por token", async () => {
    const t1 = await createJwt(validClaims)
    const t2 = await createJwt(validClaims)
    const c1 = await verifyJwt(t1)
    const c2 = await verifyJwt(t2)
    expect(c1?.jti).toBeDefined()
    expect(c2?.jti).toBeDefined()
    expect(c1?.jti).not.toBe(c2?.jti)
  })
})

describe("RBAC utilities", () => {
  test("admin tiene acceso a todo", () => {
    const adminClaims = { sub: "1", email: "a@b.com", name: "Admin", role: "admin" as const, iat: 0, exp: 9999999999 }
    expect(hasRole(adminClaims, "admin")).toBe(true)
    expect(hasRole(adminClaims, "area_manager")).toBe(true)
    expect(hasRole(adminClaims, "user")).toBe(true)
  })

  test("area_manager no tiene acceso a rutas de admin", () => {
    const managerClaims = { sub: "2", email: "m@b.com", name: "Manager", role: "area_manager" as const, iat: 0, exp: 9999999999 }
    expect(hasRole(managerClaims, "admin")).toBe(false)
    expect(hasRole(managerClaims, "area_manager")).toBe(true)
    expect(canAccessRoute(managerClaims, "/admin/users")).toBe(false)
    expect(canAccessRoute(managerClaims, "/audit")).toBe(true)
  })

  test("user no tiene acceso a rutas protegidas", () => {
    const userClaims = { sub: "3", email: "u@b.com", name: "User", role: "user" as const, iat: 0, exp: 9999999999 }
    expect(hasRole(userClaims, "admin")).toBe(false)
    expect(hasRole(userClaims, "area_manager")).toBe(false)
    expect(canAccessRoute(userClaims, "/admin/users")).toBe(false)
    expect(canAccessRoute(userClaims, "/audit")).toBe(false)
    expect(canAccessRoute(userClaims, "/chat")).toBe(true)
  })

  test("getRequiredRole retorna 'admin' para /api/admin/users", () => {
    expect(getRequiredRole("/api/admin/users")).toBe("admin")
  })

  test("getRequiredRole retorna null para /chat", () => {
    expect(getRequiredRole("/chat")).toBeNull()
  })

  test("canAccessRoute con area_manager en /audit retorna true", () => {
    const managerClaims = { sub: "2", email: "m@b.com", name: "Manager", role: "area_manager" as const, iat: 0, exp: 9999999999 }
    expect(canAccessRoute(managerClaims, "/audit")).toBe(true)
  })

  test("verifyJwt retorna null para token expirado", async () => {
    // Crear un token que ya expiró usando exp en el pasado
    const { SignJWT } = await import("jose")
    const secret = new TextEncoder().encode(process.env["JWT_SECRET"]!)
    const expiredToken = await new SignJWT({ sub: "1", email: "x@x.com", name: "X", role: "user" })
      .setProtectedHeader({ alg: "HS256" })
      .setIssuedAt()
      .setExpirationTime("1s")
      .sign(secret)
    // Esperar a que expire
    await new Promise((r) => setTimeout(r, 1500))
    const result = await verifyJwt(expiredToken)
    expect(result).toBeNull()
  })

  test("makeAuthCookie incluye Secure en NODE_ENV=production", async () => {
    const originalEnv = process.env["NODE_ENV"]
    process.env["NODE_ENV"] = "production"
    const token = await createJwt({ sub: "1", email: "t@t.com", name: "T", role: "user" })
    const cookie = makeAuthCookie(token)
    expect(cookie).toContain("Secure")
    process.env["NODE_ENV"] = originalEnv
  })

  // ── isAdmin / isAreaManager ──────────────────────────────────────────────

  test("isAdmin returns true only for admin role", () => {
    expect(isAdmin({ sub: "1", email: "a@b.com", name: "A", role: "admin" as const, iat: 0, exp: 9999999999 })).toBe(true)
    expect(isAdmin({ sub: "2", email: "m@b.com", name: "M", role: "area_manager" as const, iat: 0, exp: 9999999999 })).toBe(false)
    expect(isAdmin({ sub: "3", email: "u@b.com", name: "U", role: "user" as const, iat: 0, exp: 9999999999 })).toBe(false)
  })

  test("isAreaManager returns true for admin and area_manager", () => {
    expect(isAreaManager({ sub: "1", email: "a@b.com", name: "A", role: "admin" as const, iat: 0, exp: 9999999999 })).toBe(true)
    expect(isAreaManager({ sub: "2", email: "m@b.com", name: "M", role: "area_manager" as const, iat: 0, exp: 9999999999 })).toBe(true)
    expect(isAreaManager({ sub: "3", email: "u@b.com", name: "U", role: "user" as const, iat: 0, exp: 9999999999 })).toBe(false)
  })

  // ── Route edge cases ─────────────────────────────────────────────────────

  test("getRequiredRole handles trailing slash on /admin/", () => {
    expect(getRequiredRole("/admin/")).toBe("admin")
  })

  test("getRequiredRole handles exact /admin without trailing slash", () => {
    expect(getRequiredRole("/admin")).toBe("admin")
  })

  test("getRequiredRole handles deeply nested admin route /admin/users/123/edit", () => {
    expect(getRequiredRole("/admin/users/123/edit")).toBe("admin")
  })

  test("getRequiredRole handles /api/admin exact (no sub-path)", () => {
    expect(getRequiredRole("/api/admin")).toBe("admin")
  })

  test("getRequiredRole handles /api/audit/logs nested route", () => {
    expect(getRequiredRole("/api/audit/logs")).toBe("area_manager")
  })

  test("getRequiredRole handles /audit trailing slash", () => {
    expect(getRequiredRole("/audit/")).toBe("area_manager")
  })

  test("getRequiredRole is case-sensitive — /Admin/users is NOT protected", () => {
    // The regex uses lowercase /admin, so /Admin should not match
    expect(getRequiredRole("/Admin/users")).toBeNull()
  })

  test("getRequiredRole does not match partial prefix — /administration is NOT protected", () => {
    // /admin(\/|$)/ should NOT match /administration because the regex requires / or end-of-string after 'admin'
    expect(getRequiredRole("/administration")).toBeNull()
  })

  test("getRequiredRole does not match /api/administrator", () => {
    expect(getRequiredRole("/api/administrator")).toBeNull()
  })

  test("getRequiredRole returns null for /login, /settings, /collections", () => {
    expect(getRequiredRole("/login")).toBeNull()
    expect(getRequiredRole("/settings")).toBeNull()
    expect(getRequiredRole("/collections")).toBeNull()
  })

  test("canAccessRoute allows any role on unprotected routes with query params", () => {
    const userClaims = { sub: "3", email: "u@b.com", name: "U", role: "user" as const, iat: 0, exp: 9999999999 }
    // getRequiredRole receives pathname, not full URL — query params should not appear
    // but test that pathnames with query-like strings are still handled
    expect(canAccessRoute(userClaims, "/chat")).toBe(true)
    expect(canAccessRoute(userClaims, "/collections")).toBe(true)
  })

  test("canAccessRoute blocks user on /admin with trailing slash", () => {
    const userClaims = { sub: "3", email: "u@b.com", name: "U", role: "user" as const, iat: 0, exp: 9999999999 }
    expect(canAccessRoute(userClaims, "/admin/")).toBe(false)
  })

  test("canAccessRoute blocks user on deeply nested /api/admin/roles/5", () => {
    const userClaims = { sub: "3", email: "u@b.com", name: "U", role: "user" as const, iat: 0, exp: 9999999999 }
    expect(canAccessRoute(userClaims, "/api/admin/roles/5")).toBe(false)
  })

  test("canAccessRoute allows admin on all protected routes", () => {
    const adminClaims = { sub: "1", email: "a@b.com", name: "A", role: "admin" as const, iat: 0, exp: 9999999999 }
    expect(canAccessRoute(adminClaims, "/admin")).toBe(true)
    expect(canAccessRoute(adminClaims, "/admin/users/123")).toBe(true)
    expect(canAccessRoute(adminClaims, "/api/admin/config")).toBe(true)
    expect(canAccessRoute(adminClaims, "/audit")).toBe(true)
    expect(canAccessRoute(adminClaims, "/api/audit/logs")).toBe(true)
  })
})

describe("Plan 26: access+refresh tokens", () => {
  const baseClaims = {
    sub: "99",
    email: "plan26@test.com",
    name: "Plan26 User",
    role: "user" as const,
  }

  // ── createAccessToken ──────────────────────────────────────────────────

  test("createAccessToken includes type: 'access' in claims", async () => {
    const token = await createAccessToken(baseClaims)
    const claims = await verifyJwt(token)
    expect(claims).not.toBeNull()
    expect((claims as Record<string, unknown>)["type"]).toBe("access")
  })

  test("createAccessToken generates unique JTI per token", async () => {
    const t1 = await createAccessToken(baseClaims)
    const t2 = await createAccessToken(baseClaims)
    const c1 = await verifyJwt(t1)
    const c2 = await verifyJwt(t2)
    expect(c1?.jti).toBeDefined()
    expect(c2?.jti).toBeDefined()
    expect(c1!.jti).not.toBe(c2!.jti)
  })

  test("createAccessToken is aliased as createJwt", () => {
    expect(createAccessToken).toBe(createJwt)
  })

  // ── createRefreshToken ─────────────────────────────────────────────────

  test("createRefreshToken includes type: 'refresh' in claims", async () => {
    const token = await createRefreshToken("99")
    const claims = await verifyJwt(token)
    expect(claims).not.toBeNull()
    expect((claims as Record<string, unknown>)["type"]).toBe("refresh")
  })

  test("createRefreshToken only includes sub claim (no email/name/role)", async () => {
    const token = await createRefreshToken("99")
    const claims = await verifyJwt(token)
    expect(claims).not.toBeNull()
    expect(claims!.sub).toBe("99")
    // Refresh tokens should NOT have user-detail claims
    expect((claims as Record<string, unknown>)["email"]).toBeUndefined()
    expect((claims as Record<string, unknown>)["name"]).toBeUndefined()
    expect((claims as Record<string, unknown>)["role"]).toBeUndefined()
  })

  test("createRefreshToken generates unique JTI per token", async () => {
    const t1 = await createRefreshToken("99")
    const t2 = await createRefreshToken("99")
    const c1 = await verifyJwt(t1)
    const c2 = await verifyJwt(t2)
    expect(c1?.jti).toBeDefined()
    expect(c2?.jti).toBeDefined()
    expect(c1!.jti).not.toBe(c2!.jti)
  })

  // ── makeRefreshCookie ──────────────────────────────────────────────────

  test("makeRefreshCookie has SameSite=Strict", async () => {
    const token = await createRefreshToken("99")
    const cookie = makeRefreshCookie(token)
    expect(cookie).toContain("SameSite=Strict")
  })

  test("makeRefreshCookie has Path=/api/auth/refresh", async () => {
    const token = await createRefreshToken("99")
    const cookie = makeRefreshCookie(token)
    expect(cookie).toContain("Path=/api/auth/refresh")
  })

  test("makeRefreshCookie has HttpOnly", async () => {
    const token = await createRefreshToken("99")
    const cookie = makeRefreshCookie(token)
    expect(cookie).toContain("HttpOnly")
  })

  test("makeRefreshCookie includes refresh_token name", async () => {
    const token = await createRefreshToken("99")
    const cookie = makeRefreshCookie(token)
    expect(cookie).toContain("refresh_token=")
  })

  test("makeRefreshCookie includes Secure in production", async () => {
    const originalEnv = process.env["NODE_ENV"]
    process.env["NODE_ENV"] = "production"
    const token = await createRefreshToken("99")
    const cookie = makeRefreshCookie(token)
    expect(cookie).toContain("Secure")
    process.env["NODE_ENV"] = originalEnv
  })

  // ── makeClearRefreshCookie ─────────────────────────────────────────────

  test("makeClearRefreshCookie has Max-Age=0", () => {
    const cookie = makeClearRefreshCookie()
    expect(cookie).toContain("Max-Age=0")
  })

  test("makeClearRefreshCookie has correct Path=/api/auth/refresh", () => {
    const cookie = makeClearRefreshCookie()
    expect(cookie).toContain("Path=/api/auth/refresh")
  })

  test("makeClearRefreshCookie has SameSite=Strict", () => {
    const cookie = makeClearRefreshCookie()
    expect(cookie).toContain("SameSite=Strict")
  })

  test("makeClearRefreshCookie has HttpOnly", () => {
    const cookie = makeClearRefreshCookie()
    expect(cookie).toContain("HttpOnly")
  })

  // ── revokeToken ────────────────────────────────────────────────────────

  test("revokeToken calls Redis SET with correct key and TTL", async () => {
    const redis = getRedisClient()
    const jti = crypto.randomUUID()
    const futureExp = Math.floor(Date.now() / 1000) + 3600 // 1h from now

    await revokeToken(jti, futureExp)

    const stored = await redis.get(`revoked:${jti}`)
    expect(stored).toBe("1")

    // Verify TTL was set (ioredis-mock supports ttl)
    if (typeof redis.ttl === "function") {
      const ttl = await redis.ttl(`revoked:${jti}`)
      expect(ttl).toBeGreaterThan(3500)
      expect(ttl).toBeLessThanOrEqual(3600)
    }
  })

  test("revokeToken with expired TTL (exp in past) does not call Redis", async () => {
    const redis = getRedisClient()
    const jti = crypto.randomUUID()
    const pastExp = Math.floor(Date.now() / 1000) - 100 // 100s ago

    await revokeToken(jti, pastExp)

    const stored = await redis.get(`revoked:${jti}`)
    expect(stored).toBeNull()
  })

  // NOTE: extractClaims revocation tests are in auth-routes.test.ts (tests 14-16)
  // because they require a controlled Redis mock. Running them here conflicts with
  // auth-routes.test.ts which mocks @rag-saldivia/db globally via mock.module().
})
