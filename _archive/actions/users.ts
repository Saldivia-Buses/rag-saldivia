"use server"
/**
 * Server Actions — Gestión de usuarios
 * Solo admins pueden ejecutar estas acciones (verificado vía requireAdmin).
 */

import { revalidatePath } from "next/cache"
import { requireAdmin } from "@/lib/auth/current-user"
import {
  createUser,
  updateUser,
  deleteUser,
} from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"

export async function actionCreateUser(data: {
  email: string
  name: string
  password: string
  role?: "admin" | "area_manager" | "user"
  areaIds?: number[]
}) {
  const admin = await requireAdmin()

  const user = await createUser(data)

  log.info("user.created", {
    targetUserId: user.id,
    email: user.email,
    role: user.role,
  }, { userId: admin.id })

  revalidatePath("/admin/users")
  return user
}

export async function actionUpdateUser(
  id: number,
  data: { name?: string; role?: "admin" | "area_manager" | "user"; active?: boolean }
) {
  const admin = await requireAdmin()

  const updated = await updateUser(id, data)

  log.info("user.updated", { targetUserId: id, changes: data }, { userId: admin.id })

  revalidatePath("/admin/users")
  return updated
}

export async function actionDeleteUser(id: number) {
  const admin = await requireAdmin()

  await deleteUser(id)

  log.info("user.deleted", { targetUserId: id }, { userId: admin.id })

  revalidatePath("/admin/users")
}
