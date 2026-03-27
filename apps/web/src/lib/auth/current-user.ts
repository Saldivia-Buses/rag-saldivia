/**
 * Helper para obtener el usuario actual en Server Components.
 * Lee los headers que el middleware injectó después de verificar el JWT.
 */

import { headers } from "next/headers"
import { redirect } from "next/navigation"
import { cache } from "react"
import type { Role } from "@rag-saldivia/shared"

export type CurrentUser = {
  id: number
  email: string
  name: string
  role: Role
}

/**
 * Obtener el usuario actual desde los headers del middleware.
 * Usa React.cache() para deduplicar dentro de la misma request.
 */
const getCurrentUser = cache(async (): Promise<CurrentUser | null> => {
  const h = await headers()
  const userId = h.get("x-user-id")
  const email = h.get("x-user-email")
  const name = h.get("x-user-name")
  const role = h.get("x-user-role") as Role | null

  if (!userId || !email || !name || !role) return null

  return {
    id: Number(userId),
    email,
    name,
    role,
  }
})

/**
 * Obtener usuario o redirigir a login si no está autenticado.
 * Usar en Server Components de páginas protegidas.
 */
export async function requireUser(): Promise<CurrentUser> {
  const user = await getCurrentUser()
  if (!user) redirect("/login")
  return user
}

/**
 * Requerir rol admin.
 */
export async function requireAdmin(): Promise<CurrentUser> {
  const user = await requireUser()
  if (user.role !== "admin") redirect("/")
  return user
}
