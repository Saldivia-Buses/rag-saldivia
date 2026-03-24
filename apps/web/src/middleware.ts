/**
 * Next.js Middleware — Auth + RBAC en el edge
 *
 * Verifica JWT en cada request. Rutas protegidas redirigen a /login si no hay sesión.
 * Rutas de admin verifican que el usuario tenga el rol requerido.
 */

import { NextResponse, type NextRequest } from "next/server"
import { jwtVerify } from "jose"
import type { JwtClaims } from "@rag-saldivia/shared"

const PUBLIC_ROUTES = [
  "/login",
  "/api/auth/login",
  "/api/auth/refresh",
  "/api/log",           // Frontend puede logear sin auth
  "/_next",
  "/favicon.ico",
]

const ADMIN_ROUTES = [/^\/admin(\/|$)/, /^\/api\/admin(\/|$)/]
const AREA_MANAGER_ROUTES = [/^\/audit(\/|$)/, /^\/api\/audit(\/|$)/]

function isPublic(pathname: string): boolean {
  return PUBLIC_ROUTES.some((p) => pathname.startsWith(p))
}

function getSecret(): Uint8Array {
  const secret = process.env["JWT_SECRET"] ?? ""
  return new TextEncoder().encode(secret)
}

async function verifyClaims(token: string): Promise<JwtClaims | null> {
  try {
    const { payload } = await jwtVerify(token, getSecret())
    return payload as unknown as JwtClaims
  } catch {
    return null
  }
}

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Rutas públicas siempre pasan
  if (isPublic(pathname)) {
    return NextResponse.next()
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

  // Verificar roles para rutas protegidas
  const isAdminRoute = ADMIN_ROUTES.some((p) => p.test(pathname))
  const isAreaManagerRoute = AREA_MANAGER_ROUTES.some((p) => p.test(pathname))

  if (isAdminRoute && claims.role !== "admin") {
    if (pathname.startsWith("/api/")) {
      return NextResponse.json(
        { ok: false, error: "Acceso denegado — se requiere rol admin" },
        { status: 403 }
      )
    }
    return NextResponse.redirect(new URL("/", request.url))
  }

  if (isAreaManagerRoute && claims.role === "user") {
    if (pathname.startsWith("/api/")) {
      return NextResponse.json(
        { ok: false, error: "Acceso denegado" },
        { status: 403 }
      )
    }
    return NextResponse.redirect(new URL("/", request.url))
  }

  // Pasar claims al request como headers para Server Components
  const headers = new Headers(request.headers)
  headers.set("x-user-id", String(claims.sub))
  headers.set("x-user-email", claims.email)
  headers.set("x-user-name", claims.name)
  headers.set("x-user-role", claims.role)

  return NextResponse.next({ request: { headers } })
}

export const config = {
  matcher: [
    // Aplica a todas las rutas excepto archivos estáticos de Next.js
    "/((?!_next/static|_next/image|favicon.ico).*)",
  ],
}
