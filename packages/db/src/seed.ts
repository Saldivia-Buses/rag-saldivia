#!/usr/bin/env bun
/**
 * Seed de datos de desarrollo.
 * Crea un usuario admin por defecto y un área de ejemplo.
 *
 * Uso: bun run db:seed
 *      bun packages/db/src/seed.ts
 *
 * IMPORTANTE: Solo para desarrollo. No correr en producción.
 */

import { getDb } from "./connection.js"
import { areas, users, userAreas, areaCollections } from "./schema.js"
import { hashSync } from "bcrypt-ts"

const db = getDb()

async function seed() {
  console.log("Seeding base de datos de desarrollo...")

  const now = Date.now()

  // ── Área por defecto ───────────────────────────────────────────────────
  const [defaultArea] = await db
    .insert(areas)
    .values({
      name: "General",
      description: "Área general de acceso",
      createdAt: now,
    })
    .onConflictDoNothing()
    .returning()

  const areaId = defaultArea?.id

  if (!areaId) {
    console.log("  El área 'General' ya existe — skipping")
  } else {
    console.log(`  Área creada: General (id=${areaId})`)
  }

  // Obtener el área existente si ya existía
  const existingArea = areaId
    ? { id: areaId }
    : await db.query.areas.findFirst({ where: (a, { eq }) => eq(a.name, "General") })

  if (!existingArea) throw new Error("No se pudo crear o encontrar el área General")

  // ── Colección de ejemplo ───────────────────────────────────────────────
  await db
    .insert(areaCollections)
    .values({
      areaId: existingArea.id,
      collectionName: "tecpia",
      permission: "write",
    })
    .onConflictDoNothing()

  console.log("  Colección 'tecpia' asignada al área General")

  // ── Usuario admin ──────────────────────────────────────────────────────
  const adminPasswordHash = hashSync("changeme", 10)
  const adminApiKeyHash = Bun.hash("rsk_dev_admin_api_key_changeme").toString()

  const [adminUser] = await db
    .insert(users)
    .values({
      email: "admin@localhost",
      name: "Admin (dev)",
      role: "admin",
      apiKeyHash: adminApiKeyHash,
      passwordHash: adminPasswordHash,
      preferences: {},
      active: true,
      createdAt: now,
    })
    .onConflictDoNothing()
    .returning()

  if (!adminUser) {
    console.log("  Usuario admin@localhost ya existe — skipping")
  } else {
    console.log(`  Usuario admin creado: admin@localhost (id=${adminUser.id})`)
    console.log("    Email:    admin@localhost")
    console.log("    Password: changeme")
    console.log("    IMPORTANTE: cambiar antes de usar en producción")

    // Asignar admin al área general
    await db
      .insert(userAreas)
      .values({ userId: adminUser.id, areaId: existingArea.id })
      .onConflictDoNothing()
  }

  // ── Usuario de prueba ──────────────────────────────────────────────────
  const userPasswordHash = hashSync("test1234", 10)
  const userApiKeyHash = Bun.hash("rsk_dev_user_api_key_changeme").toString()

  const [testUser] = await db
    .insert(users)
    .values({
      email: "user@localhost",
      name: "Usuario de prueba",
      role: "user",
      apiKeyHash: userApiKeyHash,
      passwordHash: userPasswordHash,
      preferences: {},
      active: true,
      createdAt: now,
    })
    .onConflictDoNothing()
    .returning()

  if (!testUser) {
    console.log("  Usuario user@localhost ya existe — skipping")
  } else {
    console.log(`  Usuario de prueba creado: user@localhost (id=${testUser.id})`)

    await db
      .insert(userAreas)
      .values({ userId: testUser.id, areaId: existingArea.id })
      .onConflictDoNothing()
  }

  console.log("\nSeed completado.")
  console.log("Credenciales de desarrollo:")
  console.log("  admin@localhost / changeme  (admin)")
  console.log("  user@localhost / test1234   (user)")
}

seed().catch((err) => {
  console.error("Seed falló:", err)
  process.exit(1)
})
