import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getMemory, setMemory, deleteMemory } from "@rag-saldivia/db"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false }, { status: 401 })
  const memory = await getMemory(Number(claims.sub))
  return NextResponse.json({ ok: true, data: memory })
}

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false }, { status: 401 })
  const body = await request.json().catch(() => null) as { key?: string; value?: string } | null
  if (!body?.key || !body.value) return NextResponse.json({ ok: false, error: "key y value requeridos" }, { status: 400 })
  await setMemory(Number(claims.sub), body.key, body.value, "explicit")
  return NextResponse.json({ ok: true })
}

export async function DELETE(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false }, { status: 401 })
  const { searchParams } = new URL(request.url)
  const key = searchParams.get("key")
  if (!key) return NextResponse.json({ ok: false, error: "key requerido" }, { status: 400 })
  await deleteMemory(Number(claims.sub), key)
  return NextResponse.json({ ok: true })
}
