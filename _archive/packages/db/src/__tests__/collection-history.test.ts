/**
 * Tests de queries de historial de colecciones contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/collection-history.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser } from "./setup"
import { recordIngestionEvent, listHistoryByCollection } from "../queries/collection-history"
import * as schema from "../schema"
import { randomUUID } from "crypto"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM collection_history; DELETE FROM users;")
})

describe("recordIngestionEvent", () => {
  test("crea registro con todos los campos", async () => {
    const user = await insertUser(db)
    const record = await recordIngestionEvent({ collection: "docs", userId: user.id, action: "added", filename: "doc.pdf", docCount: 5 })
    expect(record.collection).toBe("docs")
    expect(record.action).toBe("added")
    expect(record.filename).toBe("doc.pdf")
    expect(record.docCount).toBe(5)
    expect(record.id).toHaveLength(36)
  })

  test("acepta action 'removed' y campos opcionales null", async () => {
    const user = await insertUser(db)
    const record = await recordIngestionEvent({ collection: "docs", userId: user.id, action: "removed" })
    expect(record.action).toBe("removed")
    expect(record.filename).toBeNull()
  })
})

describe("listHistoryByCollection", () => {
  test("retorna vacío para colección sin historial", async () => {
    expect(await listHistoryByCollection("inexistente")).toHaveLength(0)
  })

  test("retorna solo registros de la colección especificada", async () => {
    const user = await insertUser(db)
    await recordIngestionEvent({ collection: "docs", userId: user.id, action: "added" })
    await recordIngestionEvent({ collection: "otros", userId: user.id, action: "added" })
    const history = await listHistoryByCollection("docs")
    expect(history).toHaveLength(1)
    expect(history[0]!.collection).toBe("docs")
  })

  test("retorna ordenado por createdAt descendente", async () => {
    const user = await insertUser(db)
    await db.insert(schema.collectionHistory).values({ id: randomUUID(), collection: "docs", userId: user.id, action: "added", createdAt: 1000 })
    await db.insert(schema.collectionHistory).values({ id: randomUUID(), collection: "docs", userId: user.id, action: "added", createdAt: 3000 })
    const history = await listHistoryByCollection("docs")
    expect(history[0]!.createdAt).toBe(3000)
    expect(history[1]!.createdAt).toBe(1000)
  })
})
