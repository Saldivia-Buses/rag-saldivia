/**
 * Chat session sidebar list with search, create, and delete.
 *
 * Renders the collapsible left panel inside the chat area. Sessions are
 * passed as props from a Server Component (the /chat layout fetches them).
 *
 * Features:
 *   - Client-side search filtering (useMemo, no server round-trip)
 *   - New session creation via server action -> redirect to /chat/[id]
 *   - Delete with confirmation -> server action + redirect if active
 *   - Width animates to 0 when sidebar is closed (via useSidebar context)
 *   - Inner content uses a fixed 260px wrapper to prevent layout shift
 *
 * Data flow: /chat layout (RSC) -> sessions prop -> SessionList (client)
 * Depends on: ChatLayout (useSidebar), server actions (create/delete)
 */
"use client"

import Link from "next/link"
import { usePathname, useRouter } from "next/navigation"
import { useState, useMemo, useCallback } from "react"
import { Plus, MessageSquare, Trash2, Search } from "lucide-react"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import type { DbChatSession } from "@rag-saldivia/db"
import { actionDeleteSession, actionCreateSession } from "@/app/actions/chat"
import { useSidebar } from "./ChatLayout"

export function SessionList({ sessions, defaultCollection }: { sessions: DbChatSession[]; defaultCollection?: string }) {
  const pathname = usePathname()
  const router = useRouter()
  const [creating, setCreating] = useState(false)
  const [search, setSearch] = useState("")
  const { open } = useSidebar()
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)

  const filtered = useMemo(() => {
    if (!search.trim()) return sessions
    const q = search.toLowerCase()
    return sessions.filter((s) => s.title.toLowerCase().includes(q))
  }, [sessions, search])

  async function handleNew() {
    setCreating(true)
    try {
      const result = await actionCreateSession({ collection: defaultCollection ?? "default" })
      router.push(`/chat/${result?.data?.id}`)
    } finally {
      setCreating(false)
    }
  }

  function handleDeleteClick(id: string, e: React.MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    setDeleteTarget(id)
  }

  const confirmDelete = useCallback(async () => {
    if (!deleteTarget) return
    const id = deleteTarget
    setDeleteTarget(null)
    await actionDeleteSession({ id })
    if (pathname === `/chat/${id}`) router.push("/chat")
  }, [deleteTarget, pathname, router])

  return (
    <aside
      className="shrink-0 border-r border-border bg-surface flex flex-col h-full overflow-hidden transition-[width] duration-200"
      style={{ width: open ? "260px" : "0px", borderRightWidth: open ? "1px" : "0px" }}
    >
      <div style={{ width: "260px", minWidth: "260px" }} className="flex flex-col h-full">
        {/* Header */}
        <div className="flex items-center justify-between" style={{ padding: "12px 12px 8px" }}>
          <span className="text-xs font-semibold uppercase tracking-wide text-fg-subtle">
            Chats
          </span>
          <button
            type="button"
            onClick={handleNew}
            disabled={creating}
            className="flex items-center justify-center shrink-0 rounded-lg text-fg-muted hover:text-fg hover:bg-surface-2 transition-colors disabled:opacity-40"
            style={{ width: "36px", height: "36px" }}
            title="Nuevo chat"
            aria-label="Nuevo chat"
          >
            <Plus size={18} aria-hidden />
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
              tabIndex={open ? 0 : -1}
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
                tabIndex={open ? 0 : -1}
              >
                <MessageSquare size={14} className="shrink-0 opacity-50" />
                <span className="flex-1 truncate">{session.title}</span>
                <button
                  onClick={(e) => handleDeleteClick(session.id, e)}
                  className="opacity-0 group-hover:opacity-50 hover:!opacity-100 transition-opacity p-0.5 text-fg-subtle hover:text-destructive"
                  title="Eliminar"
                  tabIndex={open ? 0 : -1}
                >
                  <Trash2 size={12} />
                </button>
              </Link>
            )
          })}
        </nav>

        <ConfirmDialog
          open={deleteTarget !== null}
          onOpenChange={(o) => { if (!o) setDeleteTarget(null) }}
          title="¿Eliminar esta sesión?"
          description="Se perderán todos los mensajes de esta conversación."
          onConfirm={confirmDelete}
        />
      </div>
    </aside>
  )
}
