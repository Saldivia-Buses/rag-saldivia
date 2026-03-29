/**
 * Server actions for RBAC role management.
 *
 * All actions require appropriate permissions via requirePermission().
 *
 * Data flow: Admin UI → server action → DB queries → revalidate
 * Depends on: @rag-saldivia/db (RBAC queries), lib/auth/permissions.ts
 */

"use server"

import { revalidatePath } from "next/cache"
import { requirePermission } from "@/lib/auth/permissions"
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

/** List all roles with user counts. Requires admin.access. */
export async function actionListRoles() {
  await requirePermission("admin.access")
  return listRoles()
}

/** List all permissions. Requires admin.access. */
export async function actionListPermissions() {
  await requirePermission("admin.access")
  return listPermissions()
}

/** Create a custom role. Requires roles.manage. */
export async function actionCreateRole(data: {
  name: string
  description?: string
  level: number
  color?: string
  icon?: string
}) {
  await requirePermission("roles.manage")
  const role = await createRole(data)
  revalidatePath("/admin")
  return role
}

/** Update a role's metadata. Requires roles.manage. */
export async function actionUpdateRole(
  id: number,
  data: Partial<{ name: string; description: string; level: number; color: string; icon: string }>
) {
  await requirePermission("roles.manage")
  const updated = await updateRole(id, data)
  revalidatePath("/admin")
  return updated
}

/** Delete a role. Requires roles.manage. Cannot delete system roles. */
export async function actionDeleteRole(id: number) {
  await requirePermission("roles.manage")
  await deleteRole(id)
  revalidatePath("/admin")
}

/** Set permissions for a role. Requires roles.manage. Admin role (level 100) is protected. */
export async function actionSetRolePermissions(roleId: number, permissionKeys: string[]) {
  await requirePermission("roles.manage")
  // Protect the Admin role — get it by checking if it's the system role with level 100
  const { getRoleById } = await import("@rag-saldivia/db")
  const role = await getRoleById(roleId)
  if (role && role.isSystem && role.level >= 100) {
    throw new Error("Los permisos del rol Admin no se pueden modificar")
  }
  await setRolePermissions(roleId, permissionKeys)
  revalidatePath("/admin")
}

/** Assign roles to a user. Requires users.manage. Admin user (id=1) is protected. */
export async function actionSetUserRoles(userId: number, roleIds: number[]) {
  await requirePermission("users.manage")
  if (userId === 1) throw new Error("El usuario admin no se puede modificar")
  await setUserRoles(userId, roleIds)
  revalidatePath("/admin")
}

/** Get permission keys for a role. Requires admin.access. */
export async function actionGetRolePermissions(roleId: number) {
  await requirePermission("admin.access")
  return getRolePermissionKeys(roleId)
}
