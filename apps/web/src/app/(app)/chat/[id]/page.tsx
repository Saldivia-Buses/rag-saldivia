import { notFound } from "next/navigation"
import { requireUser } from "@/lib/auth/current-user"
import { getSessionById, listSessionsByUser, listActiveTemplates } from "@rag-saldivia/db"
import { getCachedRagCollections } from "@/lib/rag/collections-cache"
import { SessionList } from "@/components/chat/SessionList"
import { ChatInterface } from "@/components/chat/ChatInterface"
import { ChatDropZone } from "@/components/chat/ChatDropZone"

export default async function ChatSessionPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const user = await requireUser()
  const { id } = await params

  const [session, sessions, templates, availableCollections] = await Promise.all([
    getSessionById(id, user.id),
    listSessionsByUser(user.id),
    listActiveTemplates(),
    getCachedRagCollections(),
  ])

  if (!session) notFound()

  return (
    <div className="flex h-full">
      <SessionList sessions={sessions} />
      <ChatDropZone sessionId={session.id}>
        <ChatInterface
          session={session}
          userId={user.id}
          templates={templates}
          availableCollections={availableCollections}
        />
      </ChatDropZone>
    </div>
  )
}
