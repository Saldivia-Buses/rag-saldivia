/**
 * /admin/roles — Role management page.
 *
 * Server Component that fetches roles and permissions,
 * then renders AdminRoles client component.
 * Auth is handled by the admin layout.
 */

import { AdminRoles } from "@/components/admin/AdminRoles"
import { listRoles, listPermissions } from "@rag-saldivia/db"

export default async function AdminRolesPage() {
  const [roles, permissions] = await Promise.all([
    listRoles(),
    listPermissions(),
  ])

  return (
    <AdminRoles
      initialRoles={roles}
      initialPermissions={permissions}
    />
  )
}
