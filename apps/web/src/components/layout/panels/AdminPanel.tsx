"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { Users, Database, ShieldCheck, SlidersHorizontal, Activity } from "lucide-react"
import { Separator } from "@/components/ui/separator"

type NavItem = { href: string; label: string; icon: React.ReactNode }

const ADMIN_SECTIONS: Array<{ label: string; items: NavItem[] }> = [
  {
    label: "Gestión",
    items: [
      { href: "/admin/users", label: "Usuarios", icon: <Users size={14} /> },
      { href: "/admin/areas", label: "Áreas", icon: <Database size={14} /> },
      { href: "/admin/permissions", label: "Permisos", icon: <ShieldCheck size={14} /> },
      { href: "/admin/rag-config", label: "Config RAG", icon: <SlidersHorizontal size={14} /> },
    ],
  },
  {
    label: "Observabilidad",
    items: [
      { href: "/admin/system", label: "Sistema", icon: <Activity size={14} /> },
      { href: "/admin/ingestion", label: "Monitoring ingesta", icon: <Activity size={14} /> },
      { href: "/admin/analytics", label: "Analytics", icon: <Activity size={14} /> },
      { href: "/admin/knowledge-gaps", label: "Brechas", icon: <Activity size={14} /> },
    ],
  },
]

export function AdminPanel() {
  const pathname = usePathname()

  return (
    <div
      className="flex flex-col h-full"
      style={{ background: "var(--sidebar-bg)" }}
    >
      <div className="px-3 py-3 flex-shrink-0">
        <span
          className="text-xs font-semibold uppercase tracking-wider"
          style={{ color: "var(--muted-foreground)" }}
        >
          Admin
        </span>
      </div>
      <Separator />
      <div className="flex-1 overflow-y-auto py-2">
        {ADMIN_SECTIONS.map((section) => (
          <div key={section.label} className="mb-3">
            <p
              className="px-3 py-1 text-xs font-medium"
              style={{ color: "var(--muted-foreground)" }}
            >
              {section.label}
            </p>
            {section.items.map((item) => {
              const active = pathname.startsWith(item.href)
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className="flex items-center gap-2 px-3 py-1.5 text-sm rounded-md mx-1 transition-colors"
                  style={{
                    background: active ? "var(--accent)" : "transparent",
                    color: active ? "white" : "var(--foreground)",
                    opacity: active ? 1 : 0.75,
                  }}
                >
                  {item.icon}
                  {item.label}
                </Link>
              )
            })}
          </div>
        ))}
      </div>
    </div>
  )
}
