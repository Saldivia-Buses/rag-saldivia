/**
 * Tests de queries de fuentes externas contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/external-sources.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser } from "./setup"
import { createExternalSource, listExternalSources, listActiveSourcesToSync, updateSourceLastSync, deleteExternalSource } from "../queries/external-sources"
import * as schema from "../schema"
import { eq } from "drizzle-orm"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM external_sources; DELETE FROM users;")
})

describe("createExternalSource", () => {
  test("crea fuente con active=true y lastSync null", async () => {
    const user = await insertUser(db)
    const src = await createExternalSource({ userId: user.id, provider: "google_drive", name: "Mi Drive", collectionDest: "docs" })
    expect(src.provider).toBe("google_drive")
    expect(src.active).toBe(true)
    expect(src.lastSync).toBeNull()
    expect(src.id).toHaveLength(36)
  })

  test("acepta los tres providers", async () => {
    const user = await insertUser(db)
    const gd = await createExternalSource({ userId: user.id, provider: "google_drive", name: "GD", collectionDest: "c" })
    const sp = await createExternalSource({ userId: user.id, provider: "sharepoint", name: "SP", collectionDest: "c" })
    const cf = await createExternalSource({ userId: user.id, provider: "confluence", name: "CF", collectionDest: "c" })
    expect([gd.provider, sp.provider, cf.provider]).toEqual(["google_drive", "sharepoint", "confluence"])
  })
})

describe("listExternalSources", () => {
  test("retorna solo fuentes del usuario especificado", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    await createExternalSource({ userId: u1.id, provider: "google_drive", name: "GD", collectionDest: "c" })
    await createExternalSource({ userId: u2.id, provider: "sharepoint", name: "SP", collectionDest: "c" })
    expect(await listExternalSources(u1.id)).toHaveLength(1)
  })
})

describe("listActiveSourcesToSync", () => {
  test("retorna fuentes activas nunca sincronizadas", async () => {
    const user = await insertUser(db)
    await createExternalSource({ userId: user.id, provider: "google_drive", name: "Nuevo", collectionDest: "c", schedule: "daily" })
    const toSync = await listActiveSourcesToSync()
    expect(toSync.some((s) => s.name === "Nuevo")).toBe(true)
  })

  test("no retorna fuentes recientemente sincronizadas", async () => {
    const user = await insertUser(db)
    const src = await createExternalSource({ userId: user.id, provider: "google_drive", name: "Reciente", collectionDest: "c", schedule: "daily" })
    await db.update(schema.externalSources).set({ lastSync: Date.now() }).where(eq(schema.externalSources.id, src.id))
    const toSync = await listActiveSourcesToSync()
    expect(toSync.find((s) => s.id === src.id)).toBeUndefined()
  })

  test("no retorna fuentes inactivas", async () => {
    const user = await insertUser(db)
    const src = await createExternalSource({ userId: user.id, provider: "confluence", name: "Inactiva", collectionDest: "c" })
    await db.update(schema.externalSources).set({ active: false }).where(eq(schema.externalSources.id, src.id))
    const toSync = await listActiveSourcesToSync()
    expect(toSync.find((s) => s.id === src.id)).toBeUndefined()
  })
})

describe("updateSourceLastSync / deleteExternalSource", () => {
  test("updateSourceLastSync actualiza lastSync", async () => {
    const user = await insertUser(db)
    const src = await createExternalSource({ userId: user.id, provider: "google_drive", name: "GD", collectionDest: "c" })
    const before = Date.now()
    await updateSourceLastSync(src.id)
    const updated = (await listExternalSources(user.id))[0]!
    expect(updated.lastSync).toBeGreaterThanOrEqual(before)
  })

  test("deleteExternalSource no elimina fuentes de otro usuario", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    const src = await createExternalSource({ userId: u1.id, provider: "google_drive", name: "Prot", collectionDest: "c" })
    await deleteExternalSource(src.id, u2.id)
    expect(await listExternalSources(u1.id)).toHaveLength(1)
  })

  test("deleteExternalSource elimina la fuente del usuario correcto", async () => {
    const user = await insertUser(db)
    const src = await createExternalSource({ userId: user.id, provider: "google_drive", name: "Del", collectionDest: "c" })
    await deleteExternalSource(src.id, user.id)
    expect(await listExternalSources(user.id)).toHaveLength(0)
  })
})
