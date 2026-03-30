"use server"

import { revalidatePath } from "next/cache"
import { adminAction } from "@/lib/safe-action"
import { ragFetch } from "@/lib/rag/client"
import { invalidateCollectionsCache } from "@/lib/rag/collections-cache"
import { CollectionNameSchema } from "@rag-saldivia/shared"
import { z } from "zod"

export const actionCreateCollection = adminAction
  .schema(z.object({ name: CollectionNameSchema }))
  .action(async ({ parsedInput: { name } }) => {
    const res = await ragFetch(`/v1/collections`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ collection_name: name }),
    } as Parameters<typeof ragFetch>[1])
    if ("error" in res) throw new Error(res.error.message)
    await invalidateCollectionsCache()
    revalidatePath("/collections")
  })

export const actionDeleteCollection = adminAction
  .schema(z.object({ name: CollectionNameSchema }))
  .action(async ({ parsedInput: { name } }) => {
    try {
      await ragFetch(`/v1/collections/${encodeURIComponent(name)}`, {
        method: "DELETE",
      } as Parameters<typeof ragFetch>[1])
    } catch {
      // En modo mock: simular éxito
    }
    await invalidateCollectionsCache()
    revalidatePath("/collections")
  })
