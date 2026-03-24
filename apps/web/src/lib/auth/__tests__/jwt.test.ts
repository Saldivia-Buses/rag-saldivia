/**
 * Tests del flujo de autenticación JWT.
 * Corre con: bun test src/lib/auth/__tests__/jwt.test.ts
 */

import { describe, test, expect, beforeEach } from "bun:test"

// Setup env vars para tests
process.env["JWT_SECRET"] = "test-secret-key-at-least-16-chars"
process.env["JWT_EXPIRY"] = "1h"
process.env["NODE_ENV"] = "test"

// Import after env setup
const { createJwt, verifyJwt, extractClaims, makeAuthCookie, makeClearAuthCookie } =
  await import("../jwt.js")
const { hasRole, canAccessRoute, getRequiredRole } = await import("../rbac.js")

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
})
