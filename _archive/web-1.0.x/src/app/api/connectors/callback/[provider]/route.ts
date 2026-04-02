/**
 * GET /api/connectors/callback/[provider] — OAuth callback for connector setup.
 *
 * After the user authorizes access to Google Drive / SharePoint, the IdP
 * redirects here with a code. We exchange it for tokens, encrypt, and store.
 */

import { NextResponse } from "next/server"
import { SignJWT, jwtVerify } from "jose"
import { createExternalSource } from "@rag-saldivia/db"
import { extractClaims } from "@/lib/auth/jwt"

function getStateSecret(): Uint8Array {
  const secret = process.env["JWT_SECRET"]
  if (!secret) throw new Error("JWT_SECRET not configured")
  return new TextEncoder().encode(secret)
}

/** Create HMAC-signed state for OAuth CSRF protection. */
export async function generateConnectorOAuthState(data: {
  userId: number
  provider: string
  name: string
  collectionDest: string
  schedule: string
}): Promise<string> {
  return new SignJWT(data as unknown as Record<string, unknown>)
    .setProtectedHeader({ alg: "HS256" })
    .setIssuedAt()
    .setExpirationTime("10m")
    .sign(getStateSecret())
}

async function verifyConnectorState(token: string) {
  try {
    const { payload } = await jwtVerify(token, getStateSecret())
    return payload as unknown as {
      userId: number; provider: string; name: string; collectionDest: string; schedule: string
    }
  } catch { return null }
}

// Token exchange URLs
const TOKEN_URLS: Record<string, string> = {
  google_drive: "https://oauth2.googleapis.com/token",
  sharepoint: "https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token",
}

export async function GET(
  request: Request,
  { params }: { params: Promise<{ provider: string }> },
) {
  const { provider: _provider } = await params
  const url = new URL(request.url)
  const code = url.searchParams.get("code")
  const state = url.searchParams.get("state")
  const baseUrl = process.env["APP_URL"] ?? "http://localhost:3000"
  const errorRedirect = (msg: string) =>
    NextResponse.redirect(new URL(`/admin/connectors?error=${encodeURIComponent(msg)}`, baseUrl))

  if (!code || !state) return errorRedirect("Parámetros faltantes")

  // Verify HMAC state
  const stateData = await verifyConnectorState(state)
  if (!stateData) return errorRedirect("Estado expirado o inválido")

  // Verify the user is still authenticated
  const claims = await extractClaims(request)
  if (!claims || Number(claims.sub) !== stateData.userId) {
    return errorRedirect("Sesión inválida")
  }

  // Exchange code for tokens
  let tokenUrl = TOKEN_URLS[stateData.provider]
  if (!tokenUrl) return errorRedirect("Provider no soportado para OAuth")

  const clientId = stateData.provider === "google_drive"
    ? (process.env["GOOGLE_CLIENT_ID"] ?? "")
    : (process.env["AZURE_CLIENT_ID"] ?? "")
  const clientSecret = stateData.provider === "google_drive"
    ? (process.env["GOOGLE_CLIENT_SECRET"] ?? "")
    : (process.env["AZURE_CLIENT_SECRET"] ?? "")

  if (stateData.provider === "sharepoint") {
    const tenantId = process.env["AZURE_TENANT_ID"] ?? "common"
    tokenUrl = tokenUrl.replace("{tenant}", tenantId)
  }

  const callbackUrl = `${baseUrl}/api/connectors/callback/${stateData.provider}`

  try {
    const tokenRes = await fetch(tokenUrl, {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({
        client_id: clientId,
        client_secret: clientSecret,
        code,
        redirect_uri: callbackUrl,
        grant_type: "authorization_code",
      }),
    })

    if (!tokenRes.ok) {
      const text = await tokenRes.text()
      return errorRedirect(`Error del proveedor: ${text.slice(0, 100)}`)
    }

    const tokens = (await tokenRes.json()) as {
      access_token: string
      refresh_token?: string
    }

    // Store the source with encrypted credentials
    await createExternalSource({
      userId: stateData.userId,
      provider: stateData.provider as "google_drive" | "sharepoint" | "confluence" | "web_crawler",
      name: stateData.name,
      credentials: JSON.stringify({
        accessToken: tokens.access_token,
        refreshToken: tokens.refresh_token ?? "",
        clientId,
        clientSecret,
      }),
      collectionDest: stateData.collectionDest,
      schedule: stateData.schedule as "hourly" | "daily" | "weekly",
    })

    return NextResponse.redirect(new URL("/admin/connectors?success=true", baseUrl))
  } catch (err) {
    return errorRedirect(`Error: ${String(err).slice(0, 100)}`)
  }
}
