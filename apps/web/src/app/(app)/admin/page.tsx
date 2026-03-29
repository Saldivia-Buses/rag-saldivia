/**
 * /admin — Admin dashboard with system stats and presence.
 *
 * Server Component that fetches stats + user presence,
 * then renders AdminDashboard client component.
 * Auth is handled by the admin layout.
 */

import { AdminDashboard } from "@/components/admin/AdminDashboard"
import { countUsers, countSessions, countMessages, listRoles, getUsersPresence } from "@rag-saldivia/db"

export const dynamic = "force-dynamic" // always fresh presence data

export default async function AdminDashboardPage() {
  const [users, sessions, messages, roles, usersPresence] = await Promise.all([
    countUsers(),
    countSessions(),
    countMessages(),
    listRoles(),
    getUsersPresence(),
  ])

  return (
    <AdminDashboard
      stats={{
        users,
        sessions,
        messages,
        roles: roles.map((r) => ({
          id: r.id,
          name: r.name,
          color: r.color,
          level: r.level,
          userCount: r.userCount,
        })),
        usersPresence,
      }}
    />
  )
}
