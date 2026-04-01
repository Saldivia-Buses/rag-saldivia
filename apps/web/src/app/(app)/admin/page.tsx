/**
 * /admin — Admin dashboard with system stats and presence.
 *
 * Server Component that fetches stats + user presence,
 * then renders AdminDashboard client component.
 * Auth is handled by the admin layout.
 */

import { AdminDashboard } from "@/components/admin/AdminDashboard"
import { countUsers, countSessions, countMessages, listRoles, getUsersPresence, getRedisClient } from "@rag-saldivia/db"
import { ADMIN_STATS_CACHE_TTL_S } from "@rag-saldivia/config"

export const dynamic = "force-dynamic" // always fresh presence data

const CACHE_KEY = "cache:admin:stats"

async function getAdminStats(): Promise<{
  users: Awaited<ReturnType<typeof countUsers>>
  sessions: Awaited<ReturnType<typeof countSessions>>
  messages: Awaited<ReturnType<typeof countMessages>>
  roles: Awaited<ReturnType<typeof listRoles>>
  usersPresence: Awaited<ReturnType<typeof getUsersPresence>>
}> {
  // Try Redis cache first
  try {
    const cached = await getRedisClient().get(CACHE_KEY)
    if (cached) return JSON.parse(cached)
  } catch { /* Redis down — skip cache */ }

  const [users, sessions, messages, roles, usersPresence] = await Promise.all([
    countUsers(),
    countSessions(),
    countMessages(),
    listRoles(),
    getUsersPresence(),
  ])
  const stats = { users, sessions, messages, roles, usersPresence }

  // Cache for next request
  try {
    await getRedisClient().set(CACHE_KEY, JSON.stringify(stats), "EX", ADMIN_STATS_CACHE_TTL_S)
  } catch { /* Redis down */ }

  return stats
}

export default async function AdminDashboardPage() {
  const { users, sessions, messages, roles, usersPresence } = await getAdminStats()

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
