"use client"

import Link from "next/link"
import { usePathname, useRouter } from "next/navigation"
import { useState } from "react"
import { Plus, MessageSquare, Trash2 } from "lucide-react"
import type { DbChatSession } from "@rag-saldivia/db"
import { actionDeleteSession, actionCreateSession } from "@/app/actions/chat"

export function SessionList({ sessions }: { sessions: DbChatSession[] }) {
  const pathname = usePathname()
  const router = useRouter()
  const [creating, setCreating] = useState(false)

  async function handleNew() {
    setCreating(true)
    try {
      const session = await actionCreateSession({ collection: "tecpia" })
      router.push(`/chat/${session!.id}`)
    } finally {
      setCreating(false)
    }
  }

  async function handleDelete(id: string, e: React.MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    if (!confirm("¿Eliminar esta sesión?")) return
    await actionDeleteSession(id)
    if (pathname === `/chat/${id}`) router.push("/chat")
  }

  return (
    <aside className="w-60 shrink-0 border-r border-border bg-surface flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between" style={{ padding: "16px 16px 12px" }}>
        <span className="text-xs font-semibold uppercase tracking-wide text-fg-subtle">
          Chats
        </span>
        <button
          type="button"
          onClick={handleNew}
          disabled={creating}
          className="p-1 rounded-md text-fg-muted hover:text-fg hover:bg-surface-2 transition-colors disabled:opacity-40"
          title="Nuevo chat"
          aria-label="Nuevo chat"
        >
          <Plus size={16} aria-hidden />
        </button>
      </div>

      {/* Sessions */}
      <nav className="flex-1 overflow-y-auto flex flex-col" style={{ padding: "0 8px 8px", gap: "2px" }}>
        {sessions.length === 0 && (
          <p className="text-sm text-center text-fg-subtle" style={{ padding: "24px 12px" }}>
            Sin chats todavía
          </p>
        )}
        {sessions.map((session) => {
          const active = pathname === `/chat/${session.id}`

          return (
            <Link
              key={session.id}
              href={`/chat/${session.id}`}
              className={`group flex items-center rounded-lg text-sm transition-colors ${
                active
                  ? "bg-accent-subtle text-accent font-medium"
                  : "text-fg-muted hover:bg-surface-2 hover:text-fg"
              }`}
              style={{ padding: "8px 12px", gap: "8px" }}
            >
              <MessageSquare size={14} className="shrink-0 opacity-50" />
              <span className="flex-1 truncate">{session.title}</span>
              <button
                onClick={(e) => handleDelete(session.id, e)}
                className="opacity-0 group-hover:opacity-50 hover:!opacity-100 transition-opacity p-0.5 text-fg-subtle hover:text-destructive"
                title="Eliminar"
              >
                <Trash2 size={12} />
              </button>
            </Link>
          )
        })}
      </nav>
    </aside>
  )
}
