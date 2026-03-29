/**
 * Server actions for admin user management.
 *
 * All actions require admin role. Provides CRUD for users:
 * list, create, update (role/active), reset password, delete.
 *
 * Data flow: AdminUsers component → server action → DB queries → revalidate
 * Depends on: @rag-saldivia/db (users queries), lib/auth/current-user.ts
 */

"use server"

import { revalidatePath } from "next/cache"
import { requireAdmin } from "@/lib/auth/current-user"
import { listUsers, createUser, updateUser, updatePassword, deleteUser } from "@rag-saldivia/db"

/** List all users with their areas. Admin only. */
export async function actionListUsers() {
  await requireAdmin()
  return listUsers()
}

/** Create a new user. Admin only. */
export async function actionCreateUser(data: {
  email: string
  name: string
  password: string
  role?: "admin" | "area_manager" | "user"
}) {
  await requireAdmin()
  const user = await createUser(data)
  revalidatePath("/admin/users")
  return user
}

/** Update user fields (name, role, active status). Admin only. User id=1 is protected. */
export async function actionUpdateUser(
  id: number,
  data: Partial<{ name: string; role: "admin" | "area_manager" | "user"; active: boolean }>
) {
  await requireAdmin()
  if (id === 1 && (data.role || data.active === false)) {
    throw new Error("El usuario admin principal no se puede modificar")
  }
  const updated = await updateUser(id, data)
  revalidatePath("/admin/users")
  return updated
}

/** Reset a user's password. Admin only. */
export async function actionResetPassword(userId: number, newPassword: string) {
  await requireAdmin()
  await updatePassword(userId, newPassword)
  revalidatePath("/admin/users")
}

/** Delete a user permanently. Admin only. User id=1 is protected. */
export async function actionDeleteUser(id: number) {
  await requireAdmin()
  if (id === 1) throw new Error("El usuario admin principal no se puede eliminar")
  await deleteUser(id)
  revalidatePath("/admin/users")
}
