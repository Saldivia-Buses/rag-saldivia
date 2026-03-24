import { requireUser } from "@/lib/auth/current-user"
import { listSessionsByUser } from "@rag-saldivia/db"
import { SessionList } from "@/components/chat/SessionList"

export default async function ChatPage() {
  const user = await requireUser()
  const sessions = await listSessionsByUser(user.id)

  return (
    <div className="flex h-full">
      <SessionList sessions={sessions} />
      <div className="flex-1 flex items-center justify-center" style={{ color: "var(--muted-foreground)" }}>
        <div className="text-center space-y-2">
          <p className="text-lg font-medium">Seleccioná una sesión o creá una nueva</p>
          <p className="text-sm">Hacé una pregunta sobre tus documentos</p>
        </div>
      </div>
    </div>
  )
}
