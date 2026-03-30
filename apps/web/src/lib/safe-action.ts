import { createSafeActionClient } from "next-safe-action"
import { requireUser, requireAdmin } from "@/lib/auth/current-user"

/**
 * Strip undefined values from an object at runtime.
 * Bridges Zod optional (T | undefined) to exactOptionalPropertyTypes (T?).
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function clean(obj: Record<string, unknown>): any {
  return Object.fromEntries(
    Object.entries(obj).filter(([, v]) => v !== undefined)
  )
}

const baseClient = createSafeActionClient({
  handleServerError: (e) => e.message,
})

/** Authenticated action — injects user into ctx */
export const authAction = baseClient.use(async ({ next }) => {
  const user = await requireUser()
  return next({ ctx: { user } })
})

/** Admin-only action — injects admin user into ctx */
export const adminAction = baseClient.use(async ({ next }) => {
  const user = await requireAdmin()
  return next({ ctx: { user } })
})
