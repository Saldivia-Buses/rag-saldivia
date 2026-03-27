"use server"

import { requireUser } from "@/lib/auth/current-user"
import { updateUser, updatePassword, getUserById, getDb, users, setMemory, deleteMemory } from "@rag-saldivia/db"
import { eq } from "drizzle-orm"
import { log } from "@rag-saldivia/logger/backend"
import { revalidatePath } from "next/cache"

export async function actionUpdateProfile(data: { name: string }) {
  const user = await requireUser()
  await updateUser(user.id, { name: data.name })
  log.info("user.updated", { changes: { name: data.name } }, { userId: user.id })
  revalidatePath("/settings")
  revalidatePath("/", "layout") // actualiza el sidebar con el nombre nuevo
}

export async function actionUpdatePassword(currentPassword: string, newPassword: string) {
  const user = await requireUser()
  const { verifyPassword } = await import("@rag-saldivia/db")
  const valid = await verifyPassword(user.email, currentPassword)
  if (!valid) throw new Error("Contraseña actual incorrecta")
  await updatePassword(user.id, newPassword)
  log.info("auth.password_changed", {}, { userId: user.id })
}

export async function actionUpdatePreferences(preferences: Record<string, unknown>) {
  const user = await requireUser()
  const current = await getUserById(user.id)
  const merged = { ...(current?.preferences ?? {}), ...preferences }
  await updateUser(user.id, { preferences: merged })
  revalidatePath("/settings")
}

export async function actionCompleteOnboarding() {
  const user = await requireUser()
  const db = getDb()
  await db.update(users).set({ onboardingCompleted: true }).where(eq(users.id, user.id))
  revalidatePath("/")
}

export async function actionAddMemory(key: string, value: string) {
  const user = await requireUser()
  await setMemory(user.id, key, value, "explicit")
}

export async function actionDeleteMemory(key: string) {
  const user = await requireUser()
  await deleteMemory(user.id, key)
}
