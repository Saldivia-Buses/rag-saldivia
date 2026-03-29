"use server"
/**
 * Server Actions — Gestión de áreas
 */

import { revalidatePath } from "next/cache"
import { requireAdmin } from "@/lib/auth/current-user"
import {
  createArea,
  updateArea,
  deleteArea,
  countUsersInArea,
  setAreaCollections,
} from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"

export async function actionCreateArea(name: string, description?: string) {
  const admin = await requireAdmin()
  const area = await createArea(name, description)
  log.info("area.created", { name, description }, { userId: admin.id })
  revalidatePath("/admin/areas")
  revalidatePath("/admin/permissions")
  return area
}

export async function actionUpdateArea(id: number, data: { name?: string; description?: string }) {
  const admin = await requireAdmin()
  const updated = await updateArea(id, data)
  log.info("area.updated", { areaId: id, changes: data }, { userId: admin.id })
  revalidatePath("/admin/areas")
  return updated
}

export async function actionDeleteArea(id: number) {
  const admin = await requireAdmin()

  // Verificar que no haya usuarios activos en el área
  const userCount = await countUsersInArea(id)
  if (userCount > 0) {
    throw new Error(`No se puede eliminar el área: tiene ${userCount} usuario(s) asignado(s)`)
  }

  await deleteArea(id)
  log.info("area.deleted", { areaId: id }, { userId: admin.id })
  revalidatePath("/admin/areas")
  revalidatePath("/admin/permissions")
}

export async function actionSetAreaCollections(
  areaId: number,
  collections: Array<{ name: string; permission: "read" | "write" | "admin" }>
) {
  const admin = await requireAdmin()
  await setAreaCollections(areaId, collections)
  log.info("admin.config_changed", {
    type: "area_collections",
    areaId,
    collections: collections.map((c) => c.name),
  }, { userId: admin.id })
  revalidatePath("/admin/permissions")
}
