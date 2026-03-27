"use server"

import { revalidatePath, updateTag } from "next/cache"
import { requireAdmin } from "@/lib/auth/current-user"
import { ragFetch } from "@/lib/rag/client"

export async function actionCreateCollection(name: string) {
  await requireAdmin()
  const res = await ragFetch(`/v1/collections`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ collection_name: name }),
  } as Parameters<typeof ragFetch>[1])
  if ("error" in res) throw new Error(res.error.message)
  updateTag("collections")
  revalidatePath("/collections")
}

export async function actionDeleteCollection(name: string) {
  await requireAdmin()
  try {
    await ragFetch(`/v1/collections/${encodeURIComponent(name)}`, {
      method: "DELETE",
    } as Parameters<typeof ragFetch>[1])
  } catch {
    // En modo mock: simular éxito
  }
  updateTag("collections")
  revalidatePath("/collections")
}
