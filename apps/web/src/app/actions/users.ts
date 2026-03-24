"use server"
/**
 * Server Actions — Gestión de usuarios
 * Solo admins pueden ejecutar estas acciones (verificado vía requireAdmin).
 */

import { revalidatePath } from "next/cache"
import { requireAdmin } from "@/lib/auth/current-user"
import {
  listUsers,
  getUserById,
  createUser,
  updateUser,
  deleteUser,
  addUserArea,
  removeUserArea,
  updatePassword,
} from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"

export async function actionListUsers() {
  await requireAdmin()
  return listUsers()
}

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

export async function actionAssignArea(userId: number, areaId: number) {
  const admin = await requireAdmin()
  await addUserArea(userId, areaId)
  log.info("user.area_assigned", { targetUserId: userId, areaId }, { userId: admin.id })
  revalidatePath("/admin/users")
}

export async function actionRemoveArea(userId: number, areaId: number) {
  const admin = await requireAdmin()
  await removeUserArea(userId, areaId)
  log.info("user.area_removed", { targetUserId: userId, areaId }, { userId: admin.id })
  revalidatePath("/admin/users")
}

export async function actionUpdatePassword(userId: number, newPassword: string) {
  const admin = await requireAdmin()
  await updatePassword(userId, newPassword)
  log.info("auth.password_changed", { targetUserId: userId }, { userId: admin.id })
}
