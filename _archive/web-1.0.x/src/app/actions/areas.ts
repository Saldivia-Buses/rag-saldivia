"use server"

import { z } from "zod"
import { revalidatePath } from "next/cache"
import { adminAction, clean } from "@/lib/safe-action"
import {
  createArea,
  updateArea,
  deleteArea,
  countUsersInArea,
  setAreaCollections,
  addUserArea,
  removeUserArea,
} from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"

export const actionCreateArea = adminAction
  .schema(z.object({
    name: z.string().min(2),
    description: z.string().optional(),
  }))
  .action(async ({ parsedInput: data, ctx: { user } }) => {
    const area = await createArea(data.name, data.description ?? "")
    log.info("area.created", { name: data.name }, { userId: user.id })
    revalidatePath("/admin/areas")
    revalidatePath("/admin/permissions")
    return area
  })

export const actionUpdateArea = adminAction
  .schema(z.object({
    id: z.number(),
    data: z.object({
      name: z.string().min(2).optional(),
      description: z.string().optional(),
    }),
  }))
  .action(async ({ parsedInput: { id, data }, ctx: { user } }) => {
    const updated = await updateArea(id, clean(data))
    log.info("area.updated", { areaId: id, changes: data }, { userId: user.id })
    revalidatePath("/admin/areas")
    return updated
  })

export const actionDeleteArea = adminAction
  .schema(z.object({ id: z.number() }))
  .action(async ({ parsedInput: { id }, ctx: { user } }) => {
    const userCount = await countUsersInArea(id)
    if (userCount > 0) {
      throw new Error(`No se puede eliminar: tiene ${userCount} usuario(s) asignado(s)`)
    }
    await deleteArea(id)
    log.info("area.deleted", { areaId: id }, { userId: user.id })
    revalidatePath("/admin/areas")
    revalidatePath("/admin/permissions")
  })

export const actionSetAreaCollections = adminAction
  .schema(z.object({
    areaId: z.number(),
    collections: z.array(z.object({
      name: z.string(),
      permission: z.enum(["read", "write", "admin"]),
    })),
  }))
  .action(async ({ parsedInput: { areaId, collections }, ctx: { user } }) => {
    await setAreaCollections(areaId, collections)
    log.info("admin.config_changed", {
      type: "area_collections",
      areaId,
      collections: collections.map((c) => c.name),
    }, { userId: user.id })
    revalidatePath("/admin/areas")
    revalidatePath("/admin/permissions")
  })

export const actionAddUserToArea = adminAction
  .schema(z.object({ userId: z.number(), areaId: z.number() }))
  .action(async ({ parsedInput: { userId, areaId }, ctx: { user } }) => {
    await addUserArea(userId, areaId)
    log.info("user.area_assigned", { areaId, targetUserId: userId }, { userId: user.id })
    revalidatePath("/admin/areas")
  })

export const actionRemoveUserFromArea = adminAction
  .schema(z.object({ userId: z.number(), areaId: z.number() }))
  .action(async ({ parsedInput: { userId, areaId }, ctx: { user } }) => {
    await removeUserArea(userId, areaId)
    log.info("user.area_removed", { areaId, targetUserId: userId }, { userId: user.id })
    revalidatePath("/admin/areas")
  })
