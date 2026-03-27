"use server"

import { revalidatePath } from "next/cache"
import { requireUser } from "@/lib/auth/current-user"
import { createWebhook, deleteWebhook } from "@rag-saldivia/db"

export async function actionCreateWebhook(data: { url: string; events: string[] }) {
  const user = await requireUser()
  const webhook = await createWebhook({ userId: user.id, url: data.url, events: data.events })
  revalidatePath("/admin/webhooks")
  return webhook
}

export async function actionDeleteWebhook(id: string) {
  const user = await requireUser()
  await deleteWebhook(id, user.id)
  revalidatePath("/admin/webhooks")
}
