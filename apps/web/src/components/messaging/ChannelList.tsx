/**
 * Channel sidebar — lists user's channels grouped by type.
 * Highlights the active channel. Shows unread badges.
 */
"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { useMemo, useState, memo } from "react"
import { Hash, Lock, MessageCircle, Users, Plus } from "lucide-react"
import { cn } from "@/lib/utils"
import { UnreadBadge } from "./UnreadBadge"
import { ChannelCreateDialog } from "./ChannelCreateDialog"
import { DirectMessageDialog } from "./DirectMessageDialog"

type ChannelData = {
  id: string
  type: string
  name: string | null
  description: string | null
  topic: string | null
  channelMembers: Array<{ userId: number }>
}

function channelIcon(type: string) {
  switch (type) {
    case "public": return Hash
    case "private": return Lock
    case "dm": return MessageCircle
    case "group_dm": return Users
    default: return Hash
  }
}

function channelDisplayName(channel: ChannelData, _userId: number): string {
  if (channel.name) return channel.name
  // For DMs, show the other user (we don't have names here, show "Mensaje directo")
  if (channel.type === "dm") return "Mensaje directo"
  return "Canal sin nombre"
}

export const ChannelList = memo(function ChannelList({
  channels,
  unreadCounts,
  userId,
  allUsers,
}: {
  channels: ChannelData[]
  unreadCounts: Record<string, number>
  userId: number
  allUsers?: Array<{ id: number; name: string; email: string }>
}) {
  const pathname = usePathname()
  const [showCreateChannel, setShowCreateChannel] = useState(false)
  const [showCreateDM, setShowCreateDM] = useState(false)
  const activeChannelId = pathname.startsWith("/messaging/")
    ? pathname.split("/messaging/")[1]?.split("/")[0]
    : null

  const grouped = useMemo(() => {
    const groups: Record<string, ChannelData[]> = {
      public: [],
      private: [],
      dm: [],
      group_dm: [],
    }
    for (const ch of channels) {
      const key = ch.type as string
      if (groups[key]) groups[key].push(ch)
    }
    return groups
  }, [channels])

  const sections: Array<{ label: string; key: string; items: ChannelData[] }> = [
    { label: "Canales", key: "public", items: grouped.public ?? [] },
    { label: "Canales privados", key: "private", items: grouped.private ?? [] },
    { label: "Mensajes directos", key: "dm", items: [...(grouped.dm ?? []), ...(grouped.group_dm ?? [])] },
  ].filter((s) => s.items.length > 0)

  return (
    <aside className="shrink-0 border-r border-border bg-surface flex flex-col h-full" style={{ width: "272px" }}>
      <div className="border-b border-border flex items-center justify-between" style={{ padding: "14px 16px", minHeight: "52px" }}>
        <h2 className="text-sm font-semibold text-fg tracking-tight">Mensajería</h2>
        <div className="flex items-center gap-0.5">
          <button
            type="button"
            onClick={() => setShowCreateChannel(true)}
            className="flex items-center justify-center rounded-md text-fg-subtle hover:text-fg hover:bg-surface-2 transition-colors"
            style={{ width: "28px", height: "28px" }}
            title="Crear canal"
          >
            <Plus className="h-4 w-4" />
          </button>
          {allUsers && (
            <button
              type="button"
              onClick={() => setShowCreateDM(true)}
              className="flex items-center justify-center rounded-md text-fg-subtle hover:text-fg hover:bg-surface-2 transition-colors"
              style={{ width: "28px", height: "28px" }}
              title="Mensaje directo"
            >
              <MessageCircle className="h-4 w-4" />
            </button>
          )}
        </div>
      </div>

      <nav className="flex-1 overflow-y-auto" style={{ padding: "12px 10px" }}>
        {sections.map((section) => (
          <div key={section.key} style={{ marginBottom: "20px" }}>
            <div style={{ padding: "0 8px", marginBottom: "6px" }}>
              <span className="text-xs font-medium text-fg-subtle uppercase" style={{ letterSpacing: "0.05em" }}>
                {section.label}
              </span>
            </div>
            <ul className="flex flex-col" style={{ gap: "2px" }}>
              {section.items.map((ch) => {
                const Icon = channelIcon(ch.type)
                const unread = unreadCounts[ch.id] ?? 0
                const isActive = ch.id === activeChannelId

                return (
                  <li key={ch.id}>
                    <Link
                      href={`/messaging/${ch.id}`}
                      className={cn(
                        "flex items-center gap-2.5 text-sm transition-colors",
                        isActive
                          ? "bg-accent text-accent-fg"
                          : "text-fg-muted hover:bg-surface-2 hover:text-fg",
                        unread > 0 && !isActive && "font-medium text-fg",
                      )}
                      style={{ padding: "8px 12px", borderRadius: "8px" }}
                    >
                      <Icon className="h-4 w-4 shrink-0 opacity-70" />
                      <span className="truncate flex-1">
                        {channelDisplayName(ch, userId)}
                      </span>
                      {!isActive && <UnreadBadge count={unread} />}
                    </Link>
                  </li>
                )
              })}
            </ul>
          </div>
        ))}

        {channels.length === 0 && (
          <div className="px-2 py-8 text-center">
            <p className="text-sm text-fg-subtle">No hay canales todavía</p>
          </div>
        )}
      </nav>

      <ChannelCreateDialog open={showCreateChannel} onClose={() => setShowCreateChannel(false)} />
      {allUsers && (
        <DirectMessageDialog
          open={showCreateDM}
          onClose={() => setShowCreateDM(false)}
          users={allUsers}
          currentUserId={userId}
        />
      )}
    </aside>
  )
})
