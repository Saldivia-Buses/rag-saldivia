/**
 * DirectMessageDialog — create a DM or group DM with selected users.
 */
"use client"

import { useState, useTransition, useEffect } from "react"
import { useRouter } from "next/navigation"
import { X, Check } from "lucide-react"
import { Button } from "@/components/ui/button"
import { actionCreateChannel } from "@/app/actions/messaging"
import { cn } from "@/lib/utils"

type UserInfo = { id: number; name: string; email: string }

export function DirectMessageDialog({
  open,
  onClose,
  users,
  currentUserId,
}: {
  open: boolean
  onClose: () => void
  users: UserInfo[]
  currentUserId: number
}) {
  const router = useRouter()
  const [selected, setSelected] = useState<number[]>([])
  const [search, setSearch] = useState("")
  const [isPending, startTransition] = useTransition()

  // Reset on open
  useEffect(() => {
    if (open) {
      setSelected([])
      setSearch("")
    }
  }, [open])

  if (!open) return null

  const otherUsers = users.filter((u) => u.id !== currentUserId)
  const filtered = search.trim()
    ? otherUsers.filter(
        (u) =>
          u.name.toLowerCase().includes(search.toLowerCase()) ||
          u.email.toLowerCase().includes(search.toLowerCase()),
      )
    : otherUsers

  function toggleUser(userId: number) {
    setSelected((prev) =>
      prev.includes(userId) ? prev.filter((id) => id !== userId) : [...prev, userId],
    )
  }

  function handleCreate() {
    if (selected.length === 0) return

    const type = selected.length === 1 ? "dm" : "group_dm"
    const memberNames = selected
      .map((id) => users.find((u) => u.id === id)?.name)
      .filter(Boolean)
      .join(", ")

    startTransition(async () => {
      const result = await actionCreateChannel({
        type,
        memberIds: selected,
        ...(selected.length > 1 ? { name: memberNames } : {}),
      })
      if (result?.data) {
        onClose()
        router.push(`/messaging/${result.data.id}`)
      }
    })
  }

  return (
    <>
      <div className="fixed inset-0 z-50 bg-black/50" onClick={onClose} />

      <div className="fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 z-50 w-full max-w-sm bg-surface border border-border rounded-xl shadow-xl">
        <div className="flex items-center justify-between px-5 py-4 border-b border-border">
          <h2 className="text-base font-semibold text-fg">Mensaje directo</h2>
          <button type="button" onClick={onClose} className="text-fg-subtle hover:text-fg">
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="px-5 py-3">
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar usuarios..."
            className="w-full px-3 py-2 rounded-lg border border-border bg-bg text-sm text-fg placeholder:text-fg-subtle outline-none focus:border-accent"
            autoFocus
          />
        </div>

        <div className="px-2 max-h-60 overflow-y-auto">
          {filtered.map((user) => {
            const isSelected = selected.includes(user.id)
            return (
              <button
                key={user.id}
                type="button"
                onClick={() => toggleUser(user.id)}
                className={cn(
                  "w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors",
                  isSelected ? "bg-accent-subtle" : "hover:bg-surface-2",
                )}
              >
                <div className="h-7 w-7 rounded-full bg-accent-subtle text-accent flex items-center justify-center text-xs font-semibold shrink-0">
                  {user.name.charAt(0).toUpperCase()}
                </div>
                <div className="flex-1 text-left min-w-0">
                  <span className="text-fg font-medium truncate block">{user.name}</span>
                  <span className="text-fg-subtle text-xs truncate block">{user.email}</span>
                </div>
                {isSelected && <Check className="h-4 w-4 text-accent shrink-0" />}
              </button>
            )
          })}
          {filtered.length === 0 && (
            <p className="text-sm text-fg-subtle text-center py-4">No se encontraron usuarios</p>
          )}
        </div>

        <div className="flex items-center justify-between px-5 py-3 border-t border-border">
          <span className="text-xs text-fg-subtle">
            {selected.length} seleccionado{selected.length !== 1 ? "s" : ""}
          </span>
          <Button onClick={handleCreate} disabled={isPending || selected.length === 0}>
            {isPending ? "Creando..." : "Iniciar chat"}
          </Button>
        </div>
      </div>
    </>
  )
}
