/**
 * /chat/[id] — Individual chat session page.
 *
 * Server Component that fetches session data, session list, and prompt templates
 * in parallel, then passes them to client components for rendering.
 *
 * Data flow: DB queries (parallel) → SessionList + ChatInterface (client)
 */

import { notFound } from "next/navigation"
import { requireUser } from "@/lib/auth/current-user"
import { getSessionById, listSessionsByUser, listActiveTemplates, getUserById, getUserCollections } from "@rag-saldivia/db"
import { SessionList } from "@/components/chat/SessionList"
import { ChatInterface } from "@/components/chat/ChatInterface"

export default async function ChatSessionPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const user = await requireUser()
  const { id } = await params

  const [session, sessions, templates, fullUser, userCols] = await Promise.all([
    getSessionById(id, user.id),
    listSessionsByUser(user.id),
    listActiveTemplates().catch(() => []),
    getUserById(user.id),
    getUserCollections(user.id).catch(() => []),
  ])
  const defaultCollection = (fullUser?.preferences as Record<string, unknown> | undefined)?.defaultCollection as string | undefined

  if (!session) notFound()

  return (
    <>
      <SessionList sessions={sessions} {...(defaultCollection ? { defaultCollection } : {})} />
      <ChatInterface
        session={session}
        userId={user.id}
        templates={templates}
        availableCollections={userCols.map((c) => c.name)}
      />
    </>
  )
}
