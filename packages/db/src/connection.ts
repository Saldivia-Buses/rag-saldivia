/**
 * Conexión singleton a la base de datos SQLite.
 *
 * Usa @libsql/client (JavaScript puro, sin compilación nativa).
 * Compatible con Bun Y Node.js — sin problemas de plataforma.
 * Soportado oficialmente por Drizzle ORM.
 */

import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { mkdirSync } from "fs"
import { join, dirname, resolve } from "path"
import * as schema from "./schema"

const DEFAULT_DB_PATH = join(process.cwd(), "data", "app.db")

function getDbUrl(): string {
  const path = process.env["DATABASE_PATH"] ?? DEFAULT_DB_PATH
  if (path === ":memory:") return ":memory:"

  // Asegurar que el directorio existe
  try {
    mkdirSync(dirname(resolve(path)), { recursive: true })
  } catch {
    // Ignorar si ya existe
  }

  // @libsql/client usa URLs tipo file:/absolute/path
  return `file:${resolve(path)}`
}

function createConnection() {
  const url = getDbUrl()
  const client = createClient({ url })
  return drizzle(client, { schema })
}

let _db: ReturnType<typeof createConnection> | null = null

export function getDb() {
  if (!_db) {
    _db = createConnection()
  }
  return _db
}

/**
 * Reemplaza el singleton de DB con una instancia externa.
 * Solo para uso en tests — permite inyectar una DB en memoria.
 */
export function _injectDbForTesting(db: ReturnType<typeof createConnection>) {
  _db = db
}

/** Resetea el singleton. Usar en afterAll de tests para aislar suites. */
export function _resetDbForTesting() {
  _db = null
}

export const db = new Proxy({} as ReturnType<typeof createConnection>, {
  get(_, prop) {
    return getDb()[prop as keyof ReturnType<typeof createConnection>]
  },
})

export type Db = ReturnType<typeof getDb>
