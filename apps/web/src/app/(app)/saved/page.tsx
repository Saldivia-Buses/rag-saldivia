import { requireUser } from "@/lib/auth/current-user"
import { listSavedResponses } from "@rag-saldivia/db"
import { Bookmark } from "lucide-react"

export default async function SavedPage() {
  const user = await requireUser()
  const saved = await listSavedResponses(user.id)

  return (
    <div className="p-6 max-w-3xl mx-auto">
      <div className="mb-6">
        <h1 className="text-xl font-semibold">Respuestas guardadas</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          {saved.length} respuesta{saved.length !== 1 ? "s" : ""} guardada{saved.length !== 1 ? "s" : ""}
        </p>
      </div>

      {saved.length === 0 ? (
        <div
          className="flex flex-col items-center justify-center py-20 gap-3"
          style={{ color: "var(--muted-foreground)" }}
        >
          <Bookmark size={32} strokeWidth={1.5} />
          <p className="text-sm">Todavía no guardaste ninguna respuesta</p>
          <p className="text-xs">Hacé clic en el ícono 🔖 en cualquier respuesta del chat</p>
        </div>
      ) : (
        <div className="space-y-4">
          {saved.map((item) => (
            <div
              key={item.id}
              className="p-4 rounded-xl border"
              style={{ borderColor: "var(--border)", background: "var(--background)" }}
            >
              {item.sessionTitle && (
                <p
                  className="text-xs font-medium mb-2 flex items-center gap-1"
                  style={{ color: "var(--muted-foreground)" }}
                >
                  <Bookmark size={11} />
                  {item.sessionTitle}
                </p>
              )}
              <p className="text-sm whitespace-pre-wrap leading-relaxed">{item.content}</p>
              <p className="text-xs mt-3" style={{ color: "var(--muted-foreground)" }}>
                {new Date(item.createdAt).toLocaleString("es-AR")}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
