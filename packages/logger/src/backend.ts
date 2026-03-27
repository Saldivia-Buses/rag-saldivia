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
import { writeEvent } from "@rag-saldivia/db"
import { shouldLog, LEVEL_COLORS, RESET, DIM, BOLD } from "./levels"
import { getSuggestion } from "./suggestions"

export type LogContext = {
  userId?: number | null
  sessionId?: string | null
  ip?: string
  duration?: number
  requestId?: string
  [key: string]: unknown
}

type LogFn = (type: EventType, payload?: Record<string, unknown>, ctx?: LogContext) => void

// ── Helpers de formato (extraídos de formatPretty) ─────────────────────────

function formatTimestamp(): string {
  return new Date().toISOString().slice(11, 23) // HH:MM:SS.mmm
}

export function formatHeader(level: LogLevel, type: EventType): string {
  const ts = DIM + formatTimestamp() + RESET
  const lvl = LEVEL_COLORS[level] + level.padEnd(5) + RESET
  const t = BOLD + type + RESET
  return [ts, lvl, t].join("  ")
}

export function formatContext(ctx?: LogContext): string {
  if (!ctx) return ""
  const parts: string[] = []
  if (ctx.userId) parts.push(`user=${ctx.userId}`)
  if (ctx.requestId) parts.push(`req=${ctx.requestId.slice(0, 8)}`)
  if (ctx.duration !== undefined) parts.push(`${ctx.duration}ms`)
  return parts.length > 0 ? DIM + parts.join("  ") + RESET : ""
}

export function formatPayloadSummary(payload: Record<string, unknown>): string {
  if (Object.keys(payload).length === 0) return ""
  const summary = Object.entries(payload)
    .slice(0, 3)
    .map(([k, v]) => `${k}=${JSON.stringify(v)}`)
    .join("  ")
  return DIM + summary + RESET
}

export function formatSuggestion(level: LogLevel, payload: Record<string, unknown>): string {
  if (level !== "ERROR" && level !== "FATAL") return ""
  const errMsg = String(payload["error"] ?? payload["message"] ?? "")
  const suggestion = getSuggestion(errMsg)
  if (!suggestion) return ""
  return "\n" + suggestion.split("\n").map((l) => "      " + DIM + l + RESET).join("\n")
}

function formatPretty(level: LogLevel, type: EventType, payload: Record<string, unknown>, ctx?: LogContext): string {
  return [
    formatHeader(level, type),
    formatContext(ctx),
    formatPayloadSummary(payload),
  ].filter(Boolean).join("  ") + formatSuggestion(level, payload)
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

type LogFilename = "backend.log" | "errors.log" | "frontend.log"
let _writeToFile: ((filename: LogFilename, line: string) => Promise<void>) | null = null
let _shouldWriteToErrorLog: ((level: LogLevel) => boolean) | null = null

async function writeToFiles(level: LogLevel, line: string): Promise<void> {
  try {
    if (!_writeToFile) {
      const rotation = await import("./rotation")
      _writeToFile = rotation.writeToLogFile
      _shouldWriteToErrorLog = rotation.shouldWriteToErrorLog
    }
    const writeFn = _writeToFile
    const shouldError = _shouldWriteToErrorLog
    if (writeFn) {
      await writeFn("backend.log", line)
      if (shouldError?.(level)) {
        await writeFn("errors.log", line)
      }
    }
  } catch {
    // Silencioso
  }
}

// ── Write Event ────────────────────────────────────────────────────────────

/**
 * Persiste en la tabla `events` vía `writeEvent` de `@rag-saldivia/db`. El campo `type` debe ser un
 * `EventType` válido del schema compartido; un string arbitrario puede degradar el comportamiento
 * en validación. Preferir siempre tipos importados de `@rag-saldivia/shared`.
 */
async function persistEvent(level: LogLevel, type: EventType, payload: Record<string, unknown>, ctx?: LogContext) {
  try {
    // userId=0 es el sistema (SYSTEM_API_KEY) — no tiene FK en la tabla users
    const userId = ctx?.userId != null && ctx.userId > 0 ? ctx.userId : null
    await writeEvent({
      source: "backend",
      level,
      type,
      userId,
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
      process.stdout.write(`${line}\n`)
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
      write(level, "system.request", { method, path, status, duration }, ctx)
    },
  }
}

export const log = createLogger()
