"use client"

import Link from "next/link"
import { usePathname, useRouter } from "next/navigation"
import { useState, useRef, useEffect } from "react"
import { Plus, MessageSquare, Trash2, Pencil, Check, X } from "lucide-react"
import type { DbChatSession } from "@rag-saldivia/db"
import { actionDeleteSession, actionCreateSession, actionRenameSession } from "@/app/actions/chat"

export function SessionList({ sessions }: { sessions: DbChatSession[] }) {
  const pathname = usePathname()
  const router = useRouter()
  const [creating, setCreating] = useState(false)
  const [renamingId, setRenamingId] = useState<string | null>(null)
  const [renameValue, setRenameValue] = useState("")
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (renamingId && inputRef.current) {
      inputRef.current.focus()
      inputRef.current.select()
    }
  }, [renamingId])

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

  function startRename(session: DbChatSession, e: React.MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    setRenamingId(session.id)
    setRenameValue(session.title)
  }

  async function commitRename(id: string) {
    const trimmed = renameValue.trim()
    if (trimmed && trimmed !== sessions.find((s) => s.id === id)?.title) {
      await actionRenameSession(id, trimmed)
    }
    setRenamingId(null)
  }

  function cancelRename() {
    setRenamingId(null)
    setRenameValue("")
  }

  function handleRenameKeyDown(e: React.KeyboardEvent, id: string) {
    if (e.key === "Enter") {
      e.preventDefault()
      commitRename(id)
    } else if (e.key === "Escape") {
      cancelRename()
    }
  }

  return (
    <div
      className="w-64 flex-shrink-0 border-r flex flex-col"
      style={{ borderColor: "var(--border)" }}
    >
      {/* Header */}
      <div className="p-3 border-b flex items-center justify-between" style={{ borderColor: "var(--border)" }}>
        <span className="text-sm font-medium">Sesiones</span>
        <button
          onClick={handleNew}
          disabled={creating}
          className="p-1.5 rounded-md hover:opacity-80 transition-opacity disabled:opacity-40"
          title="Nueva sesión"
        >
          <Plus size={16} />
        </button>
      </div>

      {/* Sessions */}
      <div className="flex-1 overflow-y-auto p-2 space-y-0.5">
        {sessions.length === 0 && (
          <p className="px-3 py-4 text-sm text-center" style={{ color: "var(--muted-foreground)" }}>
            Sin sesiones. Creá una nueva.
          </p>
        )}
        {sessions.map((session) => {
          const active = pathname === `/chat/${session.id}`
          const isRenaming = renamingId === session.id

          return (
            <div key={session.id}>
              {isRenaming ? (
                // Modo edición inline
                <div
                  className="flex items-center gap-1 px-2 py-1.5 rounded-md"
                  style={{ background: "var(--accent)" }}
                >
                  <MessageSquare size={14} className="flex-shrink-0 opacity-60" />
                  <input
                    ref={inputRef}
                    value={renameValue}
                    onChange={(e) => setRenameValue(e.target.value)}
                    onKeyDown={(e) => handleRenameKeyDown(e, session.id)}
                    onBlur={() => commitRename(session.id)}
                    className="flex-1 min-w-0 text-sm bg-transparent outline-none border-b"
                    style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
                    maxLength={80}
                  />
                  <button
                    onMouseDown={(e) => { e.preventDefault(); commitRename(session.id) }}
                    className="p-0.5 hover:opacity-80 transition-opacity flex-shrink-0"
                    title="Confirmar"
                  >
                    <Check size={12} />
                  </button>
                  <button
                    onMouseDown={(e) => { e.preventDefault(); cancelRename() }}
                    className="p-0.5 hover:opacity-80 transition-opacity flex-shrink-0"
                    title="Cancelar"
                  >
                    <X size={12} />
                  </button>
                </div>
              ) : (
                // Modo normal
                <Link
                  href={`/chat/${session.id}`}
                  className="flex items-center gap-2 px-3 py-2 rounded-md text-sm group transition-colors"
                  style={{
                    background: active ? "var(--accent)" : "transparent",
                    color: active ? "var(--foreground)" : "var(--muted-foreground)",
                  }}
                >
                  <MessageSquare size={14} className="flex-shrink-0 opacity-60" />
                  <span className="flex-1 truncate">{session.title}</span>
                  <button
                    onClick={(e) => startRename(session, e)}
                    className="opacity-0 group-hover:opacity-60 hover:!opacity-100 transition-opacity p-0.5"
                    title="Renombrar sesión"
                  >
                    <Pencil size={12} />
                  </button>
                  <button
                    onClick={(e) => handleDelete(session.id, e)}
                    className="opacity-0 group-hover:opacity-60 hover:!opacity-100 transition-opacity p-0.5"
                    title="Eliminar sesión"
                  >
                    <Trash2 size={12} />
                  </button>
                </Link>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
