"use client"

import { NavRail } from "./NavRail"
import { SecondaryPanel } from "./SecondaryPanel"
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
  // Estado de zen mode — se agrega en Fase 1 (F1.11)
  const isZen = false

  return (
    <div className="flex h-screen overflow-hidden">
      <NavRail user={user} hidden={isZen} />
      <SecondaryPanel hidden={isZen} />
      <main className="flex-1 overflow-y-auto">{children}</main>
    </div>
  )
}
