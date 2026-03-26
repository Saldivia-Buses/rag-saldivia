import { eq, and } from "drizzle-orm"
import { getDb } from "../connection"
import { userMemory } from "../schema"

export async function setMemory(
  userId: number,
  key: string,
  value: string,
  source: "explicit" | "inferred" = "explicit"
) {
  const db = getDb()
  const now = Date.now()
  await db
    .insert(userMemory)
    .values({ userId, key, value, source, createdAt: now, updatedAt: now })
    .onConflictDoUpdate({
      target: [userMemory.userId, userMemory.key],
      set: { value, source, updatedAt: now },
    })
}

export async function getMemory(userId: number) {
  const db = getDb()
  return db.select().from(userMemory).where(eq(userMemory.userId, userId))
}

export async function deleteMemory(userId: number, key: string) {
  const db = getDb()
  await db
    .delete(userMemory)
    .where(and(eq(userMemory.userId, userId), eq(userMemory.key, key)))
}

export async function getMemoryAsContext(userId: number): Promise<string> {
  const entries = await getMemory(userId)
  if (entries.length === 0) return ""
  const lines = entries.map((e) => `${e.key}: ${e.value}`).join(", ")
  return `User context and preferences: ${lines}`
}
