import { requireUser } from "@/lib/auth/current-user"
import { listSessionsByUser } from "@rag-saldivia/db"
import { SessionList } from "@/components/chat/SessionList"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import { MessageSquare } from "lucide-react"

export default async function ChatPage() {
  const user = await requireUser()
  const sessions = await listSessionsByUser(user.id)

  return (
    <div className="flex h-full">
      <SessionList sessions={sessions} />
      <div className="flex-1 flex items-center justify-center p-8">
        <EmptyPlaceholder className="max-w-sm border-none bg-transparent">
          <h1 className="sr-only">Chat</h1>
          <EmptyPlaceholder.Icon icon={MessageSquare} />
          <EmptyPlaceholder.Title>
            Seleccioná una sesión o creá una nueva
          </EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>
            Hacé una pregunta sobre tus documentos y obtené respuestas fundamentadas.
          </EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      </div>
    </div>
  )
}
