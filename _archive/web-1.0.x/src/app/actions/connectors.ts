"use server"

import { z } from "zod"
import { adminAction } from "@/lib/safe-action"
import {
  createExternalSource,
  listExternalSources,
  deleteExternalSource,
  countSyncDocuments,
  deleteSyncDocumentsForSource,
  toggleExternalSource,
  getExternalSourceById,
} from "@rag-saldivia/db"
import { ConnectorProviderSchema } from "@rag-saldivia/shared"
import { externalSyncQueue } from "@/lib/queue"
import { revalidatePath } from "next/cache"

export const actionListConnectors = adminAction
  .schema(z.object({}))
  .action(async ({ ctx }) => {
    const sources = await listExternalSources(ctx.user.id)
    const withStats = await Promise.all(
      sources.map(async (s) => ({
        id: s.id,
        provider: s.provider,
        name: s.name,
        collectionDest: s.collectionDest,
        schedule: s.schedule,
        active: s.active,
        lastSync: s.lastSync,
        createdAt: s.createdAt,
        docCount: await countSyncDocuments(s.id),
      }))
    )
    return { connectors: withStats }
  })

export const actionCreateConnector = adminAction
  .schema(
    z.object({
      provider: ConnectorProviderSchema,
      name: z.string().min(1).max(200),
      collectionDest: z.string().min(1),
      schedule: z.enum(["hourly", "daily", "weekly"]),
      credentials: z.string().min(1),
    })
  )
  .action(async ({ parsedInput: data, ctx }) => {
    await createExternalSource({
      userId: ctx.user.id,
      provider: data.provider,
      name: data.name,
      credentials: data.credentials,
      collectionDest: data.collectionDest,
      schedule: data.schedule,
    })
    revalidatePath("/admin/connectors")
    return { ok: true }
  })

export const actionDeleteConnector = adminAction
  .schema(z.object({ id: z.string().min(1) }))
  .action(async ({ parsedInput: { id }, ctx }) => {
    await deleteSyncDocumentsForSource(id)
    await deleteExternalSource(id, ctx.user.id)
    revalidatePath("/admin/connectors")
    return { ok: true }
  })

export const actionToggleConnector = adminAction
  .schema(z.object({ id: z.string().min(1), active: z.boolean() }))
  .action(async ({ parsedInput: { id, active }, ctx }) => {
    await toggleExternalSource(id, ctx.user.id, active)
    revalidatePath("/admin/connectors")
    return { ok: true }
  })

export const actionSyncNow = adminAction
  .schema(z.object({ id: z.string().min(1) }))
  .action(async ({ parsedInput: { id }, ctx }) => {
    // Verify ownership and read provider/collection from DB (not client input)
    const source = await getExternalSourceById(id, ctx.user.id)
    if (!source) throw new Error("Conector no encontrado")
    if (!source.active) throw new Error("Conector desactivado")

    await externalSyncQueue.add("manual-sync", {
      sourceId: source.id,
      provider: source.provider,
      collectionDest: source.collectionDest,
      fullSync: false,
    })
    return { ok: true }
  })
