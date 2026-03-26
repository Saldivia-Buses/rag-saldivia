/**
 * Tests de queries de prompt templates contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/templates.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { eq, desc } from "drizzle-orm"
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
    CREATE TABLE IF NOT EXISTS prompt_templates (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      title TEXT NOT NULL,
      prompt TEXT NOT NULL,
      focus_mode TEXT NOT NULL DEFAULT 'detallado',
      created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      active INTEGER NOT NULL DEFAULT 1,
      created_at INTEGER NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_prompt_templates_active ON prompt_templates(active);
  `)
})

afterEach(async () => {
  await client.executeMultiple("DELETE FROM prompt_templates; DELETE FROM users;")
})

// ── Helpers ──────────────────────────────────────────────────────────────────

async function createUser(email = "admin@test.com") {
  const [user] = await testDb
    .insert(schema.users)
    .values({ email, name: "Admin", role: "admin", apiKeyHash: "hash", preferences: {}, active: true, createdAt: Date.now() })
    .returning()
  return user!
}

async function createTemplate(userId: number, data: { title: string; prompt: string; focusMode?: string }) {
  const [row] = await testDb
    .insert(schema.promptTemplates)
    .values({ title: data.title, prompt: data.prompt, focusMode: data.focusMode ?? "detallado", createdBy: userId, active: true, createdAt: Date.now() })
    .returning()
  return row!
}

async function listActiveTemplates() {
  return testDb
    .select()
    .from(schema.promptTemplates)
    .where(eq(schema.promptTemplates.active, true))
    .orderBy(desc(schema.promptTemplates.createdAt))
}

async function deleteTemplate(id: number) {
  await testDb.delete(schema.promptTemplates).where(eq(schema.promptTemplates.id, id))
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("createTemplate", () => {
  test("crea un template con active=true por defecto", async () => {
    const admin = await createUser()
    const tpl = await createTemplate(admin.id, { title: "Resumen ejecutivo", prompt: "Resume en 3 puntos clave" })

    expect(tpl.title).toBe("Resumen ejecutivo")
    expect(tpl.prompt).toBe("Resume en 3 puntos clave")
    expect(tpl.active).toBe(true)
    expect(tpl.createdBy).toBe(admin.id)
    expect(tpl.focusMode).toBe("detallado")
  })

  test("acepta focusMode personalizado", async () => {
    const admin = await createUser()
    const tpl = await createTemplate(admin.id, { title: "Técnico", prompt: "Explica técnicamente", focusMode: "técnico" })
    expect(tpl.focusMode).toBe("técnico")
  })
})

describe("listActiveTemplates", () => {
  test("retorna vacío si no hay templates", async () => {
    const templates = await listActiveTemplates()
    expect(templates).toHaveLength(0)
  })

  test("retorna solo los templates activos", async () => {
    const admin = await createUser()
    const t1 = await createTemplate(admin.id, { title: "Activo", prompt: "..." })
    await createTemplate(admin.id, { title: "Activo 2", prompt: "..." })

    // Desactivar uno directamente
    await testDb.update(schema.promptTemplates).set({ active: false }).where(eq(schema.promptTemplates.id, t1.id))

    const templates = await listActiveTemplates()
    expect(templates).toHaveLength(1)
    expect(templates[0]!.title).toBe("Activo 2")
  })

  test("retorna ordenados por createdAt descendente", async () => {
    const admin = await createUser()
    await testDb.insert(schema.promptTemplates).values({
      title: "Primero", prompt: "...", focusMode: "detallado", createdBy: admin.id, active: true, createdAt: 1000,
    })
    await testDb.insert(schema.promptTemplates).values({
      title: "Segundo", prompt: "...", focusMode: "detallado", createdBy: admin.id, active: true, createdAt: 2000,
    })

    const templates = await listActiveTemplates()
    expect(templates[0]!.title).toBe("Segundo")
    expect(templates[1]!.title).toBe("Primero")
  })
})

describe("deleteTemplate", () => {
  test("elimina el template y ya no aparece en la lista", async () => {
    const admin = await createUser()
    const tpl = await createTemplate(admin.id, { title: "Borrar", prompt: "..." })

    await deleteTemplate(tpl.id)

    const templates = await listActiveTemplates()
    expect(templates.find((t) => t.id === tpl.id)).toBeUndefined()
  })
})
