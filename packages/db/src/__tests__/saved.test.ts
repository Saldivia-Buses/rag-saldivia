/**
 * Tests de queries de respuestas guardadas contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/saved.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { eq, and, desc } from "drizzle-orm"
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
    CREATE TABLE IF NOT EXISTS saved_responses (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      message_id INTEGER REFERENCES chat_messages(id) ON DELETE SET NULL,
      content TEXT NOT NULL,
      session_title TEXT,
      created_at INTEGER NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_saved_responses_user ON saved_responses(user_id);
  `)
})

afterEach(async () => {
  await client.executeMultiple(
    "DELETE FROM saved_responses; DELETE FROM chat_messages; DELETE FROM chat_sessions; DELETE FROM users;"
  )
})

// ── Helpers ──────────────────────────────────────────────────────────────────

async function createTestUser(email: string) {
  const [user] = await testDb
    .insert(schema.users)
    .values({ email, name: "Test User", role: "user", apiKeyHash: "hash", preferences: {}, active: true, createdAt: Date.now() })
    .returning()
  if (!user) throw new Error("Failed to create user")
  return user
}

async function createTestSession(userId: number, id = "test-session") {
  const [session] = await testDb
    .insert(schema.chatSessions)
    .values({ id, userId, title: "Test Session", collection: "test-col", crossdoc: false, createdAt: Date.now(), updatedAt: Date.now() })
    .returning()
  if (!session) throw new Error("Failed to create session")
  return session
}

async function createTestMessage(sessionId: string) {
  const [msg] = await testDb
    .insert(schema.chatMessages)
    .values({ sessionId, role: "assistant", content: "Test content", timestamp: Date.now() })
    .returning()
  if (!msg) throw new Error("Failed to create message")
  return msg
}

// Helpers locales que replican las queries de saved.ts usando testDb
async function save(data: { userId: number; messageId?: number; content: string; sessionTitle?: string }) {
  const [row] = await testDb
    .insert(schema.savedResponses)
    .values({ ...data, createdAt: Date.now() })
    .returning()
  return row!
}

async function unsaveById(id: number, userId: number) {
  await testDb
    .delete(schema.savedResponses)
    .where(and(eq(schema.savedResponses.id, id), eq(schema.savedResponses.userId, userId)))
}

async function unsaveByMsgId(messageId: number, userId: number) {
  await testDb
    .delete(schema.savedResponses)
    .where(and(eq(schema.savedResponses.messageId, messageId), eq(schema.savedResponses.userId, userId)))
}

async function listSaved(userId: number) {
  return testDb
    .select()
    .from(schema.savedResponses)
    .where(eq(schema.savedResponses.userId, userId))
    .orderBy(desc(schema.savedResponses.createdAt))
}

async function checkIsSaved(messageId: number, userId: number): Promise<boolean> {
  const rows = await testDb
    .select({ id: schema.savedResponses.id })
    .from(schema.savedResponses)
    .where(and(eq(schema.savedResponses.messageId, messageId), eq(schema.savedResponses.userId, userId)))
    .limit(1)
  return rows.length > 0
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("saveResponse", () => {
  test("guarda una respuesta con todos los campos", async () => {
    const user = await createTestUser("save1@test.com")
    const row = await save({ userId: user.id, content: "Mi respuesta", sessionTitle: "Mi sesión" })

    expect(row.content).toBe("Mi respuesta")
    expect(row.sessionTitle).toBe("Mi sesión")
    expect(row.userId).toBe(user.id)
    expect(row.createdAt).toBeGreaterThan(0)
  })

  test("guarda una respuesta sin messageId (null)", async () => {
    const user = await createTestUser("save2@test.com")
    const row = await save({ userId: user.id, content: "Sin mensaje asociado" })

    expect(row.messageId).toBeNil()
    expect(row.content).toBe("Sin mensaje asociado")
  })

  test("guarda una respuesta con messageId FK válido", async () => {
    const user = await createTestUser("save3@test.com")
    const session = await createTestSession(user.id, "s3")
    const msg = await createTestMessage(session.id)
    const row = await save({ userId: user.id, messageId: msg.id, content: msg.content })

    expect(row.messageId).toBe(msg.id)
  })
})

describe("listSavedResponses", () => {
  test("retorna array vacío si el usuario no tiene guardados", async () => {
    const user = await createTestUser("list1@test.com")
    const result = await listSaved(user.id)

    expect(result).toBeArray()
    expect(result.length).toBe(0)
  })

  test("retorna las respuestas del usuario ordenadas por fecha descendente", async () => {
    const user = await createTestUser("list2@test.com")
    // Insertar con timestamps distintos
    await testDb.insert(schema.savedResponses).values({ userId: user.id, content: "Primero", createdAt: 1000 })
    await testDb.insert(schema.savedResponses).values({ userId: user.id, content: "Segundo", createdAt: 2000 })
    await testDb.insert(schema.savedResponses).values({ userId: user.id, content: "Tercero", createdAt: 3000 })

    const result = await listSaved(user.id)

    expect(result.length).toBe(3)
    expect(result[0]!.content).toBe("Tercero")
    expect(result[1]!.content).toBe("Segundo")
    expect(result[2]!.content).toBe("Primero")
  })

  test("solo retorna las respuestas del usuario solicitado", async () => {
    const user1 = await createTestUser("list3a@test.com")
    const user2 = await createTestUser("list3b@test.com")
    await save({ userId: user1.id, content: "De user1" })
    await save({ userId: user2.id, content: "De user2" })

    const result = await listSaved(user1.id)

    expect(result.length).toBe(1)
    expect(result[0]!.content).toBe("De user1")
  })
})

describe("unsaveResponse (por id)", () => {
  test("elimina la respuesta del usuario correcto", async () => {
    const user = await createTestUser("unsave1@test.com")
    const row = await save({ userId: user.id, content: "A eliminar" })

    await unsaveById(row.id, user.id)
    const result = await listSaved(user.id)

    expect(result.length).toBe(0)
  })

  test("no elimina respuestas de otros usuarios con el mismo id", async () => {
    const user1 = await createTestUser("unsave2a@test.com")
    const user2 = await createTestUser("unsave2b@test.com")
    const row = await save({ userId: user1.id, content: "De user1" })

    // user2 intenta eliminar el guardado de user1
    await unsaveById(row.id, user2.id)
    const result = await listSaved(user1.id)

    expect(result.length).toBe(1)
  })
})

describe("unsaveByMessageId", () => {
  test("elimina la respuesta asociada al messageId del usuario", async () => {
    const user = await createTestUser("unsavemsg1@test.com")
    const session = await createTestSession(user.id, "sm1")
    const msg = await createTestMessage(session.id)
    await save({ userId: user.id, messageId: msg.id, content: "Guardada" })

    await unsaveByMsgId(msg.id, user.id)
    const result = await listSaved(user.id)

    expect(result.length).toBe(0)
  })

  test("no afecta respuestas de otros usuarios con el mismo messageId", async () => {
    const user1 = await createTestUser("unsavemsg2a@test.com")
    const user2 = await createTestUser("unsavemsg2b@test.com")
    const session = await createTestSession(user1.id, "sm2")
    const msg = await createTestMessage(session.id)
    await save({ userId: user1.id, messageId: msg.id, content: "De user1" })

    await unsaveByMsgId(msg.id, user2.id)
    const result = await listSaved(user1.id)

    expect(result.length).toBe(1)
  })
})

describe("isSaved", () => {
  test("retorna true si el mensaje está guardado por el usuario", async () => {
    const user = await createTestUser("issaved1@test.com")
    const session = await createTestSession(user.id, "is1")
    const msg = await createTestMessage(session.id)
    await save({ userId: user.id, messageId: msg.id, content: "Content" })

    const result = await checkIsSaved(msg.id, user.id)

    expect(result).toBe(true)
  })

  test("retorna false si el mensaje no está guardado", async () => {
    const user = await createTestUser("issaved2@test.com")

    const result = await checkIsSaved(9999, user.id)

    expect(result).toBe(false)
  })

  test("retorna false si otro usuario guardó el mismo mensaje", async () => {
    const user1 = await createTestUser("issaved3a@test.com")
    const user2 = await createTestUser("issaved3b@test.com")
    const session = await createTestSession(user1.id, "is3")
    const msg = await createTestMessage(session.id)
    await save({ userId: user1.id, messageId: msg.id, content: "Content" })

    const result = await checkIsSaved(msg.id, user2.id)

    expect(result).toBe(false)
  })
})
