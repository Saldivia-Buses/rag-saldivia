import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { listReportsByUser, createReport, deleteReport } from "@rag-saldivia/db"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Solo admins" }, { status: 403 })

  const reports = await listReportsByUser(Number(claims.sub))
  return NextResponse.json({ ok: true, data: reports })
}

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Solo admins" }, { status: 403 })

  const body = await request.json().catch(() => null) as {
    query?: string
    collection?: string
    schedule?: "daily" | "weekly" | "monthly"
    destination?: "saved" | "email"
    email?: string
  } | null
  if (!body?.query || !body.collection || !body.schedule || !body.destination) {
    return NextResponse.json({ ok: false, error: "Campos requeridos: query, collection, schedule, destination" }, { status: 400 })
  }

  const report = await createReport({
    userId: Number(claims.sub),
    query: body.query,
    collection: body.collection,
    schedule: body.schedule,
    destination: body.destination,
    email: body.email ?? null,
  })

  return NextResponse.json({ ok: true, data: report })
}

export async function DELETE(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Solo admins" }, { status: 403 })

  const { searchParams } = new URL(request.url)
  const id = searchParams.get("id")
  if (!id) return NextResponse.json({ ok: false, error: "id requerido" }, { status: 400 })

  await deleteReport(id, Number(claims.sub))
  return NextResponse.json({ ok: true })
}
