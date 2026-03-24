/**
 * POST /api/upload
 * Recibe un archivo multipart, lo guarda en disco y crea un job en ingestion_queue.
 */

import { NextResponse } from "next/server"
import { writeFile, mkdir } from "fs/promises"
import { join } from "path"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb, ingestionQueue } from "@rag-saldivia/db"
import { canAccessCollection } from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"

const UPLOAD_DIR = join(process.cwd(), "data", "uploads")
const MAX_FILE_SIZE_MB = 100

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) {
    return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  }

  const userId = Number(claims.sub)

  try {
    const formData = await request.formData()
    const file = formData.get("file") as File | null
    const collection = formData.get("collection") as string | null

    if (!file) {
      return NextResponse.json({ ok: false, error: "No se recibió archivo" }, { status: 400 })
    }
    if (!collection) {
      return NextResponse.json({ ok: false, error: "Colección requerida" }, { status: 400 })
    }

    // Verificar permisos de escritura sobre la colección
    const hasAccess = await canAccessCollection(userId, collection, "write")
    if (!hasAccess && claims.role !== "admin") {
      return NextResponse.json(
        { ok: false, error: `Sin permiso de escritura sobre '${collection}'` },
        { status: 403 }
      )
    }

    // Validar tamaño
    const sizeMB = file.size / 1024 / 1024
    if (sizeMB > MAX_FILE_SIZE_MB) {
      return NextResponse.json(
        { ok: false, error: `Archivo demasiado grande (${sizeMB.toFixed(1)}MB, máximo ${MAX_FILE_SIZE_MB}MB)` },
        { status: 413 }
      )
    }

    // Sanitizar nombre de archivo
    const safeName = file.name.replace(/[^a-zA-Z0-9._-]/g, "_")

    // Guardar archivo
    await mkdir(UPLOAD_DIR, { recursive: true })
    const filePath = join(UPLOAD_DIR, `${Date.now()}_${safeName}`)
    const buffer = await file.arrayBuffer()
    await writeFile(filePath, Buffer.from(buffer))

    // Crear job en la cola
    const jobId = crypto.randomUUID()
    const db = getDb()
    await db.insert(ingestionQueue).values({
      id: jobId,
      collection,
      filePath,
      userId,
      priority: 0,
      status: "pending",
      createdAt: Date.now(),
    })

    log.info("ingestion.started", {
      jobId,
      filename: safeName,
      collection,
      sizeMB: sizeMB.toFixed(2),
    }, { userId })

    return NextResponse.json({
      ok: true,
      data: { jobId, filename: safeName, collection },
    })
  } catch (error) {
    log.error("system.error", { error: String(error), endpoint: "POST /api/upload" }, { userId })
    return NextResponse.json({ ok: false, error: "Error al procesar el archivo" }, { status: 500 })
  }
}
