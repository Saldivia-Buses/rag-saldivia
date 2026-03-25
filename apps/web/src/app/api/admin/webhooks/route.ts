import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { listAllWebhooks, createWebhook, deleteWebhook } from "@rag-saldivia/db"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims || claims.role !== "admin") return NextResponse.json({ ok: false, error: "Solo admins" }, { status: 403 })
  const hooks = await listAllWebhooks()
  return NextResponse.json({ ok: true, data: hooks })
}

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims || claims.role !== "admin") return NextResponse.json({ ok: false, error: "Solo admins" }, { status: 403 })

  const body = await request.json().catch(() => null) as { url?: string; events?: string[] } | null
  if (!body?.url || !body.events?.length) {
    return NextResponse.json({ ok: false, error: "url y events son requeridos" }, { status: 400 })
  }

  const webhook = await createWebhook({ userId: Number(claims.sub), url: body.url, events: body.events })
  return NextResponse.json({ ok: true, data: webhook })
}

export async function DELETE(request: Request) {
  const claims = await extractClaims(request)
  if (!claims || claims.role !== "admin") return NextResponse.json({ ok: false, error: "Solo admins" }, { status: 403 })

  const { searchParams } = new URL(request.url)
  const id = searchParams.get("id")
  if (!id) return NextResponse.json({ ok: false, error: "id requerido" }, { status: 400 })

  await deleteWebhook(id, Number(claims.sub))
  return NextResponse.json({ ok: true })
}
