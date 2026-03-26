/**
 * Tests de queries de áreas contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/areas.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema } from "./setup"
import { createArea, listAreas, getAreaById, updateArea, deleteArea, setAreaCollections, addAreaCollection, removeAreaCollection, countUsersInArea } from "../queries/areas"
import * as schema from "../schema"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM area_collections; DELETE FROM user_areas; DELETE FROM areas; DELETE FROM users;")
})

describe("createArea", () => {
  test("crea un área con nombre y descripción", async () => {
    const area = await createArea("Ventas", "Área comercial")
    expect(area!.name).toBe("Ventas")
    expect(area!.description).toBe("Área comercial")
    expect(area!.createdAt).toBeGreaterThan(0)
  })

  test("crea área con descripción vacía por defecto", async () => {
    const area = await createArea("IT")
    expect(area!.description).toBe("")
  })
})

describe("listAreas", () => {
  test("retorna vacío si no hay áreas", async () => {
    const areas = await listAreas()
    expect(areas).toHaveLength(0)
  })

  test("retorna áreas ordenadas por nombre", async () => {
    await createArea("Zeta")
    await createArea("Alpha")
    const areas = await listAreas()
    expect(areas[0]!.name).toBe("Alpha")
    expect(areas[1]!.name).toBe("Zeta")
  })
})

describe("setAreaCollections", () => {
  test("reemplaza todas las colecciones del área", async () => {
    const area = await createArea("Finanzas")
    await addAreaCollection(area!.id, "old-col")
    await setAreaCollections(area!.id, [{ name: "new-col", permission: "read" }])

    const found = await getAreaById(area!.id)
    expect(found!.areaCollections).toHaveLength(1)
    expect(found!.areaCollections[0]!.collectionName).toBe("new-col")
  })

  test("acepta array vacío — borra todas las colecciones", async () => {
    const area = await createArea("RR.HH.")
    await addAreaCollection(area!.id, "hr-docs")
    await setAreaCollections(area!.id, [])

    const found = await getAreaById(area!.id)
    expect(found!.areaCollections).toHaveLength(0)
  })
})

describe("addAreaCollection", () => {
  test("agrega colección con permiso read por defecto", async () => {
    const area = await createArea("Soporte")
    await addAreaCollection(area!.id, "support-docs")

    const found = await getAreaById(area!.id)
    expect(found!.areaCollections[0]!.permission).toBe("read")
  })

  test("actualiza permiso si la colección ya existe (upsert)", async () => {
    const area = await createArea("Legal")
    await addAreaCollection(area!.id, "legal-docs", "read")
    await addAreaCollection(area!.id, "legal-docs", "admin")

    const found = await getAreaById(area!.id)
    expect(found!.areaCollections).toHaveLength(1)
    expect(found!.areaCollections[0]!.permission).toBe("admin")
  })
})

describe("removeAreaCollection", () => {
  test("elimina solo la colección especificada, no las demás del área", async () => {
    const area = await createArea("Ops")
    await addAreaCollection(area!.id, "ops-a")
    await addAreaCollection(area!.id, "ops-b")
    await removeAreaCollection(area!.id, "ops-a")

    const found = await getAreaById(area!.id)
    expect(found!.areaCollections).toHaveLength(1)
    expect(found!.areaCollections[0]!.collectionName).toBe("ops-b")
  })

  test("no afecta colecciones con el mismo nombre en otras áreas", async () => {
    const a1 = await createArea("Area1")
    const a2 = await createArea("Area2")
    await addAreaCollection(a1!.id, "shared-col")
    await addAreaCollection(a2!.id, "shared-col")
    await removeAreaCollection(a1!.id, "shared-col")

    const found2 = await getAreaById(a2!.id)
    expect(found2!.areaCollections).toHaveLength(1)
  })
})

describe("updateArea / deleteArea", () => {
  test("actualiza nombre y descripción", async () => {
    const area = await createArea("Old Name")
    const updated = await updateArea(area!.id, { name: "New Name", description: "Desc" })
    expect(updated!.name).toBe("New Name")
    expect(updated!.description).toBe("Desc")
  })

  test("elimina el área", async () => {
    const area = await createArea("ToDelete")
    await deleteArea(area!.id)
    const found = await getAreaById(area!.id)
    expect(found).toBeUndefined()
  })
})
