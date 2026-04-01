/**
 * sso.ts — SSO core logic using arctic for OAuth2/OIDC.
 *
 * Arctic provides pure functions that generate authorization URLs and parse
 * callbacks. After the callback, we use our existing createAccessToken() +
 * createRefreshToken() — the JWT pipeline is unchanged.
 *
 * Used by: /api/auth/sso/[provider]/route.ts, /api/auth/callback/[provider]/route.ts
 */

import { Google, MicrosoftEntraId, GitHub } from "arctic"
import { SignJWT, jwtVerify } from "jose"
import { SSO_STATE_TTL_S, SSO_CALLBACK_PATH } from "@rag-saldivia/config"
import { getSsoProviderByType, type DbSsoProvider } from "@rag-saldivia/db"
// decryptSecret is used by sso queries (re-exported from db), not directly here

type SsoProviderType = "google" | "microsoft" | "github" | "oidc_generic"

// ── Arctic provider factory ───────────────────────────────────────────────

function getBaseUrl(): string {
  return process.env["NEXT_PUBLIC_APP_URL"] ?? process.env["APP_URL"] ?? "http://localhost:3000"
}

function buildCallbackUrl(providerType: string): string {
  return `${getBaseUrl()}${SSO_CALLBACK_PATH}/${providerType}`
}

export type ArcticProvider = {
  createAuthorizationURL(state: string, codeVerifier: string, scopes: string[]): URL
  validateAuthorizationCode(code: string, codeVerifier: string): Promise<unknown>
  /** true if this provider doesn't use PKCE (e.g. GitHub) */
  noPkce?: boolean | undefined
}

export function createArcticProvider(
  type: SsoProviderType,
  clientId: string,
  clientSecret: string,
  tenantId?: string | null,
): ArcticProvider {
  const callbackUrl = buildCallbackUrl(type)
  switch (type) {
    case "google":
      return new Google(clientId, clientSecret, callbackUrl)
    case "microsoft":
      return new MicrosoftEntraId(tenantId ?? "common", clientId, clientSecret, callbackUrl)
    case "github": {
      const gh = new GitHub(clientId, clientSecret, callbackUrl)
      // GitHub doesn't use PKCE — wrap to match our interface
      return {
        createAuthorizationURL: (state: string, _codeVerifier: string, scopes: string[]) =>
          gh.createAuthorizationURL(state, scopes),
        validateAuthorizationCode: (code: string) =>
          gh.validateAuthorizationCode(code),
        noPkce: true,
      }
    }
    case "oidc_generic":
      return new Google(clientId, clientSecret, callbackUrl)
  }
}

// ── State token (CSRF protection) ─────────────────────────────────────────

function getStateSecret(): Uint8Array {
  const secret = process.env["JWT_SECRET"]
  if (!secret) throw new Error("JWT_SECRET not configured")
  return new TextEncoder().encode(secret)
}

/** Create a signed state token containing provider and optional redirect. */
export async function createStateToken(provider: string, redirectTo?: string | undefined): Promise<string> {
  const builder = new SignJWT({ provider, ...(redirectTo ? { redirectTo } : {}) })
    .setProtectedHeader({ alg: "HS256" })
    .setIssuedAt()
    .setExpirationTime(`${SSO_STATE_TTL_S}s`)
  return builder.sign(getStateSecret())
}

/** Verify a state token. Returns null if invalid or expired. */
export async function verifyStateToken(token: string): Promise<{ provider: string; redirectTo?: string | undefined } | null> {
  try {
    const { payload } = await jwtVerify(token, getStateSecret())
    const result: { provider: string; redirectTo?: string | undefined } = { provider: payload.provider as string }
    if (payload.redirectTo) result.redirectTo = payload.redirectTo as string
    return result
  } catch {
    return null
  }
}

// ── User info extraction ──────────────────────────────────────────────────

export type SsoUserInfo = {
  email: string
  name: string
  sub: string // external user ID from the IdP
}

/** Extract user info from OAuth tokens — provider-specific claim mapping. */
export async function extractUserInfo(
  type: SsoProviderType,
  accessToken: string,
): Promise<SsoUserInfo | null> {
  try {
    switch (type) {
      case "google": {
        const res = await fetch("https://openidconnect.googleapis.com/v1/userinfo", {
          headers: { Authorization: `Bearer ${accessToken}` },
        })
        if (!res.ok) return null
        const data = await res.json() as { sub: string; email: string; name: string }
        return { email: data.email, name: data.name, sub: data.sub }
      }
      case "microsoft": {
        const res = await fetch("https://graph.microsoft.com/v1.0/me", {
          headers: { Authorization: `Bearer ${accessToken}` },
        })
        if (!res.ok) return null
        const data = await res.json() as { id: string; mail?: string; userPrincipalName: string; displayName: string }
        return {
          email: data.mail ?? data.userPrincipalName,
          name: data.displayName,
          sub: data.id,
        }
      }
      case "github": {
        const res = await fetch("https://api.github.com/user", {
          headers: { Authorization: `Bearer ${accessToken}`, Accept: "application/json" },
        })
        if (!res.ok) return null
        const data = await res.json() as { id: number; login: string; name?: string; email?: string }
        // GitHub may not return email in profile — fetch from /user/emails
        let email = data.email
        if (!email) {
          const emailRes = await fetch("https://api.github.com/user/emails", {
            headers: { Authorization: `Bearer ${accessToken}`, Accept: "application/json" },
          })
          if (emailRes.ok) {
            const emails = await emailRes.json() as Array<{ email: string; primary: boolean }>
            email = emails.find((e) => e.primary)?.email ?? emails[0]?.email
          }
        }
        if (!email) return null
        return { email, name: data.name ?? data.login, sub: String(data.id) }
      }
      default:
        return null
    }
  } catch {
    return null
  }
}

// ── Provider config loader ────────────────────────────────────────────────

export type LoadedProvider = {
  config: Omit<DbSsoProvider, "clientSecretEncrypted"> & { clientSecret: string }
  arctic: ReturnType<typeof createArcticProvider>
}

/** Load provider config from DB + create arctic instance. */
export async function loadProvider(type: SsoProviderType): Promise<LoadedProvider | null> {
  const config = await getSsoProviderByType(type)
  if (!config || !config.clientSecret) return null

  const arctic = createArcticProvider(type, config.clientId, config.clientSecret, config.tenantId)
  return { config: config as LoadedProvider["config"], arctic }
}
