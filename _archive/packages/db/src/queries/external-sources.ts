import { eq, and, sql, count } from "drizzle-orm"
import { getDb } from "../connection"
import { externalSources, syncDocuments } from "../schema"
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

export async function toggleExternalSource(id: string, userId: number, active: boolean) {
  const db = getDb()
  await db
    .update(externalSources)
    .set({ active })
    .where(and(eq(externalSources.id, id), eq(externalSources.userId, userId)))
}

export async function getExternalSourceById(id: string, userId: number) {
  const db = getDb()
  const [row] = await db
    .select()
    .from(externalSources)
    .where(and(eq(externalSources.id, id), eq(externalSources.userId, userId)))
    .limit(1)
  return row ?? null
}

// ── Sync Documents (change detection) ─────────────────────────────────────

export async function upsertSyncDocument(data: {
  sourceId: string
  externalId: string
  title: string
  contentHash: string
  mimeType?: string
  sizeBytes?: number | null
  lastModifiedExternal?: number | null
  status?: "synced" | "failed" | "pending"
  errorMessage?: string | null
}) {
  const db = getDb()
  const now = Date.now()
  const [row] = await db
    .insert(syncDocuments)
    .values({
      sourceId: data.sourceId,
      externalId: data.externalId,
      title: data.title,
      contentHash: data.contentHash,
      mimeType: data.mimeType ?? "application/octet-stream",
      sizeBytes: data.sizeBytes ?? null,
      lastModifiedExternal: data.lastModifiedExternal ?? null,
      lastSyncedAt: now,
      status: data.status ?? "synced",
      errorMessage: data.errorMessage ?? null,
    })
    .onConflictDoUpdate({
      target: [syncDocuments.sourceId, syncDocuments.externalId],
      set: {
        title: sql`excluded.title`,
        contentHash: sql`excluded.content_hash`,
        mimeType: sql`excluded.mime_type`,
        sizeBytes: sql`excluded.size_bytes`,
        lastModifiedExternal: sql`excluded.last_modified_external`,
        lastSyncedAt: now,
        status: sql`excluded.status`,
        errorMessage: sql`excluded.error_message`,
      },
    })
    .returning()
  return row!
}

export async function getSyncDocumentsBySource(sourceId: string) {
  const db = getDb()
  return db.select().from(syncDocuments).where(eq(syncDocuments.sourceId, sourceId))
}

export async function getSyncDocumentByExternalId(sourceId: string, externalId: string) {
  const db = getDb()
  const [row] = await db
    .select()
    .from(syncDocuments)
    .where(and(eq(syncDocuments.sourceId, sourceId), eq(syncDocuments.externalId, externalId)))
    .limit(1)
  return row ?? null
}

export async function deleteSyncDocumentsForSource(sourceId: string) {
  const db = getDb()
  await db.delete(syncDocuments).where(eq(syncDocuments.sourceId, sourceId))
}

export async function countSyncDocuments(sourceId: string): Promise<number> {
  const db = getDb()
  const [result] = await db
    .select({ total: count() })
    .from(syncDocuments)
    .where(eq(syncDocuments.sourceId, sourceId))
  return result?.total ?? 0
}
