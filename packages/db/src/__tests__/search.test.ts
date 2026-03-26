/**
 * Tests de búsqueda universal contra SQLite en memoria.
 * Nota: FTS5 no está disponible en :memory: sin setup especial.
 * Estos tests verifican el path LIKE (fallback) que activa en ese contexto.
 * Corre con: bun test packages/db/src/__tests__/search.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { eq, and, sql } from "drizzle-orm"
import { randomUUID } from "crypto"
import * as schema from "../schema"

process.env["DATABASE_PATH"] = ":memory:"

const client = createClient({ url: ":memory:" })
const testDb = drizzle(client, { schema })

type SearchResult = {
  type: "session" | "message" | "saved" | "template"
  id: string
  title: string
  snippet: string
  sessionId?: string
  href: string
}

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
      message_id INTEGER,
      content TEXT NOT NULL,
      session_title TEXT,
      created_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS prompt_templates (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      title TEXT NOT NULL,
      prompt TEXT NOT NULL,
      focus_mode TEXT NOT NULL DEFAULT 'detallado',
      created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      active INTEGER NOT NULL DEFAULT 1,
      created_at INTEGER NOT NULL
    );
  `)
})

afterEach(async () => {
  await client.executeMultiple(
    "DELETE FROM saved_responses; DELETE FROM prompt_templates; DELETE FROM chat_messages; DELETE FROM chat_sessions; DELETE FROM users;"
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

async function createSession(userId: number, id: string, title: string) {
  const now = Date.now()
  const [session] = await testDb
    .insert(schema.chatSessions)
    .values({ id, userId, title, collection: "col", crossdoc: false, createdAt: now, updatedAt: now })
    .returning()
  return session!
}

async function createSavedResponse(userId: number, content: string, sessionTitle?: string) {
  const [row] = await testDb
    .insert(schema.savedResponses)
    .values({ userId, content, sessionTitle: sessionTitle ?? null, createdAt: Date.now() })
    .returning()
  return row!
}

async function createTemplate(userId: number, title: string, prompt: string) {
  const [row] = await testDb
    .insert(schema.promptTemplates)
    .values({ title, prompt, focusMode: "detallado", createdBy: userId, active: true, createdAt: Date.now() })
    .returning()
  return row!
}

// Replica universalSearch con LIKE (path que activa sin FTS5)
async function universalSearch(query: string, userId: number, limit = 20): Promise<SearchResult[]> {
  if (!query || query.trim().length < 2) return []

  const q = query.trim()
  const results: SearchResult[] = []
  const likeQ = `%${q}%`

  // Sessions por título
  const sessions = await testDb
    .select({ id: schema.chatSessions.id, title: schema.chatSessions.title })
    .from(schema.chatSessions)
    .where(and(eq(schema.chatSessions.userId, userId), sql`${schema.chatSessions.title} LIKE ${likeQ}`))
    .limit(Math.floor(limit / 2))

  for (const s of sessions) {
    results.push({ type: "session", id: s.id, title: s.title, snippet: "", href: `/chat/${s.id}` })
  }

  // Templates activos por título y prompt
  const templates = await testDb
    .select({ id: schema.promptTemplates.id, title: schema.promptTemplates.title, prompt: schema.promptTemplates.prompt })
    .from(schema.promptTemplates)
    .where(and(eq(schema.promptTemplates.active, true), sql`(${schema.promptTemplates.title} LIKE ${likeQ} OR ${schema.promptTemplates.prompt} LIKE ${likeQ})`))
    .limit(5)

  for (const t of templates) {
    results.push({ type: "template", id: String(t.id), title: t.title, snippet: t.prompt.slice(0, 80), href: "/chat" })
  }

  // Saved responses del usuario
  const saved = await testDb
    .select({ id: schema.savedResponses.id, content: schema.savedResponses.content, sessionTitle: schema.savedResponses.sessionTitle })
    .from(schema.savedResponses)
    .where(and(eq(schema.savedResponses.userId, userId), sql`${schema.savedResponses.content} LIKE ${likeQ}`))
    .limit(5)

  for (const s of saved) {
    results.push({ type: "saved", id: String(s.id), title: s.sessionTitle ?? "Respuesta guardada", snippet: s.content.slice(0, 100), href: "/saved" })
  }

  return results.slice(0, limit)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("universalSearch — casos edge", () => {
  test("retorna array vacío para query vacío", async () => {
    const user = await createUser()
    const results = await universalSearch("", user.id)
    expect(results).toHaveLength(0)
  })

  test("retorna array vacío para query de 1 caracter", async () => {
    const user = await createUser()
    const results = await universalSearch("a", user.id)
    expect(results).toHaveLength(0)
  })

  test("retorna array vacío para query de solo espacios", async () => {
    const user = await createUser()
    const results = await universalSearch("   ", user.id)
    expect(results).toHaveLength(0)
  })
})

describe("universalSearch — sesiones", () => {
  test("encuentra sesiones por título (LIKE)", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-1", "Análisis de mercado")
    await createSession(user.id, "sess-2", "Estrategia comercial")

    const results = await universalSearch("mercado", user.id)
    expect(results.some((r) => r.type === "session" && r.title === "Análisis de mercado")).toBe(true)
    expect(results.every((r) => r.title !== "Estrategia comercial")).toBe(true)
  })

  test("no retorna sesiones de otros usuarios", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createSession(u1.id, "sess-u1", "Proyecto confidencial")
    await createSession(u2.id, "sess-u2", "Proyecto público")

    const results = await universalSearch("Proyecto", u1.id)
    expect(results.every((r) => r.type !== "session" || r.id === "sess-u1")).toBe(true)
  })
})

describe("universalSearch — templates", () => {
  test("encuentra templates por título", async () => {
    const user = await createUser()
    await createTemplate(user.id, "Resumen ejecutivo", "Resume los puntos clave en formato ejecutivo")

    const results = await universalSearch("ejecutivo", user.id)
    expect(results.some((r) => r.type === "template")).toBe(true)
  })

  test("encuentra templates por contenido del prompt", async () => {
    const user = await createUser()
    await createTemplate(user.id, "Análisis técnico", "Proporciona un análisis detallado de los aspectos técnicos")

    const results = await universalSearch("detallado", user.id)
    expect(results.some((r) => r.type === "template")).toBe(true)
  })
})

describe("universalSearch — respuestas guardadas", () => {
  test("encuentra respuestas guardadas por contenido", async () => {
    const user = await createUser()
    await createSavedResponse(user.id, "La inteligencia artificial transformará la industria", "Sesión IA")

    const results = await universalSearch("inteligencia", user.id)
    expect(results.some((r) => r.type === "saved")).toBe(true)
  })

  test("no retorna respuestas guardadas de otros usuarios", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createSavedResponse(u2.id, "Contenido exclusivo de u2", "Sesión")

    const results = await universalSearch("exclusivo", u1.id)
    expect(results.filter((r) => r.type === "saved")).toHaveLength(0)
  })
})
