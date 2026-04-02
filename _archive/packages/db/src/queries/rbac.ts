/**
 * RBAC queries — roles, permissions, and resolution.
 *
 * Provides the data layer for the granular permission system:
 *   - CRUD for roles (with system role protection)
 *   - Permission catalog and role-permission matrix
 *   - User ↔ Role assignments (many-to-many)
 *   - Permission resolution (union of all user roles)
 *   - Legacy fallback from users.role field
 *
 * Used by: actions/roles.ts, lib/auth/permissions.ts
 * Depends on: schema.ts (roles, permissions, rolePermissions, userRoleAssignments)
 */

import { eq, sql, and, inArray, count } from "drizzle-orm"
import { getDb } from "../connection"
import {
  roles,
  permissions,
  rolePermissions,
  userRoleAssignments,
  users,
  chatSessions,
  chatMessages,
} from "../schema"

function now() {
  return Date.now()
}

// ── Roles ────────────────────────────────────────────────────────────────

/** List all roles with count of assigned users. */
export async function listRoles() {
  const db = getDb()
  const allRoles = await db.select().from(roles).orderBy(roles.level)
  const counts = await db
    .select({ roleId: userRoleAssignments.roleId, count: count() })
    .from(userRoleAssignments)
    .groupBy(userRoleAssignments.roleId)

  const countMap = Object.fromEntries(counts.map((c) => [c.roleId, c.count]))
  return allRoles.map((r) => ({ ...r, userCount: countMap[r.id] ?? 0 }))
}

/** Get a single role with its permission keys. */
export async function getRoleById(id: number) {
  const db = getDb()
  const role = await db.query.roles.findFirst({
    where: (r, { eq }) => eq(r.id, id),
    with: {
      rolePermissions: {
        with: { permission: true },
      },
    },
  })
  if (!role) return null
  return {
    ...role,
    permissionKeys: role.rolePermissions.map((rp) => rp.permission.key),
  }
}

/** Create a custom role. */
export async function createRole(data: {
  name: string
  description?: string
  level: number
  color?: string
  icon?: string
}) {
  const [role] = await getDb()
    .insert(roles)
    .values({
      name: data.name,
      description: data.description ?? "",
      level: data.level,
      color: data.color ?? "#6e6c69",
      icon: data.icon ?? "user",
      isSystem: false,
      createdAt: now(),
    })
    .returning()
  return role!
}

/** Update a role's metadata. */
export async function updateRole(
  id: number,
  data: Partial<{ name: string; description: string; level: number; color: string; icon: string }>
) {
  const [updated] = await getDb()
    .update(roles)
    .set(data)
    .where(eq(roles.id, id))
    .returning()
  return updated
}

/** Delete a role. Fails if it's a system role or has users assigned. */
export async function deleteRole(id: number) {
  const db = getDb()
  const role = await db.query.roles.findFirst({ where: (r, { eq }) => eq(r.id, id) })
  if (!role) throw new Error("Role not found")
  if (role.isSystem) throw new Error("Cannot delete system role")

  const [assignment] = await db
    .select({ count: count() })
    .from(userRoleAssignments)
    .where(eq(userRoleAssignments.roleId, id))

  if (assignment && assignment.count > 0) {
    throw new Error("Cannot delete role with assigned users")
  }

  await db.delete(roles).where(eq(roles.id, id))
}

// ── Permissions ──────────────────────────────────────────────────────────

/** List all permissions, ordered by category. */
export async function listPermissions() {
  return getDb().select().from(permissions).orderBy(permissions.category, permissions.key)
}

/** Get permission keys for a role. */
export async function getRolePermissionKeys(roleId: number): Promise<string[]> {
  const db = getDb()
  const rows = await db
    .select({ key: permissions.key })
    .from(rolePermissions)
    .innerJoin(permissions, eq(rolePermissions.permissionId, permissions.id))
    .where(eq(rolePermissions.roleId, roleId))

  return rows.map((r) => r.key)
}

/** Replace all permissions for a role (transactional). */
export async function setRolePermissions(roleId: number, permissionKeys: string[]) {
  const db = getDb()

  // Get permission IDs from keys
  const perms = permissionKeys.length > 0
    ? await db.select().from(permissions).where(inArray(permissions.key, permissionKeys))
    : []

  // Replace in transaction
  await db.transaction(async (tx) => {
    await tx.delete(rolePermissions).where(eq(rolePermissions.roleId, roleId))
    if (perms.length > 0) {
      await tx.insert(rolePermissions).values(
        perms.map((p) => ({ roleId, permissionId: p.id }))
      )
    }
  })
}

