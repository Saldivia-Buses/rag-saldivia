/**
 * Tests de queries de eventos (black box) contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/events.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { eq, and, gte, desc, asc } from "drizzle-orm"
import * as schema from "../schema"

process.env["DATABASE_PATH"] = ":memory:"

const client = createClient({ url: ":memory:" })
const testDb = drizzle(client, { schema })

let _seq = 0

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
    CREATE INDEX IF NOT EXISTS idx_events_ts ON events(ts);
    CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
    CREATE INDEX IF NOT EXISTS idx_events_level ON events(level);
    CREATE INDEX IF NOT EXISTS idx_events_sequence ON events(sequence);
  `)
})

afterEach(async () => {
  await client.executeMultiple("DELETE FROM events; DELETE FROM users;")
  _seq = 0
})

// ── Helpers ──────────────────────────────────────────────────────────────────

async function createUser(email = "user@test.com") {
  const [user] = await testDb
    .insert(schema.users)
    .values({ email, name: "Test", role: "user", apiKeyHash: "hash", preferences: {}, active: true, createdAt: Date.now() })
    .returning()
  return user!
}

async function writeEvent(data: {
  source?: "frontend" | "backend"
  level?: "INFO" | "WARN" | "ERROR" | "DEBUG" | "TRACE" | "FATAL"
  type?: string
  userId?: number | null
  payload?: Record<string, unknown>
  ts?: number
}) {
  const [event] = await testDb
    .insert(schema.events)
    .values({
      id: crypto.randomUUID(),
      ts: data.ts ?? Date.now(),
      source: data.source ?? "backend",
      level: data.level ?? "INFO",
      type: data.type ?? "system.test",
      userId: data.userId ?? null,
      sessionId: null,
      payload: data.payload ?? {},
      sequence: ++_seq,
    })
    .returning()
  return event!
}

async function queryEvents(filters: {
  source?: "frontend" | "backend"
  level?: "INFO" | "WARN" | "ERROR" | "DEBUG" | "TRACE" | "FATAL"
  type?: string
  userId?: number
  fromTs?: number
  limit?: number
  order?: "asc" | "desc"
}) {
  const conditions = []
  if (filters.source) conditions.push(eq(schema.events.source, filters.source))
  if (filters.level) conditions.push(eq(schema.events.level, filters.level))
  if (filters.type) conditions.push(eq(schema.events.type, filters.type))
  if (filters.userId) conditions.push(eq(schema.events.userId!, filters.userId))
  if (filters.fromTs) conditions.push(gte(schema.events.ts, filters.fromTs))

  return testDb.query.events.findMany({
    where: conditions.length > 0 ? and(...conditions) : undefined,
    orderBy: filters.order === "asc"
      ? (e, { asc }) => [asc(e.sequence)]
      : (e, { desc }) => [desc(e.sequence)],
    limit: filters.limit ?? 100,
  })
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("writeEvent", () => {
  test("persiste un evento con todos los campos", async () => {
    const ev = await writeEvent({
      source: "backend",
      level: "INFO",
      type: "auth.login",
      payload: { email: "test@example.com" },
    })

    expect(ev.source).toBe("backend")
    expect(ev.level).toBe("INFO")
    expect(ev.type).toBe("auth.login")
    expect(ev.payload).toMatchObject({ email: "test@example.com" })
    expect(ev.ts).toBeGreaterThan(0)
    expect(ev.id).toHaveLength(36) // UUID format
  })

  test("asigna sequence monotónicamente creciente", async () => {
    const e1 = await writeEvent({ type: "ev.one" })
    const e2 = await writeEvent({ type: "ev.two" })
    const e3 = await writeEvent({ type: "ev.three" })

    expect(e2.sequence).toBeGreaterThan(e1.sequence)
    expect(e3.sequence).toBeGreaterThan(e2.sequence)
  })

  test("permite userId null para eventos de sistema", async () => {
    const ev = await writeEvent({ type: "system.startup", userId: null })
    expect(ev.userId).toBeNull()
  })

  test("asocia userId cuando se provee", async () => {
    const user = await createUser()
    const ev = await writeEvent({ type: "auth.login", userId: user.id })
    expect(ev.userId).toBe(user.id)
  })
})

describe("queryEvents", () => {
  test("retorna todos los eventos sin filtros", async () => {
    await writeEvent({ type: "a" })
    await writeEvent({ type: "b" })
    await writeEvent({ type: "c" })

    const results = await queryEvents({})
    expect(results.length).toBe(3)
  })

  test("filtra por type", async () => {
    await writeEvent({ type: "auth.login" })
    await writeEvent({ type: "rag.query" })
    await writeEvent({ type: "auth.login" })

    const results = await queryEvents({ type: "auth.login" })
    expect(results.length).toBe(2)
    expect(results.every((e) => e.type === "auth.login")).toBe(true)
  })

  test("filtra por level", async () => {
    await writeEvent({ level: "ERROR", type: "system.error" })
    await writeEvent({ level: "INFO", type: "auth.login" })
    await writeEvent({ level: "ERROR", type: "db.error" })

    const results = await queryEvents({ level: "ERROR" })
    expect(results.length).toBe(2)
    expect(results.every((e) => e.level === "ERROR")).toBe(true)
  })

  test("filtra por source", async () => {
    await writeEvent({ source: "frontend", type: "ui.click" })
    await writeEvent({ source: "backend", type: "auth.login" })

    const frontend = await queryEvents({ source: "frontend" })
    expect(frontend.length).toBe(1)
    expect(frontend[0]!.type).toBe("ui.click")
  })

  test("filtra por userId", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await writeEvent({ userId: u1.id, type: "auth.login" })
    await writeEvent({ userId: u2.id, type: "auth.login" })
    await writeEvent({ userId: u1.id, type: "rag.query" })

    const results = await queryEvents({ userId: u1.id })
    expect(results.length).toBe(2)
    expect(results.every((e) => e.userId === u1.id)).toBe(true)
  })

  test("ordena por sequence ascendente cuando se pide", async () => {
    await writeEvent({ type: "first" })
    await writeEvent({ type: "second" })
    await writeEvent({ type: "third" })

    const results = await queryEvents({ order: "asc" })
    expect(results[0]!.type).toBe("first")
    expect(results[2]!.type).toBe("third")
  })

  test("respeta el límite de resultados", async () => {
    for (let i = 0; i < 5; i++) await writeEvent({ type: "ev" })
    const results = await queryEvents({ limit: 3 })
    expect(results.length).toBe(3)
  })
})
