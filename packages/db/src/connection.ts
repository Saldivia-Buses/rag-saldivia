/**
 * Conexión singleton a la base de datos SQLite.
 *
 * Usa better-sqlite3 (Node.js compatible) para Next.js.
 * En scripts puros de Bun (migrate.ts, seed.ts, worker) se usa bun:sqlite via init.ts.
 */

import Database from "better-sqlite3"
import { drizzle } from "drizzle-orm/better-sqlite3"
import { mkdirSync } from "fs"
import { join, dirname } from "path"
import * as schema from "./schema.js"

const DEFAULT_DB_PATH = join(process.cwd(), "data", "app.db")

function getDbPath(): string {
  return process.env["DATABASE_PATH"] ?? DEFAULT_DB_PATH
}

function createConnection() {
  const dbPath = getDbPath()

  if (dbPath !== ":memory:") {
    try {
      mkdirSync(dirname(dbPath), { recursive: true })
    } catch {
      // Ignorar si ya existe
    }
  }

  const sqlite = new Database(dbPath)

  sqlite.pragma("journal_mode = WAL")
  sqlite.pragma("synchronous = NORMAL")
  sqlite.pragma("foreign_keys = ON")
  sqlite.pragma("cache_size = -32768")
  sqlite.pragma("temp_store = MEMORY")

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
