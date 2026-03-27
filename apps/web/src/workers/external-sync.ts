#!/usr/bin/env bun
/**
 * Worker de auto-ingesta desde fuentes externas.
 * F3.48 — Google Drive, SharePoint, Confluence.
 *
 * F8.26 — Master lock via Redis SET NX EX:
 * Solo la instancia maestra sincroniza. Si múltiples instancias corren,
 * las no-master esperan hasta obtener el lock.
 *
 * Prerequisito: F2.38 (webhooks).
 * Las credenciales OAuth se configuran en /admin/external-sources.
 */

import { listActiveSourcesToSync, updateSourceLastSync, getRedisClient } from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"

const SYNC_INTERVAL_MS = 5 * 60 * 1000 // 5 minutos
const WORKER_ID = `external-sync-${process.pid}-${Date.now()}`
const MASTER_LOCK_KEY = "worker:master:external-sync"
const MASTER_LOCK_TTL_S = 60

async function acquireExternalSyncLock(): Promise<boolean> {
  try {
    const result = await getRedisClient().set(
      MASTER_LOCK_KEY,
      WORKER_ID,
      "EX",
      MASTER_LOCK_TTL_S,
      "NX"
    )
    return result === "OK"
  } catch {
    // Si Redis no está disponible, asumir que somos master (degradado seguro)
    return true
  }
}

// Renovar el lock cada 30s mientras el loop está activo
const lockRenewalInterval = setInterval(() => {
  getRedisClient().expire(MASTER_LOCK_KEY, MASTER_LOCK_TTL_S).catch(() => {})
}, 30_000)

async function syncSource(source: Awaited<ReturnType<typeof listActiveSourcesToSync>>[number]) {
  log.info("ingestion.started", {
    provider: source.provider,
    name: source.name,
    collection: source.collectionDest,
    sourceId: source.id,
  })

  try {
    switch (source.provider) {
      case "google_drive": {
        log.info("system.warning", { message: `Google Drive sync not yet implemented`, sourceId: source.id })
        break
      }
      case "sharepoint": {
        log.info("system.warning", { message: `SharePoint sync not yet implemented`, sourceId: source.id })
        break
      }
      case "confluence": {
        log.info("system.warning", { message: `Confluence sync not yet implemented`, sourceId: source.id })
        break
      }
    }

    await updateSourceLastSync(source.id)

    log.info("ingestion.completed", {
      provider: source.provider,
      name: source.name,
      collection: source.collectionDest,
      sourceId: source.id,
    })
  } catch (err) {
    log.error("ingestion.failed", {
      provider: source.provider,
      sourceId: source.id,
      error: String(err).slice(0, 200),
    })
  }
}

async function syncAllSources() {
  const sources = await listActiveSourcesToSync()
  if (sources.length > 0) {
    log.info("system.start", { message: `[external-sync] Syncing ${sources.length} source(s)` })
    await Promise.allSettled(sources.map(syncSource))
  }
}

async function syncLoop() {
  log.info("system.start", { message: `[external-sync] Worker started (id: ${WORKER_ID})` })

  while (true) {
    try {
      // Solo el master sincroniza — evita duplicar sync si hay múltiples instancias
      if (await acquireExternalSyncLock()) {
        await syncAllSources()
      }
    } catch (err) {
      log.error("system.error", { error: `[external-sync] Loop error: ${String(err).slice(0, 200)}` })
    }
    await new Promise((r) => setTimeout(r, SYNC_INTERVAL_MS))
  }
}

process.on("SIGTERM", () => {
  clearInterval(lockRenewalInterval)
  getRedisClient().del(MASTER_LOCK_KEY).catch(() => {})
})

syncLoop().catch((e) => {
  log.fatal("system.error", { error: `[external-sync] Fatal: ${String(e)}` })
  process.exit(1)
})
