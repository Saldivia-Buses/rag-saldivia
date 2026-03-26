/**
 * Tests de queries de rate limits contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/rate-limits.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { eq, and, gte, sql } from "drizzle-orm"
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
    CREATE TABLE IF NOT EXISTS areas (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL UNIQUE,
      description TEXT NOT NULL DEFAULT '',
      created_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS user_areas (
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      area_id INTEGER NOT NULL REFERENCES areas(id) ON DELETE CASCADE,
      PRIMARY KEY (user_id, area_id)
    );
    CREATE TABLE IF NOT EXISTS events (
      id TEXT PRIMARY KEY,
      ts INTEGER NOT NULL,
      source TEXT NOT NULL,
      level TEXT NOT NULL,
      type TEXT NOT NULL,
      user_id INTEGER REFERENCES users(id),
      session_id TEXT,
      payload TEXT NOT NULL DEFAULT '{}',
      sequence INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS rate_limits (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      target_type TEXT NOT NULL,
      target_id INTEGER NOT NULL,
      max_queries_per_hour INTEGER NOT NULL,
      active INTEGER NOT NULL DEFAULT 1,
      created_at INTEGER NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_rate_limits_target ON rate_limits(target_type, target_id);
  `)
})

afterEach(async () => {
  await client.executeMultiple(
    "DELETE FROM rate_limits; DELETE FROM events; DELETE FROM user_areas; DELETE FROM areas; DELETE FROM users;"
  )
})

// ── Helpers ──────────────────────────────────────────────────────────────────

async function createUser(email = "user@test.com") {
  const [user] = await testDb
    .insert(schema.users)
    .values({ email, name: "Test", role: "user", apiKeyHash: "hash", preferences: {}, active: true, createdAt: Date.now() })
    .returning()
  return user!
}

async function createArea(name: string) {
  const [area] = await testDb
    .insert(schema.areas)
    .values({ name, description: "", createdAt: Date.now() })
    .returning()
  return area!
}

async function assignUserToArea(userId: number, areaId: number) {
  await testDb.insert(schema.userAreas).values({ userId, areaId })
}

async function createRateLimit(data: { targetType: "user" | "area"; targetId: number; maxQueriesPerHour: number }) {
  const [row] = await testDb
    .insert(schema.rateLimits)
    .values({ ...data, active: true, createdAt: Date.now() })
    .returning()
  return row!
}

async function getRateLimit(userId: number): Promise<number | null> {
  // Primero nivel usuario
  const userLimit = await testDb
    .select({ max: schema.rateLimits.maxQueriesPerHour })
    .from(schema.rateLimits)
    .where(and(eq(schema.rateLimits.targetType, "user"), eq(schema.rateLimits.targetId, userId), eq(schema.rateLimits.active, true)))
    .limit(1)
  if (userLimit[0]) return userLimit[0].max

  // Luego nivel área
  const areas = await testDb
    .select({ areaId: schema.userAreas.areaId })
    .from(schema.userAreas)
    .where(eq(schema.userAreas.userId, userId))
  if (areas.length === 0) return null

  for (const { areaId } of areas) {
    const areaLimit = await testDb
      .select({ max: schema.rateLimits.maxQueriesPerHour })
      .from(schema.rateLimits)
      .where(and(eq(schema.rateLimits.targetType, "area"), eq(schema.rateLimits.targetId, areaId), eq(schema.rateLimits.active, true)))
      .limit(1)
    if (areaLimit[0]) return areaLimit[0].max
  }
  return null
}

async function countQueriesLastHour(userId: number): Promise<number> {
  const oneHourAgo = Date.now() - 60 * 60 * 1000
  const rows = await testDb
    .select({ count: sql<number>`count(*)` })
    .from(schema.events)
    .where(and(eq(schema.events.type, "rag.stream_started"), eq(schema.events.userId!, userId), gte(schema.events.ts, oneHourAgo)))
  return rows[0]?.count ?? 0
}

async function listRateLimits() {
  return testDb.select().from(schema.rateLimits)
}

async function deleteRateLimit(id: number) {
  await testDb.delete(schema.rateLimits).where(eq(schema.rateLimits.id, id))
}

let _seq = 0
async function insertEvent(userId: number, type = "rag.stream_started", tsOffset = 0) {
  await testDb.insert(schema.events).values({
    id: crypto.randomUUID(),
    ts: Date.now() + tsOffset,
    source: "backend",
    level: "INFO",
    type,
    userId,
    payload: {},
    sequence: ++_seq,
  })
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("createRateLimit", () => {
  test("crea un límite a nivel usuario", async () => {
    const user = await createUser()
    const limit = await createRateLimit({ targetType: "user", targetId: user.id, maxQueriesPerHour: 100 })

    expect(limit.targetType).toBe("user")
    expect(limit.targetId).toBe(user.id)
    expect(limit.maxQueriesPerHour).toBe(100)
    expect(limit.active).toBe(true)
  })

  test("crea un límite a nivel área", async () => {
    const area = await createArea("Marketing")
    const limit = await createRateLimit({ targetType: "area", targetId: area.id, maxQueriesPerHour: 50 })
    expect(limit.targetType).toBe("area")
    expect(limit.targetId).toBe(area.id)
  })
})

describe("getRateLimit", () => {
  test("retorna null si no hay límite configurado", async () => {
    const user = await createUser()
    const limit = await getRateLimit(user.id)
    expect(limit).toBeNull()
  })

  test("retorna el límite a nivel usuario", async () => {
    const user = await createUser()
    await createRateLimit({ targetType: "user", targetId: user.id, maxQueriesPerHour: 200 })

    const limit = await getRateLimit(user.id)
    expect(limit).toBe(200)
  })

  test("prioriza el límite de usuario sobre el de área", async () => {
    const user = await createUser()
    const area = await createArea("Ventas")
    await assignUserToArea(user.id, area.id)
    await createRateLimit({ targetType: "area", targetId: area.id, maxQueriesPerHour: 30 })
    await createRateLimit({ targetType: "user", targetId: user.id, maxQueriesPerHour: 500 })

    const limit = await getRateLimit(user.id)
    expect(limit).toBe(500) // user-level tiene prioridad
  })

  test("retorna límite de área si no hay límite de usuario", async () => {
    const user = await createUser()
    const area = await createArea("Soporte")
    await assignUserToArea(user.id, area.id)
    await createRateLimit({ targetType: "area", targetId: area.id, maxQueriesPerHour: 75 })

    const limit = await getRateLimit(user.id)
    expect(limit).toBe(75)
  })
})

describe("countQueriesLastHour", () => {
  test("retorna 0 si no hay queries en la última hora", async () => {
    const user = await createUser()
    const count = await countQueriesLastHour(user.id)
    expect(count).toBe(0)
  })

  test("cuenta solo eventos de tipo rag.stream_started del usuario", async () => {
    const user = await createUser()
    await insertEvent(user.id, "rag.stream_started")
    await insertEvent(user.id, "rag.stream_started")
    await insertEvent(user.id, "auth.login") // diferente tipo, no cuenta

    const count = await countQueriesLastHour(user.id)
    expect(count).toBe(2)
  })

  test("no cuenta eventos de hace más de 1 hora", async () => {
    const user = await createUser()
    const twoHoursAgo = -(2 * 60 * 60 * 1000)
    await insertEvent(user.id, "rag.stream_started", twoHoursAgo)

    const count = await countQueriesLastHour(user.id)
    expect(count).toBe(0)
  })

  test("no cuenta eventos de otros usuarios", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await insertEvent(u2.id, "rag.stream_started")

    const count = await countQueriesLastHour(u1.id)
    expect(count).toBe(0)
  })
})

describe("listRateLimits / deleteRateLimit", () => {
  test("lista todos los rate limits", async () => {
    const user = await createUser("a@test.com")
    const area = await createArea("QA")
    await createRateLimit({ targetType: "user", targetId: user.id, maxQueriesPerHour: 100 })
    await createRateLimit({ targetType: "area", targetId: area.id, maxQueriesPerHour: 50 })

    const limits = await listRateLimits()
    expect(limits).toHaveLength(2)
  })

  test("elimina un rate limit específico", async () => {
    const user = await createUser()
    const limit = await createRateLimit({ targetType: "user", targetId: user.id, maxQueriesPerHour: 100 })

    await deleteRateLimit(limit.id)

    const limits = await listRateLimits()
    expect(limits.find((l) => l.id === limit.id)).toBeUndefined()
  })
})
