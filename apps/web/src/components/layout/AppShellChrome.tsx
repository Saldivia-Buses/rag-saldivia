"use client"

import { useState } from "react"
import { NavRail } from "./NavRail"
import { SecondaryPanel } from "./SecondaryPanel"
import { CommandPalette } from "./CommandPalette"
import { useZenMode } from "@/hooks/useZenMode"
import { useNotifications } from "@/hooks/useNotifications"
import { useGlobalHotkeys } from "@/hooks/useGlobalHotkeys"
import type { CurrentUser } from "@/lib/auth/current-user"
import type { DbChatSession, DbProject } from "@rag-saldivia/db"

type Changelog = { version: string; entries: { version: string; content: string }[] }

type Props = {
  user: CurrentUser
  children: React.ReactNode
  initialSessions: Pick<DbChatSession, "id" | "title" | "collection">[]
  initialProjects: DbProject[]
  changelog: Changelog
}

/**
 * Client Component wrapper de la shell.
 * Concentra todo el estado de UI: zen mode, notificaciones, hotkeys.
 * AppShell.tsx sigue siendo Server Component y solo renderiza este componente.
 */
export function AppShellChrome({
  user,
  children,
  initialSessions,
  initialProjects,
  changelog,
}: Props) {
  const { isZen, toggleZen } = useZenMode()
  const { unreadCount } = useNotifications()
  const [paletteOpen, setPaletteOpen] = useState(false)
  useGlobalHotkeys({ onOpenPalette: () => setPaletteOpen(true) })

  return (
    <div className="flex h-screen overflow-hidden bg-bg">
      <NavRail user={user} hidden={isZen} unreadCount={unreadCount} changelog={changelog} />
      <SecondaryPanel hidden={isZen} initialProjects={initialProjects} />
      <main className="flex-1 overflow-y-auto bg-bg">{children}</main>
      <CommandPalette
        open={paletteOpen}
        onClose={() => setPaletteOpen(false)}
        user={user}
        onToggleZen={toggleZen}
        initialSessions={initialSessions}
      />
      {isZen && (
        <div className="fixed bottom-4 right-4 px-3 py-1.5 rounded-full text-xs font-medium shadow-lg pointer-events-none bg-fg text-bg">
          ESC para salir del modo Zen
        </div>
      )}
    </div>
  )
}
