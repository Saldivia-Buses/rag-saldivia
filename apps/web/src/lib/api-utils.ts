/**
 * api-utils.ts — Shared helpers for API route handlers.
 *
 * Eliminates repeated patterns across all /api/ routes:
 * - Consistent JSON response format: { ok, data?, error? }
 * - Auth check with automatic 401 response
 * - Admin-only guard with automatic 403 response
 * - Error wrapping with logging
 *
 * Usage:
 *   const claims = await requireAuth(request)
 *   if (claims instanceof NextResponse) return claims
 *
 *   return apiOk({ users: [...] })
 *
 * Used by: all API route handlers in app/api/
 * Depends on: lib/auth/jwt.ts (extractClaims)
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import type { JwtClaims } from "@rag-saldivia/shared"
import { log } from "@rag-saldivia/logger/backend"

// ── Response helpers ──────────────────────────────────────────────────────

/** Success response with optional data payload */
export function apiOk<T>(data?: T, status = 200) {
  return NextResponse.json({ ok: true, ...(data !== undefined && { data }) }, { status })
}

/** Error response with message and optional details */
export function apiError(error: string, status = 400, details?: unknown) {
  return NextResponse.json(
    { ok: false, error, ...(details !== undefined && { details }) },
    { status }
  )
}

/** 500 error with logging — use in catch blocks */
export function apiServerError(error: unknown, endpoint: string, userId?: number) {
  log.error("system.error", { error: String(error), endpoint }, userId ? { userId } : undefined)
  return NextResponse.json({ ok: false, error: "Error interno del servidor" }, { status: 500 })
}

// ── Auth guards ───────────────────────────────────────────────────────────

/**
 * Extract and validate JWT claims from request.
 * Returns claims on success, or a 401 NextResponse on failure.
 *
 * Pattern:
 *   const claims = await requireAuth(request)
 *   if (claims instanceof NextResponse) return claims
 *   // claims is now JwtClaims, guaranteed
 */
export async function requireAuth(request: Request): Promise<JwtClaims | NextResponse> {
  const claims = await extractClaims(request)
  if (!claims) return apiError("No autenticado", 401)
  return claims
}

/**
 * Require admin role. Returns claims or 401/403 response.
 *
 * Pattern:
 *   const claims = await requireAdmin(request)
 *   if (claims instanceof NextResponse) return claims
 */
export async function requireAdmin(request: Request): Promise<JwtClaims | NextResponse> {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims
  if (claims.role !== "admin") return apiError("Solo administradores", 403)
  return claims
}
