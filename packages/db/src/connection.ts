/**
 * Conexión singleton a la base de datos SQLite.
 *
 * Usa better-sqlite3 (síncrono) que es significativamente más rápido que
 * el módulo sqlite3 async para el patrón de acceso de este sistema.
 * SQLite serializa writes internamente — no hay race conditions entre workers.
 */

import Database from "better-sqlite3"
import { drizzle } from "drizzle-orm/better-sqlite3"
import { join } from "path"
import * as schema from "./schema.js"

const DEFAULT_DB_PATH = join(process.cwd(), "data", "app.db")

function getDbPath(): string {
  return process.env["DATABASE_PATH"] ?? DEFAULT_DB_PATH
}

function createConnection() {
  const dbPath = getDbPath()

  // Crear directorio si no existe
  const dir = dbPath.split("/").slice(0, -1).join("/")
  if (dir && dir !== "." && dbPath !== ":memory:") {
    try {
      const { mkdirSync } = require("fs")
      mkdirSync(dir, { recursive: true })
    } catch {
      // En Bun, usar Bun.spawnSync o asumir que el dir existe
    }
  }

  const sqlite = new Database(dbPath)

  // Performance pragmas (seguras — no comprometen durabilidad para este caso de uso)
  sqlite.pragma("journal_mode = WAL")      // Write-Ahead Logging: mejor concurrencia
  sqlite.pragma("synchronous = NORMAL")    // Más rápido, durabilidad suficiente con WAL
  sqlite.pragma("foreign_keys = ON")       // Enforce FK constraints
  sqlite.pragma("cache_size = -32768")     // 32MB de cache
  sqlite.pragma("temp_store = MEMORY")     // Tablas temporales en memoria

  return drizzle(sqlite, { schema })
}

// Singleton: reutiliza la misma conexión en toda la aplicación
let _db: ReturnType<typeof createConnection> | null = null

export function getDb() {
  if (!_db) {
    _db = createConnection()
  }
  return _db
}

export const db = new Proxy({} as ReturnType<typeof createConnection>, {
  get(_, prop) {
    return getDb()[prop as keyof ReturnType<typeof createConnection>]
  },
})

export type Db = ReturnType<typeof getDb>
