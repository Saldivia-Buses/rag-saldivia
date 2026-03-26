/**
 * Tests de queries de memoria de usuario contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/memory.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { eq, and } from "drizzle-orm"
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
    CREATE TABLE IF NOT EXISTS user_memory (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      key TEXT NOT NULL,
      value TEXT NOT NULL,
      source TEXT NOT NULL DEFAULT 'explicit',
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL,
      UNIQUE(user_id, key)
    );
  `)
})

afterEach(async () => {
  await client.executeMultiple("DELETE FROM user_memory; DELETE FROM users;")
})

// ── Helpers ──────────────────────────────────────────────────────────────────

async function createUser(email = "user@test.com") {
  const [user] = await testDb
    .insert(schema.users)
    .values({ email, name: "Test", role: "user", apiKeyHash: "hash", preferences: {}, active: true, createdAt: Date.now() })
    .returning()
  return user!
}

async function setMemory(userId: number, key: string, value: string, source: "explicit" | "inferred" = "explicit") {
  const now = Date.now()
  await testDb
    .insert(schema.userMemory)
    .values({ userId, key, value, source, createdAt: now, updatedAt: now })
    .onConflictDoUpdate({
      target: [schema.userMemory.userId, schema.userMemory.key],
      set: { value, source, updatedAt: now },
    })
}

async function getMemory(userId: number) {
  return testDb.select().from(schema.userMemory).where(eq(schema.userMemory.userId, userId))
}

async function deleteMemory(userId: number, key: string) {
  await testDb
    .delete(schema.userMemory)
    .where(and(eq(schema.userMemory.userId, userId), eq(schema.userMemory.key, key)))
}

async function getMemoryAsContext(userId: number): Promise<string> {
  const entries = await getMemory(userId)
  if (entries.length === 0) return ""
  const lines = entries.map((e) => `${e.key}: ${e.value}`).join(", ")
  return `User context and preferences: ${lines}`
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("setMemory", () => {
  test("inserta una nueva entrada de memoria", async () => {
    const user = await createUser()
    await setMemory(user.id, "idioma", "español")

    const entries = await getMemory(user.id)
    expect(entries).toHaveLength(1)
    expect(entries[0]!.key).toBe("idioma")
    expect(entries[0]!.value).toBe("español")
    expect(entries[0]!.source).toBe("explicit")
  })

  test("actualiza el valor si la key ya existe (upsert)", async () => {
    const user = await createUser()
    await setMemory(user.id, "idioma", "español")
    await setMemory(user.id, "idioma", "inglés")

    const entries = await getMemory(user.id)
    expect(entries).toHaveLength(1)
    expect(entries[0]!.value).toBe("inglés")
  })

  test("guarda source 'inferred' cuando se indica", async () => {
    const user = await createUser()
    await setMemory(user.id, "tema", "tecnología", "inferred")

    const entries = await getMemory(user.id)
    expect(entries[0]!.source).toBe("inferred")
  })

  test("permite múltiples keys por usuario", async () => {
    const user = await createUser()
    await setMemory(user.id, "idioma", "español")
    await setMemory(user.id, "formato", "conciso")

    const entries = await getMemory(user.id)
    expect(entries).toHaveLength(2)
  })
})

describe("getMemory", () => {
  test("retorna vacío si el usuario no tiene memoria", async () => {
    const user = await createUser()
    const entries = await getMemory(user.id)
    expect(entries).toHaveLength(0)
  })

  test("no retorna entradas de otros usuarios", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await setMemory(u1.id, "idioma", "español")
    await setMemory(u2.id, "idioma", "inglés")

    const entriesU1 = await getMemory(u1.id)
    expect(entriesU1).toHaveLength(1)
    expect(entriesU1[0]!.value).toBe("español")
  })
})

describe("deleteMemory", () => {
  test("elimina la key especificada", async () => {
    const user = await createUser()
    await setMemory(user.id, "idioma", "español")
    await setMemory(user.id, "formato", "conciso")

    await deleteMemory(user.id, "idioma")

    const entries = await getMemory(user.id)
    expect(entries).toHaveLength(1)
    expect(entries[0]!.key).toBe("formato")
  })

  test("no elimina keys de otros usuarios con el mismo nombre", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await setMemory(u1.id, "idioma", "español")
    await setMemory(u2.id, "idioma", "inglés")

    await deleteMemory(u1.id, "idioma")

    const entriesU2 = await getMemory(u2.id)
    expect(entriesU2).toHaveLength(1)
  })
})

describe("getMemoryAsContext", () => {
  test("retorna string vacío si no hay entradas", async () => {
    const user = await createUser()
    const ctx = await getMemoryAsContext(user.id)
    expect(ctx).toBe("")
  })

  test("retorna el formato 'User context and preferences: key: value'", async () => {
    const user = await createUser()
    await setMemory(user.id, "idioma", "español")

    const ctx = await getMemoryAsContext(user.id)
    expect(ctx).toContain("User context and preferences:")
    expect(ctx).toContain("idioma: español")
  })

  test("incluye múltiples entradas separadas por coma", async () => {
    const user = await createUser()
    await setMemory(user.id, "idioma", "español")
    await setMemory(user.id, "formato", "conciso")

    const ctx = await getMemoryAsContext(user.id)
    expect(ctx).toContain("idioma: español")
    expect(ctx).toContain("formato: conciso")
  })
})
