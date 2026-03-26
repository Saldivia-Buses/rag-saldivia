/**
 * Tests de queries de sesiones y mensajes de chat contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/sessions.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { eq, and, desc, asc } from "drizzle-orm"
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
    CREATE TABLE IF NOT EXISTS message_feedback (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      message_id INTEGER NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
      user_id INTEGER NOT NULL REFERENCES users(id),
      rating TEXT NOT NULL,
      created_at INTEGER NOT NULL,
      UNIQUE(message_id, user_id)
    );
  `)
})

afterEach(async () => {
  await client.executeMultiple(
    "DELETE FROM message_feedback; DELETE FROM chat_messages; DELETE FROM chat_sessions; DELETE FROM users;"
  )
})

// ── Helpers ──────────────────────────────────────────────────────────────────

async function createUser(email = "user@test.com") {
  const [user] = await testDb
    .insert(schema.users)
    .values({ email, name: "Test", role: "user", apiKeyHash: "hash", preferences: {}, active: true, createdAt: Date.now() })
    .returning()
  if (!user) throw new Error("Failed to create user")
  return user
}

async function createSession(userId: number, id = crypto.randomUUID(), title = "Test session") {
  const now = Date.now()
  const [session] = await testDb
    .insert(schema.chatSessions)
    .values({ id, userId, title, collection: "test-col", crossdoc: false, createdAt: now, updatedAt: now })
    .returning()
  if (!session) throw new Error("Failed to create session")
  return session
}

async function addMessage(sessionId: string, role: "user" | "assistant" = "user", content = "Hello") {
  const [msg] = await testDb
    .insert(schema.chatMessages)
    .values({ sessionId, role, content, timestamp: Date.now() })
    .returning()
  if (!msg) throw new Error("Failed to create message")
  return msg
}

async function listSessions(userId: number) {
  return testDb.query.chatSessions.findMany({
    where: (s, { eq }) => eq(s.userId, userId),
    orderBy: (s, { desc }) => [desc(s.updatedAt)],
  })
}

async function getSession(id: string, userId?: number) {
  return testDb.query.chatSessions.findFirst({
    where: (s, { and, eq }) =>
      userId ? and(eq(s.id, id), eq(s.userId, userId)) : eq(s.id, id),
    with: { messages: { orderBy: (m, { asc }) => [asc(m.timestamp)] } },
  })
}

async function updateTitle(id: string, userId: number, title: string) {
  const [updated] = await testDb
    .update(schema.chatSessions)
    .set({ title, updatedAt: Date.now() })
    .where(and(eq(schema.chatSessions.id, id), eq(schema.chatSessions.userId, userId)))
    .returning()
  return updated
}

async function deleteSession(id: string, userId: number) {
  await testDb.delete(schema.chatSessions).where(
    and(eq(schema.chatSessions.id, id), eq(schema.chatSessions.userId, userId))
  )
}

async function addFeedback(messageId: number, userId: number, rating: "up" | "down") {
  await testDb
    .insert(schema.messageFeedback)
    .values({ messageId, userId, rating, createdAt: Date.now() })
    .onConflictDoUpdate({
      target: [schema.messageFeedback.messageId, schema.messageFeedback.userId],
      set: { rating },
    })
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("createSession", () => {
  test("crea una sesión con defaults correctos", async () => {
    const user = await createUser()
    const session = await createSession(user.id, "sess-1", "Mi sesión")

    expect(session.id).toBe("sess-1")
    expect(session.userId).toBe(user.id)
    expect(session.title).toBe("Mi sesión")
    expect(session.collection).toBe("test-col")
    expect(session.crossdoc).toBe(false)
    expect(session.createdAt).toBeGreaterThan(0)
    expect(session.updatedAt).toBeGreaterThan(0)
  })

  test("crea sesiones independientes para usuarios distintos", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createSession(u1.id, "sess-a")
    await createSession(u2.id, "sess-b")

    const sessions1 = await listSessions(u1.id)
    const sessions2 = await listSessions(u2.id)
    expect(sessions1).toHaveLength(1)
    expect(sessions2).toHaveLength(1)
    expect(sessions1[0]!.id).toBe("sess-a")
    expect(sessions2[0]!.id).toBe("sess-b")
  })
})

describe("listSessionsByUser", () => {
  test("retorna vacío si el usuario no tiene sesiones", async () => {
    const user = await createUser()
    const sessions = await listSessions(user.id)
    expect(sessions).toHaveLength(0)
  })

  test("retorna sesiones ordenadas por updatedAt descendente", async () => {
    const user = await createUser()
    await testDb.insert(schema.chatSessions).values({
      id: "old", userId: user.id, title: "Old", collection: "c",
      crossdoc: false, createdAt: 1000, updatedAt: 1000,
    })
    await testDb.insert(schema.chatSessions).values({
      id: "new", userId: user.id, title: "New", collection: "c",
      crossdoc: false, createdAt: 2000, updatedAt: 2000,
    })
    const sessions = await listSessions(user.id)
    expect(sessions[0]!.id).toBe("new")
    expect(sessions[1]!.id).toBe("old")
  })
})

describe("getSessionById", () => {
  test("retorna sesión con mensajes incluidos", async () => {
    const user = await createUser()
    const session = await createSession(user.id, "sess-x")
    await addMessage("sess-x", "user", "Hola")
    await addMessage("sess-x", "assistant", "Hola también")

    const found = await getSession("sess-x")
    expect(found).toBeDefined()
    expect(found!.messages).toHaveLength(2)
  })

  test("retorna undefined si el userId no coincide", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createSession(u1.id, "sess-y")

    const found = await getSession("sess-y", u2.id)
    expect(found).toBeUndefined()
  })
})

describe("updateSessionTitle", () => {
  test("actualiza el título y el updatedAt", async () => {
    const user = await createUser()
    const session = await createSession(user.id, "sess-t")
    const before = session.updatedAt

    await new Promise((r) => setTimeout(r, 5))
    const updated = await updateTitle("sess-t", user.id, "Nuevo título")

    expect(updated!.title).toBe("Nuevo título")
    expect(updated!.updatedAt).toBeGreaterThanOrEqual(before)
  })

  test("no actualiza sesiones de otro usuario", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createSession(u1.id, "sess-z", "Original")

    const result = await updateTitle("sess-z", u2.id, "Hack")
    expect(result).toBeUndefined()

    const session = await getSession("sess-z")
    expect(session!.title).toBe("Original")
  })
})

describe("deleteSession", () => {
  test("elimina la sesión y sus mensajes en cascade", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-del")
    await addMessage("sess-del")

    await deleteSession("sess-del", user.id)

    const session = await getSession("sess-del")
    expect(session).toBeUndefined()

    const msgs = await testDb
      .select()
      .from(schema.chatMessages)
      .where(eq(schema.chatMessages.sessionId, "sess-del"))
    expect(msgs).toHaveLength(0)
  })
})

describe("addMessage", () => {
  test("crea un mensaje y persiste el rol y contenido", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-m")
    const msg = await addMessage("sess-m", "assistant", "Respuesta del asistente")

    expect(msg.role).toBe("assistant")
    expect(msg.content).toBe("Respuesta del asistente")
    expect(msg.sessionId).toBe("sess-m")
    expect(msg.timestamp).toBeGreaterThan(0)
  })
})

describe("addFeedback", () => {
  test("agrega feedback a un mensaje", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-f")
    const msg = await addMessage("sess-f")
    await addFeedback(msg.id, user.id, "up")

    const fb = await testDb
      .select()
      .from(schema.messageFeedback)
      .where(eq(schema.messageFeedback.messageId, msg.id))
    expect(fb[0]!.rating).toBe("up")
  })

  test("actualiza el rating si ya existe (upsert)", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-fb2")
    const msg = await addMessage("sess-fb2")
    await addFeedback(msg.id, user.id, "up")
    await addFeedback(msg.id, user.id, "down")

    const fb = await testDb
      .select()
      .from(schema.messageFeedback)
      .where(eq(schema.messageFeedback.messageId, msg.id))
    expect(fb).toHaveLength(1)
    expect(fb[0]!.rating).toBe("down")
  })
})
