/**
 * Next.js Middleware — Auth + RBAC en el edge
 *
 * Verifica JWT en cada request. Rutas protegidas redirigen a /login si no hay sesión.
 * Rutas de admin verifican que el usuario tenga el rol requerido.
 */

import { NextResponse, type NextRequest } from "next/server"
import { jwtVerify } from "jose"
import type { JwtClaims } from "@rag-saldivia/shared"
import { canAccessRoute } from "@/lib/auth/rbac"

const PUBLIC_ROUTES = [
  "/login",
  "/api/auth/login",
  "/api/auth/refresh",
  "/api/health",        // Health check público para CLI y monitoreo
  "/api/log",           // Frontend puede logear sin auth
  "/api/auth/sso",      // SSO authorization redirect + providers listing
  "/api/auth/callback", // SSO callback from IdP (user has no JWT yet)
  "/_next",
  "/favicon.ico",
]

function isPublic(pathname: string): boolean {
  return PUBLIC_ROUTES.some((p) => pathname.startsWith(p))
}

function getSecret(): Uint8Array {
  const secret = process.env["JWT_SECRET"] ?? ""
  return new TextEncoder().encode(secret)
}

function isSystemApiKey(token: string): boolean {
  const key = process.env["SYSTEM_API_KEY"]
  return !!key && token === key
}

async function verifyClaims(token: string): Promise<JwtClaims | null> {
  try {
    const { payload } = await jwtVerify(token, getSecret())
    return payload as unknown as JwtClaims
  } catch {
    return null
  }
}

/**
 * Punto de entrada del middleware de Next.js (auth + RBAC en el edge). Genera `x-request-id`
 * para correlación de logs. Propaga `x-user-jti` para que los route handlers puedan comprobar
 * revocación de JWT con Redis. Corre en Edge — sin ioredis, SQLite ni `fs`.
 */
export async function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Generar requestId de correlación para todos los requests
  const requestId = crypto.randomUUID()

  // Rutas públicas siempre pasan (con requestId propagado)
  if (isPublic(pathname)) {
    const headers = new Headers(request.headers)
    headers.set("x-request-id", requestId)
    return NextResponse.next({ request: { headers } })
  }

  // Extraer JWT desde cookie o Authorization header
  const cookieToken = request.cookies.get("auth_token")?.value
  const bearerToken = request.headers.get("authorization")?.replace("Bearer ", "")
  const token = cookieToken ?? bearerToken

  if (!token) {
    // API routes → 401
    if (pathname.startsWith("/api/")) {
      return NextResponse.json(
        { ok: false, error: "No autenticado" },
        { status: 401 }
      )
    }
    // Páginas → redirect a login
    const url = request.nextUrl.clone()
    url.pathname = "/login"
    url.searchParams.set("from", pathname)
    return NextResponse.redirect(url)
  }

  // SYSTEM_API_KEY: acceso de servicio a servicio con rol admin
  if (isSystemApiKey(token)) {
    const headers = new Headers(request.headers)
    headers.set("x-user-id", "0")
    headers.set("x-user-email", "system@internal")
    headers.set("x-user-name", "System")
    headers.set("x-user-role", "admin")
    headers.set("x-request-id", requestId)
    return NextResponse.next({ request: { headers } })
  }

  const claims = await verifyClaims(token)

  if (!claims) {
    if (pathname.startsWith("/api/")) {
      return NextResponse.json(
        { ok: false, error: "Token inválido o expirado" },
        { status: 401 }
      )
    }
    const url = request.nextUrl.clone()
    url.pathname = "/login"
    url.searchParams.set("from", pathname)
    return NextResponse.redirect(url)
  }

  // Verificar roles para rutas protegidas (single source of truth: lib/auth/rbac.ts)
  if (!canAccessRoute(claims, pathname)) {
    if (pathname.startsWith("/api/")) {
      return NextResponse.json(
        { ok: false, error: "Acceso denegado" },
        { status: 403 }
      )
    }
    return NextResponse.redirect(new URL("/", request.url))
  }

  // Pasar claims y requestId al request como headers para Server Components
  const headers = new Headers(request.headers)
  headers.set("x-user-id", String(claims.sub))
  headers.set("x-user-email", claims.email)
  headers.set("x-user-name", claims.name)
  headers.set("x-user-role", claims.role)
  headers.set("x-request-id", requestId)
  // Propagar jti para que extractClaims() verifique revocación en route handlers
  if (claims.jti) headers.set("x-user-jti", claims.jti)

  return NextResponse.next({ request: { headers } })
}

export const config = {
  matcher: [
    // Aplica a todas las rutas excepto archivos estáticos de Next.js
    "/((?!_next/static|_next/image|favicon.ico).*)",
  ],
}
