import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { listExternalSources, createExternalSource, deleteExternalSource } from "@rag-saldivia/db"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims || claims.role !== "admin") return NextResponse.json({ ok: false }, { status: 403 })
  const sources = await listExternalSources(Number(claims.sub))
  return NextResponse.json({ ok: true, data: sources })
}

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims || claims.role !== "admin") return NextResponse.json({ ok: false }, { status: 403 })
  const body = await request.json().catch(() => null) as {
    provider?: "google_drive" | "sharepoint" | "confluence"
    name?: string
    collectionDest?: string
    schedule?: "hourly" | "daily" | "weekly"
  } | null
  if (!body?.provider || !body.name || !body.collectionDest) {
    return NextResponse.json({ ok: false, error: "provider, name y collectionDest requeridos" }, { status: 400 })
  }
  const source = await createExternalSource({
    userId: Number(claims.sub),
    provider: body.provider,
    name: body.name,
    collectionDest: body.collectionDest,
    schedule: body.schedule ?? "daily",
    credentials: "{}",
  })
  return NextResponse.json({ ok: true, data: source })
}

export async function DELETE(request: Request) {
  const claims = await extractClaims(request)
  if (!claims || claims.role !== "admin") return NextResponse.json({ ok: false }, { status: 403 })
  const { searchParams } = new URL(request.url)
  const id = searchParams.get("id")
  if (!id) return NextResponse.json({ ok: false }, { status: 400 })
  await deleteExternalSource(id, Number(claims.sub))
  return NextResponse.json({ ok: true })
}
