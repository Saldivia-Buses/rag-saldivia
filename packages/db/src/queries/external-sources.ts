import { eq, and, lte } from "drizzle-orm"
import { getDb } from "../connection"
import { externalSources } from "../schema"
import type { NewExternalSource } from "../schema"
import { randomUUID } from "crypto"

export async function createExternalSource(data: Omit<NewExternalSource, "id" | "createdAt" | "lastSync" | "active">) {
  const db = getDb()
  const [row] = await db
    .insert(externalSources)
    .values({ id: randomUUID(), ...data, active: true, createdAt: Date.now() })
    .returning()
  return row!
}

export async function listExternalSources(userId: number) {
  const db = getDb()
  return db.select().from(externalSources).where(eq(externalSources.userId, userId))
}

export async function listActiveSourcesToSync() {
  const db = getDb()
  const now = Date.now()
  const allActive = await db.select().from(externalSources).where(eq(externalSources.active, true))
  // Filtrar los que necesitan sync según su schedule
  return allActive.filter((s) => {
    const lastSync = s.lastSync ?? 0
    const intervalMs = s.schedule === "hourly" ? 3600_000 : s.schedule === "weekly" ? 7 * 86400_000 : 86400_000
    return now - lastSync >= intervalMs
  })
}

export async function updateSourceLastSync(id: string) {
  const db = getDb()
  await db.update(externalSources).set({ lastSync: Date.now() }).where(eq(externalSources.id, id))
}

export async function deleteExternalSource(id: string, userId: number) {
  const db = getDb()
  await db.delete(externalSources).where(and(eq(externalSources.id, id), eq(externalSources.userId, userId)))
}
