/**
 * External sync worker — processes BullMQ jobs to sync external sources.
 *
 * For each job:
 * 1. Load connector from registry by provider name
 * 2. Decrypt credentials from DB
 * 3. Authenticate with the external service
 * 4. Detect changes (if supported) or list all documents
 * 5. For each new/modified doc: compute SHA-256, compare with sync_documents
 * 6. If changed: fetch content → enqueue to ingestion queue
 * 7. Update sync_documents with new hash
 * 8. Update lastSync on the external source
 */

import { createHash } from "crypto"
import { getConnector } from "@/lib/connectors/registry"
import {
  updateSourceLastSync,
  upsertSyncDocument,
  getSyncDocumentByExternalId,
} from "@rag-saldivia/db"
import { decryptSecret } from "@rag-saldivia/db"
import { ingestionQueue, startExternalSyncWorker, type ExternalSyncJobData, type Job } from "@/lib/queue"

const MAX_DOC_SIZE_BYTES = 10 * 1024 * 1024 // 10MB

function sha256(content: string | Buffer): string {
  return createHash("sha256").update(content).digest("hex")
}

async function processExternalSyncJob(job: Job<ExternalSyncJobData>): Promise<void> {
  const { sourceId, provider, collectionDest, fullSync } = job.data

  const connector = getConnector(provider)

  // Load credentials from DB (they're stored encrypted)
  // We need to find the source to get credentials
  // The worker gets sourceId — look it up directly
  const { getDb } = await import("@rag-saldivia/db")
  const db = getDb()
  const { externalSources } = await import("@rag-saldivia/db")
  const { eq } = await import("drizzle-orm")
  const [source] = await db.select().from(externalSources).where(eq(externalSources.id, sourceId)).limit(1)

  if (!source || !source.active) {
    await job.log(`Source ${sourceId} not found or inactive, skipping`)
    return
  }

  const credentials = JSON.parse(decryptSecret(source.credentials)) as Record<string, unknown>

  // 1. Authenticate
  try {
    await connector.authenticate(credentials)
    await job.log(`Authenticated with ${provider}`)
  } catch (err) {
    await job.log(`Authentication failed: ${String(err)}`)
    throw new Error(`Authentication failed for ${provider}: ${String(err)}`)
  }

  // 2. Get documents to sync
  let documentsToProcess: Array<{ externalId: string; title: string }> = []

  if (!fullSync && connector.detectChanges) {
    // Incremental: only changed docs since last sync
    const since = source.lastSync ?? 0
    try {
      const changes = await connector.detectChanges(credentials, since)
      documentsToProcess = changes
        .filter((c) => c.changeType !== "deleted")
        .map((c) => ({ externalId: c.externalId, title: c.title }))

      // Handle deletions
      for (const deleted of changes.filter((c) => c.changeType === "deleted")) {
        await upsertSyncDocument({
          sourceId,
          externalId: deleted.externalId,
          title: deleted.title,
          contentHash: "deleted",
          status: "synced",
        })
      }
      await job.log(`Detected ${changes.length} changes (${documentsToProcess.length} to process)`)
    } catch {
      // Fallback to full list if change detection fails
      await job.log("Change detection failed, falling back to full list")
      const result = await connector.listDocuments(credentials)
      documentsToProcess = result.documents.map((d) => ({ externalId: d.externalId, title: d.title }))
    }
  } else {
    // Full sync: list all documents
    let pageToken: string | undefined
    do {
      const result = await connector.listDocuments(credentials, { pageToken })
      documentsToProcess.push(...result.documents.map((d) => ({ externalId: d.externalId, title: d.title })))
      pageToken = result.nextPageToken
    } while (pageToken)
    await job.log(`Listed ${documentsToProcess.length} documents for full sync`)
  }

  // 3. Process each document
  let synced = 0
  let skipped = 0
  let failed = 0

  for (const doc of documentsToProcess) {
    try {
      // Check if content changed via hash
      const existing = await getSyncDocumentByExternalId(sourceId, doc.externalId)

      // Fetch content
      const content = await connector.fetchDocument(credentials, doc.externalId)

      // Size check
      if (content.sizeBytes > MAX_DOC_SIZE_BYTES) {
        await upsertSyncDocument({
          sourceId,
          externalId: doc.externalId,
          title: doc.title,
          contentHash: "too_large",
          sizeBytes: content.sizeBytes,
          status: "failed",
          errorMessage: `Document too large: ${(content.sizeBytes / 1024 / 1024).toFixed(1)}MB (max ${MAX_DOC_SIZE_BYTES / 1024 / 1024}MB)`,
        })
        failed++
        continue
      }

      const contentStr = typeof content.content === "string" ? content.content : content.content.toString("utf8")
      const hash = sha256(contentStr)

      // Skip if content unchanged
      if (existing && existing.contentHash === hash && !fullSync) {
        skipped++
        continue
      }

      // Enqueue for ingestion into RAG server
      await ingestionQueue.add("external-doc", {
        filePath: `external://${provider}/${doc.externalId}`,
        collection: collectionDest,
        userId: source.userId,
        filename: doc.title,
        // The ingestion worker will need to handle "external://" prefix
        // by reading the content from the sync document or re-fetching
      })

      // Record sync
      await upsertSyncDocument({
        sourceId,
        externalId: doc.externalId,
        title: doc.title,
        contentHash: hash,
        mimeType: content.mimeType,
        sizeBytes: content.sizeBytes,
        status: "synced",
      })

      synced++
    } catch (err) {
      await upsertSyncDocument({
        sourceId,
        externalId: doc.externalId,
        title: doc.title,
        contentHash: "error",
        status: "failed",
        errorMessage: String(err),
      })
      failed++
    }
  }

  // 4. Update lastSync
  await updateSourceLastSync(sourceId)
  await job.log(`Sync complete: ${synced} synced, ${skipped} skipped, ${failed} failed`)
}

/** Start the external sync worker. Call once from the worker entry point. */
export function initExternalSyncWorker() {
  return startExternalSyncWorker(processExternalSyncJob)
}
