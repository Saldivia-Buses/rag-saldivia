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
  Settings,
  LogOut,
} from "lucide-react"
import { cn } from "@/lib/utils"
import type { CurrentUser } from "@/lib/auth/current-user"
import { actionLogout } from "@/app/actions/auth"

type Changelog = { version: string; entries: { version: string; content: string }[] }

type NavItem = {
  href: string
  label: string
  icon: React.ReactNode
}

const NAV_ITEMS: NavItem[] = [
  { href: "/chat",        label: "Chat",          icon: <MessageSquare size={16} /> },
  { href: "/collections", label: "Colecciones",   icon: <FolderOpen size={16} /> },
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
  user: _user,
  changelog: _changelog,
}: {
  user: CurrentUser
  changelog: Changelog
}) {
  const pathname = usePathname()

  return (
    <TooltipProvider delayDuration={100}>
      <nav
        className="flex flex-col items-center py-3 gap-1 h-screen shrink-0 bg-surface border-r border-border"
        style={{ width: 44 }}
      >
        {/* Brand */}
        <div className="mb-2">
          <div className="w-7 h-7 rounded-md flex items-center justify-center bg-accent">
            <span className="text-xs font-bold text-accent-fg select-none">S</span>
          </div>
        </div>

        {/* Separador */}
        <div className="w-5 h-px bg-border mb-1" />

        {/* Nav items */}
        <div className="flex flex-col items-center gap-0.5 flex-1">
          {NAV_ITEMS.map((item) => (
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
