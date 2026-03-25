import { eq, desc } from "drizzle-orm"
import { getDb } from "../connection"
import { collectionHistory } from "../schema"
import type { NewCollectionHistory } from "../schema"
import { randomUUID } from "crypto"

export async function recordIngestionEvent(data: Omit<NewCollectionHistory, "id" | "createdAt">) {
  const db = getDb()
  const [row] = await db
    .insert(collectionHistory)
    .values({ id: randomUUID(), ...data, createdAt: Date.now() })
    .returning()
  return row!
}

export async function listHistoryByCollection(collection: string) {
  const db = getDb()
  return db
    .select()
    .from(collectionHistory)
    .where(eq(collectionHistory.collection, collection))
    .orderBy(desc(collectionHistory.createdAt))
    .limit(50)
}
