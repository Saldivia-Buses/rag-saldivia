import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { listProjects, createProject, deleteProject } from "@rag-saldivia/db"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  const ps = await listProjects(Number(claims.sub))
  return NextResponse.json({ ok: true, data: ps })
}

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })

  const body = await request.json().catch(() => null) as { name?: string; description?: string; instructions?: string } | null
  if (!body?.name) return NextResponse.json({ ok: false, error: "name requerido" }, { status: 400 })

  const project = await createProject({
    userId: Number(claims.sub),
    name: body.name,
    description: body.description ?? "",
    instructions: body.instructions ?? "",
  })
  return NextResponse.json({ ok: true, data: project })
}

export async function DELETE(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  const { searchParams } = new URL(request.url)
  const id = searchParams.get("id")
  if (!id) return NextResponse.json({ ok: false, error: "id requerido" }, { status: 400 })
  await deleteProject(id, Number(claims.sub))
  return NextResponse.json({ ok: true })
}
