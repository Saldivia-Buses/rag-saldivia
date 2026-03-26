/**
 * Tests de queries de eventos (black box) contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/events.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser } from "./setup"
import { writeEvent, queryEvents, getEventsForReplay } from "../queries/events"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM events; DELETE FROM users;")
})

describe("writeEvent", () => {
  test("persiste un evento con todos los campos", async () => {
    const ev = await writeEvent({
      source: "backend",
      level: "INFO",
      type: "auth.login",
      payload: { email: "test@example.com" },
    })
    expect(ev!.source).toBe("backend")
    expect(ev!.level).toBe("INFO")
    expect(ev!.type).toBe("auth.login")
    expect(ev!.payload).toMatchObject({ email: "test@example.com" })
    expect(ev!.id).toHaveLength(36)
  })

  test("asigna sequence monotónicamente creciente", async () => {
    const e1 = await writeEvent({ source: "backend", level: "INFO", type: "ev.one" })
    const e2 = await writeEvent({ source: "backend", level: "INFO", type: "ev.two" })
    expect(e2!.sequence).toBeGreaterThan(e1!.sequence)
  })

  test("permite userId null", async () => {
    const ev = await writeEvent({ source: "backend", level: "INFO", type: "system.startup", userId: null })
    expect(ev!.userId).toBeNull()
  })

  test("asocia userId cuando se provee", async () => {
    const user = await insertUser(db)
    const ev = await writeEvent({ source: "backend", level: "INFO", type: "auth.login", userId: user.id })
    expect(ev!.userId).toBe(user.id)
  })
})

describe("queryEvents", () => {
  test("retorna todos los eventos sin filtros", async () => {
    await writeEvent({ source: "backend", level: "INFO", type: "a" })
    await writeEvent({ source: "frontend", level: "WARN", type: "b" })
    const results = await queryEvents({})
    expect(results.length).toBeGreaterThanOrEqual(2)
  })

  test("filtra por type", async () => {
    await writeEvent({ source: "backend", level: "INFO", type: "auth.login" })
    await writeEvent({ source: "backend", level: "INFO", type: "rag.query" })
    const results = await queryEvents({ type: "auth.login" })
    expect(results.every((e) => e.type === "auth.login")).toBe(true)
  })

  test("filtra por level", async () => {
    await writeEvent({ source: "backend", level: "ERROR", type: "system.error" })
    await writeEvent({ source: "backend", level: "INFO", type: "other" })
    const errors = await queryEvents({ level: "ERROR" })
    expect(errors.every((e) => e.level === "ERROR")).toBe(true)
  })

  test("filtra por source", async () => {
    await writeEvent({ source: "frontend", level: "INFO", type: "ui.click" })
    await writeEvent({ source: "backend", level: "INFO", type: "api.call" })
    const frontend = await queryEvents({ source: "frontend" })
    expect(frontend.every((e) => e.source === "frontend")).toBe(true)
  })

  test("filtra por userId", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    await writeEvent({ source: "backend", level: "INFO", type: "auth.login", userId: u1.id })
    await writeEvent({ source: "backend", level: "INFO", type: "auth.login", userId: u2.id })
    const results = await queryEvents({ userId: u1.id })
    expect(results.every((e) => e.userId === u1.id)).toBe(true)
  })

  test("respeta el límite de resultados", async () => {
    for (let i = 0; i < 5; i++) {
      await writeEvent({ source: "backend", level: "INFO", type: "ev" })
    }
    const results = await queryEvents({ limit: 3 })
    expect(results.length).toBeLessThanOrEqual(3)
  })
})

describe("getEventsForReplay", () => {
  test("retorna eventos desde el timestamp indicado", async () => {
    const before = Date.now()
    await writeEvent({ source: "backend", level: "INFO", type: "replay.test" })
    const events = await getEventsForReplay(before)
    expect(events.some((e) => e.type === "replay.test")).toBe(true)
  })
})
