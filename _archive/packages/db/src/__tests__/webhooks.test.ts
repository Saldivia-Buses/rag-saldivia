/**
 * Tests de queries de webhooks contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/webhooks.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser } from "./setup"
import { createWebhook, listWebhooksByUser, listWebhooksByEvent, deleteWebhook, listAllWebhooks } from "../queries/webhooks"
import * as schema from "../schema"
import { eq } from "drizzle-orm"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM webhooks; DELETE FROM users;")
})

describe("createWebhook", () => {
  test("crea webhook con secret aleatorio y active=true", async () => {
    const user = await insertUser(db)
    const wh = await createWebhook({ userId: user.id, url: "https://a.com", events: ["ingestion.completed"] })
    expect(wh.url).toBe("https://a.com")
    expect(wh.active).toBe(true)
    expect(wh.secret).toHaveLength(32)
    expect(wh.id).toHaveLength(36)
  })

  test("dos webhooks tienen secrets distintos", async () => {
    const user = await insertUser(db)
    const w1 = await createWebhook({ userId: user.id, url: "https://a.com", events: ["*"] })
    const w2 = await createWebhook({ userId: user.id, url: "https://b.com", events: ["*"] })
    expect(w1.secret).not.toBe(w2.secret)
  })
})

describe("listWebhooksByUser", () => {
  test("retorna solo los webhooks del usuario especificado", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    await createWebhook({ userId: u1.id, url: "https://u1.com", events: ["*"] })
    await createWebhook({ userId: u2.id, url: "https://u2.com", events: ["*"] })
    const hooks = await listWebhooksByUser(u1.id)
    expect(hooks).toHaveLength(1)
    expect(hooks[0]!.url).toBe("https://u1.com")
  })
})

describe("listWebhooksByEvent", () => {
  test("retorna webhooks que escuchan el evento específico", async () => {
    const user = await insertUser(db)
    await createWebhook({ userId: user.id, url: "https://a.com", events: ["ingestion.completed"] })
    await createWebhook({ userId: user.id, url: "https://b.com", events: ["user.created"] })
    const hooks = await listWebhooksByEvent("ingestion.completed")
    expect(hooks.every((h) => (h.events as string[]).includes("ingestion.completed"))).toBe(true)
    expect(hooks.some((h) => h.url === "https://b.com")).toBe(false)
  })

  test("incluye webhooks con wildcard '*'", async () => {
    const user = await insertUser(db)
    await createWebhook({ userId: user.id, url: "https://wildcard.com", events: ["*"] })
    const hooks = await listWebhooksByEvent("cualquier.evento")
    expect(hooks.some((h) => h.url === "https://wildcard.com")).toBe(true)
  })

  test("excluye webhooks inactivos", async () => {
    const user = await insertUser(db)
    const wh = await createWebhook({ userId: user.id, url: "https://inactive.com", events: ["ingestion.completed"] })
    await db.update(schema.webhooks).set({ active: false }).where(eq(schema.webhooks.id, wh.id))
    const hooks = await listWebhooksByEvent("ingestion.completed")
    expect(hooks.find((h) => h.url === "https://inactive.com")).toBeUndefined()
  })
})

describe("deleteWebhook", () => {
  test("elimina el webhook de la lista", async () => {
    const user = await insertUser(db)
    const wh = await createWebhook({ userId: user.id, url: "https://del.com", events: ["*"] })
    await deleteWebhook(wh.id, user.id)
    expect((await listAllWebhooks()).find((w) => w.id === wh.id)).toBeUndefined()
  })
})
