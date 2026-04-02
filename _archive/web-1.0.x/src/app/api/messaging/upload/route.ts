/**
 * /api/messaging/upload — Upload files for messaging.
 * Saves to data/messaging-files/, returns URL + metadata.
 * 10MB limit, server-side validation.
 */

import { NextResponse } from "next/server"
import { requireAuth, apiOk, apiError, apiServerError } from "@/lib/api-utils"
import { writeFile, mkdir } from "fs/promises"
import { join } from "path"
import { randomUUID } from "crypto"

const MAX_FILE_SIZE = 10 * 1024 * 1024 // 10MB
const UPLOAD_DIR = join(process.cwd(), "../../data/messaging-files")

export async function POST(request: Request) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  try {
    const formData = await request.formData()
    const file = formData.get("file") as File | null
    if (!file) return apiError("No se envió archivo")

    if (file.size > MAX_FILE_SIZE) {
      return apiError(`Archivo demasiado grande. Máximo ${MAX_FILE_SIZE / 1024 / 1024}MB`)
    }

    // Ensure upload directory exists
    await mkdir(UPLOAD_DIR, { recursive: true })

    // Generate unique filename
    const ext = file.name.split(".").pop() ?? ""
    const uniqueName = `${randomUUID()}${ext ? `.${ext}` : ""}`
    const filePath = join(UPLOAD_DIR, uniqueName)

    // Write file
    const buffer = Buffer.from(await file.arrayBuffer())
    await writeFile(filePath, buffer)

    const metadata = {
      fileName: file.name,
      fileSize: file.size,
      mimeType: file.type || "application/octet-stream",
      url: `/api/messaging/upload/${uniqueName}`,
    }

    return apiOk(metadata, 201)
  } catch (error) {
    return apiServerError(error, "POST /api/messaging/upload", Number(claims.sub))
  }
}
