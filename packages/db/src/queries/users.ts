/**
 * Queries de usuarios — reemplaza los métodos de AuthDB en database.py relacionados
 * con usuarios: verify_user, get_user, create_user, update_user, delete_user, etc.
 */

import { eq, and } from "drizzle-orm"
import { createHash } from "crypto"
import { getDb } from "../connection"
import { users, userAreas, areas } from "../schema"
import { compare, hash } from "bcrypt-ts"

function now() {
  return Date.now()
}

// ── Lectura ────────────────────────────────────────────────────────────────

export async function getUserById(id: number) {
  return getDb().query.users.findFirst({
    where: (u, { eq }) => eq(u.id, id),
    with: { userAreas: { with: { area: true } } },
  })
}

export async function getUserByEmail(email: string) {
  return getDb().query.users.findFirst({
    where: (u, { eq }) => eq(u.email, email.toLowerCase()),
    with: { userAreas: { with: { area: true } } },
  })
}

export async function getUserByApiKey(apiKeyHash: string) {
  return getDb().query.users.findFirst({
    where: (u, { and, eq }) => and(eq(u.apiKeyHash, apiKeyHash), eq(u.active, true)),
  })
}

export async function listUsers() {
  return getDb().query.users.findMany({
    with: { userAreas: { with: { area: true } } },
    orderBy: (u, { asc }) => [asc(u.name)],
  })
}

// ── Autenticación ──────────────────────────────────────────────────────────

export async function verifyPassword(email: string, password: string) {
  const user = await getUserByEmail(email)
  if (!user || !user.active || !user.passwordHash) return null
  const valid = await compare(password, user.passwordHash)
  if (!valid) return null

  await getDb()
    .update(users)
    .set({ lastLogin: now() })
    .where(eq(users.id, user.id))

  return user
}

// ── Escritura ──────────────────────────────────────────────────────────────

export async function createUser(data: {
  email: string
  name: string
  password: string
  role?: "admin" | "area_manager" | "user"
  areaIds?: number[]
}) {
  const db = getDb()
  const passwordHash = await hash(data.password, 10)
  const apiKeyHash = createHash("sha256").update(`rsk_${crypto.randomUUID()}`).digest("hex").slice(0, 32)

  const [user] = await db
    .insert(users)
    .values({
      email: data.email.toLowerCase(),
      name: data.name,
      role: data.role ?? "user",
      apiKeyHash,
      passwordHash,
      preferences: {},
      active: true,
      createdAt: now(),
    })
    .returning()

  if (!user) throw new Error("Failed to create user")

  if (data.areaIds && data.areaIds.length > 0) {
    await db.insert(userAreas).values(
      data.areaIds.map((areaId) => ({ userId: user.id, areaId }))
    )
  }

  return user
}

export async function updateUser(
  id: number,
  data: Partial<{
    name: string
    role: "admin" | "area_manager" | "user"
    active: boolean
    preferences: Record<string, unknown>
  }>
) {
  const [updated] = await getDb()
    .update(users)
    .set(data)
    .where(eq(users.id, id))
    .returning()

  return updated
}

export async function updatePassword(id: number, newPassword: string) {
  const passwordHash = await hash(newPassword, 10)
  await getDb().update(users).set({ passwordHash }).where(eq(users.id, id))
}

export async function deleteUser(id: number) {
  await getDb().delete(users).where(eq(users.id, id))
}

export async function getUserAreas(userId: number) {
  const rows = await getDb().query.userAreas.findMany({
    where: (ua, { eq }) => eq(ua.userId, userId),
    with: { area: true },
  })
  return rows.map((r) => r.area)
}

export async function addUserArea(userId: number, areaId: number) {
  await getDb()
    .insert(userAreas)
    .values({ userId, areaId })
    .onConflictDoNothing()
}

export async function removeUserArea(userId: number, areaId: number) {
  await getDb()
    .delete(userAreas)
    .where(and(eq(userAreas.userId, userId), eq(userAreas.areaId, areaId)))
}

export async function getUserCollections(userId: number): Promise<Array<{ name: string; permission: string }>> {
  const userAreaRows = await getDb().query.userAreas.findMany({
    where: (ua, { eq }) => eq(ua.userId, userId),
    with: { area: { with: { areaCollections: true } } },
  })

  const seen = new Map<string, string>()
  for (const ua of userAreaRows) {
    for (const ac of ua.area.areaCollections) {
      const existing = seen.get(ac.collectionName)
      if (!existing || ac.permission === "admin" || (ac.permission === "write" && existing === "read")) {
        seen.set(ac.collectionName, ac.permission)
      }
    }
  }

  return Array.from(seen.entries()).map(([name, permission]) => ({ name, permission }))
}

export async function canAccessCollection(
  userId: number,
  collectionName: string,
  permission: "read" | "write" | "admin" = "read"
): Promise<boolean> {
  const collections = await getUserCollections(userId)
  const col = collections.find((c) => c.name === collectionName)
  if (!col) return false

  const levels = { read: 1, write: 2, admin: 3 }
  return (levels[col.permission as keyof typeof levels] ?? 0) >= (levels[permission] ?? 1)
}
