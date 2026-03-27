#!/usr/bin/env bun
/**
 * Worker de auto-ingesta desde fuentes externas.
 * F3.48 — Google Drive, SharePoint, Confluence.
 *
 * Prerequisito: F2.38 (webhooks).
 * Las credenciales OAuth se configuran en /admin/external-sources.
 * En esta versión MVP, el worker hace lo que puede sin SDKs instalados
 * (la instalación de googleapis y @microsoft/microsoft-graph-client
 * se realiza cuando se configuran las credenciales reales).
 */

import { listActiveSourcesToSync, updateSourceLastSync } from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"

const SYNC_INTERVAL_MS = 5 * 60 * 1000 // 5 minutos

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
        // MVP: requiere googleapis instalado y credenciales OAuth configuradas
        // Implementación completa pendiente de credenciales reales
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

async function syncLoop() {
  while (true) {
    try {
      const sources = await listActiveSourcesToSync()
      if (sources.length > 0) {
        log.info("system.start", { message: `[external-sync] Syncing ${sources.length} source(s)` })
        await Promise.allSettled(sources.map(syncSource))
      }
    } catch (err) {
      log.error("system.error", { error: `[external-sync] Loop error: ${String(err).slice(0, 200)}` })
    }
    await new Promise((r) => setTimeout(r, SYNC_INTERVAL_MS))
  }
}

syncLoop().catch((e) => {
  log.fatal("system.error", { error: `[external-sync] Fatal: ${String(e)}` })
  process.exit(1)
})
