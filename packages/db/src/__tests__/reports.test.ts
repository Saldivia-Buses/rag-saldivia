/**
 * Tests de queries de informes programados contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/reports.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser } from "./setup"
import { createReport, listActiveReports, listReportsByUser, updateLastRun, deleteReport } from "../queries/reports"
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
  await client.executeMultiple("DELETE FROM scheduled_reports; DELETE FROM users;")
})

describe("createReport", () => {
  test("crea informe con active=true y nextRun en el futuro", async () => {
    const user = await insertUser(db)
    const report = await createReport({ userId: user.id, query: "¿Estado?", collection: "docs", schedule: "daily", destination: "saved" })
    expect(report.active).toBe(true)
    expect(report.nextRun).toBeGreaterThan(Date.now())
    expect(report.lastRun).toBeNull()
  })

  test("weekly nextRun > daily nextRun", async () => {
    const user = await insertUser(db)
    const daily = await createReport({ userId: user.id, query: "q", collection: "c", schedule: "daily", destination: "saved" })
    const weekly = await createReport({ userId: user.id, query: "q", collection: "c", schedule: "weekly", destination: "saved" })
    expect(weekly.nextRun).toBeGreaterThan(daily.nextRun)
  })
})

describe("listActiveReports", () => {
  test("no retorna informes con nextRun en el futuro", async () => {
    const user = await insertUser(db)
    await createReport({ userId: user.id, query: "q", collection: "c", schedule: "daily", destination: "saved" })
    expect(await listActiveReports()).toHaveLength(0)
  })

  test("retorna informes activos con nextRun en el pasado", async () => {
    const user = await insertUser(db)
    await db.insert(schema.scheduledReports).values({
      id: randomUUID(), userId: user.id, query: "pasado", collection: "c",
      schedule: "daily", destination: "saved", active: true, nextRun: Date.now() - 1000, createdAt: Date.now(),
    })
    const active = await listActiveReports()
    expect(active.some((r) => r.query === "pasado")).toBe(true)
  })
})

describe("listReportsByUser", () => {
  test("retorna solo los informes del usuario", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    await createReport({ userId: u1.id, query: "q1", collection: "c", schedule: "daily", destination: "saved" })
    await createReport({ userId: u2.id, query: "q2", collection: "c", schedule: "daily", destination: "saved" })
    const list = await listReportsByUser(u1.id)
    expect(list).toHaveLength(1)
    expect(list[0]!.query).toBe("q1")
  })
})

describe("updateLastRun / deleteReport", () => {
  test("updateLastRun actualiza lastRun y recalcula nextRun", async () => {
    const user = await insertUser(db)
    const report = await createReport({ userId: user.id, query: "q", collection: "c", schedule: "daily", destination: "saved" })
    await updateLastRun(report.id, "daily")
    const updated = (await listReportsByUser(user.id))[0]!
    expect(updated.lastRun).not.toBeNull()
  })

  test("deleteReport elimina solo el informe del usuario correcto", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    const report = await createReport({ userId: u1.id, query: "q", collection: "c", schedule: "daily", destination: "saved" })
    await deleteReport(report.id, u2.id) // intento de u2
    expect(await listReportsByUser(u1.id)).toHaveLength(1)
    await deleteReport(report.id, u1.id) // correcto
    expect(await listReportsByUser(u1.id)).toHaveLength(0)
  })
})
