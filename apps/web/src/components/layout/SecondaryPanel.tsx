"use client"

import { usePathname } from "next/navigation"
import { ChatPanel } from "./panels/ChatPanel"
import { AdminPanel } from "./panels/AdminPanel"

const PANEL_WIDTH = 168

export function SecondaryPanel({ hidden }: { hidden?: boolean }) {
  const pathname = usePathname()

  if (hidden) return null

  const panel = pathname.startsWith("/chat")
    ? <ChatPanel />
    : pathname.startsWith("/admin")
    ? <AdminPanel />
    : null

  if (!panel) return null

  return (
    <div
      className="flex-shrink-0 border-r h-screen overflow-hidden"
      style={{ width: PANEL_WIDTH, borderColor: "var(--border)" }}
    >
      {panel}
    </div>
  )
}
