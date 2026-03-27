import { createSafeActionClient } from "next-safe-action"
import { requireUser, requireAdmin } from "@/lib/auth/current-user"

/**
 * Cliente para Server Actions que requieren usuario autenticado.
 * Inyecta `ctx.user` en cada action sin boilerplate.
 */
export const authClient = createSafeActionClient().use(async ({ next }) => {
  const user = await requireUser()
  return next({ ctx: { user } })
})

/**
 * Cliente para Server Actions que requieren rol admin.
 * Inyecta `ctx.admin` en cada action.
 */
export const adminClient = createSafeActionClient().use(async ({ next }) => {
  const admin = await requireAdmin()
  return next({ ctx: { admin } })
})
