/**
 * JWT utilities — createJwt, verifyJwt, extractClaims
 * Usa `jose` (misma lib que el frontend SvelteKit actual).
 *
 * F8.25 — JWT revocation list via Redis:
 * - createJwt agrega .setJti(crypto.randomUUID()) para identificar cada token
 * - logout escribe `SET revoked:{jti} 1 EX {ttl}` en Redis
 * - extractClaims verifica blacklist antes de retornar claims
 *
 * NOTA: la verificación de revocación NO está en middleware.ts porque
 * el middleware corre en Edge runtime y ioredis requiere Node.js APIs.
 * extractClaims() es llamado desde route handlers (Node.js) — ahí funciona.
 */

import { SignJWT, jwtVerify } from "jose"
import type { JwtClaims } from "@rag-saldivia/shared"
import { getRedisClient } from "@rag-saldivia/db"

function getSecret(): Uint8Array {
  const secret = process.env["JWT_SECRET"]
  if (!secret) throw new Error("JWT_SECRET no está configurado")
  return new TextEncoder().encode(secret)
}

function getExpiry(): string {
  return process.env["JWT_EXPIRY"] ?? "24h"
}

export async function createJwt(claims: Omit<JwtClaims, "iat" | "exp" | "jti">): Promise<string> {
  const expiry = getExpiry()
  return new SignJWT({ ...claims })
    .setProtectedHeader({ alg: "HS256" })
    .setIssuedAt()
    .setExpirationTime(expiry)
    .setJti(crypto.randomUUID())
    .sign(getSecret())
}

export async function verifyJwt(token: string): Promise<JwtClaims | null> {
  try {
    const { payload } = await jwtVerify(token, getSecret())
    return payload as unknown as JwtClaims
  } catch {
    return null
  }
}

async function isRevoked(jti: string): Promise<boolean> {
  try {
    const result = await getRedisClient().get(`revoked:${jti}`)
    return result !== null
  } catch {
    // Si Redis no está disponible, no bloquear — asumir no revocado
    return false
  }
}

export async function extractClaims(request: Request): Promise<JwtClaims | null> {
  // Si el middleware ya autenticó la request (JWT o SYSTEM_API_KEY),
  // los claims están en los headers x-user-* — usarlos directamente.
  const userId = request.headers.get("x-user-id")
  const userRole = request.headers.get("x-user-role")
  if (userId && userRole) {
    // Verificar revocación usando el jti que el middleware propagó
    const jti = request.headers.get("x-user-jti")
    if (jti && await isRevoked(jti)) return null

    return {
      sub: userId,
      email: request.headers.get("x-user-email") ?? "",
      name: request.headers.get("x-user-name") ?? "",
      role: userRole as JwtClaims["role"],
      jti: jti ?? undefined,
      iat: 0,
      exp: 0,
    }
  }

  // Fallback: verificar JWT desde cookie o Authorization header
  let token: string | null = null

  const cookieHeader = request.headers.get("cookie")
  if (cookieHeader) {
    const match = cookieHeader.match(/(?:^|;\s*)auth_token=([^;]+)/)
    if (match?.[1]) token = decodeURIComponent(match[1])
  }

  if (!token) {
    const authHeader = request.headers.get("authorization")
    if (authHeader?.startsWith("Bearer ")) token = authHeader.slice(7)
  }

  if (!token) return null

  const claims = await verifyJwt(token)
  if (!claims) return null

  // Verificar blacklist de revocación
  if (claims.jti && await isRevoked(claims.jti)) return null

  return claims
}

export function makeAuthCookie(token: string): string {
  const isProduction = process.env["NODE_ENV"] === "production"
  const maxAge = 60 * 60 * 24 // 24 horas en segundos
  return [
    `auth_token=${encodeURIComponent(token)}`,
    `Max-Age=${maxAge}`,
    "Path=/",
    "HttpOnly",
    "SameSite=Lax",
    ...(isProduction ? ["Secure"] : []),
  ].join("; ")
}

export function makeClearAuthCookie(): string {
  return "auth_token=; Max-Age=0; Path=/; HttpOnly; SameSite=Lax"
}
