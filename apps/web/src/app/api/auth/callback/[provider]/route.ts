/**
 * GET /api/auth/callback/[provider] — OIDC callback from IdP.
 * POST /api/auth/callback/saml — SAML callback (IdPs POST the SAMLResponse).
 *
 * Verifies state, exchanges code for tokens, extracts user info,
 * finds or provisions user, creates our JWT, sets cookies, redirects to /chat.
 */

import { NextResponse } from "next/server"
import { timingSafeEqual } from "crypto"
import { loadProvider, loadSamlProvider, verifyStateToken, extractUserInfo, validateSamlResponse } from "@/lib/auth/sso"
import { createAccessToken, createRefreshToken, makeAuthCookie, makeRefreshCookie } from "@/lib/auth/jwt"
import { findUserBySso, createSsoUser, getUserByEmail, getUserById } from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"
import type { SsoProviderType } from "@rag-saldivia/shared"

function getCookieValue(request: Request, name: string): string | null {
  const cookieHeader = request.headers.get("cookie")
  if (!cookieHeader) return null
  const match = cookieHeader.match(new RegExp(`(?:^|;\\s*)${name}=([^;]+)`))
  return match?.[1] ? decodeURIComponent(match[1]) : null
}

function safeEqual(a: string, b: string): boolean {
  if (a.length !== b.length) return false
  return timingSafeEqual(Buffer.from(a), Buffer.from(b))
}

function clearSsoCookies(response: NextResponse): void {
  response.headers.append("Set-Cookie", "sso_state=; Max-Age=0; Path=/; HttpOnly; SameSite=Lax")
  response.headers.append("Set-Cookie", "sso_verifier=; Max-Age=0; Path=/; HttpOnly; SameSite=Lax")
  response.headers.append("Set-Cookie", "sso_token=; Max-Age=0; Path=/; HttpOnly; SameSite=Lax")
}

function errorRedirect(message: string, code: string): NextResponse {
  const response = NextResponse.redirect(new URL(`/login?sso_error=${encodeURIComponent(code)}`, process.env["APP_URL"] ?? "http://localhost:3000"))
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
  if (!savedState || !safeEqual(savedState, state)) {
    log.warn("auth.failed", { reason: "sso_state_mismatch", provider: providerType })
    return errorRedirect("Estado de sesión inválido", "invalid_state")
  }

  // 3. Verify signed state token
  const stateToken = getCookieValue(request, "sso_token")
  if (!stateToken) {
    return errorRedirect("Token de estado faltante", "invalid_state")
  }
  const tokenPayload = await verifyStateToken(stateToken)
  if (!tokenPayload || tokenPayload.provider !== providerType || tokenPayload.state !== state) {
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
    log.error("auth.failed", { reason: "sso_token_exchange", provider: providerType, error: String(err) })
    return errorRedirect("Error al intercambiar código con el proveedor", "provider_error")
  }

  // 6. Extract user info from IdP
  const accessToken = tokens.accessToken()
  const userInfo = await extractUserInfo(providerType as SsoProviderType, accessToken)
  if (!userInfo || !userInfo.email) {
    log.error("auth.failed", { reason: "sso_user_info", provider: providerType })
    return errorRedirect("No se pudo obtener información del usuario", "provider_error")
  }

  // 7. Find or provision user
  const user = await findOrProvisionUser(providerType, userInfo, loaded.config)
  if (!user) return user as unknown as NextResponse

  // 8. Issue session and redirect
  return issueSessionAndRedirect(user, providerType)
}

