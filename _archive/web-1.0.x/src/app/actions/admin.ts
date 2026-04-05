/**
 * Server actions for admin user management.
 * All actions require admin role via adminAction middleware.
 */

"use server"

import { z } from "zod"
import { revalidatePath } from "next/cache"
import { adminAction, clean } from "@/lib/safe-action"
import { listUsers, createUser, updateUser, updatePassword, deleteUser } from "@rag-saldivia/db"

const RoleEnum = z.enum(["admin", "area_manager", "user"])

export const actionListUsers = adminAction
  .action(async () => {
    return listUsers()
  })

export const actionCreateUser = adminAction
  .schema(z.object({
    email: z.string().email(),
    name: z.string().min(1),
    password: z.string().min(8),
    role: RoleEnum.optional(),
  }))
  .action(async ({ parsedInput: data }) => {
    const user = await createUser(clean(data))
    revalidatePath("/admin/users")
    return user
  })

export const actionUpdateUser = adminAction
  .schema(z.object({
    id: z.number(),
    data: z.object({
      name: z.string().optional(),
      role: RoleEnum.optional(),
      active: z.boolean().optional(),
    }),
  }))
  .action(async ({ parsedInput: { id, data } }) => {
    if (id === 1 && (data.role || data.active === false)) {
      throw new Error("El usuario admin principal no se puede modificar")
    }
    const updated = await updateUser(id, clean(data))
    revalidatePath("/admin/users")
    return updated
  })

export const actionResetPassword = adminAction
  .schema(z.object({ userId: z.number(), newPassword: z.string().min(8) }))
  .action(async ({ parsedInput: { userId, newPassword } }) => {
    await updatePassword(userId, newPassword)
    revalidatePath("/admin/users")
  })

export const actionDeleteUser = adminAction
  .schema(z.object({ id: z.number() }))
  .action(async ({ parsedInput: { id } }) => {
    if (id === 1) throw new Error("El usuario admin principal no se puede eliminar")
    await deleteUser(id)
    revalidatePath("/admin/users")
  })
