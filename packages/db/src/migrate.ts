#!/usr/bin/env bun
/**
 * Corre las migraciones de la base de datos.
 * Para desarrollo local: usa init.ts (SQL puro, sin drizzle-kit).
 * Para producción con migraciones incrementales: usar drizzle-kit migrate.
 *
 * Uso: bun run db:migrate
 */

import "./init.js"
