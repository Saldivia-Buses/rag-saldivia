/**
 * Queries de áreas.
 */

import { and, eq } from "drizzle-orm"
import { getDb } from "../connection"
import { areas, areaCollections, userAreas } from "../schema"

function now() {
  return Date.now()
}

export async function listAreas() {
  return getDb().query.areas.findMany({
    with: { areaCollections: true },
    orderBy: (a, { asc }) => [asc(a.name)],
  })
}

export async function getAreaById(id: number) {
  return getDb().query.areas.findFirst({
    where: (a, { eq }) => eq(a.id, id),
    with: { areaCollections: true },
  })
}

export async function createArea(name: string, description = "") {
  const [area] = await getDb()
    .insert(areas)
    .values({ name, description, createdAt: now() })
    .returning()
  return area
}

export async function updateArea(id: number, data: { name?: string; description?: string }) {
  const [updated] = await getDb().update(areas).set(data).where(eq(areas.id, id)).returning()
  return updated
}

export async function deleteArea(id: number) {
  await getDb().delete(areas).where(eq(areas.id, id))
}

export async function countUsersInArea(areaId: number): Promise<number> {
  const rows = await getDb().query.userAreas.findMany({
    where: (ua, { eq }) => eq(ua.areaId, areaId),
  })
  return rows.length
}

export async function setAreaCollections(
  areaId: number,
  collections: Array<{ name: string; permission: "read" | "write" | "admin" }>
) {
  const db = getDb()
  await db.delete(areaCollections).where(eq(areaCollections.areaId, areaId))
  if (collections.length > 0) {
    await db.insert(areaCollections).values(
      collections.map((c) => ({ areaId, collectionName: c.name, permission: c.permission }))
    )
  }
}

export async function addAreaCollection(
  areaId: number,
  collectionName: string,
  permission: "read" | "write" | "admin" = "read"
) {
  await getDb()
    .insert(areaCollections)
    .values({ areaId, collectionName, permission })
    .onConflictDoUpdate({
      target: [areaCollections.areaId, areaCollections.collectionName],
      set: { permission },
    })
}

export async function removeAreaCollection(areaId: number, collectionName: string) {
  await getDb()
    .delete(areaCollections)
    .where(and(eq(areaCollections.areaId, areaId), eq(areaCollections.collectionName, collectionName)))
}
