"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { Plus, FolderKanban } from "lucide-react"
import { Separator } from "@/components/ui/separator"
import type { DbProject } from "@rag-saldivia/db"

export function ProjectsPanel() {
  const pathname = usePathname()
  const [projects, setProjects] = useState<DbProject[]>([])

  useEffect(() => {
    fetch("/api/projects")
      .then((r) => r.json())
      .then((d: { ok: boolean; data?: DbProject[] }) => {
        if (d.ok) setProjects(d.data ?? [])
      })
      .catch(() => {})
  }, [])

  return (
    <div className="flex flex-col h-full" style={{ background: "var(--sidebar-bg)" }}>
      <div className="px-3 py-3 flex items-center justify-between flex-shrink-0">
        <span className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Proyectos
        </span>
        <Link href="/projects" className="w-6 h-6 flex items-center justify-center rounded-md transition-colors hover:opacity-80" style={{ color: "var(--muted-foreground)" }}>
          <Plus size={14} />
        </Link>
      </div>
      <Separator />
      <div className="flex-1 overflow-y-auto py-2 px-2 space-y-0.5">
        {projects.length === 0 ? (
          <p className="text-xs px-2 py-2" style={{ color: "var(--muted-foreground)" }}>Sin proyectos</p>
        ) : (
          projects.map((p) => {
            const active = pathname.startsWith(`/projects/${p.id}`)
            return (
              <Link
                key={p.id}
                href={`/projects/${p.id}`}
                className="flex items-center gap-2 px-2 py-1.5 rounded-md text-sm transition-colors"
                style={{
                  background: active ? "var(--accent)" : "transparent",
                  color: active ? "white" : "var(--foreground)",
                  opacity: active ? 1 : 0.75,
                }}
              >
                <FolderKanban size={13} className="shrink-0" />
                <span className="truncate">{p.name}</span>
              </Link>
            )
          })
        )}
      </div>
    </div>
  )
}
