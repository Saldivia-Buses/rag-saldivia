/**
 * Tests de queries de fuentes externas contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/external-sources.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { eq, and } from "drizzle-orm"
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
    CREATE TABLE IF NOT EXISTS external_sources (
      id TEXT PRIMARY KEY,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      provider TEXT NOT NULL,
      name TEXT NOT NULL,
      credentials TEXT NOT NULL DEFAULT '{}',
      collection_dest TEXT NOT NULL,
      schedule TEXT NOT NULL DEFAULT 'daily',
      active INTEGER NOT NULL DEFAULT 1,
      last_sync INTEGER,
      created_at INTEGER NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_external_sources_user ON external_sources(user_id);
  `)
})

afterEach(async () => {
  await client.executeMultiple("DELETE FROM external_sources; DELETE FROM users;")
})

// ── Helpers ──────────────────────────────────────────────────────────────────

async function createUser(email = "user@test.com") {
  const [user] = await testDb
    .insert(schema.users)
    .values({ email, name: "Test", role: "user", apiKeyHash: "hash", preferences: {}, active: true, createdAt: Date.now() })
    .returning()
  return user!
}

async function createExternalSource(userId: number, data: {
  provider: "google_drive" | "sharepoint" | "confluence"
  name: string
  collectionDest: string
  schedule?: "hourly" | "daily" | "weekly"
}) {
  const [row] = await testDb
    .insert(schema.externalSources)
    .values({
      id: randomUUID(),
      userId,
      provider: data.provider,
      name: data.name,
      credentials: "{}",
      collectionDest: data.collectionDest,
      schedule: data.schedule ?? "daily",
      active: true,
      createdAt: Date.now(),
    })
    .returning()
  return row!
}

async function listExternalSources(userId: number) {
  return testDb.select().from(schema.externalSources).where(eq(schema.externalSources.userId, userId))
}

async function listActiveSourcesToSync() {
  const now = Date.now()
  const allActive = await testDb.select().from(schema.externalSources).where(eq(schema.externalSources.active, true))
  return allActive.filter((s) => {
    const lastSync = s.lastSync ?? 0
    const intervalMs = s.schedule === "hourly" ? 3_600_000 : s.schedule === "weekly" ? 7 * 86_400_000 : 86_400_000
    return now - lastSync >= intervalMs
  })
}

async function updateSourceLastSync(id: string) {
  await testDb.update(schema.externalSources).set({ lastSync: Date.now() }).where(eq(schema.externalSources.id, id))
}

async function deleteExternalSource(id: string, userId: number) {
  await testDb.delete(schema.externalSources).where(
    and(eq(schema.externalSources.id, id), eq(schema.externalSources.userId, userId))
  )
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("createExternalSource", () => {
  test("crea una fuente con active=true", async () => {
    const user = await createUser()
    const source = await createExternalSource(user.id, {
      provider: "google_drive",
      name: "Mi Drive",
      collectionDest: "docs",
    })

    expect(source.provider).toBe("google_drive")
    expect(source.name).toBe("Mi Drive")
    expect(source.active).toBe(true)
    expect(source.lastSync).toBeNull()
    expect(source.id).toHaveLength(36)
  })

  test("acepta los tres providers soportados", async () => {
    const user = await createUser()
    const gd = await createExternalSource(user.id, { provider: "google_drive", name: "GD", collectionDest: "c" })
    const sp = await createExternalSource(user.id, { provider: "sharepoint", name: "SP", collectionDest: "c" })
    const cf = await createExternalSource(user.id, { provider: "confluence", name: "CF", collectionDest: "c" })

    expect(gd.provider).toBe("google_drive")
    expect(sp.provider).toBe("sharepoint")
    expect(cf.provider).toBe("confluence")
  })
})

describe("listExternalSources", () => {
  test("retorna solo fuentes del usuario especificado", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createExternalSource(u1.id, { provider: "google_drive", name: "De u1", collectionDest: "c" })
    await createExternalSource(u2.id, { provider: "sharepoint", name: "De u2", collectionDest: "c" })

    const sources = await listExternalSources(u1.id)
    expect(sources).toHaveLength(1)
    expect(sources[0]!.name).toBe("De u1")
  })

  test("retorna vacío si el usuario no tiene fuentes", async () => {
    const user = await createUser()
    const sources = await listExternalSources(user.id)
    expect(sources).toHaveLength(0)
  })
})

describe("listActiveSourcesToSync", () => {
  test("retorna fuentes activas que nunca se sincronizaron (lastSync = null)", async () => {
    const user = await createUser()
    await createExternalSource(user.id, { provider: "google_drive", name: "Nuevo", collectionDest: "c", schedule: "daily" })

    const toSync = await listActiveSourcesToSync()
    expect(toSync.length).toBeGreaterThanOrEqual(1)
    expect(toSync.some((s) => s.name === "Nuevo")).toBe(true)
  })

  test("no retorna fuentes recientemente sincronizadas", async () => {
    const user = await createUser()
    const source = await createExternalSource(user.id, { provider: "google_drive", name: "Reciente", collectionDest: "c", schedule: "daily" })

    // Simular sync reciente
    await testDb.update(schema.externalSources).set({ lastSync: Date.now() }).where(eq(schema.externalSources.id, source.id))

    const toSync = await listActiveSourcesToSync()
    expect(toSync.find((s) => s.id === source.id)).toBeUndefined()
  })

  test("no retorna fuentes inactivas", async () => {
    const user = await createUser()
    const source = await createExternalSource(user.id, { provider: "confluence", name: "Inactiva", collectionDest: "c" })
    await testDb.update(schema.externalSources).set({ active: false }).where(eq(schema.externalSources.id, source.id))

    const toSync = await listActiveSourcesToSync()
    expect(toSync.find((s) => s.id === source.id)).toBeUndefined()
  })
})

describe("updateSourceLastSync", () => {
  test("actualiza lastSync al momento actual", async () => {
    const user = await createUser()
    const source = await createExternalSource(user.id, { provider: "google_drive", name: "GD", collectionDest: "c" })
    expect(source.lastSync).toBeNull()

    const before = Date.now()
    await updateSourceLastSync(source.id)

    const updated = await testDb.select().from(schema.externalSources).where(eq(schema.externalSources.id, source.id))
    expect(updated[0]!.lastSync).toBeGreaterThanOrEqual(before)
  })
})

describe("deleteExternalSource", () => {
  test("elimina la fuente del usuario correcto", async () => {
    const user = await createUser()
    const source = await createExternalSource(user.id, { provider: "google_drive", name: "Borrar", collectionDest: "c" })

    await deleteExternalSource(source.id, user.id)

    const sources = await listExternalSources(user.id)
    expect(sources.find((s) => s.id === source.id)).toBeUndefined()
  })

  test("no elimina fuentes de otro usuario", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    const source = await createExternalSource(u1.id, { provider: "google_drive", name: "Protegida", collectionDest: "c" })

    await deleteExternalSource(source.id, u2.id) // intento de otro usuario

    const sources = await listExternalSources(u1.id)
    expect(sources.find((s) => s.id === source.id)).toBeDefined()
  })
})
