/**
 * Tests de queries de rate limits contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/rate-limits.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser } from "./setup"
import { createRateLimit, getRateLimit, countQueriesLastHour, listRateLimits, deleteRateLimit } from "../queries/rate-limits"
import { createArea } from "../queries/areas"
import { addUserArea } from "../queries/users"
import { writeEvent } from "../queries/events"
import * as schema from "../schema"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM rate_limits; DELETE FROM events; DELETE FROM user_areas; DELETE FROM area_collections; DELETE FROM areas; DELETE FROM users;")
})

describe("createRateLimit", () => {
  test("crea límite a nivel usuario", async () => {
    const user = await insertUser(db)
    const limit = await createRateLimit({ targetType: "user", targetId: user.id, maxQueriesPerHour: 100 })
    expect(limit.targetType).toBe("user")
    expect(limit.maxQueriesPerHour).toBe(100)
    expect(limit.active).toBe(true)
  })
})

describe("getRateLimit", () => {
  test("retorna null si no hay límite", async () => {
    const user = await insertUser(db)
    expect(await getRateLimit(user.id)).toBeNull()
  })

  test("retorna límite a nivel usuario", async () => {
    const user = await insertUser(db)
    await createRateLimit({ targetType: "user", targetId: user.id, maxQueriesPerHour: 200 })
    expect(await getRateLimit(user.id)).toBe(200)
  })

  test("prioriza límite de usuario sobre área", async () => {
    const user = await insertUser(db)
    const area = await createArea("Ventas")
    await addUserArea(user.id, area!.id)
    await createRateLimit({ targetType: "area", targetId: area!.id, maxQueriesPerHour: 30 })
    await createRateLimit({ targetType: "user", targetId: user.id, maxQueriesPerHour: 500 })
    expect(await getRateLimit(user.id)).toBe(500)
  })

  test("retorna límite de área si no hay de usuario", async () => {
    const user = await insertUser(db)
    const area = await createArea("Soporte")
    await addUserArea(user.id, area!.id)
    await createRateLimit({ targetType: "area", targetId: area!.id, maxQueriesPerHour: 75 })
    expect(await getRateLimit(user.id)).toBe(75)
  })
})

describe("countQueriesLastHour", () => {
  test("retorna 0 sin queries", async () => {
    const user = await insertUser(db)
    expect(await countQueriesLastHour(user.id)).toBe(0)
  })

  test("cuenta solo eventos rag.stream_started del usuario", async () => {
    const user = await insertUser(db)
    await writeEvent({ source: "backend", level: "INFO", type: "rag.stream_started", userId: user.id })
    await writeEvent({ source: "backend", level: "INFO", type: "rag.stream_started", userId: user.id })
    await writeEvent({ source: "backend", level: "INFO", type: "auth.login", userId: user.id })
    expect(await countQueriesLastHour(user.id)).toBe(2)
  })

  test("no cuenta eventos de hace más de 1 hora", async () => {
    const user = await insertUser(db)
    // Insertar evento con timestamp de hace 2 horas
    await db.insert(schema.events).values({
      id: crypto.randomUUID(),
      ts: Date.now() - 2 * 60 * 60 * 1000,
      source: "backend", level: "INFO", type: "rag.stream_started",
      userId: user.id, payload: {}, sequence: 9999,
    })
    expect(await countQueriesLastHour(user.id)).toBe(0)
  })
})

describe("listRateLimits / deleteRateLimit", () => {
  test("lista todos los rate limits", async () => {
    const user = await insertUser(db)
    await createRateLimit({ targetType: "user", targetId: user.id, maxQueriesPerHour: 100 })
    expect(await listRateLimits()).toHaveLength(1)
  })

  test("deleteRateLimit elimina el límite", async () => {
    const user = await insertUser(db)
    const limit = await createRateLimit({ targetType: "user", targetId: user.id, maxQueriesPerHour: 50 })
    await deleteRateLimit(limit.id)
    expect(await listRateLimits()).toHaveLength(0)
  })
})
