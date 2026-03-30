import { requireUser } from "@/lib/auth/current-user"
import { listSessionsByUser, getUserById } from "@rag-saldivia/db"
import { SessionList } from "@/components/chat/SessionList"

export default async function ChatPage() {
  const user = await requireUser()
  const [sessions, fullUser] = await Promise.all([
    listSessionsByUser(user.id),
    getUserById(user.id),
  ])
  const defaultCollection = (fullUser?.preferences as Record<string, unknown> | undefined)?.defaultCollection as string | undefined

  return (
    <>
      <SessionList sessions={sessions} {...(defaultCollection ? { defaultCollection } : {})} />
      <div className="flex-1 flex flex-col items-center justify-center bg-bg">
        <h1 className="sr-only">Chat</h1>
        <svg
          width="32" height="32" viewBox="0 0 16 16" fill="none"
          className="text-fg-subtle"
          style={{ marginBottom: "16px", opacity: 0.4 }}
        >
          <path
            d="M8 1L9.5 6.5L15 8L9.5 9.5L8 15L6.5 9.5L1 8L6.5 6.5L8 1Z"
            fill="currentColor"
          />
        </svg>
        <p className="text-sm text-fg-subtle">
          Seleccioná una sesión o creá una nueva
        </p>
      </div>
    </>
  )
}
