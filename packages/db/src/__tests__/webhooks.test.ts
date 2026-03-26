/**
 * Tests de queries de webhooks contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/webhooks.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { eq } from "drizzle-orm"
import { randomUUID, randomBytes } from "crypto"
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
    CREATE TABLE IF NOT EXISTS webhooks (
      id TEXT PRIMARY KEY,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      url TEXT NOT NULL,
      events TEXT NOT NULL DEFAULT '[]',
      secret TEXT NOT NULL,
      active INTEGER NOT NULL DEFAULT 1,
      created_at INTEGER NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_webhooks_active ON webhooks(active);
  `)
})

afterEach(async () => {
  await client.executeMultiple("DELETE FROM webhooks; DELETE FROM users;")
})

// ── Helpers ──────────────────────────────────────────────────────────────────

async function createUser(email = "user@test.com") {
  const [user] = await testDb
    .insert(schema.users)
    .values({ email, name: "Test", role: "user", apiKeyHash: "hash", preferences: {}, active: true, createdAt: Date.now() })
    .returning()
  return user!
}

async function createWebhook(userId: number, data: { url: string; events: string[] }) {
  const [row] = await testDb
    .insert(schema.webhooks)
    .values({
      id: randomUUID(),
      userId,
      url: data.url,
      events: data.events,
      secret: randomBytes(16).toString("hex"),
      active: true,
      createdAt: Date.now(),
    })
    .returning()
  return row!
}

async function listWebhooksByUser(userId: number) {
  return testDb.select().from(schema.webhooks).where(eq(schema.webhooks.userId, userId))
}

async function listWebhooksByEvent(eventType: string) {
  const all = await testDb.select().from(schema.webhooks).where(eq(schema.webhooks.active, true))
  return all.filter((w) => {
    const evts = w.events as string[]
    return evts.includes(eventType) || evts.includes("*")
  })
}

async function deleteWebhook(id: string) {
  await testDb.delete(schema.webhooks).where(eq(schema.webhooks.id, id))
}

async function listAllWebhooks() {
  return testDb.select().from(schema.webhooks)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("createWebhook", () => {
  test("crea un webhook con secret aleatorio y active=true", async () => {
    const user = await createUser()
    const wh = await createWebhook(user.id, { url: "https://example.com/hook", events: ["ingestion.completed"] })

    expect(wh.url).toBe("https://example.com/hook")
    expect(wh.active).toBe(true)
    expect(wh.secret).toHaveLength(32) // 16 bytes → 32 hex chars
    expect(wh.userId).toBe(user.id)
    expect(wh.id).toHaveLength(36) // UUID
  })

  test("dos webhooks tienen secrets distintos", async () => {
    const user = await createUser()
    const w1 = await createWebhook(user.id, { url: "https://a.com", events: ["*"] })
    const w2 = await createWebhook(user.id, { url: "https://b.com", events: ["*"] })

    expect(w1.secret).not.toBe(w2.secret)
  })
})

describe("listWebhooksByUser", () => {
  test("retorna solo los webhooks del usuario especificado", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createWebhook(u1.id, { url: "https://u1.com", events: ["*"] })
    await createWebhook(u2.id, { url: "https://u2.com", events: ["*"] })

    const hooks = await listWebhooksByUser(u1.id)
    expect(hooks).toHaveLength(1)
    expect(hooks[0]!.url).toBe("https://u1.com")
  })
})

describe("listWebhooksByEvent", () => {
  test("retorna webhooks que escuchan el evento específico", async () => {
    const user = await createUser()
    await createWebhook(user.id, { url: "https://a.com", events: ["ingestion.completed", "query.completed"] })
    await createWebhook(user.id, { url: "https://b.com", events: ["user.created"] })

    const hooks = await listWebhooksByEvent("ingestion.completed")
    expect(hooks).toHaveLength(1)
    expect(hooks[0]!.url).toBe("https://a.com")
  })

  test("retorna webhooks con wildcard '*'", async () => {
    const user = await createUser()
    await createWebhook(user.id, { url: "https://wildcard.com", events: ["*"] })
    await createWebhook(user.id, { url: "https://specific.com", events: ["user.created"] })

    const hooks = await listWebhooksByEvent("ingestion.completed")
    expect(hooks.some((h) => h.url === "https://wildcard.com")).toBe(true)
  })

  test("excluye webhooks inactivos", async () => {
    const user = await createUser()
    const wh = await createWebhook(user.id, { url: "https://inactive.com", events: ["ingestion.completed"] })
    await testDb.update(schema.webhooks).set({ active: false }).where(eq(schema.webhooks.id, wh.id))

    const hooks = await listWebhooksByEvent("ingestion.completed")
    expect(hooks.find((h) => h.url === "https://inactive.com")).toBeUndefined()
  })
})

describe("deleteWebhook", () => {
  test("elimina el webhook y ya no aparece en la lista", async () => {
    const user = await createUser()
    const wh = await createWebhook(user.id, { url: "https://borrar.com", events: ["*"] })

    await deleteWebhook(wh.id)

    const all = await listAllWebhooks()
    expect(all.find((w) => w.id === wh.id)).toBeUndefined()
  })
})
