/**
 * GET /api/admin/templates — listar templates activos (cualquier usuario autenticado)
 * POST /api/admin/templates — crear template (solo admin)
 * DELETE /api/admin/templates?id=X — eliminar template (solo admin)
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { listActiveTemplates, createTemplate, deleteTemplate } from "@rag-saldivia/db"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })

  const templates = await listActiveTemplates()
  return NextResponse.json({ ok: true, data: templates })
}

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Solo admins" }, { status: 403 })

  const body = await request.json().catch(() => null)
  if (!body?.title || !body?.prompt) {
    return NextResponse.json({ ok: false, error: "title y prompt son requeridos" }, { status: 400 })
  }

  const template = await createTemplate({
    title: body.title as string,
    prompt: body.prompt as string,
    focusMode: (body.focusMode as string) ?? "detallado",
    createdBy: Number(claims.sub),
  })

  return NextResponse.json({ ok: true, data: template })
}

export async function DELETE(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Solo admins" }, { status: 403 })

  const { searchParams } = new URL(request.url)
  const id = searchParams.get("id")
  if (!id) return NextResponse.json({ ok: false, error: "id requerido" }, { status: 400 })

  await deleteTemplate(Number(id))
  return NextResponse.json({ ok: true })
}
