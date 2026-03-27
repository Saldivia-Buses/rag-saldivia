"use server"

import { revalidatePath } from "next/cache"
import { requireUser } from "@/lib/auth/current-user"
import { createProject, deleteProject } from "@rag-saldivia/db"

export async function actionCreateProject(data: {
  name: string
  description?: string
  instructions?: string
}) {
  const user = await requireUser()
  const project = await createProject({ userId: user.id, ...data })
  revalidatePath("/projects")
  return project
}

export async function actionDeleteProject(id: string) {
  const user = await requireUser()
  await deleteProject(id, user.id)
  revalidatePath("/projects")
}
