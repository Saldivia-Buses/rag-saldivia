/**
 * Tests de queries de áreas contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/areas.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { and, eq } from "drizzle-orm"
import * as schema from "../schema"

process.env["DATABASE_PATH"] = ":memory:"

const client = createClient({ url: ":memory:" })
const testDb = drizzle(client, { schema })

beforeAll(async () => {
  await client.executeMultiple(`
    CREATE TABLE IF NOT EXISTS areas (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL UNIQUE,
      description TEXT NOT NULL DEFAULT '',
      created_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS area_collections (
      area_id INTEGER NOT NULL REFERENCES areas(id) ON DELETE CASCADE,
      collection_name TEXT NOT NULL,
      permission TEXT NOT NULL DEFAULT 'read',
      PRIMARY KEY (area_id, collection_name)
    );
  `)
})

afterEach(async () => {
  await client.executeMultiple("DELETE FROM area_collections; DELETE FROM areas;")
})

// Helpers locales que usan testDb en lugar del singleton

async function createTestArea(name: string, description = "") {
  const [area] = await testDb
    .insert(schema.areas)
    .values({ name, description, createdAt: Date.now() })
    .returning()
  if (!area) throw new Error("Failed to create area")
  return area
}

async function addCollection(
  areaId: number,
  collectionName: string,
  permission: "read" | "write" | "admin" = "read"
) {
  await testDb
    .insert(schema.areaCollections)
    .values({ areaId, collectionName, permission })
    .onConflictDoUpdate({
      target: [schema.areaCollections.areaId, schema.areaCollections.collectionName],
      set: { permission },
    })
}

async function removeCollection(areaId: number, collectionName: string) {
  await testDb
    .delete(schema.areaCollections)
    .where(
      and(
        eq(schema.areaCollections.areaId, areaId),
        eq(schema.areaCollections.collectionName, collectionName)
      )
    )
}

async function getCollections(areaId: number) {
  return testDb.query.areaCollections.findMany({
    where: (ac, { eq }) => eq(ac.areaId, areaId),
  })
}

// ── Tests ──────────────────────────────────────────────────────────────────

describe("removeAreaCollection", () => {
  test("elimina solo la colección especificada, no las demás del área", async () => {
    const area = await createTestArea("Área Test")

    await addCollection(area.id, "col-a", "read")
    await addCollection(area.id, "col-b", "write")
    await addCollection(area.id, "col-c", "admin")

    const antes = await getCollections(area.id)
    expect(antes).toHaveLength(3)

    // Eliminar solo col-b
    await removeCollection(area.id, "col-b")

    const despues = await getCollections(area.id)
    expect(despues).toHaveLength(2)

    const nombres = despues.map((c) => c.collectionName)
    expect(nombres).toContain("col-a")
    expect(nombres).toContain("col-c")
    expect(nombres).not.toContain("col-b")
  })

  test("no afecta colecciones con el mismo nombre en otras áreas", async () => {
    const area1 = await createTestArea("Área 1")
    const area2 = await createTestArea("Área 2")

    await addCollection(area1.id, "col-compartida", "read")
    await addCollection(area2.id, "col-compartida", "read")

    // Eliminar de area1
    await removeCollection(area1.id, "col-compartida")

    const cols1 = await getCollections(area1.id)
    const cols2 = await getCollections(area2.id)

    expect(cols1).toHaveLength(0)
    expect(cols2).toHaveLength(1)
    expect(cols2[0]?.collectionName).toBe("col-compartida")
  })

  test("no falla si la colección no existe en el área", async () => {
    const area = await createTestArea("Área Vacía")

    // No debe lanzar error
    await expect(removeCollection(area.id, "col-inexistente")).resolves.toBeUndefined()
  })

  test("deja el área sin colecciones si se elimina la última", async () => {
    const area = await createTestArea("Área Una Colección")
    await addCollection(area.id, "unica", "read")

    await removeCollection(area.id, "unica")

    const cols = await getCollections(area.id)
    expect(cols).toHaveLength(0)
  })
})

describe("setAreaCollections", () => {
  test("reemplaza todas las colecciones del área", async () => {
    const area = await createTestArea("Área Set")

    await addCollection(area.id, "vieja-1", "read")
    await addCollection(area.id, "vieja-2", "read")

    // Reemplazar con nuevas
    await testDb.delete(schema.areaCollections).where(eq(schema.areaCollections.areaId, area.id))
    await testDb.insert(schema.areaCollections).values([
      { areaId: area.id, collectionName: "nueva-1", permission: "write" },
      { areaId: area.id, collectionName: "nueva-2", permission: "admin" },
    ])

    const cols = await getCollections(area.id)
    expect(cols).toHaveLength(2)

    const nombres = cols.map((c) => c.collectionName)
    expect(nombres).toContain("nueva-1")
    expect(nombres).toContain("nueva-2")
    expect(nombres).not.toContain("vieja-1")
    expect(nombres).not.toContain("vieja-2")
  })

  test("acepta array vacío (borra todas las colecciones del área)", async () => {
    const area = await createTestArea("Área A Vaciar")
    await addCollection(area.id, "col-1", "read")
    await addCollection(area.id, "col-2", "read")

    // Reemplazar con array vacío
    await testDb.delete(schema.areaCollections).where(eq(schema.areaCollections.areaId, area.id))

    const cols = await getCollections(area.id)
    expect(cols).toHaveLength(0)
  })
})

describe("addAreaCollection", () => {
  test("agrega una colección con permiso por defecto read", async () => {
    const area = await createTestArea("Área Add")
    await addCollection(area.id, "mi-coleccion")

    const cols = await getCollections(area.id)
    expect(cols).toHaveLength(1)
    expect(cols[0]?.permission).toBe("read")
  })

  test("actualiza el permiso si la colección ya existe (upsert)", async () => {
    const area = await createTestArea("Área Upsert")
    await addCollection(area.id, "col-upsert", "read")
    await addCollection(area.id, "col-upsert", "admin")

    const cols = await getCollections(area.id)
    expect(cols).toHaveLength(1)
    expect(cols[0]?.permission).toBe("admin")
  })
})
