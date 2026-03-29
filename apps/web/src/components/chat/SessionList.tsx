"use client"

import Link from "next/link"
import { usePathname, useRouter } from "next/navigation"
import { useState, useMemo } from "react"
import { Plus, MessageSquare, Trash2, Search, PanelLeftClose, PanelLeft } from "lucide-react"
import type { DbChatSession } from "@rag-saldivia/db"
import { actionDeleteSession, actionCreateSession } from "@/app/actions/chat"

export function SessionList({ sessions }: { sessions: DbChatSession[] }) {
  const pathname = usePathname()
  const router = useRouter()
  const [creating, setCreating] = useState(false)
  const [search, setSearch] = useState("")
  const [collapsed, setCollapsed] = useState(false)

  const filtered = useMemo(() => {
    if (!search.trim()) return sessions
    const q = search.toLowerCase()
    return sessions.filter((s) => s.title.toLowerCase().includes(q))
  }, [sessions, search])

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

  if (collapsed) {
    return (
      <aside className="shrink-0 border-r border-border bg-surface flex flex-col items-center transition-all duration-200" style={{ width: "48px", padding: "12px 0" }}>
        <button
          type="button"
          onClick={() => setCollapsed(false)}
          className="p-1.5 rounded-md text-fg-muted hover:text-fg hover:bg-surface-2 transition-colors"
          title="Mostrar panel"
          aria-label="Mostrar panel de chats"
        >
          <PanelLeft size={16} />
        </button>
        <div style={{ marginTop: "8px" }}>
          <button
            type="button"
            onClick={handleNew}
            disabled={creating}
            className="p-1.5 rounded-md text-fg-muted hover:text-fg hover:bg-surface-2 transition-colors disabled:opacity-40"
            title="Nuevo chat"
            aria-label="Nuevo chat"
          >
            <Plus size={16} />
          </button>
        </div>
      </aside>
    )
  }

  return (
    <aside className="w-60 shrink-0 border-r border-border bg-surface flex flex-col h-full transition-all duration-200">
      {/* Header */}
      <div className="flex items-center justify-between" style={{ padding: "12px 12px 8px" }}>
        <div className="flex items-center" style={{ gap: "6px" }}>
          <button
            type="button"
            onClick={() => setCollapsed(true)}
            className="p-1 rounded-md text-fg-muted hover:text-fg hover:bg-surface-2 transition-colors"
            title="Ocultar panel"
            aria-label="Ocultar panel de chats"
          >
            <PanelLeftClose size={16} />
          </button>
          <span className="text-xs font-semibold uppercase tracking-wide text-fg-subtle">
            Chats
          </span>
        </div>
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

      {/* Search */}
      <div style={{ padding: "0 12px 8px" }}>
        <div className="flex items-center rounded-md border border-border bg-bg transition-colors focus-within:border-accent" style={{ padding: "4px 8px", gap: "6px" }}>
          <Search size={12} className="text-fg-subtle shrink-0" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar chats..."
            className="flex-1 bg-transparent text-xs text-fg placeholder:text-fg-subtle outline-none"
            style={{ minWidth: 0 }}
          />
        </div>
      </div>

      {/* Sessions */}
      <nav aria-label="Lista de chats" className="flex-1 overflow-y-auto flex flex-col" style={{ padding: "0 8px 8px", gap: "2px" }}>
        {filtered.length === 0 && (
          <p className="text-xs text-center text-fg-subtle" style={{ padding: "24px 12px" }}>
            {search ? "Sin resultados" : "Sin chats todavía"}
          </p>
        )}
        {filtered.map((session) => {
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
