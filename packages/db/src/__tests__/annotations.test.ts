/**
 * Tests de queries de anotaciones contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/annotations.test.ts
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
    CREATE TABLE IF NOT EXISTS annotations (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
      message_id INTEGER REFERENCES chat_messages(id) ON DELETE SET NULL,
      selected_text TEXT NOT NULL,
      note TEXT,
      created_at INTEGER NOT NULL
    );
  `)
})

afterEach(async () => {
  await client.executeMultiple(
    "DELETE FROM annotations; DELETE FROM chat_messages; DELETE FROM chat_sessions; DELETE FROM users;"
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

async function saveAnnotation(data: { userId: number; sessionId: string; messageId?: number; selectedText: string; note?: string }) {
  const [row] = await testDb
    .insert(schema.annotations)
    .values({ ...data, messageId: data.messageId ?? null, note: data.note ?? null, createdAt: Date.now() })
    .returning()
  return row!
}

async function listAnnotationsBySession(sessionId: string, userId: number) {
  return testDb
    .select()
    .from(schema.annotations)
    .where(and(eq(schema.annotations.sessionId, sessionId), eq(schema.annotations.userId, userId)))
    .orderBy(desc(schema.annotations.createdAt))
}

async function deleteAnnotation(id: number, userId: number) {
  await testDb
    .delete(schema.annotations)
    .where(and(eq(schema.annotations.id, id), eq(schema.annotations.userId, userId)))
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("saveAnnotation", () => {
  test("crea una anotación con texto seleccionado", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-a")
    const ann = await saveAnnotation({ userId: user.id, sessionId: "sess-a", selectedText: "texto importante" })

    expect(ann.selectedText).toBe("texto importante")
    expect(ann.userId).toBe(user.id)
    expect(ann.sessionId).toBe("sess-a")
    expect(ann.note).toBeNull()
    expect(ann.createdAt).toBeGreaterThan(0)
  })

  test("guarda la nota cuando se provee", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-b")
    const ann = await saveAnnotation({ userId: user.id, sessionId: "sess-b", selectedText: "frase", note: "recordar esto" })

    expect(ann.note).toBe("recordar esto")
  })
})

describe("listAnnotationsBySession", () => {
  test("retorna vacío si no hay anotaciones", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-c")
    const result = await listAnnotationsBySession("sess-c", user.id)
    expect(result).toHaveLength(0)
  })

  test("filtra por sessionId Y userId simultáneamente", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createSession(u1.id, "sess-u1")
    await createSession(u2.id, "sess-u2")

    await saveAnnotation({ userId: u1.id, sessionId: "sess-u1", selectedText: "de u1" })
    await saveAnnotation({ userId: u2.id, sessionId: "sess-u2", selectedText: "de u2" })

    const result = await listAnnotationsBySession("sess-u1", u1.id)
    expect(result).toHaveLength(1)
    expect(result[0]!.selectedText).toBe("de u1")
  })

  test("no retorna anotaciones de otro usuario en la misma sesión", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createSession(u1.id, "sess-shared")

    await saveAnnotation({ userId: u1.id, sessionId: "sess-shared", selectedText: "de u1" })

    const result = await listAnnotationsBySession("sess-shared", u2.id)
    expect(result).toHaveLength(0)
  })
})

describe("deleteAnnotation", () => {
  test("elimina la anotación del usuario correcto", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-del")
    const ann = await saveAnnotation({ userId: user.id, sessionId: "sess-del", selectedText: "borrar" })

    await deleteAnnotation(ann.id, user.id)

    const remaining = await listAnnotationsBySession("sess-del", user.id)
    expect(remaining).toHaveLength(0)
  })

  test("no elimina anotaciones de otro usuario", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createSession(u1.id, "sess-prot")
    const ann = await saveAnnotation({ userId: u1.id, sessionId: "sess-prot", selectedText: "protegida" })

    await deleteAnnotation(ann.id, u2.id)

    const remaining = await listAnnotationsBySession("sess-prot", u1.id)
    expect(remaining).toHaveLength(1)
  })
})
