/**
 * Permission checking utility for server-side code.
 *
 * Usage in server actions:
 *   const user = await requirePermission("users.manage")
 *
 * Usage in API routes:
 *   if (!await checkPermission(userId, "collections.manage")) return apiError(...)
 *
 * Depends on: @rag-saldivia/db (RBAC queries), lib/auth/current-user.ts
 */

import { getUserEffectivePermissions, hasPermission as dbHasPermission } from "@rag-saldivia/db"
import { requireUser, type CurrentUser } from "./current-user"

/** Check if a specific user has a permission. */
export async function checkPermission(userId: number, permKey: string): Promise<boolean> {
  return dbHasPermission(userId, permKey)
}

/**
 * Require a permission for the current user.
 * Throws if the user doesn't have the permission.
 */
export async function requirePermission(permKey: string): Promise<CurrentUser> {
  const user = await requireUser()
  const has = await dbHasPermission(user.id, permKey)
  if (!has) throw new Error(`Missing permission: ${permKey}`)
  return user
}

/** Get all effective permissions for the current user. */
export async function getCurrentPermissions(): Promise<Set<string>> {
  const user = await requireUser()
  return getUserEffectivePermissions(user.id)
}
