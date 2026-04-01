import { eq, and } from "drizzle-orm"
import { getDb } from "../connection"
import { externalSources } from "../schema"
import type { NewExternalSource } from "../schema"
import { randomUUID } from "crypto"
import { encryptSecret, decryptSecret } from "../crypto"

// Re-export with legacy names for backwards compatibility
export const encryptCredentials = encryptSecret
export const decryptCredentials = decryptSecret

// --- Query functions ---

export async function createExternalSource(data: Omit<NewExternalSource, "id" | "createdAt" | "lastSync" | "active">) {
  const db = getDb()
  const values = {
    id: randomUUID(),
    ...data,
    credentials: data.credentials ? encryptSecret(data.credentials) : "{}",
    active: true,
    createdAt: Date.now(),
  }
  const [row] = await db.insert(externalSources).values(values).returning()
  return { ...row!, credentials: data.credentials ?? "{}" }
}

export async function listExternalSources(userId: number) {
  const db = getDb()
  const rows = await db.select().from(externalSources).where(eq(externalSources.userId, userId))
  return rows.map((r) => ({ ...r, credentials: decryptSecret(r.credentials) }))
}

export async function listActiveSourcesToSync() {
  const db = getDb()
  const now = Date.now()
  const allActive = await db.select().from(externalSources).where(eq(externalSources.active, true))
  return allActive
    .filter((s) => {
      const lastSync = s.lastSync ?? 0
      const intervalMs = s.schedule === "hourly" ? 3600_000 : s.schedule === "weekly" ? 7 * 86400_000 : 86400_000
      return now - lastSync >= intervalMs
    })
    .map((r) => ({ ...r, credentials: decryptSecret(r.credentials) }))
}

export async function updateSourceLastSync(id: string) {
  const db = getDb()
  await db.update(externalSources).set({ lastSync: Date.now() }).where(eq(externalSources.id, id))
}

export async function deleteExternalSource(id: string, userId: number) {
  const db = getDb()
  await db.delete(externalSources).where(and(eq(externalSources.id, id), eq(externalSources.userId, userId)))
}
