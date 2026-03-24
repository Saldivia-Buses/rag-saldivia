/**
 * Conexión singleton a la base de datos SQLite.
 *
 * Usa bun:sqlite (nativo de Bun) — sin compilación nativa, funciona en todos
 * los sistemas donde corra Bun. Más rápido que better-sqlite3.
 * SQLite serializa writes internamente — no hay race conditions entre workers.
 */

import { Database } from "bun:sqlite"
import { drizzle } from "drizzle-orm/bun-sqlite"
import { join, dirname } from "path"
import { mkdirSync } from "fs"
import * as schema from "./schema.js"

const DEFAULT_DB_PATH = join(process.cwd(), "data", "app.db")

function getDbPath(): string {
  return process.env["DATABASE_PATH"] ?? DEFAULT_DB_PATH
}

function createConnection() {
  const dbPath = getDbPath()

  // Crear directorio si no existe
  if (dbPath !== ":memory:") {
    try {
      mkdirSync(dirname(dbPath), { recursive: true })
    } catch {
      // Ignorar si ya existe
    }
  }

  const sqlite = new Database(dbPath)

  // Performance pragmas
  sqlite.exec("PRAGMA journal_mode = WAL")     // Write-Ahead Logging: mejor concurrencia
  sqlite.exec("PRAGMA synchronous = NORMAL")   // Más rápido, durabilidad suficiente con WAL
  sqlite.exec("PRAGMA foreign_keys = ON")      // Enforce FK constraints
  sqlite.exec("PRAGMA cache_size = -32768")    // 32MB de cache
  sqlite.exec("PRAGMA temp_store = MEMORY")    // Tablas temporales en memoria

  return drizzle(sqlite, { schema })
}

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
