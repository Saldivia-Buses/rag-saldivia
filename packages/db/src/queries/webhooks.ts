import { eq, inArray } from "drizzle-orm"
import { getDb } from "../connection"
import { webhooks } from "../schema"
import type { NewWebhook } from "../schema"
import { randomUUID, randomBytes } from "crypto"

export async function createWebhook(data: { userId: number; url: string; events: string[] }) {
  const db = getDb()
  const [row] = await db
    .insert(webhooks)
    .values({
      id: randomUUID(),
      userId: data.userId,
      url: data.url,
      events: data.events,
      secret: randomBytes(16).toString("hex"),
      active: true,
      createdAt: Date.now(),
    })
    .returning()
  return row!
}

export async function listWebhooksByUser(userId: number) {
  const db = getDb()
  return db.select().from(webhooks).where(eq(webhooks.userId, userId))
}

export async function listWebhooksByEvent(eventType: string) {
  const db = getDb()
  // Filtrar webhooks activos que escuchan este tipo de evento
  const all = await db.select().from(webhooks).where(eq(webhooks.active, true))
  return all.filter((w) => {
    const evts = w.events as string[]
    return evts.includes(eventType) || evts.includes("*")
  })
}

export async function deleteWebhook(id: string, userId: number) {
  const db = getDb()
  await db
    .delete(webhooks)
    .where(eq(webhooks.id, id))
}

export async function listAllWebhooks() {
  const db = getDb()
  return db.select().from(webhooks)
}
