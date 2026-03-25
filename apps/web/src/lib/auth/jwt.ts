/**
 * JWT utilities — createJwt, verifyJwt, extractClaims
 * Usa `jose` (misma lib que el frontend SvelteKit actual).
 */

import { SignJWT, jwtVerify } from "jose"
import type { JwtClaims } from "@rag-saldivia/shared"

function getSecret(): Uint8Array {
  const secret = process.env["JWT_SECRET"]
  if (!secret) throw new Error("JWT_SECRET no está configurado")
  return new TextEncoder().encode(secret)
}

function getExpiry(): string {
  return process.env["JWT_EXPIRY"] ?? "24h"
}

export async function createJwt(claims: Omit<JwtClaims, "iat" | "exp">): Promise<string> {
  const expiry = getExpiry()
  return new SignJWT({ ...claims })
    .setProtectedHeader({ alg: "HS256" })
    .setIssuedAt()
    .setExpirationTime(expiry)
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

export async function extractClaims(request: Request): Promise<JwtClaims | null> {
  // Si el middleware ya autenticó la request (JWT o SYSTEM_API_KEY),
  // los claims están en los headers x-user-* — usarlos directamente.
  const userId = request.headers.get("x-user-id")
  const userRole = request.headers.get("x-user-role")
  if (userId && userRole) {
    return {
      sub: userId,
      email: request.headers.get("x-user-email") ?? "",
      name: request.headers.get("x-user-name") ?? "",
      role: userRole as JwtClaims["role"],
      iat: 0,
      exp: 0,
    }
  }

  // Fallback: verificar JWT desde cookie o Authorization header
  const cookieHeader = request.headers.get("cookie")
  if (cookieHeader) {
    const match = cookieHeader.match(/(?:^|;\s*)auth_token=([^;]+)/)
    if (match?.[1]) {
      return verifyJwt(decodeURIComponent(match[1]))
    }
  }

  const authHeader = request.headers.get("authorization")
  if (authHeader?.startsWith("Bearer ")) {
    return verifyJwt(authHeader.slice(7))
  }

  return null
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
