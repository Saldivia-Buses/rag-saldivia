"use client"

import { usePathname } from "next/navigation"
import { ChatPanel } from "./panels/ChatPanel"
import { AdminPanel } from "./panels/AdminPanel"
import { ProjectsPanel } from "./panels/ProjectsPanel"
import type { DbProject } from "@rag-saldivia/db"

const PANEL_WIDTH = 168

export function SecondaryPanel({
  hidden,
  initialProjects,
}: {
  hidden?: boolean
  initialProjects: DbProject[]
}) {
  const pathname = usePathname()

  if (hidden) return null

  const panel = pathname.startsWith("/chat")
    ? <ChatPanel />
    : pathname.startsWith("/admin")
    ? <AdminPanel />
    : pathname.startsWith("/projects")
    ? <ProjectsPanel initialProjects={initialProjects} />
    : null

  if (!panel) return null

  return (
    <div
      className="flex-shrink-0 border-r border-border h-screen overflow-hidden bg-surface"
      style={{ width: PANEL_WIDTH }}
    >
      {panel}
    </div>
  )
}
