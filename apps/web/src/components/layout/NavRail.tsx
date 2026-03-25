"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { ThemeToggle } from "@/components/ui/theme-toggle"
import {
  MessageSquare,
  FolderOpen,
  Upload,
  FileText,
  Settings,
  Users,
  LogOut,
  Bookmark,
} from "lucide-react"
import type { CurrentUser } from "@/lib/auth/current-user"

type NavItem = {
  href: string
  label: string
  icon: React.ReactNode
  roles?: Array<"admin" | "area_manager" | "user">
}

const NAV_ITEMS: NavItem[] = [
  { href: "/chat", label: "Chat", icon: <MessageSquare size={18} /> },
  { href: "/collections", label: "Colecciones", icon: <FolderOpen size={18} /> },
  { href: "/upload", label: "Subir docs", icon: <Upload size={18} /> },
  { href: "/saved", label: "Guardados", icon: <Bookmark size={18} /> },
  { href: "/audit", label: "Audit", icon: <FileText size={18} />, roles: ["admin", "area_manager"] },
  { href: "/admin", label: "Admin", icon: <Users size={18} />, roles: ["admin"] },
  { href: "/settings", label: "Configuración", icon: <Settings size={18} /> },
]

function NavIcon({
  item,
  active,
}: {
  item: NavItem
  active: boolean
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Link
          href={item.href}
          className="w-9 h-9 flex items-center justify-center rounded-md transition-colors"
          style={{
            background: active ? "var(--accent)" : "transparent",
            color: active ? "white" : "rgba(255,255,255,0.45)",
          }}
          onMouseEnter={(e) => {
            if (!active) e.currentTarget.style.color = "rgba(255,255,255,0.85)"
          }}
          onMouseLeave={(e) => {
            if (!active) e.currentTarget.style.color = "rgba(255,255,255,0.45)"
          }}
        >
          {item.icon}
        </Link>
      </TooltipTrigger>
      <TooltipContent side="right" sideOffset={8}>
        {item.label}
      </TooltipContent>
    </Tooltip>
  )
}

export function NavRail({
  user,
  hidden,
}: {
  user: CurrentUser
  hidden?: boolean
}) {
  const pathname = usePathname()
  const visible = NAV_ITEMS.filter(
    (i) => !i.roles || i.roles.includes(user.role)
  )

  if (hidden) return null

  return (
    <TooltipProvider delayDuration={100}>
      <nav
        className="flex flex-col items-center py-3 gap-1 h-screen flex-shrink-0"
        style={{ width: 44, background: "var(--nav-bg)" }}
      >
        {/* Brand */}
        <div
          className="w-7 h-7 rounded-md flex items-center justify-center mb-2 flex-shrink-0"
          style={{ background: "var(--accent)" }}
        >
          <span className="text-xs font-bold text-white select-none">R</span>
        </div>

        {/* Nav items */}
        <div className="flex flex-col items-center gap-1 flex-1">
          {visible.map((item) => (
            <NavIcon
              key={item.href}
              item={item}
              active={pathname.startsWith(item.href)}
            />
          ))}
        </div>

        {/* Bottom: theme toggle + logout */}
        <div className="flex flex-col items-center gap-1 flex-shrink-0">
          <div style={{ color: "rgba(255,255,255,0.45)" }}>
            <ThemeToggle />
          </div>
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                onClick={async () => {
                  await fetch("/api/auth/logout", { method: "DELETE" })
                  window.location.href = "/login"
                }}
                className="w-9 h-9 flex items-center justify-center rounded-md transition-colors"
                style={{ color: "rgba(255,255,255,0.35)" }}
                onMouseEnter={(e) =>
                  (e.currentTarget.style.color = "rgba(255,255,255,0.75)")
                }
                onMouseLeave={(e) =>
                  (e.currentTarget.style.color = "rgba(255,255,255,0.35)")
                }
              >
                <LogOut size={16} />
              </button>
            </TooltipTrigger>
            <TooltipContent side="right" sideOffset={8}>
              Cerrar sesión
            </TooltipContent>
          </Tooltip>
        </div>
      </nav>
    </TooltipProvider>
  )
}
