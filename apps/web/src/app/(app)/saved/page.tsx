import { requireUser } from "@/lib/auth/current-user"
import { listSavedResponses } from "@rag-saldivia/db"
import { Bookmark } from "lucide-react"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"

export default async function SavedPage() {
  const user = await requireUser()
  const saved = await listSavedResponses(user.id)

  return (
    <div className="p-6 max-w-3xl mx-auto">
      <div className="mb-6">
        <h1 className="text-lg font-semibold text-fg">Respuestas guardadas</h1>
        <p className="text-sm text-fg-muted mt-0.5">
          {saved.length} respuesta{saved.length !== 1 ? "s" : ""} guardada{saved.length !== 1 ? "s" : ""}
        </p>
      </div>

      {saved.length === 0 ? (
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={Bookmark} />
          <EmptyPlaceholder.Title>Sin respuestas guardadas</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>
            Hacé clic en el ícono 🔖 en cualquier respuesta del chat para guardarla acá.
          </EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      ) : (
        <div className="space-y-3">
          {saved.map((item) => (
            <div key={item.id} className="p-4 rounded-xl border border-border bg-surface">
              {item.sessionTitle && (
                <p className="text-xs font-medium text-fg-muted mb-2 flex items-center gap-1">
                  <Bookmark size={11} />
                  {item.sessionTitle}
                </p>
              )}
              <p className="text-sm whitespace-pre-wrap leading-relaxed text-fg">{item.content}</p>
              <p className="text-xs text-fg-subtle mt-3">
                {new Date(item.createdAt).toLocaleString("es-AR")}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
