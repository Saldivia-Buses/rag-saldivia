/**
 * GET /api/auth/sso/[provider] — Initiate SSO flow.
 *
 * For OIDC: loads config from DB, generates authorization URL via arctic,
 * stores state + codeVerifier in HttpOnly cookies, redirects to IdP.
 * For SAML: generates AuthnRequest, redirects to IdP entry point.
 */

import { NextResponse } from "next/server"
import { generateCodeVerifier, generateState } from "arctic"
import { loadProvider, loadSamlProvider, createStateToken, createSamlAuthorizeUrl } from "@/lib/auth/sso"
import { SSO_STATE_TTL_S } from "@rag-saldivia/config"
import type { SsoProviderType } from "@rag-saldivia/shared"

const VALID_TYPES = new Set(["google", "microsoft", "github", "oidc_generic", "saml"])

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ provider: string }> },
) {
  const { provider: providerType } = await params

  if (!VALID_TYPES.has(providerType)) {
    return NextResponse.json({ ok: false, error: "Provider no soportado" }, { status: 400 })
  }

  const isProduction = process.env["NODE_ENV"] === "production"
  // OIDC uses Lax (GET callback from same-site redirect). SAML uses None (POST from cross-origin IdP).
  const oidcCookieOpts = `Path=/; HttpOnly; SameSite=Lax; Max-Age=${SSO_STATE_TTL_S}${isProduction ? "; Secure" : ""}`
  const samlCookieOpts = `Path=/; HttpOnly; SameSite=None; Secure; Max-Age=${SSO_STATE_TTL_S}`

  // ── SAML flow ────────────────────────────────────────────────────────
  if (providerType === "saml") {
    const samlProvider = await loadSamlProvider()
    if (!samlProvider) {
      return NextResponse.json({ ok: false, error: "SAML provider no configurado" }, { status: 404 })
    }

    const state = generateState()
    const stateToken = await createStateToken(providerType, state)
    const url = await createSamlAuthorizeUrl(samlProvider.saml, state)

    const response = NextResponse.redirect(url)
    response.headers.append("Set-Cookie", `sso_state=${encodeURIComponent(state)}; ${samlCookieOpts}`)
    response.headers.append("Set-Cookie", `sso_token=${encodeURIComponent(stateToken)}; ${samlCookieOpts}`)
    return response
  }

  // ── OIDC flow ────────────────────────────────────────────────────────
  const loaded = await loadProvider(providerType as SsoProviderType)
  if (!loaded) {
    return NextResponse.json({ ok: false, error: "Provider no configurado" }, { status: 404 })
  }

  const state = generateState()
  const codeVerifier = generateCodeVerifier()
  const stateToken = await createStateToken(providerType, state)

  const scopes = loaded.config.scopes.split(" ").filter(Boolean)
  const url = loaded.arctic.createAuthorizationURL(state, codeVerifier, scopes)

  const response = NextResponse.redirect(url.toString())
  response.headers.append("Set-Cookie", `sso_state=${encodeURIComponent(state)}; ${oidcCookieOpts}`)
  response.headers.append("Set-Cookie", `sso_verifier=${encodeURIComponent(codeVerifier)}; ${oidcCookieOpts}`)
  response.headers.append("Set-Cookie", `sso_token=${encodeURIComponent(stateToken)}; ${oidcCookieOpts}`)

  return response
}
