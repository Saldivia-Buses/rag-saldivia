"use client"

import Link from "next/link"
import { usePathname, useRouter } from "next/navigation"
import { useState, useRef, useEffect } from "react"
import { Plus, MessageSquare, Trash2, Pencil, Check, X, Tag, Download, GitBranch } from "lucide-react"
import type { DbChatSession } from "@rag-saldivia/db"
import { actionDeleteSession, actionCreateSession, actionRenameSession, actionAddTag, actionRemoveTag } from "@/app/actions/chat"
import { Badge } from "@/components/ui/badge"
import { exportToMarkdown, downloadFile } from "@/lib/export"

export function SessionList({ sessions }: { sessions: DbChatSession[] }) {
  const pathname = usePathname()
  const router = useRouter()
  const [creating, setCreating] = useState(false)
  const [renamingId, setRenamingId] = useState<string | null>(null)
  const [renameValue, setRenameValue] = useState("")
  const inputRef = useRef<HTMLInputElement>(null)

  // Tags y bulk
  const [sessionTags, setSessionTags] = useState<Record<string, string[]>>({})
  const [tagInput, setTagInput] = useState<Record<string, string>>({})
  const [showTagInput, setShowTagInput] = useState<string | null>(null)
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [filterTag, setFilterTag] = useState<string | null>(null)

  const allTags = [...new Set(Object.values(sessionTags).flat())]
  const filteredSessions = filterTag
    ? sessions.filter((s) => sessionTags[s.id]?.includes(filterTag))
    : sessions

  useEffect(() => {
    if (renamingId && inputRef.current) {
      inputRef.current.focus()
      inputRef.current.select()
    }
  }, [renamingId])

  async function handleAddTag(sessionId: string) {
    const tag = (tagInput[sessionId] ?? "").trim()
    if (!tag) return
    await actionAddTag(sessionId, tag)
    setSessionTags((prev) => ({ ...prev, [sessionId]: [...(prev[sessionId] ?? []), tag] }))
    setTagInput((prev) => ({ ...prev, [sessionId]: "" }))
    setShowTagInput(null)
  }

  async function handleRemoveTag(sessionId: string, tag: string, e: React.MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    await actionRemoveTag(sessionId, tag)
    setSessionTags((prev) => ({
      ...prev,
      [sessionId]: (prev[sessionId] ?? []).filter((t) => t !== tag),
    }))
  }

  function toggleSelect(id: string, e: React.MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  async function bulkDelete() {
    if (!confirm(`¿Eliminar ${selected.size} sesiones?`)) return
    for (const id of selected) {
      await actionDeleteSession(id)
    }
    setSelected(new Set())
    if ([...selected].some((id) => pathname === `/chat/${id}`)) router.push("/chat")
  }

  function bulkExport() {
    for (const id of selected) {
      const session = sessions.find((s) => s.id === id)
      if (!session) continue
      const md = exportToMarkdown({ title: session.title, collection: session.collection, createdAt: session.createdAt, messages: [] })
      downloadFile(md, `${session.title}.md`, "text/markdown")
    }
    setSelected(new Set())
  }

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
      className="w-64 shrink-0 border-r flex flex-col"
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

      {/* Filtro por tag */}
      {allTags.length > 0 && (
        <div className="px-3 py-2 border-b flex flex-wrap gap-1" style={{ borderColor: "var(--border)" }}>
          {allTags.map((t) => (
            <button
              key={t}
              onClick={() => setFilterTag(filterTag === t ? null : t)}
              className="text-xs px-2 py-0.5 rounded-full border transition-colors"
              style={{
                background: filterTag === t ? "var(--accent)" : "transparent",
                color: filterTag === t ? "white" : "var(--muted-foreground)",
                borderColor: "var(--border)",
              }}
            >
              #{t}
            </button>
          ))}
        </div>
      )}

      {/* Bulk toolbar */}
      {selected.size > 0 && (
        <div className="px-3 py-2 border-b flex items-center gap-2" style={{ borderColor: "var(--border)" }}>
          <span className="text-xs flex-1" style={{ color: "var(--muted-foreground)" }}>
            {selected.size} seleccionadas
          </span>
          <button onClick={bulkExport} title="Exportar" className="p-1 hover:opacity-80">
            <Download size={13} />
          </button>
          <button onClick={bulkDelete} title="Eliminar" className="p-1 hover:opacity-80" style={{ color: "var(--destructive)" }}>
            <Trash2 size={13} />
          </button>
          <button onClick={() => setSelected(new Set())} className="p-1 hover:opacity-80">
            <X size={13} />
          </button>
        </div>
      )}

      {/* Sessions */}
      <div className="flex-1 overflow-y-auto p-2 space-y-0.5">
        {filteredSessions.length === 0 && (
          <p className="px-3 py-4 text-sm text-center" style={{ color: "var(--muted-foreground)" }}>
            {filterTag ? `Sin sesiones con #${filterTag}` : "Sin sesiones. Creá una nueva."}
          </p>
        )}
        {filteredSessions.map((session) => {
          const active = pathname === `/chat/${session.id}`
          const isRenaming = renamingId === session.id
          const tags = sessionTags[session.id] ?? []

          return (
            <div key={session.id}>
              {isRenaming ? (
                // Modo edición inline
                <div
                  className="flex items-center gap-1 px-2 py-1.5 rounded-md"
                  style={{ background: "var(--accent)" }}
                >
                  <MessageSquare size={14} className="shrink-0 opacity-60" />
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
                    className="p-0.5 hover:opacity-80 transition-opacity shrink-0"
                    title="Confirmar"
                  >
                    <Check size={12} />
                  </button>
                  <button
                    onMouseDown={(e) => { e.preventDefault(); cancelRename() }}
                    className="p-0.5 hover:opacity-80 transition-opacity shrink-0"
                    title="Cancelar"
                  >
                    <X size={12} />
                  </button>
                </div>
              ) : (
                // Modo normal
                <div>
                  <div className="flex items-center">
                    <input
                      type="checkbox"
                      checked={selected.has(session.id)}
                      onChange={(e) => { e.stopPropagation(); toggleSelect(session.id, e as unknown as React.MouseEvent) }}
                      className="mr-1.5 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
                    />
                    <Link
                      href={`/chat/${session.id}`}
                      className="flex items-center gap-2 flex-1 px-2 py-1.5 rounded-md text-sm group transition-colors"
                      style={{
                        background: active ? "var(--accent)" : "transparent",
                        color: active ? "var(--foreground)" : "var(--muted-foreground)",
                      }}
                    >
                      <MessageSquare size={14} className="shrink-0 opacity-60" />
                      <span className="flex-1 truncate">{session.title}</span>
                      {(session as DbChatSession & { forkedFrom?: string | null }).forkedFrom && (
                        <GitBranch size={11} className="shrink-0 opacity-40" title="Sesión bifurcada" />
                      )}
                      <button
                        onClick={(e) => { e.preventDefault(); e.stopPropagation(); setShowTagInput(showTagInput === session.id ? null : session.id) }}
                        className="opacity-0 group-hover:opacity-60 hover:opacity-100! transition-opacity p-0.5"
                        title="Agregar etiqueta"
                      >
                        <Tag size={12} />
                      </button>
                      <button
                        onClick={(e) => startRename(session, e)}
                        className="opacity-0 group-hover:opacity-60 hover:opacity-100! transition-opacity p-0.5"
                        title="Renombrar sesión"
                      >
                        <Pencil size={12} />
                      </button>
                      <button
                        onClick={(e) => handleDelete(session.id, e)}
                        className="opacity-0 group-hover:opacity-60 hover:opacity-100! transition-opacity p-0.5"
                        title="Eliminar sesión"
                      >
                        <Trash2 size={12} />
                      </button>
                    </Link>
                  </div>
                  {/* Tags badges */}
                  {tags.length > 0 && (
                    <div className="flex flex-wrap gap-1 pl-6 pb-1">
                      {tags.map((t) => (
                        <Badge
                          key={t}
                          variant="outline"
                          className="text-xs px-1.5 py-0 h-4 cursor-pointer hover:opacity-70"
                          onClick={(e) => handleRemoveTag(session.id, t, e)}
                          title="Clic para eliminar tag"
                        >
                          #{t}
                        </Badge>
                      ))}
                    </div>
                  )}
                  {/* Tag input inline */}
                  {showTagInput === session.id && (
                    <div className="flex items-center gap-1 pl-6 pr-2 pb-1">
                      <input
                        autoFocus
                        value={tagInput[session.id] ?? ""}
                        onChange={(e) => setTagInput((p) => ({ ...p, [session.id]: e.target.value }))}
                        onKeyDown={(e) => { if (e.key === "Enter") handleAddTag(session.id); if (e.key === "Escape") setShowTagInput(null) }}
                        placeholder="nueva-etiqueta"
                        className="flex-1 text-xs px-2 py-0.5 rounded border outline-none"
                        style={{ borderColor: "var(--border)", background: "var(--muted)", color: "var(--foreground)" }}
                      />
                      <button onClick={() => handleAddTag(session.id)} className="p-0.5 hover:opacity-80"><Check size={11} /></button>
                      <button onClick={() => setShowTagInput(null)} className="p-0.5 hover:opacity-80"><X size={11} /></button>
                    </div>
                  )}
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