/** SAML callback — IdPs POST the SAMLResponse. */
export async function POST(request: Request) {
  const samlProvider = await loadSamlProvider()
  if (!samlProvider) {
    return errorRedirect("SAML provider no configurado", "provider_error")
  }

  // Parse form body (SAMLResponse + RelayState)
  const formData = await request.formData()
  const samlResponse = formData.get("SAMLResponse") as string | null
  const relayState = formData.get("RelayState") as string | null

  if (!samlResponse) {
    return errorRedirect("SAMLResponse faltante", "missing_params")
  }

  // Verify state
  const savedState = getCookieValue(request, "sso_state")
  if (!savedState || !relayState || !safeEqual(savedState, relayState)) {
    log.warn("auth.failed", { reason: "saml_state_mismatch" })
    return errorRedirect("Estado de sesión SAML inválido", "invalid_state")
  }

  const stateToken = getCookieValue(request, "sso_token")
  if (!stateToken) {
    return errorRedirect("Token de estado SAML faltante", "invalid_state")
  }
  const tokenPayload = await verifyStateToken(stateToken)
  if (!tokenPayload || tokenPayload.provider !== "saml" || tokenPayload.state !== relayState) {
    return errorRedirect("Token de estado SAML expirado", "expired_state")
  }

  // Validate SAML assertion
  const userInfo = await validateSamlResponse(samlProvider.saml, {
    SAMLResponse: samlResponse,
  })
  if (!userInfo || !userInfo.email) {
    log.error("auth.failed", { reason: "saml_assertion_invalid" })
    return errorRedirect("Aserción SAML inválida", "provider_error")
  }

  // Find or provision user (same logic as OIDC)
  const user = await findOrProvisionUser("saml", userInfo, samlProvider.config)
  if (!user) return user as unknown as NextResponse // errorRedirect already returned

  return issueSessionAndRedirect(user, "saml")
}

// ── Shared helpers ────────────────────────────────────────────────────────

async function findOrProvisionUser(
  providerType: string,
  userInfo: { email: string; name: string; sub: string },
  config: { autoProvision: boolean; defaultRole: string },
) {
  let user = await findUserBySso(providerType, userInfo.sub)

  if (!user) {
    // Check if a local account with this email exists — do NOT auto-link
    // (prevents account takeover if attacker registers victim's email on an IdP)
    const existingUser = await getUserByEmail(userInfo.email)
    if (existingUser) {
      // Account exists but is not SSO-linked — user must login with password
      // and link SSO from settings (or admin links it manually)
      return errorRedirect("Ya existe una cuenta con este email. Iniciá sesión con contraseña para vincular SSO.", "account_exists") as unknown as null
    } else if (config.autoProvision) {
      const created = await createSsoUser({
        email: userInfo.email,
        name: userInfo.name,
        ssoProvider: providerType,
        ssoSubject: userInfo.sub,
        role: config.defaultRole as "admin" | "area_manager" | "user",
      })
      user = await getUserById(created.id)
      log.info("user.created", { method: "sso_provision", provider: providerType, email: userInfo.email }, { userId: user!.id })
    } else {
      return errorRedirect("No se encontró una cuenta. Contactá al administrador.", "no_account") as unknown as null
    }
  }

  if (!user || !user.active) {
    return errorRedirect(user ? "Tu cuenta está desactivada" : "Error interno", user ? "inactive" : "provider_error") as unknown as null
  }

  return user
}

async function issueSessionAndRedirect(
  user: { id: number; email: string; name: string; role: string },
  providerType: string,
): Promise<NextResponse> {
  const accessJwt = await createAccessToken({
    sub: String(user.id),
    email: user.email,
    name: user.name,
    role: user.role as "admin" | "area_manager" | "user",
  })
  const refreshJwt = await createRefreshToken(String(user.id))

  log.info("auth.login", { method: "sso", provider: providerType }, { userId: user.id })

  const baseUrl = process.env["APP_URL"] ?? "http://localhost:3000"
  const response = NextResponse.redirect(new URL("/chat", baseUrl))
  response.headers.append("Set-Cookie", makeAuthCookie(accessJwt))
  response.headers.append("Set-Cookie", makeRefreshCookie(refreshJwt))
  clearSsoCookies(response)
  return response
}
