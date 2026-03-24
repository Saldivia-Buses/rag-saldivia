#!/usr/bin/env bun
/**
 * Ejecuta las migraciones de la base de datos.
 * Uso: bun run db:migrate
 *      bun packages/db/src/migrate.ts
 */

import { migrate } from "drizzle-orm/better-sqlite3/migrator"
import { getDb } from "./connection.js"
import { join } from "path"

const db = getDb()
const migrationsFolder = join(import.meta.dir, "..", "drizzle")

console.log("Corriendo migraciones...")
migrate(db, { migrationsFolder })
console.log("Migraciones completadas.")
