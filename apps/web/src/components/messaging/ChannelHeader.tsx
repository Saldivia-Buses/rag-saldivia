/**
 * Channel header — shows channel name, topic, and member count.
 */

import { Hash, Lock, MessageCircle, Users } from "lucide-react"

type ChannelData = {
  id: string
  type: string
  name: string | null
  description: string | null
  topic: string | null
}

const CHANNEL_ICONS = {
  public: Hash,
  private: Lock,
  dm: MessageCircle,
  group_dm: Users,
} as const

export function ChannelHeader({
  channel,
  memberCount,
}: {
  channel: ChannelData
  memberCount: number
}) {
  const Icon = CHANNEL_ICONS[channel.type as keyof typeof CHANNEL_ICONS] ?? Hash

  return (
    <header className="shrink-0 border-b border-border bg-surface px-4 py-3 flex items-center gap-3">
      <Icon className="h-5 w-5 text-fg-muted shrink-0" />
      <div className="flex-1 min-w-0">
        <h1 className="text-sm font-semibold text-fg truncate">
          {channel.name ?? "Mensaje directo"}
        </h1>
        {channel.topic && (
          <p className="text-xs text-fg-subtle truncate">{channel.topic}</p>
        )}
      </div>
      <div className="shrink-0 flex items-center gap-1 text-xs text-fg-subtle">
        <Users className="h-3.5 w-3.5" />
        <span>{memberCount}</span>
      </div>
    </header>
  )
}
