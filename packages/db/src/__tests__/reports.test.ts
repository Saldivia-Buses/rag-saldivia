/**
 * Tests de queries de informes programados contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/reports.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { eq, and, lte } from "drizzle-orm"
import { randomUUID } from "crypto"
import * as schema from "../schema"

process.env["DATABASE_PATH"] = ":memory:"

const client = createClient({ url: ":memory:" })
const testDb = drizzle(client, { schema })

beforeAll(async () => {
  await client.executeMultiple(`
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      email TEXT NOT NULL UNIQUE,
      name TEXT NOT NULL,
      role TEXT NOT NULL DEFAULT 'user',
      api_key_hash TEXT NOT NULL,
      password_hash TEXT,
      preferences TEXT NOT NULL DEFAULT '{}',
      active INTEGER NOT NULL DEFAULT 1,
      onboarding_completed INTEGER NOT NULL DEFAULT 0,
      sso_provider TEXT,
      sso_subject TEXT,
      created_at INTEGER NOT NULL,
      last_login INTEGER
    );
    CREATE TABLE IF NOT EXISTS scheduled_reports (
      id TEXT PRIMARY KEY,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      query TEXT NOT NULL,
      collection TEXT NOT NULL,
      schedule TEXT NOT NULL,
      destination TEXT NOT NULL,
      email TEXT,
      active INTEGER NOT NULL DEFAULT 1,
      last_run INTEGER,
      next_run INTEGER NOT NULL,
      created_at INTEGER NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_reports_active_next_run ON scheduled_reports(active, next_run);
  `)
})

afterEach(async () => {
  await client.executeMultiple("DELETE FROM scheduled_reports; DELETE FROM users;")
})

// ── Helpers ──────────────────────────────────────────────────────────────────

function calcNextRun(schedule: "daily" | "weekly" | "monthly"): number {
  const now = Date.now()
  switch (schedule) {
    case "daily": return now + 24 * 60 * 60 * 1000
    case "weekly": return now + 7 * 24 * 60 * 60 * 1000
    case "monthly": return now + 30 * 24 * 60 * 60 * 1000
  }
}

async function createUser(email = "user@test.com") {
  const [user] = await testDb
    .insert(schema.users)
    .values({ email, name: "Test", role: "user", apiKeyHash: "hash", preferences: {}, active: true, createdAt: Date.now() })
    .returning()
  return user!
}

async function createReport(userId: number, data: {
  query: string
  collection: string
  schedule: "daily" | "weekly" | "monthly"
  destination: "saved" | "email"
  email?: string
}) {
  const [row] = await testDb
    .insert(schema.scheduledReports)
    .values({
      id: randomUUID(),
      userId,
      ...data,
      active: true,
      nextRun: calcNextRun(data.schedule),
      createdAt: Date.now(),
    })
    .returning()
  return row!
}

async function listActiveReports() {
  const now = Date.now()
  return testDb
    .select()
    .from(schema.scheduledReports)
    .where(and(eq(schema.scheduledReports.active, true), lte(schema.scheduledReports.nextRun, now)))
}

async function listReportsByUser(userId: number) {
  return testDb.select().from(schema.scheduledReports).where(eq(schema.scheduledReports.userId, userId))
}

async function updateLastRun(id: string, schedule: "daily" | "weekly" | "monthly") {
  await testDb
    .update(schema.scheduledReports)
    .set({ lastRun: Date.now(), nextRun: calcNextRun(schedule) })
    .where(eq(schema.scheduledReports.id, id))
}

async function deleteReport(id: string, userId: number) {
  await testDb.delete(schema.scheduledReports).where(
    and(eq(schema.scheduledReports.id, id), eq(schema.scheduledReports.userId, userId))
  )
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("createReport", () => {
  test("crea un informe con active=true y nextRun en el futuro", async () => {
    const user = await createUser()
    const report = await createReport(user.id, {
      query: "¿Cuál es el estado del proyecto?",
      collection: "proyectos",
      schedule: "daily",
      destination: "saved",
    })

    expect(report.query).toBe("¿Cuál es el estado del proyecto?")
    expect(report.active).toBe(true)
    expect(report.nextRun).toBeGreaterThan(Date.now())
    expect(report.lastRun).toBeNull()
  })

  test("calcula nextRun según el schedule", async () => {
    const user = await createUser()
    const daily = await createReport(user.id, { query: "q", collection: "c", schedule: "daily", destination: "saved" })
    const weekly = await createReport(user.id, { query: "q", collection: "c", schedule: "weekly", destination: "saved" })

    expect(weekly.nextRun).toBeGreaterThan(daily.nextRun)
  })
})

describe("listActiveReports", () => {
  test("no retorna informes cuyo nextRun es en el futuro", async () => {
    const user = await createUser()
    await createReport(user.id, { query: "q", collection: "c", schedule: "daily", destination: "saved" })

    const active = await listActiveReports()
    expect(active).toHaveLength(0) // nextRun está en el futuro
  })

  test("retorna informes activos con nextRun en el pasado", async () => {
    const user = await createUser()
    // Insertar directamente con nextRun en el pasado
    await testDb.insert(schema.scheduledReports).values({
      id: randomUUID(),
      userId: user.id,
      query: "q pasado",
      collection: "c",
      schedule: "daily",
      destination: "saved",
      active: true,
      nextRun: Date.now() - 1000,
      createdAt: Date.now() - 10000,
    })

    const active = await listActiveReports()
    expect(active).toHaveLength(1)
    expect(active[0]!.query).toBe("q pasado")
  })
})

describe("listReportsByUser", () => {
  test("retorna solo informes del usuario especificado", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createReport(u1.id, { query: "q1", collection: "c", schedule: "daily", destination: "saved" })
    await createReport(u2.id, { query: "q2", collection: "c", schedule: "weekly", destination: "saved" })

    const reports = await listReportsByUser(u1.id)
    expect(reports).toHaveLength(1)
    expect(reports[0]!.query).toBe("q1")
  })
})

describe("updateLastRun", () => {
  test("actualiza lastRun y recalcula nextRun", async () => {
    const user = await createUser()
    const report = await createReport(user.id, { query: "q", collection: "c", schedule: "daily", destination: "saved" })
    const beforeNextRun = report.nextRun

    await updateLastRun(report.id, "daily")

    const updated = await testDb.select().from(schema.scheduledReports).where(eq(schema.scheduledReports.id, report.id))
    expect(updated[0]!.lastRun).not.toBeNull()
    expect(updated[0]!.nextRun).toBeGreaterThanOrEqual(beforeNextRun)
  })
})

describe("deleteReport", () => {
  test("elimina el informe del usuario correcto", async () => {
    const user = await createUser()
    const report = await createReport(user.id, { query: "q", collection: "c", schedule: "daily", destination: "saved" })

    await deleteReport(report.id, user.id)

    const remaining = await listReportsByUser(user.id)
    expect(remaining).toHaveLength(0)
  })

  test("no elimina informes de otro usuario", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    const report = await createReport(u1.id, { query: "q", collection: "c", schedule: "daily", destination: "saved" })

    await deleteReport(report.id, u2.id) // intento de otro usuario

    const remaining = await listReportsByUser(u1.id)
    expect(remaining).toHaveLength(1)
  })
})
