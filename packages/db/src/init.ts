#!/usr/bin/env bun
/**
 * Inicializa la base de datos aplicando las migraciones de Drizzle Kit.
 * La fuente de verdad es src/schema.ts — los cambios se propagan con db:generate + db:push.
 *
 * Uso: bun packages/db/src/init.ts
 *      bun run db:migrate
 */

import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { migrate } from "drizzle-orm/libsql/migrator"
import { mkdirSync } from "fs"
import { join, dirname, resolve } from "path"

const dbPath = process.env["DATABASE_PATH"] ?? join(process.cwd(), "data", "app.db")
const dbUrl = dbPath === ":memory:" ? ":memory:" : `file:${resolve(dbPath)}`

if (dbPath !== ":memory:") {
  try { mkdirSync(dirname(resolve(dbPath)), { recursive: true }) } catch { /* ya existe */ }
}

const client = createClient({ url: dbUrl })
const db = drizzle(client)

const migrationsFolder = join(import.meta.dir, "../drizzle")
await migrate(db, { migrationsFolder })

console.log("Base de datos migrada correctamente")
