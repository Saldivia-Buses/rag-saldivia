"use server"

import { revalidatePath } from "next/cache"
import { requireAdmin } from "@/lib/auth/current-user"
import { createReport, deleteReport } from "@rag-saldivia/db"

export async function actionCreateReport(data: {
  query: string
  collection: string
  schedule: "daily" | "weekly" | "monthly"
  destination: "saved" | "email"
  email?: string
}) {
  const admin = await requireAdmin()
  const report = await createReport({
    query: data.query,
    collection: data.collection,
    schedule: data.schedule,
    destination: data.destination,
    email: data.email ?? null,
    userId: admin.id,
  })
  revalidatePath("/admin/reports")
  return report
}

export async function actionDeleteReport(id: string) {
  const admin = await requireAdmin()
  await deleteReport(id, admin.id)
  revalidatePath("/admin/reports")
}
