/**
 * Tests de queries de etiquetas de sesión contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/tags.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { eq, inArray } from "drizzle-orm"
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
    CREATE TABLE IF NOT EXISTS chat_sessions (
      id TEXT PRIMARY KEY,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      title TEXT NOT NULL,
      collection TEXT NOT NULL,
      crossdoc INTEGER NOT NULL DEFAULT 0,
      forked_from TEXT,
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS session_tags (
      session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
      tag TEXT NOT NULL,
      PRIMARY KEY (session_id, tag)
    );
  `)
})

afterEach(async () => {
  await client.executeMultiple("DELETE FROM session_tags; DELETE FROM chat_sessions; DELETE FROM users;")
})

// ── Helpers ──────────────────────────────────────────────────────────────────

async function createUser(email = "user@test.com") {
  const [user] = await testDb
    .insert(schema.users)
    .values({ email, name: "Test", role: "user", apiKeyHash: "hash", preferences: {}, active: true, createdAt: Date.now() })
    .returning()
  return user!
}

async function createSession(userId: number, id: string) {
  const now = Date.now()
  const [session] = await testDb
    .insert(schema.chatSessions)
    .values({ id, userId, title: "Test", collection: "col", crossdoc: false, createdAt: now, updatedAt: now })
    .returning()
  return session!
}

async function addTag(sessionId: string, tag: string) {
  await testDb
    .insert(schema.sessionTags)
    .values({ sessionId, tag: tag.toLowerCase().trim() })
    .onConflictDoNothing()
}

async function removeTag(sessionId: string, tag: string) {
  await testDb.delete(schema.sessionTags).where(eq(schema.sessionTags.sessionId, sessionId))
}

async function listTagsBySession(sessionId: string) {
  const rows = await testDb
    .select({ tag: schema.sessionTags.tag })
    .from(schema.sessionTags)
    .where(eq(schema.sessionTags.sessionId, sessionId))
  return rows.map((r) => r.tag)
}

async function listTagsByUser(userId: number): Promise<string[]> {
  const sessions = await testDb
    .select({ id: schema.chatSessions.id })
    .from(schema.chatSessions)
    .where(eq(schema.chatSessions.userId, userId))
  if (sessions.length === 0) return []
  const sessionIds = sessions.map((s) => s.id)
  const rows = await testDb
    .selectDistinct({ tag: schema.sessionTags.tag })
    .from(schema.sessionTags)
    .where(inArray(schema.sessionTags.sessionId, sessionIds))
  return rows.map((r) => r.tag)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("addTag", () => {
  test("agrega un tag a una sesión", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-1")
    await addTag("sess-1", "importante")

    const tags = await listTagsBySession("sess-1")
    expect(tags).toContain("importante")
  })

  test("normaliza a minúsculas y trimea espacios", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-2")
    await addTag("sess-2", "  URGENTE  ")

    const tags = await listTagsBySession("sess-2")
    expect(tags).toContain("urgente")
  })

  test("es idempotente — no duplica si se llama dos veces", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-3")
    await addTag("sess-3", "trabajo")
    await addTag("sess-3", "trabajo")

    const tags = await listTagsBySession("sess-3")
    expect(tags.filter((t) => t === "trabajo")).toHaveLength(1)
  })

  test("agrega múltiples tags diferentes", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-4")
    await addTag("sess-4", "urgente")
    await addTag("sess-4", "trabajo")

    const tags = await listTagsBySession("sess-4")
    expect(tags).toHaveLength(2)
  })
})

describe("listTagsBySession", () => {
  test("retorna vacío si la sesión no tiene tags", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-empty")
    const tags = await listTagsBySession("sess-empty")
    expect(tags).toHaveLength(0)
  })
})

describe("removeTag", () => {
  test("elimina tags de la sesión", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-rm")
    await addTag("sess-rm", "borrar")
    await addTag("sess-rm", "mantener")

    await removeTag("sess-rm", "borrar")

    const tags = await listTagsBySession("sess-rm")
    expect(tags).not.toContain("borrar")
  })
})

describe("listTagsByUser", () => {
  test("retorna vacío si el usuario no tiene sesiones", async () => {
    const user = await createUser()
    const tags = await listTagsByUser(user.id)
    expect(tags).toHaveLength(0)
  })

  test("retorna tags únicos de todas las sesiones del usuario", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-a")
    await createSession(user.id, "sess-b")
    await addTag("sess-a", "trabajo")
    await addTag("sess-b", "trabajo") // mismo tag en otra sesión
    await addTag("sess-b", "urgente")

    const tags = await listTagsByUser(user.id)
    expect(tags).toHaveLength(2) // distinct
    expect(tags).toContain("trabajo")
    expect(tags).toContain("urgente")
  })

  test("no retorna tags de sesiones de otro usuario", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createSession(u1.id, "sess-u1")
    await createSession(u2.id, "sess-u2")
    await addTag("sess-u1", "privado")
    await addTag("sess-u2", "ajeno")

    const tags = await listTagsByUser(u1.id)
    expect(tags).toContain("privado")
    expect(tags).not.toContain("ajeno")
  })
})
