/**
 * GET /api/auth/callback/[provider] — SSO callback from IdP.
 *
 * Verifies state, exchanges code for tokens via arctic, extracts user info,
 * finds or provisions user, creates our JWT, sets cookies, redirects to /chat.
 */

import { NextResponse } from "next/server"
import { loadProvider, verifyStateToken, extractUserInfo } from "@/lib/auth/sso"
import { createAccessToken, createRefreshToken, makeAuthCookie, makeRefreshCookie } from "@/lib/auth/jwt"
import { findUserBySso, linkSsoToUser, createSsoUser, getUserByEmail, getUserById } from "@rag-saldivia/db"
 
const ssoLog = {
  warn: (type: string, data: Record<string, unknown>) => console.warn(`[SSO] ${type}`, data),
  error: (type: string, data: Record<string, unknown>) => console.error(`[SSO] ${type}`, data),
  info: (type: string, data: Record<string, unknown>) => console.error(`[SSO] ${type}`, data), // console.error is allowed
}
 
import type { SsoProviderType } from "@rag-saldivia/shared"

function getCookieValue(request: Request, name: string): string | null {
  const cookieHeader = request.headers.get("cookie")
  if (!cookieHeader) return null
  const match = cookieHeader.match(new RegExp(`(?:^|;\\s*)${name}=([^;]+)`))
  return match?.[1] ? decodeURIComponent(match[1]) : null
}

function clearSsoCookies(response: NextResponse): void {
  response.headers.append("Set-Cookie", "sso_state=; Max-Age=0; Path=/; HttpOnly; SameSite=Lax")
  response.headers.append("Set-Cookie", "sso_verifier=; Max-Age=0; Path=/; HttpOnly; SameSite=Lax")
  response.headers.append("Set-Cookie", "sso_token=; Max-Age=0; Path=/; HttpOnly; SameSite=Lax")
}

function errorRedirect(message: string, code: string): NextResponse {
  const response = NextResponse.redirect(new URL(`/login?sso_error=${code}`, process.env["APP_URL"] ?? "http://localhost:3000"))
  clearSsoCookies(response)
  return response
}

export async function GET(
  request: Request,
  { params }: { params: Promise<{ provider: string }> },
) {
  const { provider: providerType } = await params
  const url = new URL(request.url)
  const code = url.searchParams.get("code")
  const state = url.searchParams.get("state")

  // 1. Validate params
  if (!code || !state) {
    return errorRedirect("Parámetros faltantes del proveedor SSO", "missing_params")
  }

  // 2. Verify state against cookie (CSRF protection)
  const savedState = getCookieValue(request, "sso_state")
  if (!savedState || savedState !== state) {
    ssoLog.warn("sso.state_mismatch", { provider: providerType })
    return errorRedirect("Estado de sesión inválido", "invalid_state")
  }

  // 3. Verify signed state token
  const stateToken = getCookieValue(request, "sso_token")
  if (!stateToken) {
    return errorRedirect("Token de estado faltante", "invalid_state")
  }
  const tokenPayload = await verifyStateToken(stateToken)
  if (!tokenPayload || tokenPayload.provider !== providerType) {
    return errorRedirect("Token de estado expirado o inválido", "expired_state")
  }

  // 4. Load provider config
  const loaded = await loadProvider(providerType as SsoProviderType)
  if (!loaded) {
    return errorRedirect("Proveedor no configurado", "provider_error")
  }

  // 5. Exchange code for tokens
  const codeVerifier = getCookieValue(request, "sso_verifier") ?? ""
  let tokens: { accessToken: () => string }
  try {
    tokens = await loaded.arctic.validateAuthorizationCode(code, codeVerifier) as { accessToken: () => string }
  } catch (err) {
    ssoLog.error("sso.token_exchange_failed", { provider: providerType, error: String(err) })
    return errorRedirect("Error al intercambiar código con el proveedor", "provider_error")
  }

  // 6. Extract user info from IdP
  const accessToken = tokens.accessToken()
  const userInfo = await extractUserInfo(providerType as SsoProviderType, accessToken)
  if (!userInfo || !userInfo.email) {
    ssoLog.error("sso.user_info_failed", { provider: providerType })
    return errorRedirect("No se pudo obtener información del usuario", "provider_error")
  }

  // 7. Find or provision user
  let user = await findUserBySso(providerType, userInfo.sub)

  if (!user) {
    // Try account linking by email
    const existingUser = await getUserByEmail(userInfo.email)
    if (existingUser) {
      // Link SSO to existing account (one-time)
      if (existingUser.ssoProvider && existingUser.ssoProvider !== providerType) {
        // Already linked to a different provider
        return errorRedirect("Esta cuenta ya está vinculada a otro proveedor SSO", "already_linked")
      }
      await linkSsoToUser(existingUser.id, providerType, userInfo.sub)
      user = Object.assign({}, existingUser, { ssoProvider: providerType, ssoSubject: userInfo.sub })
      ssoLog.info("sso.account_linked", { userId: existingUser.id, provider: providerType })
    } else if (loaded.config.autoProvision) {
      // Auto-provision new user
      const created = await createSsoUser({
        email: userInfo.email,
        name: userInfo.name,
        ssoProvider: providerType,
        ssoSubject: userInfo.sub,
        role: loaded.config.defaultRole as "admin" | "area_manager" | "user",
      })
      user = await getUserById(created.id)
      ssoLog.info("sso.user_provisioned", { userId: user!.id, provider: providerType, email: userInfo.email })
    } else {
      return errorRedirect("No se encontró una cuenta. Contactá al administrador.", "no_account")
    }
  }

  if (!user) {
    return errorRedirect("Error interno al procesar usuario", "provider_error")
  }

  // Check user is active
  if (!user.active) {
    return errorRedirect("Tu cuenta está desactivada", "inactive")
  }

  // 8. Create our JWT (same pipeline as email/password login)
  const accessJwt = await createAccessToken({
    sub: String(user.id),
    email: user.email,
    name: user.name,
    role: user.role as "admin" | "area_manager" | "user",
  })
  const refreshJwt = await createRefreshToken(String(user.id))

  ssoLog.info("sso.login_success", { userId: user.id, provider: providerType })

  // 9. Set cookies and redirect
  const baseUrl = process.env["APP_URL"] ?? "http://localhost:3000"
  const response = NextResponse.redirect(new URL("/chat", baseUrl))
  response.headers.append("Set-Cookie", makeAuthCookie(accessJwt))
  response.headers.append("Set-Cookie", makeRefreshCookie(refreshJwt))
  clearSsoCookies(response)

  return response
}
