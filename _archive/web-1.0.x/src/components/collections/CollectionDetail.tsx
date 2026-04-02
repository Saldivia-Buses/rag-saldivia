"use client"

import Link from "next/link"
import { MessageSquare, ArrowLeft, Clock, FileText } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { useRouter } from "next/navigation"
import { actionCreateSession } from "@/app/actions/chat"

type HistoryEvent = {
  id: string
  collection: string
  action: "added" | "removed"
  filename: string | null
  docCount: number | null
  userId: number
  createdAt: number
}

type AreaInfo = { name: string; permission: string }

const PERM_COLORS: Record<string, string> = {
  read: "var(--accent)",
  write: "var(--success)",
  admin: "var(--warning)",
}

function formatDate(ts: number): string {
  return new Date(ts).toLocaleDateString("es-AR", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  })
}

export function CollectionDetail({
  name,
  userPermission,
  isAdmin,
  areas,
  history,
}: {
  name: string
  userPermission: string | null
  isAdmin: boolean
  areas: AreaInfo[]
  history: HistoryEvent[]
}) {
  const router = useRouter()

  async function handleChat() {
    const result = await actionCreateSession({ collection: name })
    if (result?.data?.id) router.push(`/chat/${result.data.id}`)
  }

  return (
    <div className="flex flex-col" style={{ gap: "24px" }}>
      {/* Back link */}
      <Link
        href="/collections"
        className="flex items-center text-sm text-fg-muted hover:text-fg transition-colors"
        style={{ gap: "4px" }}
      >
        <ArrowLeft size={14} />
        Volver a colecciones
      </Link>

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-fg">{name}</h1>
          <div className="flex items-center" style={{ gap: "8px", marginTop: "8px" }}>
            {userPermission && (
              <Badge
                variant="outline"
                style={{ color: PERM_COLORS[userPermission] ?? "var(--fg-subtle)" }}
              >
                {userPermission}
              </Badge>
            )}
            {isAdmin && !userPermission && (
              <Badge variant="outline" style={{ color: "var(--warning)" }}>
                admin
              </Badge>
            )}
          </div>
        </div>
        <Button onClick={handleChat} style={{ gap: "6px" }}>
          <MessageSquare size={16} />
          Chatear con esta colección
        </Button>
      </div>

      {/* Areas (admin only) */}
      {isAdmin && areas.length > 0 && (
        <div className="rounded-xl border border-border bg-surface" style={{ padding: "16px 20px" }}>
          <p className="text-xs font-semibold uppercase tracking-wider text-fg-subtle" style={{ marginBottom: "8px" }}>
            Áreas con acceso
          </p>
          <div className="flex flex-wrap" style={{ gap: "6px" }}>
            {areas.map((a) => (
              <Badge
                key={a.name}
                variant="outline"
                style={{ color: PERM_COLORS[a.permission] ?? "var(--fg-subtle)" }}
              >
                {a.name} ({a.permission})
              </Badge>
            ))}
          </div>
        </div>
      )}

      {/* Ingestion history */}
      <div>
        <h2 className="text-base font-semibold text-fg" style={{ marginBottom: "12px" }}>
          Historial de ingesta
        </h2>
        {history.length === 0 ? (
          <p className="text-sm text-fg-muted" style={{ padding: "24px 0", textAlign: "center" }}>
            Sin eventos de ingesta registrados.
          </p>
        ) : (
          <div className="rounded-xl border border-border overflow-hidden">
            <table className="w-full text-sm">
              <thead style={{ backgroundColor: "var(--surface)" }}>
                <tr>
                  <th className="text-left text-xs font-semibold text-fg-muted" style={{ padding: "10px 16px" }}>
                    Fecha
                  </th>
                  <th className="text-left text-xs font-semibold text-fg-muted" style={{ padding: "10px 16px" }}>
                    Evento
                  </th>
                  <th className="text-left text-xs font-semibold text-fg-muted" style={{ padding: "10px 16px" }}>
                    Archivo
                  </th>
                  <th className="text-right text-xs font-semibold text-fg-muted" style={{ padding: "10px 16px" }}>
                    Docs
                  </th>
                </tr>
              </thead>
              <tbody>
                {history.map((event) => (
                  <tr key={event.id} style={{ borderTop: "1px solid var(--border)" }}>
                    <td className="text-fg-muted" style={{ padding: "10px 16px" }}>
                      <span className="flex items-center" style={{ gap: "6px" }}>
                        <Clock size={13} className="text-fg-subtle" />
                        {formatDate(event.createdAt)}
                      </span>
                    </td>
                    <td style={{ padding: "10px 16px" }}>
                      <Badge
                        variant="outline"
                        className="text-xs"
                        style={{ color: event.action === "added" ? "var(--success)" : "var(--destructive)" }}
                      >
                        {event.action === "added" ? "ingesta" : "eliminado"}
                      </Badge>
                    </td>
                    <td className="text-fg" style={{ padding: "10px 16px" }}>
                      {event.filename ? (
                        <span className="flex items-center" style={{ gap: "4px" }}>
                          <FileText size={13} className="text-fg-subtle" />
                          {event.filename}
                        </span>
                      ) : "—"}
                    </td>
                    <td className="text-right text-fg-muted" style={{ padding: "10px 16px" }}>
                      {event.docCount ?? "—"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
