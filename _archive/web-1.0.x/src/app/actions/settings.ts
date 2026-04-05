"use server"

import { z } from "zod"
import { authAction } from "@/lib/safe-action"
import { updateUser, updatePassword, getUserById, getDb, users, setMemory, deleteMemory } from "@rag-saldivia/db"
import { eq } from "drizzle-orm"
import { log } from "@rag-saldivia/logger/backend"
import { revalidatePath } from "next/cache"

export const actionUpdateProfile = authAction
  .schema(z.object({ name: z.string().min(2) }))
  .action(async ({ parsedInput: data, ctx: { user } }) => {
    await updateUser(user.id, { name: data.name })
    log.info("user.updated", { changes: { name: data.name } }, { userId: user.id })
    revalidatePath("/settings")
    revalidatePath("/", "layout")
  })

export const actionUpdatePassword = authAction
  .schema(z.object({
    currentPassword: z.string().min(1),
    newPassword: z.string().min(8),
  }))
  .action(async ({ parsedInput: { currentPassword, newPassword }, ctx: { user } }) => {
    const { verifyPassword } = await import("@rag-saldivia/db")
    const valid = await verifyPassword(user.email, currentPassword)
    if (!valid) throw new Error("Contraseña actual incorrecta")
    await updatePassword(user.id, newPassword)
    log.info("auth.password_changed", {}, { userId: user.id })
  })

export const actionUpdatePreferences = authAction
  .schema(z.record(z.string(), z.unknown()))
  .action(async ({ parsedInput: preferences, ctx: { user } }) => {
    const current = await getUserById(user.id)
    const merged = { ...(current?.preferences ?? {}), ...preferences }
    await updateUser(user.id, { preferences: merged })
    revalidatePath("/settings")
  })

export const actionCompleteOnboarding = authAction
  .action(async ({ ctx: { user } }) => {
    const db = getDb()
    await db.update(users).set({ onboardingCompleted: true }).where(eq(users.id, user.id))
    revalidatePath("/")
  })

export const actionAddMemory = authAction
  .schema(z.object({ key: z.string().min(1), value: z.string().min(1) }))
  .action(async ({ parsedInput: { key, value }, ctx: { user } }) => {
    await setMemory(user.id, key, value, "explicit")
    revalidatePath("/settings")
  })

export const actionDeleteMemory = authAction
  .schema(z.object({ key: z.string().min(1) }))
  .action(async ({ parsedInput: { key }, ctx: { user } }) => {
    await deleteMemory(user.id, key)
    revalidatePath("/settings")
  })
