"use server"

import { revalidatePath } from "next/cache"
import { requireAdmin } from "@/lib/auth/current-user"
import { createExternalSource, deleteExternalSource } from "@rag-saldivia/db"

export async function actionCreateExternalSource(data: {
  provider: "google_drive" | "sharepoint" | "confluence"
  name: string
  collectionDest: string
  schedule: "hourly" | "daily" | "weekly"
}) {
  const admin = await requireAdmin()
  const source = await createExternalSource({
    provider: data.provider,
    name: data.name,
    collectionDest: data.collectionDest,
    schedule: data.schedule,
    userId: admin.id,
  })
  revalidatePath("/admin/external-sources")
  return source
}

export async function actionDeleteExternalSource(id: string) {
  const admin = await requireAdmin()
  await deleteExternalSource(id, admin.id)
  revalidatePath("/admin/external-sources")
}
