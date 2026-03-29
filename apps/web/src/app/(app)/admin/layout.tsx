/**
 * Admin layout — wraps all /admin/* pages with tab navigation.
 * Requires admin role.
 */

import { requireAdmin } from "@/lib/auth/current-user"
import { AdminLayout } from "@/components/admin/AdminLayout"

export default async function AdminLayoutPage({
  children,
}: {
  children: React.ReactNode
}) {
  await requireAdmin()
  return <AdminLayout>{children}</AdminLayout>
}
