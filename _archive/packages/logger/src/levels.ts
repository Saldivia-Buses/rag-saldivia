import type { LogLevel } from "@rag-saldivia/shared"

export const LOG_LEVEL_PRIORITY: Record<LogLevel, number> = {
  TRACE: 0,
  DEBUG: 1,
  INFO: 2,
  WARN: 3,
  ERROR: 4,
  FATAL: 5,
}

export function shouldLog(eventLevel: LogLevel, configuredLevel: LogLevel): boolean {
  return LOG_LEVEL_PRIORITY[eventLevel] >= LOG_LEVEL_PRIORITY[configuredLevel]
}

// ANSI colors para el logger de dev
export const LEVEL_COLORS: Record<LogLevel, string> = {
  TRACE: "\x1b[2m",    // dim
  DEBUG: "\x1b[36m",   // cyan
  INFO:  "\x1b[32m",   // green
  WARN:  "\x1b[33m",   // yellow
  ERROR: "\x1b[31m",   // red
  FATAL: "\x1b[35m",   // magenta
}

export const RESET = "\x1b[0m"
export const DIM = "\x1b[2m"
export const BOLD = "\x1b[1m"