// ── User ↔ Roles ─────────────────────────────────────────────────────────

/** Get all roles assigned to a user. */
export async function getUserRoles(userId: number) {
  const db = getDb()
  const rows = await db
    .select()
    .from(userRoleAssignments)
    .innerJoin(roles, eq(userRoleAssignments.roleId, roles.id))
    .where(eq(userRoleAssignments.userId, userId))

  return rows.map((r) => r.roles)
}

/** Replace all roles for a user (transactional). */
export async function setUserRoles(userId: number, roleIds: number[]) {
  const db = getDb()
  await db.transaction(async (tx) => {
    await tx.delete(userRoleAssignments).where(eq(userRoleAssignments.userId, userId))
    if (roleIds.length > 0) {
      await tx.insert(userRoleAssignments).values(
        roleIds.map((roleId) => ({ userId, roleId, assignedAt: now() }))
      )
    }
  })
}

// ── Resolution ───────────────────────────────────────────────────────────

/**
 * Get the effective permission set for a user — union of all assigned roles.
 * Falls back to users.role if no RBAC assignments exist.
 */
export async function getUserEffectivePermissions(userId: number): Promise<Set<string>> {
  const db = getDb()

  // Get roles via RBAC assignments
  const userRoles = await getUserRoles(userId)

  // Fallback: if no RBAC roles, use legacy users.role field
  if (userRoles.length === 0) {
    const user = await db.query.users.findFirst({
      where: (u, { eq }) => eq(u.id, userId),
      columns: { role: true },
    })
    if (!user) return new Set()

    const legacyMap: Record<string, string> = { admin: "Admin", area_manager: "Manager", user: "Usuario" }
    const roleName = legacyMap[user.role] ?? "Usuario"
    const role = await db.query.roles.findFirst({ where: (r, { eq }) => eq(r.name, roleName) })
    if (!role) return new Set()

    const keys = await getRolePermissionKeys(role.id)
    return new Set(keys)
  }

  // Union of all role permissions
  const allKeys = new Set<string>()
  for (const role of userRoles) {
    const keys = await getRolePermissionKeys(role.id)
    for (const k of keys) allKeys.add(k)
  }
  return allKeys
}

/** Check if a user has a specific permission. */
export async function hasPermission(userId: number, permKey: string): Promise<boolean> {
  const perms = await getUserEffectivePermissions(userId)
  return perms.has(permKey)
}

/** Get the primary role (highest level) for a user. */
export async function getUserPrimaryRole(userId: number) {
  const userRoles = await getUserRoles(userId)
  if (userRoles.length === 0) return null
  return userRoles.reduce((highest, r) => (r.level > highest.level ? r : highest))
}

// ── Stats (for admin dashboard) ──────────────────────────────────────────

export async function countUsers() {
  const db = getDb()
  const [total] = await db.select({ count: count() }).from(users)
  const [active] = await db.select({ count: count() }).from(users).where(eq(users.active, true))
  const [inactive] = await db.select({ count: count() }).from(users).where(eq(users.active, false))
  return {
    total: total?.count ?? 0,
    active: active?.count ?? 0,
    inactive: inactive?.count ?? 0,
  }
}

export async function countSessions() {
  const [result] = await getDb().select({ count: count() }).from(chatSessions)
  return result?.count ?? 0
}

export async function countMessages() {
  const [result] = await getDb().select({ count: count() }).from(chatMessages)
  return result?.count ?? 0
}

// ── Presence ─────────────────────────────────────────────────────────────

/** Update lastSeen for a user (called on each authenticated request). */
export async function touchUserPresence(userId: number) {
  await getDb()
    .update(users)
    .set({ lastSeen: Date.now() })
    .where(eq(users.id, userId))
}

/** Get online users (lastSeen within the last 2 minutes). */
export async function getOnlineUsers() {
  const threshold = Date.now() - 2 * 60 * 1000
  const db = getDb()
  return db
    .select({
      id: users.id,
      name: users.name,
      email: users.email,
      lastSeen: users.lastSeen,
    })
    .from(users)
    .where(
      and(
        eq(users.active, true),
        sql`${users.lastSeen} > ${threshold}`
      )
    )
}

/** Get all users with their lastSeen for the admin dashboard. */
export async function getUsersPresence() {
  return getDb()
    .select({
      id: users.id,
      name: users.name,
      email: users.email,
      lastSeen: users.lastSeen,
      active: users.active,
    })
    .from(users)
    .orderBy(sql`${users.lastSeen} DESC NULLS LAST`)
}
