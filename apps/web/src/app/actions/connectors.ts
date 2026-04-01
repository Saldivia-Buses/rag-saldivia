"use server"

import { z } from "zod"
import { adminAction } from "@/lib/safe-action"
import {
  createExternalSource,
  listExternalSources,
  deleteExternalSource,
  countSyncDocuments,
  deleteSyncDocumentsForSource,
} from "@rag-saldivia/db"
import { ConnectorProviderSchema } from "@rag-saldivia/shared"
import { externalSyncQueue } from "@/lib/queue"
import { revalidatePath } from "next/cache"

export const actionListConnectors = adminAction
  .schema(z.object({}))
  .action(async ({ ctx }) => {
    const sources = await listExternalSources(ctx.user.id)
    // Add doc count for each source
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
      credentials: z.string().min(1), // JSON string with provider-specific creds
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
    const { getDb } = await import("@rag-saldivia/db")
    const { externalSources } = await import("@rag-saldivia/db")
    const { eq, and } = await import("drizzle-orm")
    const db = getDb()
    await db
      .update(externalSources)
      .set({ active })
      .where(and(eq(externalSources.id, id), eq(externalSources.userId, ctx.user.id)))
    revalidatePath("/admin/connectors")
    return { ok: true }
  })

export const actionSyncNow = adminAction
  .schema(z.object({ id: z.string().min(1), provider: z.string().min(1), collectionDest: z.string().min(1) }))
  .action(async ({ parsedInput: { id, provider, collectionDest } }) => {
    await externalSyncQueue.add("manual-sync", {
      sourceId: id,
      provider,
      collectionDest,
      fullSync: false,
    })
    return { ok: true }
  })
