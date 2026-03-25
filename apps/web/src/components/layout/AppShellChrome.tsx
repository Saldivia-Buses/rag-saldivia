"use client"

import { NavRail } from "./NavRail"
import { SecondaryPanel } from "./SecondaryPanel"
import { useZenMode } from "@/hooks/useZenMode"
import { useNotifications } from "@/hooks/useNotifications"
import type { CurrentUser } from "@/lib/auth/current-user"

/**
 * Client Component wrapper de la shell.
 * Concentra todo el estado de UI: zen mode (F1.11), notificaciones (F1.12), hotkeys (F1.14).
 * AppShell.tsx sigue siendo Server Component y solo renderiza este componente.
 */
export function AppShellChrome({
  user,
  children,
}: {
  user: CurrentUser
  children: React.ReactNode
}) {
  const { isZen } = useZenMode()
  const { unreadCount } = useNotifications()

  return (
    <div className="flex h-screen overflow-hidden">
      <NavRail user={user} hidden={isZen} unreadCount={unreadCount} />
      <SecondaryPanel hidden={isZen} />
      <main className="flex-1 overflow-y-auto">{children}</main>
      {isZen && (
        <div
          className="fixed bottom-4 right-4 px-3 py-1.5 rounded-full text-xs font-medium shadow-lg pointer-events-none"
          style={{ background: "var(--nav-bg)", color: "rgba(255,255,255,0.7)" }}
        >
          ESC para salir del modo Zen
        </div>
      )}
    </div>
  )
}
