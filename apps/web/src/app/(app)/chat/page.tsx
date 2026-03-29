import { requireUser } from "@/lib/auth/current-user"
import { listSessionsByUser } from "@rag-saldivia/db"
import { SessionList } from "@/components/chat/SessionList"

export default async function ChatPage() {
  const user = await requireUser()
  const sessions = await listSessionsByUser(user.id)

  return (
    <>
      <SessionList sessions={sessions} />
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
