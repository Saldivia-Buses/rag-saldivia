/**
 * GET /api/changelog
 *
 * Lee CHANGELOG.md del repo y retorna:
 * - version: versión actual del package.json
 * - entries: últimas 5 entradas del CHANGELOG
 */

import { NextResponse } from "next/server"
import { readFileSync } from "fs"
import { join } from "path"
import { parseChangelog } from "@/lib/changelog"

function getVersion(): string {
  try {
    const pkgPath = join(process.cwd(), "../../package.json")
    const pkg = JSON.parse(readFileSync(pkgPath, "utf-8")) as { version?: string }
    return pkg.version ?? "0.1.0"
  } catch {
    return "0.1.0"
  }
}

export async function GET() {
  try {
    const changelogPath = join(process.cwd(), "../../CHANGELOG.md")
    const raw = readFileSync(changelogPath, "utf-8")
    const entries = parseChangelog(raw)
    const version = getVersion()

    return NextResponse.json({ ok: true, version, entries })
  } catch {
    return NextResponse.json({ ok: false, error: "No se pudo leer el CHANGELOG" }, { status: 500 })
  }
}
