/**
 * Tests de queries de historial de colecciones contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/collection-history.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { eq, desc } from "drizzle-orm"
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
    CREATE TABLE IF NOT EXISTS collection_history (
      id TEXT PRIMARY KEY,
      collection TEXT NOT NULL,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      action TEXT NOT NULL,
      filename TEXT,
      doc_count INTEGER,
      created_at INTEGER NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_collection_history_collection ON collection_history(collection);
  `)
})

afterEach(async () => {
  await client.executeMultiple("DELETE FROM collection_history; DELETE FROM users;")
})

// ── Helpers ──────────────────────────────────────────────────────────────────

async function createUser(email = "user@test.com") {
  const [user] = await testDb
    .insert(schema.users)
    .values({ email, name: "Test", role: "user", apiKeyHash: "hash", preferences: {}, active: true, createdAt: Date.now() })
    .returning()
  return user!
}

async function recordEvent(data: { collection: string; userId: number; action: "added" | "removed"; filename?: string; docCount?: number }) {
  const [row] = await testDb
    .insert(schema.collectionHistory)
    .values({ id: randomUUID(), ...data, filename: data.filename ?? null, docCount: data.docCount ?? null, createdAt: Date.now() })
    .returning()
  return row!
}

async function listHistoryByCollection(collection: string) {
  return testDb
    .select()
    .from(schema.collectionHistory)
    .where(eq(schema.collectionHistory.collection, collection))
    .orderBy(desc(schema.collectionHistory.createdAt))
    .limit(50)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("recordIngestionEvent", () => {
  test("crea un registro con todos los campos", async () => {
    const user = await createUser()
    const record = await recordEvent({ collection: "docs", userId: user.id, action: "added", filename: "doc.pdf", docCount: 10 })

    expect(record.collection).toBe("docs")
    expect(record.action).toBe("added")
    expect(record.filename).toBe("doc.pdf")
    expect(record.docCount).toBe(10)
    expect(record.userId).toBe(user.id)
    expect(record.createdAt).toBeGreaterThan(0)
    expect(record.id).toHaveLength(36) // UUID
  })

  test("acepta action 'removed'", async () => {
    const user = await createUser()
    const record = await recordEvent({ collection: "docs", userId: user.id, action: "removed" })
    expect(record.action).toBe("removed")
  })

  test("permite filename y docCount null", async () => {
    const user = await createUser()
    const record = await recordEvent({ collection: "docs", userId: user.id, action: "added" })
    expect(record.filename).toBeNull()
    expect(record.docCount).toBeNull()
  })
})

describe("listHistoryByCollection", () => {
  test("retorna vacío para colección sin historial", async () => {
    const history = await listHistoryByCollection("inexistente")
    expect(history).toHaveLength(0)
  })

  test("retorna solo registros de la colección especificada", async () => {
    const user = await createUser()
    await recordEvent({ collection: "docs", userId: user.id, action: "added" })
    await recordEvent({ collection: "otros", userId: user.id, action: "added" })

    const history = await listHistoryByCollection("docs")
    expect(history).toHaveLength(1)
    expect(history[0]!.collection).toBe("docs")
  })

  test("retorna en orden descendente por createdAt", async () => {
    const user = await createUser()

    await testDb.insert(schema.collectionHistory).values({
      id: randomUUID(), collection: "docs", userId: user.id, action: "added", createdAt: 1000,
    })
    await testDb.insert(schema.collectionHistory).values({
      id: randomUUID(), collection: "docs", userId: user.id, action: "added", createdAt: 3000,
    })
    await testDb.insert(schema.collectionHistory).values({
      id: randomUUID(), collection: "docs", userId: user.id, action: "removed", createdAt: 2000,
    })

    const history = await listHistoryByCollection("docs")
    expect(history[0]!.createdAt).toBe(3000)
    expect(history[1]!.createdAt).toBe(2000)
    expect(history[2]!.createdAt).toBe(1000)
  })

  test("respeta el límite de 50 registros", async () => {
    const user = await createUser()
    for (let i = 0; i < 60; i++) {
      await testDb.insert(schema.collectionHistory).values({
        id: randomUUID(), collection: "bulk", userId: user.id, action: "added", createdAt: Date.now() + i,
      })
    }

    const history = await listHistoryByCollection("bulk")
    expect(history).toHaveLength(50)
  })
})
