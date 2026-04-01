/**
 * Tests for RBAC queries — roles, permissions, user-role assignments, and resolution.
 *
 * This is the most critical untested module: bugs here mean permission bypass.
 * Runs with: bun test packages/db/src/__tests__/rbac.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser, insertRole, insertPermission, insertSession, insertMessage } from "./setup"
import {
  listRoles,
  getRoleById,
  createRole,
  updateRole,
  deleteRole,
  listPermissions,
  getRolePermissionKeys,
  setRolePermissions,
  getUserRoles,
  setUserRoles,
  getUserEffectivePermissions,
  hasPermission,
  getUserPrimaryRole,
  countUsers,
  countSessions,
  countMessages,
  touchUserPresence,
  getOnlineUsers,
  getUsersPresence,
} from "../queries/rbac"
import { eq } from "drizzle-orm"
import { users } from "../schema"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => {
  _resetDbForTesting()
})

afterEach(async () => {
  await client.executeMultiple(
    "DELETE FROM user_role_assignments; DELETE FROM role_permissions; DELETE FROM permissions; DELETE FROM roles; DELETE FROM chat_messages; DELETE FROM chat_sessions; DELETE FROM users;"
  )
})

// ── Roles CRUD ────────────────────────────────────────────────────────────────

describe("createRole", () => {
  test("creates a role with provided values and correct defaults", async () => {
    const role = await createRole({ name: "Editor", level: 50 })
    expect(role.name).toBe("Editor")
    expect(role.level).toBe(50)
    expect(role.isSystem).toBe(false)
    expect(role.color).toBe("#6e6c69")
    expect(role.icon).toBe("user")
    expect(role.description).toBe("")
    expect(role.createdAt).toBeGreaterThan(0)
  })

  test("creates a role with custom color and icon", async () => {
    const role = await createRole({ name: "VIP", level: 30, color: "#ff0000", icon: "star" })
    expect(role.color).toBe("#ff0000")
    expect(role.icon).toBe("star")
  })
})

describe("listRoles", () => {
  test("returns roles ordered by level with userCount=0 for new roles", async () => {
    await insertRole(db, "Admin", 100, true)
    await insertRole(db, "User", 10, true)
    await insertRole(db, "Manager", 50)

    const result = await listRoles()
    expect(result).toHaveLength(3)
    // Ordered by level ascending
    expect(result[0]!.name).toBe("User")
    expect(result[1]!.name).toBe("Manager")
    expect(result[2]!.name).toBe("Admin")
    // All have userCount 0
    for (const r of result) {
      expect(r.userCount).toBe(0)
    }
  })

  test("returns correct userCount when users are assigned", async () => {
    const role = await insertRole(db, "Admin", 100)
    const u1 = await insertUser(db, "u1@test.com")
    const u2 = await insertUser(db, "u2@test.com")
    await setUserRoles(u1.id, [role.id])
    await setUserRoles(u2.id, [role.id])

    const result = await listRoles()
    expect(result[0]!.userCount).toBe(2)
  })
})

describe("getRoleById", () => {
  test("returns role with permissionKeys", async () => {
    const role = await insertRole(db, "Admin", 100)
    const p1 = await insertPermission(db, "users.manage")
    const p2 = await insertPermission(db, "chat.manage")
    await setRolePermissions(role.id, [p1.key, p2.key])

    const result = await getRoleById(role.id)
    expect(result).not.toBeNull()
    expect(result!.name).toBe("Admin")
    expect(result!.permissionKeys).toContain("users.manage")
    expect(result!.permissionKeys).toContain("chat.manage")
    expect(result!.permissionKeys).toHaveLength(2)
  })

  test("returns null for non-existent role", async () => {
    const result = await getRoleById(99999)
    expect(result).toBeNull()
  })
})

describe("updateRole", () => {
  test("updates partial fields without affecting others", async () => {
    const role = await createRole({ name: "Editor", level: 50, description: "Original" })
    const updated = await updateRole(role.id, { description: "Updated", color: "#00ff00" })

    expect(updated!.description).toBe("Updated")
    expect(updated!.color).toBe("#00ff00")
    expect(updated!.name).toBe("Editor")
    expect(updated!.level).toBe(50)
  })
})

describe("deleteRole", () => {
  test("throws for system role", async () => {
    const sysRole = await insertRole(db, "Admin", 100, true)
    expect(deleteRole(sysRole.id)).rejects.toThrow("Cannot delete system role")
  })

  test("throws for role with assigned users", async () => {
    const role = await insertRole(db, "Custom", 30)
    const user = await insertUser(db)
    await setUserRoles(user.id, [role.id])

    expect(deleteRole(role.id)).rejects.toThrow("Cannot delete role with assigned users")
  })

  test("deletes unassigned non-system role successfully", async () => {
    const role = await insertRole(db, "Temporary", 10)
    await deleteRole(role.id)

    const result = await getRoleById(role.id)
    expect(result).toBeNull()
  })

  test("throws for non-existent role", async () => {
    expect(deleteRole(99999)).rejects.toThrow("Role not found")
  })
})

// ── Permissions ───────────────────────────────────────────────────────────────

describe("listPermissions", () => {
  test("returns permissions ordered by category then key", async () => {
    await insertPermission(db, "users.manage", "Manage users", "Admin")
    await insertPermission(db, "chat.read", "Read chat", "Chat")
    await insertPermission(db, "chat.write", "Write chat", "Chat")
    await insertPermission(db, "admin.config", "Config", "Admin")

    const result = await listPermissions()
    expect(result).toHaveLength(4)
    // Admin comes before Chat alphabetically
    expect(result[0]!.key).toBe("admin.config")
    expect(result[1]!.key).toBe("users.manage")
    expect(result[2]!.key).toBe("chat.read")
    expect(result[3]!.key).toBe("chat.write")
  })
})

describe("getRolePermissionKeys", () => {
  test("returns permission keys for a role", async () => {
    const role = await insertRole(db, "Editor", 50)
    const p1 = await insertPermission(db, "docs.read")
    const p2 = await insertPermission(db, "docs.write")
    await setRolePermissions(role.id, [p1.key, p2.key])

    const keys = await getRolePermissionKeys(role.id)
    expect(keys).toContain("docs.read")
    expect(keys).toContain("docs.write")
    expect(keys).toHaveLength(2)
  })

  test("returns empty array for role with no permissions", async () => {
    const role = await insertRole(db, "Empty", 0)
    const keys = await getRolePermissionKeys(role.id)
    expect(keys).toEqual([])
  })
})

describe("setRolePermissions", () => {
  test("replaces all permissions for a role", async () => {
    const role = await insertRole(db, "Editor", 50)
    const p1 = await insertPermission(db, "a.read")
    const p2 = await insertPermission(db, "b.write")
    const p3 = await insertPermission(db, "c.manage")

    // Set initial permissions
    await setRolePermissions(role.id, [p1.key, p2.key])
    let keys = await getRolePermissionKeys(role.id)
    expect(keys).toHaveLength(2)

    // Replace with different set
    await setRolePermissions(role.id, [p2.key, p3.key])
    keys = await getRolePermissionKeys(role.id)
    expect(keys).toHaveLength(2)
    expect(keys).toContain("b.write")
    expect(keys).toContain("c.manage")
    expect(keys).not.toContain("a.read")
  })

  test("clears all permissions when given empty array", async () => {
    const role = await insertRole(db, "Editor", 50)
    const p1 = await insertPermission(db, "docs.read")
    await setRolePermissions(role.id, [p1.key])

    // Clear all
    await setRolePermissions(role.id, [])
    const keys = await getRolePermissionKeys(role.id)
    expect(keys).toEqual([])
  })
})

// ── User-Role ─────────────────────────────────────────────────────────────────

describe("getUserRoles", () => {
  test("returns assigned roles for user", async () => {
    const user = await insertUser(db)
    const role1 = await insertRole(db, "Admin", 100)
    const role2 = await insertRole(db, "Editor", 50)
    await setUserRoles(user.id, [role1.id, role2.id])

    const roles = await getUserRoles(user.id)
    expect(roles).toHaveLength(2)
    const names = roles.map((r) => r.name)
    expect(names).toContain("Admin")
    expect(names).toContain("Editor")
  })

  test("returns empty array for user without roles", async () => {
    const user = await insertUser(db)
    const roles = await getUserRoles(user.id)
    expect(roles).toEqual([])
  })
})

describe("setUserRoles", () => {
  test("replaces all roles for a user", async () => {
    const user = await insertUser(db)
    const role1 = await insertRole(db, "Admin", 100)
    const role2 = await insertRole(db, "Editor", 50)
    const role3 = await insertRole(db, "Viewer", 10)

    await setUserRoles(user.id, [role1.id, role2.id])
    let roles = await getUserRoles(user.id)
    expect(roles).toHaveLength(2)

    // Replace with different set
    await setUserRoles(user.id, [role3.id])
    roles = await getUserRoles(user.id)
    expect(roles).toHaveLength(1)
    expect(roles[0]!.name).toBe("Viewer")
  })

  test("removes all roles when given empty array", async () => {
    const user = await insertUser(db)
    const role = await insertRole(db, "Admin", 100)
    await setUserRoles(user.id, [role.id])

    await setUserRoles(user.id, [])
    const roles = await getUserRoles(user.id)
    expect(roles).toEqual([])
  })
})

// ── Resolution — MOST CRITICAL ───────────────────────────────────────────────

describe("getUserEffectivePermissions", () => {
  test("returns union of permissions from multiple roles", async () => {
    const user = await insertUser(db)
    const roleA = await insertRole(db, "RoleA", 50)
    const roleB = await insertRole(db, "RoleB", 30)
    const p1 = await insertPermission(db, "users.read")
    const p2 = await insertPermission(db, "users.write")
    const p3 = await insertPermission(db, "chat.read")

    // RoleA gets p1, p2. RoleB gets p2, p3. Union should be all three.
    await setRolePermissions(roleA.id, [p1.key, p2.key])
    await setRolePermissions(roleB.id, [p2.key, p3.key])
    await setUserRoles(user.id, [roleA.id, roleB.id])

    const perms = await getUserEffectivePermissions(user.id)
    expect(perms.size).toBe(3)
    expect(perms.has("users.read")).toBe(true)
    expect(perms.has("users.write")).toBe(true)
    expect(perms.has("chat.read")).toBe(true)
  })

  test("falls back to legacy users.role field when no RBAC assignments exist", async () => {
    // Create a user with legacy role "admin"
    const user = await insertUser(db, "admin@test.com", "admin")

    // Create the "Admin" role with a permission (this is the legacy mapping target)
    const adminRole = await insertRole(db, "Admin", 100)
    const perm = await insertPermission(db, "admin.dashboard")
    await setRolePermissions(adminRole.id, [perm.key])

    // No RBAC assignments — should fall back to legacy
    const perms = await getUserEffectivePermissions(user.id)
    expect(perms.size).toBe(1)
    expect(perms.has("admin.dashboard")).toBe(true)
  })

  test("falls back correctly for area_manager legacy role", async () => {
    const user = await insertUser(db, "manager@test.com", "area_manager")

    // Legacy mapping: area_manager -> "Manager"
    const managerRole = await insertRole(db, "Manager", 50)
    const perm = await insertPermission(db, "areas.manage")
    await setRolePermissions(managerRole.id, [perm.key])

    const perms = await getUserEffectivePermissions(user.id)
    expect(perms.size).toBe(1)
    expect(perms.has("areas.manage")).toBe(true)
  })

  test("falls back to 'Usuario' role for legacy user role", async () => {
    const user = await insertUser(db, "regular@test.com", "user")

    // Legacy mapping: user -> "Usuario"
    const userRole = await insertRole(db, "Usuario", 10)
    const perm = await insertPermission(db, "chat.read")
    await setRolePermissions(userRole.id, [perm.key])

    const perms = await getUserEffectivePermissions(user.id)
    expect(perms.size).toBe(1)
    expect(perms.has("chat.read")).toBe(true)
  })

  test("returns empty Set for user with no roles and no matching legacy role", async () => {
    const user = await insertUser(db)
    // No RBAC assignments, and no "Usuario" role created
    const perms = await getUserEffectivePermissions(user.id)
    expect(perms.size).toBe(0)
  })

  test("returns empty Set for non-existent user", async () => {
    const perms = await getUserEffectivePermissions(99999)
    expect(perms.size).toBe(0)
  })
})

describe("hasPermission", () => {
  test("returns true when user has the permission", async () => {
    const user = await insertUser(db)
    const role = await insertRole(db, "Admin", 100)
    const perm = await insertPermission(db, "users.manage")
    await setRolePermissions(role.id, [perm.key])
    await setUserRoles(user.id, [role.id])

    expect(await hasPermission(user.id, "users.manage")).toBe(true)
  })

  test("returns false when user does not have the permission", async () => {
    const user = await insertUser(db)
    const role = await insertRole(db, "Viewer", 10)
    const p1 = await insertPermission(db, "chat.read")
    await insertPermission(db, "users.manage") // exists but not assigned
    await setRolePermissions(role.id, [p1.key])
    await setUserRoles(user.id, [role.id])

    expect(await hasPermission(user.id, "users.manage")).toBe(false)
  })

  test("returns false for user with no roles at all", async () => {
    const user = await insertUser(db)
    expect(await hasPermission(user.id, "anything")).toBe(false)
  })
})

describe("getUserPrimaryRole", () => {
  test("returns the highest level role when user has multiple", async () => {
    const user = await insertUser(db)
    const low = await insertRole(db, "Viewer", 10)
    const high = await insertRole(db, "Admin", 100)
    const mid = await insertRole(db, "Editor", 50)
    await setUserRoles(user.id, [low.id, high.id, mid.id])

    const primary = await getUserPrimaryRole(user.id)
    expect(primary).not.toBeNull()
    expect(primary!.name).toBe("Admin")
    expect(primary!.level).toBe(100)
  })

  test("returns null for user with no roles", async () => {
    const user = await insertUser(db)
    const primary = await getUserPrimaryRole(user.id)
    expect(primary).toBeNull()
  })
})

// ── Stats ─────────────────────────────────────────────────────────────────────

describe("countUsers", () => {
  test("returns total, active, and inactive counts", async () => {
    await insertUser(db, "a@test.com") // active by default
    await insertUser(db, "b@test.com")
    // Insert an inactive user manually
    await db.insert(users).values({
      email: "c@test.com",
      name: "Inactive",
      role: "user",
      apiKeyHash: "hash",
      preferences: {},
      active: false,
      createdAt: Date.now(),
    })

    const counts = await countUsers()
    expect(counts.total).toBe(3)
    expect(counts.active).toBe(2)
    expect(counts.inactive).toBe(1)
  })
})

describe("countSessions", () => {
  test("returns correct session count", async () => {
    const user = await insertUser(db)
    await insertSession(db, user.id)
    await insertSession(db, user.id)

    const count = await countSessions()
    expect(count).toBe(2)
  })

  test("returns 0 when no sessions exist", async () => {
    const count = await countSessions()
    expect(count).toBe(0)
  })
})

describe("countMessages", () => {
  test("returns correct message count", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    await insertMessage(db, session.id, "user", "Hello")
    await insertMessage(db, session.id, "assistant", "Hi there")

    const count = await countMessages()
    expect(count).toBe(2)
  })
})

// ── Presence ──────────────────────────────────────────────────────────────────

describe("touchUserPresence", () => {
  test("updates lastSeen for a user", async () => {
    const user = await insertUser(db)
    expect(user.lastSeen).toBeNull()

    await touchUserPresence(user.id)

    const [updated] = await db
      .select({ lastSeen: users.lastSeen })
      .from(users)
      .where(eq(users.id, user.id))
    expect(updated!.lastSeen).toBeGreaterThan(0)
  })
})

describe("getOnlineUsers", () => {
  test("returns only users seen within last 2 minutes", async () => {
    const recent = await insertUser(db, "recent@test.com")
    const stale = await insertUser(db, "stale@test.com")

    // Recent: last seen just now
    await db
      .update(users)
      .set({ lastSeen: Date.now() })
      .where(eq(users.id, recent.id))

    // Stale: last seen 5 minutes ago
    await db
      .update(users)
      .set({ lastSeen: Date.now() - 5 * 60 * 1000 })
      .where(eq(users.id, stale.id))

    const online = await getOnlineUsers()
    expect(online).toHaveLength(1)
    expect(online[0]!.email).toBe("recent@test.com")
  })

  test("excludes inactive users even if recently seen", async () => {
    // Insert an inactive user with recent lastSeen
    await db.insert(users).values({
      email: "inactive@test.com",
      name: "Inactive",
      role: "user",
      apiKeyHash: "hash",
      preferences: {},
      active: false,
      lastSeen: Date.now(),
      createdAt: Date.now(),
    })

    const online = await getOnlineUsers()
    expect(online).toHaveLength(0)
  })
})

describe("getUsersPresence", () => {
  test("returns all users with presence data", async () => {
    await insertUser(db, "a@test.com")
    await insertUser(db, "b@test.com")

    const result = await getUsersPresence()
    expect(result).toHaveLength(2)
    // Each result should have expected fields
    expect(result[0]).toHaveProperty("id")
    expect(result[0]).toHaveProperty("name")
    expect(result[0]).toHaveProperty("email")
    expect(result[0]).toHaveProperty("lastSeen")
    expect(result[0]).toHaveProperty("active")
  })
})
