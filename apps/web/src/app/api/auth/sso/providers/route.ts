/**
 * GET /api/auth/sso/providers — List active SSO providers (public).
 *
 * Returns only public fields (id, name, type) for the login page.
 * This route is PUBLIC — no auth required.
 */

import { NextResponse } from "next/server"
import { listPublicSsoProviders } from "@rag-saldivia/db"

export async function GET() {
  try {
    const providers = await listPublicSsoProviders()
    return NextResponse.json({ ok: true, data: providers })
  } catch {
    return NextResponse.json({ ok: true, data: [] })
  }
}
