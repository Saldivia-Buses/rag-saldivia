import { eq, and, desc } from "drizzle-orm"
import { getDb } from "../connection"
import { annotations } from "../schema"
import type { NewAnnotation } from "../schema"

export async function saveAnnotation(data: Omit<NewAnnotation, "createdAt">) {
  const db = getDb()
  const [row] = await db
    .insert(annotations)
    .values({ ...data, createdAt: Date.now() })
    .returning()
  return row!
}

export async function listAnnotationsBySession(sessionId: string, userId: number) {
  const db = getDb()
  return db
    .select()
    .from(annotations)
    .where(and(eq(annotations.sessionId, sessionId), eq(annotations.userId, userId)))
    .orderBy(desc(annotations.createdAt))
}

export async function deleteAnnotation(id: number, userId: number) {
  const db = getDb()
  await db
    .delete(annotations)
    .where(and(eq(annotations.id, id), eq(annotations.userId, userId)))
}
