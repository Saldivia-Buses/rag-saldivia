/**
 * Admin layout with tab navigation.
 *
 * Renders a header with title + tabs (Dashboard, Usuarios, Roles).
 * Active tab is highlighted based on current pathname with smooth transitions.
 *
 * Rendered by: app/(app)/admin/layout.tsx
 */

"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { LayoutDashboard, Users, Shield, Map, Lock, FolderOpen, SlidersHorizontal, KeyRound } from "lucide-react"
import { cn } from "@/lib/utils"

const TABS = [
  { href: "/admin", label: "Dashboard", exact: true, Icon: LayoutDashboard },
  { href: "/admin/users", label: "Usuarios", exact: false, Icon: Users },
  { href: "/admin/roles", label: "Roles", exact: false, Icon: Shield },
  { href: "/admin/areas", label: "Áreas", exact: false, Icon: Map },
  { href: "/admin/permissions", label: "Permisos", exact: false, Icon: Lock },
  { href: "/admin/collections", label: "Colecciones", exact: false, Icon: FolderOpen },
  { href: "/admin/config", label: "Config RAG", exact: false, Icon: SlidersHorizontal },
  { href: "/admin/sso", label: "SSO", exact: false, Icon: KeyRound },
] as const

export function AdminLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()

  function isActive(tab: (typeof TABS)[number]) {
    if (tab.exact) return pathname === tab.href
    return pathname.startsWith(tab.href)
  }

  return (
    <div className="h-full overflow-y-auto bg-bg">
      <div style={{ maxWidth: "960px", marginLeft: "auto", marginRight: "auto", padding: "32px" }}>
        {/* Header */}
        <h1 className="text-xl font-semibold text-fg" style={{ marginBottom: "20px" }}>
          Administración
        </h1>

        {/* Tabs */}
        <div
          className="flex overflow-x-auto"
          style={{ marginBottom: "28px", gap: "4px", borderBottom: "1px solid var(--border)" }}
        >
          {TABS.map((tab) => {
            const active = isActive(tab)
            return (
              <Link
                key={tab.href}
                href={tab.href}
                className={cn(
                  "flex items-center text-sm font-medium transition-all duration-200 relative",
                  active
                    ? "text-accent"
                    : "text-fg-muted hover:text-fg"
                )}
                style={{ padding: "10px 16px", gap: "6px" }}
              >
                <tab.Icon size={15} />
                {tab.label}
                {/* Active indicator — animated bar */}
                <span
                  className="absolute bottom-0 left-2 right-2 rounded-full transition-all duration-200"
                  style={{
                    height: active ? "2px" : "0px",
                    background: "var(--accent)",
                    opacity: active ? 1 : 0,
                  }}
                />
              </Link>
            )
          })}
        </div>

        {/* Content */}
        {children}
      </div>
    </div>
  )
}
