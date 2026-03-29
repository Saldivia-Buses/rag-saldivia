import { requireUser } from "@/lib/auth/current-user"
import { getMemory } from "@rag-saldivia/db"
import { MemoryClient } from "@/components/settings/MemoryClient"

export default async function MemoryPage() {
  const user = await requireUser()
  const entries = await getMemory(user.id)
  return <MemoryClient entries={entries} />
}
