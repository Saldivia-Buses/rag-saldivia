/**
 * Rotación de archivos de log.
 * Mantiene tres archivos:
 *   logs/backend.log   — todos los eventos del servidor
 *   logs/errors.log    — solo ERROR y FATAL
 *   logs/frontend.log  — eventos del browser
 *
 * Rota automáticamente al superar MAX_SIZE_BYTES (default 10MB).
 * Mantiene MAX_ROTATIONS archivos históricos (.1, .2, .3...)
 */

import { appendFile, rename, stat, mkdir } from "fs/promises"
import { existsSync } from "fs"
import { join } from "path"
import type { LogLevel } from "@rag-saldivia/shared"

const MAX_SIZE_BYTES = 10 * 1024 * 1024 // 10MB
const MAX_ROTATIONS = 5
const LOG_DIR = join(process.cwd(), "logs")

// Caché de tamaños para no hacer stat en cada write
const _sizeCache = new Map<string, number>()

async function ensureLogDir(): Promise<void> {
  if (!existsSync(LOG_DIR)) {
    await mkdir(LOG_DIR, { recursive: true })
  }
}

async function rotateIfNeeded(filePath: string): Promise<void> {
  const cached = _sizeCache.get(filePath) ?? 0
  if (cached < MAX_SIZE_BYTES) return

  try {
    const s = await stat(filePath)
    if (s.size < MAX_SIZE_BYTES) {
      _sizeCache.set(filePath, s.size)
      return
    }

    // Rotar: .4 → eliminar, .3 → .4, .2 → .3, .1 → .2, actual → .1
    for (let i = MAX_ROTATIONS - 1; i >= 1; i--) {
      const from = `${filePath}.${i}`
      const to = `${filePath}.${i + 1}`
      if (existsSync(from)) {
        await rename(from, to).catch(() => {})
      }
    }
    await rename(filePath, `${filePath}.1`).catch(() => {})
    _sizeCache.set(filePath, 0)
  } catch {
    // Ignorar errores de rotación — no romper el servidor
  }
}

export async function writeToLogFile(
  filename: "backend.log" | "errors.log" | "frontend.log",
  line: string
): Promise<void> {
  try {
    await ensureLogDir()
    const filePath = join(LOG_DIR, filename)
    await rotateIfNeeded(filePath)
    const entry = line.endsWith("\n") ? line : line + "\n"
    await appendFile(filePath, entry, "utf-8")
    _sizeCache.set(filePath, (_sizeCache.get(filePath) ?? 0) + entry.length)
  } catch {
    // Silencioso — el logger no debe crashear el servidor
  }
}

export function shouldWriteToErrorLog(level: LogLevel): boolean {
  return level === "ERROR" || level === "FATAL"
}
