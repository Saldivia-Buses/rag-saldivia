/**
 * Server actions for RBAC role management.
 * Uses authAction with inline permission checks.
 */

"use server"

import { z } from "zod"
import { revalidatePath } from "next/cache"
import { authAction, clean } from "@/lib/safe-action"
import { checkPermission } from "@/lib/auth/permissions"
import {
  listRoles,
  createRole,
  updateRole,
  deleteRole,
  listPermissions,
  setRolePermissions,
  setUserRoles,
  getRolePermissionKeys,
} from "@rag-saldivia/db"

async function requirePerm(userId: number, key: string) {
  const has = await checkPermission(userId, key)
  if (!has) throw new Error(`Missing permission: ${key}`)
}

export const actionListRoles = authAction
  .action(async ({ ctx: { user } }) => {
    await requirePerm(user.id, "admin.access")
    return listRoles()
  })

export const actionListPermissions = authAction
  .action(async ({ ctx: { user } }) => {
    await requirePerm(user.id, "admin.access")
    return listPermissions()
  })

export const actionCreateRole = authAction
  .schema(z.object({
    name: z.string().min(1),
    description: z.string().optional(),
    level: z.number(),
    color: z.string().optional(),
    icon: z.string().optional(),
  }))
  .action(async ({ parsedInput: data, ctx: { user } }) => {
    await requirePerm(user.id, "roles.manage")
    const role = await createRole(clean(data))
    revalidatePath("/admin")
    return role
  })

export const actionUpdateRole = authAction
  .schema(z.object({
    id: z.number(),
    data: z.object({
      name: z.string().optional(),
      description: z.string().optional(),
      level: z.number().optional(),
      color: z.string().optional(),
      icon: z.string().optional(),
    }),
  }))
  .action(async ({ parsedInput: { id, data }, ctx: { user } }) => {
    await requirePerm(user.id, "roles.manage")
    const updated = await updateRole(id, clean(data))
    revalidatePath("/admin")
    return updated
  })

export const actionDeleteRole = authAction
  .schema(z.object({ id: z.number() }))
  .action(async ({ parsedInput: { id }, ctx: { user } }) => {
    await requirePerm(user.id, "roles.manage")
    await deleteRole(id)
    revalidatePath("/admin")
  })

export const actionSetRolePermissions = authAction
  .schema(z.object({ roleId: z.number(), permissionKeys: z.array(z.string()) }))
  .action(async ({ parsedInput: { roleId, permissionKeys }, ctx: { user } }) => {
    await requirePerm(user.id, "roles.manage")
    const { getRoleById } = await import("@rag-saldivia/db")
    const role = await getRoleById(roleId)
    if (role && role.isSystem && role.level >= 100) {
      throw new Error("Los permisos del rol Admin no se pueden modificar")
    }
    await setRolePermissions(roleId, permissionKeys)
    revalidatePath("/admin")
  })

export const actionSetUserRoles = authAction
  .schema(z.object({ userId: z.number(), roleIds: z.array(z.number()) }))
  .action(async ({ parsedInput: { userId, roleIds }, ctx: { user } }) => {
    await requirePerm(user.id, "users.manage")
    if (userId === 1) throw new Error("El usuario admin no se puede modificar")
    await setUserRoles(userId, roleIds)
    revalidatePath("/admin")
  })

export const actionGetRolePermissions = authAction
  .schema(z.object({ roleId: z.number() }))
  .action(async ({ parsedInput: { roleId }, ctx: { user } }) => {
    await requirePerm(user.id, "admin.access")
    return getRolePermissionKeys(roleId)
  })
