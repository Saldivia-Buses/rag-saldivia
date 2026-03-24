/**
 * Logger del servidor (Node/Bun).
 *
 * Escribe a:
 *   - Consola (formato legible en dev, JSON en producción)
 *   - Tabla events en SQLite (via @rag-saldivia/db, lazy-loaded)
 *   - Archivo logs/backend.log (rotado cada 10MB)
 *   - Archivo logs/errors.log (solo ERROR y FATAL)
 *
 * Implementación completa en Fase 5. Esta versión provee la API completa
 * con escritura a consola y tabla events. Los archivos de log se agregan en Fase 5.
 */

import type { EventType, LogLevel, EventSource } from "@rag-saldivia/shared"
import { shouldLog, LEVEL_COLORS, RESET, DIM, BOLD } from "./levels.js"
import { getSuggestion } from "./suggestions.js"

export type LogContext = {
  userId?: number | null
  sessionId?: string | null
  ip?: string
  duration?: number
  [key: string]: unknown
}

type LogFn = (type: EventType, payload?: Record<string, unknown>, ctx?: LogContext) => void

// ── Formato ────────────────────────────────────────────────────────────────

function formatTimestamp(): string {
  return new Date().toISOString().slice(11, 23) // HH:MM:SS.mmm
}

function formatPretty(level: LogLevel, type: EventType, payload: Record<string, unknown>, ctx?: LogContext): string {
  const ts = DIM + formatTimestamp() + RESET
  const lvl = LEVEL_COLORS[level] + level.padEnd(5) + RESET
  const t = BOLD + type + RESET

  const parts = [ts, lvl, t]

  if (ctx?.userId) parts.push(DIM + `user=${ctx.userId}` + RESET)
  if (ctx?.duration !== undefined) parts.push(DIM + `${ctx.duration}ms` + RESET)
  if (Object.keys(payload).length > 0) {
    const summary = Object.entries(payload)
      .slice(0, 3)
      .map(([k, v]) => `${k}=${JSON.stringify(v)}`)
      .join("  ")
    parts.push(DIM + summary + RESET)
  }

  let line = parts.join("  ")

  // Sugerencias para errores
  if (level === "ERROR" || level === "FATAL") {
    const errMsg = String(payload["error"] ?? payload["message"] ?? "")
    const suggestion = getSuggestion(errMsg)
    if (suggestion) {
      line += "\n" + suggestion.split("\n").map((l) => "      " + DIM + l + RESET).join("\n")
    }
  }

  return line
}

function formatJson(level: LogLevel, type: EventType, payload: Record<string, unknown>, ctx?: LogContext): string {
  return JSON.stringify({
    ts: Date.now(),
    level,
    type,
    ...ctx,
    ...payload,
  })
}

// ── File logging (con rotación) ────────────────────────────────────────────

let _writeToFile: ((filename: string, line: string) => Promise<void>) | null = null

async function writeToFiles(level: LogLevel, line: string): Promise<void> {
  try {
    if (!_writeToFile) {
      const { writeToLogFile, shouldWriteToErrorLog } = await import("./rotation.js")
      _writeToFile = writeToLogFile
    }
    await _writeToFile("backend.log", line)
    const { shouldWriteToErrorLog } = await import("./rotation.js")
    if (shouldWriteToErrorLog(level)) {
      await _writeToFile("errors.log", line)
    }
  } catch {
    // Silencioso
  }
}

// ── Write Event (lazy-load db para evitar circular deps) ───────────────────

let _writeToDb: ((data: {
  source: EventSource
  level: LogLevel
  type: EventType
  userId?: number | null
  sessionId?: string | null
  payload?: Record<string, unknown>
}) => Promise<void>) | null = null

async function persistEvent(level: LogLevel, type: EventType, payload: Record<string, unknown>, ctx?: LogContext) {
  try {
    if (!_writeToDb) {
      // Lazy-load para evitar dependencias circulares
      const { writeEvent } = await import("@rag-saldivia/db")
      _writeToDb = writeEvent as typeof _writeToDb
    }
    await _writeToDb?.({
      source: "backend",
      level,
      type,
      userId: ctx?.userId ?? null,
      sessionId: ctx?.sessionId ?? null,
      payload,
    })
  } catch {
    // No fallar el logger si la DB no está disponible
  }
}

// ── Core ───────────────────────────────────────────────────────────────────

function createLogger() {
  const configuredLevel = (process.env["LOG_LEVEL"] as LogLevel | undefined) ?? "INFO"
  const isDev = process.env["NODE_ENV"] !== "production"

  function write(level: LogLevel, type: EventType, payload: Record<string, unknown> = {}, ctx?: LogContext) {
    if (!shouldLog(level, configuredLevel)) return

    const line = isDev
      ? formatPretty(level, type, payload, ctx)
      : formatJson(level, type, payload, ctx)

    if (level === "ERROR" || level === "FATAL") {
      console.error(line)
    } else {
      console.log(line)
    }

    // Persistir en DB y archivo de forma asíncrona
    persistEvent(level, type, payload, ctx).catch(() => {})
    writeToFiles(level, line).catch(() => {})
  }

  return {
    trace: ((type, payload, ctx) => write("TRACE", type, payload, ctx)) as LogFn,
    debug: ((type, payload, ctx) => write("DEBUG", type, payload, ctx)) as LogFn,
    info: ((type, payload, ctx) => write("INFO", type, payload, ctx)) as LogFn,
    warn: ((type, payload, ctx) => write("WARN", type, payload, ctx)) as LogFn,
    error: ((type, payload, ctx) => write("ERROR", type, payload, ctx)) as LogFn,
    fatal: ((type, payload, ctx) => write("FATAL", type, payload, ctx)) as LogFn,

    // Helpers de uso común
    request: (method: string, path: string, status: number, duration: number, ctx?: LogContext) => {
      const level: LogLevel = status >= 500 ? "ERROR" : status >= 400 ? "WARN" : "INFO"
      write(level, "system.warning", { method, path, status, duration }, ctx)
    },
  }
}

export const log = createLogger()
