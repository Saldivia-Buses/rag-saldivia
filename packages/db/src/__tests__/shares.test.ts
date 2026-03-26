/**
 * Tests de queries de compartición de sesiones contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/shares.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { eq, and, gt } from "drizzle-orm"
import { randomBytes } from "crypto"
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
    CREATE TABLE IF NOT EXISTS chat_messages (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
      role TEXT NOT NULL,
      content TEXT NOT NULL,
      sources TEXT,
      timestamp INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS session_shares (
      id TEXT PRIMARY KEY,
      session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      token TEXT NOT NULL UNIQUE,
      expires_at INTEGER NOT NULL,
      created_at INTEGER NOT NULL
    );
    CREATE UNIQUE INDEX IF NOT EXISTS idx_session_shares_token ON session_shares(token);
  `)
})

afterEach(async () => {
  await client.executeMultiple(
    "DELETE FROM session_shares; DELETE FROM chat_messages; DELETE FROM chat_sessions; DELETE FROM users;"
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

async function createSession(userId: number, id = "sess-1") {
  const now = Date.now()
  const [session] = await testDb
    .insert(schema.chatSessions)
    .values({ id, userId, title: "Test", collection: "col", crossdoc: false, createdAt: now, updatedAt: now })
    .returning()
  return session!
}

async function createShare(sessionId: string, userId: number, ttlDays = 7) {
  const token = randomBytes(32).toString("hex")
  const now = Date.now()
  const expiresAt = now + ttlDays * 24 * 60 * 60 * 1000
  const id = randomBytes(16).toString("hex")
  const [row] = await testDb
    .insert(schema.sessionShares)
    .values({ id, sessionId, userId, token, expiresAt, createdAt: now })
    .returning()
  return row!
}

async function getShareByToken(token: string) {
  const now = Date.now()
  const rows = await testDb
    .select()
    .from(schema.sessionShares)
    .where(and(eq(schema.sessionShares.token, token), gt(schema.sessionShares.expiresAt, now)))
    .limit(1)
  return rows[0] ?? null
}

async function revokeShare(id: string, userId: number) {
  await testDb.delete(schema.sessionShares).where(
    and(eq(schema.sessionShares.id, id), eq(schema.sessionShares.userId, userId))
  )
}

async function listSharesByUser(userId: number) {
  return testDb.select().from(schema.sessionShares).where(eq(schema.sessionShares.userId, userId))
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("createShare", () => {
  test("crea un share con token único y expiresAt en el futuro", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-s1")
    const share = await createShare("sess-s1", user.id)

    expect(share.token).toHaveLength(64) // 32 bytes → 64 hex chars
    expect(share.expiresAt).toBeGreaterThan(Date.now())
    expect(share.sessionId).toBe("sess-s1")
    expect(share.userId).toBe(user.id)
  })

  test("tokens de shares distintos son únicos", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-s2")
    const s1 = await createShare("sess-s2", user.id)
    const s2 = await createShare("sess-s2", user.id)

    expect(s1.token).not.toBe(s2.token)
  })

  test("acepta TTL personalizado", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-s3")
    const share = await createShare("sess-s3", user.id, 30) // 30 días

    const thirtyDaysMs = 30 * 24 * 60 * 60 * 1000
    expect(share.expiresAt).toBeGreaterThan(Date.now() + thirtyDaysMs - 5000)
  })
})

describe("getShareByToken", () => {
  test("retorna el share si el token es válido y no expiró", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-s4")
    const share = await createShare("sess-s4", user.id)

    const found = await getShareByToken(share.token)
    expect(found).not.toBeNull()
    expect(found!.id).toBe(share.id)
  })

  test("retorna null para token inexistente", async () => {
    const found = await getShareByToken("token-inexistente")
    expect(found).toBeNull()
  })

  test("retorna null para token expirado", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-s5")

    // Insertar share expirado directamente
    const token = randomBytes(32).toString("hex")
    await testDb.insert(schema.sessionShares).values({
      id: randomBytes(16).toString("hex"),
      sessionId: "sess-s5",
      userId: user.id,
      token,
      expiresAt: Date.now() - 1000, // ya expiró
      createdAt: Date.now() - 10000,
    })

    const found = await getShareByToken(token)
    expect(found).toBeNull()
  })
})

describe("revokeShare", () => {
  test("elimina el share y el token ya no es válido", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-s6")
    const share = await createShare("sess-s6", user.id)

    await revokeShare(share.id, user.id)

    const found = await getShareByToken(share.token)
    expect(found).toBeNull()
  })

  test("no elimina shares de otro usuario", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createSession(u1.id, "sess-prot")
    const share = await createShare("sess-prot", u1.id)

    await revokeShare(share.id, u2.id) // intento de otro usuario

    const found = await getShareByToken(share.token)
    expect(found).not.toBeNull() // sigue válido
  })
})

describe("listSharesByUser", () => {
  test("retorna solo los shares del usuario especificado", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createSession(u1.id, "sess-list1")
    await createSession(u2.id, "sess-list2")
    await createShare("sess-list1", u1.id)
    await createShare("sess-list2", u2.id)

    const shares = await listSharesByUser(u1.id)
    expect(shares).toHaveLength(1)
    expect(shares[0]!.userId).toBe(u1.id)
  })
})
