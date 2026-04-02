import { eq, and, desc } from "drizzle-orm"
import { getDb } from "../connection"
import { savedResponses } from "../schema"
import type { NewSavedResponse } from "../schema"

export async function saveResponse(data: Omit<NewSavedResponse, "createdAt">) {
  const db = getDb()
  const now = Date.now()
  const [row] = await db
    .insert(savedResponses)
    .values({ ...data, createdAt: now })
    .returning()
  return row!
}

export async function unsaveResponse(id: number, userId: number) {
  const db = getDb()
  await db
    .delete(savedResponses)
    .where(and(eq(savedResponses.id, id), eq(savedResponses.userId, userId)))
}

export async function unsaveByMessageId(messageId: number, userId: number) {
  const db = getDb()
  await db
    .delete(savedResponses)
    .where(and(eq(savedResponses.messageId, messageId), eq(savedResponses.userId, userId)))
}

export async function listSavedResponses(userId: number) {
  const db = getDb()
  return db
    .select()
    .from(savedResponses)
    .where(eq(savedResponses.userId, userId))
    .orderBy(desc(savedResponses.createdAt))
}

export async function isSaved(messageId: number, userId: number): Promise<boolean> {
  const db = getDb()
  const rows = await db
    .select({ id: savedResponses.id })
    .from(savedResponses)
    .where(and(eq(savedResponses.messageId, messageId), eq(savedResponses.userId, userId)))
    .limit(1)
  return rows.length > 0
}
