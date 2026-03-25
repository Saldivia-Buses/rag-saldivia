import { notFound } from "next/navigation"
import { requireUser } from "@/lib/auth/current-user"
import { getSessionById, listSessionsByUser } from "@rag-saldivia/db"
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

  const [session, sessions] = await Promise.all([
    getSessionById(id, user.id),
    listSessionsByUser(user.id),
  ])

  if (!session) notFound()

  return (
    <div className="flex h-full">
      <SessionList sessions={sessions} />
      <ChatDropZone sessionId={session.id}>
        <ChatInterface session={session} userId={user.id} />
      </ChatDropZone>
    </div>
  )
}
