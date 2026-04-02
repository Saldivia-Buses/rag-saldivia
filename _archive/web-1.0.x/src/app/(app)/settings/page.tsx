import { requireUser } from "@/lib/auth/current-user"
import { getUserById, getUserCollections } from "@rag-saldivia/db"
import { SettingsClient } from "@/components/settings/SettingsClient"

export default async function SettingsPage() {
  const current = await requireUser()
  const [user, userCollections] = await Promise.all([
    getUserById(current.id),
    getUserCollections(current.id),
  ])
  if (!user) return null

  return (
    <div style={{ padding: "32px 24px" }}>
      <div className="max-w-xl mx-auto">
        <SettingsClient user={user} userCollections={userCollections} />
      </div>
    </div>
  )
}
