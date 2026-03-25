import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { listHistoryByCollection } from "@rag-saldivia/db"

export async function GET(
  request: Request,
  { params }: { params: Promise<{ name: string }> }
) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })

  const { name } = await params
  const history = await listHistoryByCollection(decodeURIComponent(name))
  return NextResponse.json({ ok: true, data: history })
}
