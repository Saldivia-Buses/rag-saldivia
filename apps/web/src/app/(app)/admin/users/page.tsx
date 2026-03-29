/**
 * /admin/users — User management page (admin only).
 *
 * Server Component that fetches all users, then renders AdminUsers client component.
 * Auth is handled by the admin layout.
 *
 * Data flow: DB (listUsers, listRoles) → this page → AdminUsers (client)
 */

import { requireAdmin } from "@/lib/auth/current-user"
import { listUsers, listRoles, getUserRoles } from "@rag-saldivia/db"
import { AdminUsers } from "@/components/admin/AdminUsers"

export default async function AdminUsersPage() {
  const user = await requireAdmin()
  const [users, roles] = await Promise.all([listUsers(), listRoles()])

  // Get role assignments for each user
  const userRoleMap: Record<number, number[]> = {}
  for (const u of users) {
    const userRoles = await getUserRoles(u.id)
    userRoleMap[u.id] = userRoles.map((r) => r.id)
  }

  return (
    <AdminUsers
      initialUsers={users}
      currentUserId={user.id}
      allRoles={roles}
      userRoleMap={userRoleMap}
    />
  )
}
