/**
 * GET /api/auth/sso/[provider] — Initiate SSO flow.
 *
 * Loads provider config from DB, generates authorization URL via arctic,
 * stores state + codeVerifier in HttpOnly cookies, redirects to IdP.
 */

import { NextResponse } from "next/server"
import { generateCodeVerifier, generateState } from "arctic"
import { loadProvider, createStateToken } from "@/lib/auth/sso"
import { SSO_STATE_TTL_S } from "@rag-saldivia/config"
import type { SsoProviderType } from "@rag-saldivia/shared"

const VALID_TYPES = new Set(["google", "microsoft", "github", "oidc_generic"])

export async function GET(
  request: Request,
  { params }: { params: Promise<{ provider: string }> },
) {
  const { provider: providerType } = await params

  if (!VALID_TYPES.has(providerType)) {
    return NextResponse.json({ ok: false, error: "Provider no soportado" }, { status: 400 })
  }

  const loaded = await loadProvider(providerType as SsoProviderType)
  if (!loaded) {
    return NextResponse.json({ ok: false, error: "Provider no configurado" }, { status: 404 })
  }

  const state = generateState()
  const codeVerifier = generateCodeVerifier()

  // Sign the state so we can verify it in the callback
  const stateToken = await createStateToken(providerType, request.url)

  // Generate authorization URL with scopes from config
  const scopes = loaded.config.scopes.split(" ").filter(Boolean)
  const url = loaded.arctic.createAuthorizationURL(state, codeVerifier, scopes)

  // Store state + codeVerifier in HttpOnly cookies
  const isProduction = process.env["NODE_ENV"] === "production"
  const cookieOpts = `Path=/; HttpOnly; SameSite=Lax; Max-Age=${SSO_STATE_TTL_S}${isProduction ? "; Secure" : ""}`

  const response = NextResponse.redirect(url.toString())
  response.headers.append("Set-Cookie", `sso_state=${encodeURIComponent(state)}; ${cookieOpts}`)
  response.headers.append("Set-Cookie", `sso_verifier=${encodeURIComponent(codeVerifier)}; ${cookieOpts}`)
  response.headers.append("Set-Cookie", `sso_token=${encodeURIComponent(stateToken)}; ${cookieOpts}`)

  return response
}
