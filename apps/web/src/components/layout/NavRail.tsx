"use client"

import { useState } from "react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { ThemeToggle } from "@/components/ui/theme-toggle"
import { WhatsNewPanel, useHasNewVersion } from "@/components/layout/WhatsNewPanel"
import {
  MessageSquare,
  FolderOpen,
  Upload,
  FileText,
  Settings,
  Users,
  LogOut,
  Bookmark,
  FolderKanban,
  Table2,
} from "lucide-react"
import { cn } from "@/lib/utils"
import type { CurrentUser } from "@/lib/auth/current-user"
import { actionLogout } from "@/app/actions/auth"

type Changelog = { version: string; entries: { version: string; content: string }[] }

type NavItem = {
  href: string
  label: string
  icon: React.ReactNode
  roles?: Array<"admin" | "area_manager" | "user">
}

const NAV_ITEMS: NavItem[] = [
  { href: "/chat",        label: "Chat",          icon: <MessageSquare size={16} /> },
  { href: "/collections", label: "Colecciones",   icon: <FolderOpen size={16} /> },
  { href: "/upload",      label: "Subir docs",    icon: <Upload size={16} /> },
  { href: "/saved",       label: "Guardados",     icon: <Bookmark size={16} /> },
  { href: "/projects",    label: "Proyectos",     icon: <FolderKanban size={16} /> },
  { href: "/extract",     label: "Extraer datos", icon: <Table2 size={16} /> },
  { href: "/audit",       label: "Audit",         icon: <FileText size={16} />, roles: ["admin", "area_manager"] },
  { href: "/admin",       label: "Admin",         icon: <Users size={16} />,    roles: ["admin"] },
  { href: "/settings",    label: "Configuración", icon: <Settings size={16} /> },
]

function NavIcon({ item, active }: { item: NavItem; active: boolean }) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Link
          href={item.href}
          className={cn(
            "w-9 h-9 flex items-center justify-center rounded-md transition-colors",
            active
              ? "bg-accent-subtle text-accent"
              : "text-fg-muted hover:bg-surface-2 hover:text-fg"
          )}
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
  unreadCount = 0,
  changelog,
}: {
  user: CurrentUser
  hidden?: boolean
  unreadCount?: number
  changelog: Changelog
}) {
  const pathname = usePathname()
  const [whatsNewOpen, setWhatsNewOpen] = useState(false)
  const hasNewVersion = useHasNewVersion(changelog.version)
  const visible = NAV_ITEMS.filter(
    (i) => !i.roles || i.roles.includes(user.role)
  )

  if (hidden) return null

  return (
    <TooltipProvider delayDuration={100}>
      <nav
        className="flex flex-col items-center py-3 gap-1 h-screen shrink-0 bg-surface border-r border-border"
        style={{ width: 44 }}
      >
        {/* Brand */}
        <div className="relative mb-2">
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                onClick={() => setWhatsNewOpen(true)}
                className="w-7 h-7 rounded-md flex items-center justify-center bg-accent hover:bg-accent-hover transition-colors"
              >
                <span className="text-xs font-bold text-accent-fg select-none">R</span>
              </button>
            </TooltipTrigger>
            <TooltipContent side="right" sideOffset={8}>¿Qué hay de nuevo?</TooltipContent>
          </Tooltip>
          {(unreadCount > 0 || hasNewVersion) && (
            <div className="absolute -top-1 -right-1 w-4 h-4 rounded-full flex items-center justify-center text-white bg-destructive text-[9px]">
              {unreadCount > 0 ? (unreadCount > 9 ? "9+" : unreadCount) : "·"}
            </div>
          )}
        </div>
        <WhatsNewPanel open={whatsNewOpen} onClose={() => setWhatsNewOpen(false)} changelog={changelog} />

        {/* Separador */}
        <div className="w-5 h-px bg-border mb-1" />

        {/* Nav items */}
        <div className="flex flex-col items-center gap-0.5 flex-1">
          {visible.map((item) => (
            <NavIcon
              key={item.href}
              item={item}
              active={pathname.startsWith(item.href)}
            />
          ))}
        </div>

        {/* Bottom: theme toggle + logout */}
        <div className="flex flex-col items-center gap-0.5 shrink-0">
          <div className="text-fg-muted">
            <ThemeToggle />
          </div>
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                type="button"
                aria-label="Cerrar sesión"
                onClick={() => actionLogout()}
                className="w-9 h-9 flex items-center justify-center rounded-md text-fg-subtle hover:text-fg hover:bg-surface-2 transition-colors"
              >
                <LogOut size={15} aria-hidden />
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
