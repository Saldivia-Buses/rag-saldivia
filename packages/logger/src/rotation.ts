/**
 * Rotación de archivos de log.
 * Mantiene tres archivos:
 *   logs/backend.log   — todos los eventos del servidor
 *   logs/errors.log    — solo ERROR y FATAL
 *   logs/frontend.log  — eventos del browser
 *
 * Rota automáticamente al superar MAX_SIZE_BYTES (default 10MB).
 * Mantiene MAX_ROTATIONS archivos históricos (.1, .2, .3...)
 *
 * F8.29 — Elimina _sizeCache Map in-memory.
 * Los tamaños se persisten en Redis HSET `log:sizes` para sobrevivir
 * reinicios del proceso y compartirse entre instancias.
 */

import { appendFile, rename, stat, mkdir } from "fs/promises"
import { existsSync } from "fs"
import { join } from "path"
import type { LogLevel } from "@rag-saldivia/shared"
import { getRedisClient } from "@rag-saldivia/db"

const MAX_SIZE_BYTES = 10 * 1024 * 1024 // 10MB
const MAX_ROTATIONS = 5
const LOG_DIR = join(process.cwd(), "logs")
const SIZES_HASH_KEY = "log:sizes"

async function getLogFileSize(filePath: string): Promise<number> {
  try {
    return Number(await getRedisClient().hget(SIZES_HASH_KEY, filePath) ?? "0")
  } catch {
    return 0
  }
}

async function setLogFileSize(filePath: string, size: number): Promise<void> {
  try {
    await getRedisClient().hset(SIZES_HASH_KEY, filePath, size)
  } catch {
    // Silencioso — no romper el logger si Redis no está disponible
  }
}

async function ensureLogDir(): Promise<void> {
  if (!existsSync(LOG_DIR)) {
    await mkdir(LOG_DIR, { recursive: true })
  }
}

async function rotateIfNeeded(filePath: string): Promise<void> {
  const cached = await getLogFileSize(filePath)
  if (cached < MAX_SIZE_BYTES) return

  try {
    const s = await stat(filePath)
    if (s.size < MAX_SIZE_BYTES) {
      await setLogFileSize(filePath, s.size)
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
    await setLogFileSize(filePath, 0)
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
    const currentSize = await getLogFileSize(filePath)
    await setLogFileSize(filePath, currentSize + entry.length)
  } catch {
    // Silencioso — el logger no debe crashear el servidor
  }
}

export function shouldWriteToErrorLog(level: LogLevel): boolean {
  return level === "ERROR" || level === "FATAL"
}
