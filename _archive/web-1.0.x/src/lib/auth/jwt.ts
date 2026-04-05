/**
 * JWT utilities — createJwt, verifyJwt, extractClaims
 * Usa `jose` (misma lib que el frontend SvelteKit actual).
 *
 * F8.25 — JWT revocation list via Redis:
 * - createJwt agrega .setJti(crypto.randomUUID()) para identificar cada token
 * - logout escribe `SET revoked:{jti} 1 EX {ttl}` en Redis
 * - extractClaims verifica blacklist antes de retornar claims
 *
 * NOTA: la verificación de revocación NO está en `proxy.ts` (middleware Edge) porque
 * ioredis requiere Node.js APIs.
 * extractClaims() es llamado desde route handlers (Node.js) — ahí funciona.
 */

import { SignJWT, jwtVerify } from "jose"
import type { JwtClaims } from "@rag-saldivia/shared"
import { getRedisClient } from "@rag-saldivia/db"
import {
  ACCESS_TOKEN_EXPIRY,
  ACCESS_TOKEN_MAX_AGE_S,
  REFRESH_TOKEN_EXPIRY,
  REFRESH_TOKEN_MAX_AGE_S,
} from "@rag-saldivia/config"

function getSecret(): Uint8Array {
  const secret = process.env["JWT_SECRET"]
  if (!secret) throw new Error("JWT_SECRET no está configurado")
  return new TextEncoder().encode(secret)
}

/**
 * Crea un access token (15m) con claims completos.
 * El `jti` es requerido para revocación en Redis.
 * El `jti` se propaga en el header `x-user-jti` desde el middleware en `proxy.ts`.
 */
export async function createAccessToken(claims: Omit<JwtClaims, "iat" | "exp" | "jti">): Promise<string> {
  return new SignJWT({ ...claims, type: "access" })
    .setProtectedHeader({ alg: "HS256" })
    .setIssuedAt()
    .setExpirationTime(ACCESS_TOKEN_EXPIRY)
    .setJti(crypto.randomUUID())
    .sign(getSecret())
}

/**
 * Crea un refresh token (7d) con claims mínimos (solo sub).
 * Se usa para obtener un nuevo access token sin re-autenticar.
 */
export async function createRefreshToken(sub: string): Promise<string> {
  return new SignJWT({ sub, type: "refresh" })
    .setProtectedHeader({ alg: "HS256" })
    .setIssuedAt()
    .setExpirationTime(REFRESH_TOKEN_EXPIRY)
    .setJti(crypto.randomUUID())
    .sign(getSecret())
}

/** Backward-compatible alias — creates an access token. */
export const createJwt = createAccessToken

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

/**
 * Verifica el JWT y comprueba la blacklist en Redis. Corre en runtime Node.js (route handlers).
 * No corre en `proxy.ts` (Edge) porque ioredis no es compatible con Edge. La revocación se
 * resuelve aquí usando el `jti` que el proxy añade en `x-user-jti`.
 */
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
  return [
    `auth_token=${encodeURIComponent(token)}`,
    `Max-Age=${ACCESS_TOKEN_MAX_AGE_S}`,
    "Path=/",
    "HttpOnly",
    "SameSite=Lax",
    ...(isProduction ? ["Secure"] : []),
  ].join("; ")
}

export function makeRefreshCookie(token: string): string {
  const isProduction = process.env["NODE_ENV"] === "production"
  return [
    `refresh_token=${encodeURIComponent(token)}`,
    `Max-Age=${REFRESH_TOKEN_MAX_AGE_S}`,
    "Path=/api/auth/refresh",
    "HttpOnly",
    "SameSite=Strict",
    ...(isProduction ? ["Secure"] : []),
  ].join("; ")
}

export function makeClearAuthCookie(): string {
  return "auth_token=; Max-Age=0; Path=/; HttpOnly; SameSite=Lax"
}

export function makeClearRefreshCookie(): string {
  return "refresh_token=; Max-Age=0; Path=/api/auth/refresh; HttpOnly; SameSite=Strict"
}

/** Revoke a token's JTI in Redis with TTL = remaining lifetime. */
export async function revokeToken(jti: string, expSeconds: number): Promise<void> {
  const ttl = expSeconds - Math.floor(Date.now() / 1000)
  if (ttl > 0) {
    await getRedisClient().set(`revoked:${jti}`, "1", "EX", ttl).catch(() => {})
  }
}
