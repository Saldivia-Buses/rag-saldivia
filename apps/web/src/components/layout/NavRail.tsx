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
  SquarePen,
} from "lucide-react"
import { cn } from "@/lib/utils"
import type { CurrentUser } from "@/lib/auth/current-user"
import { actionLogout } from "@/app/actions/auth"
import { actionCreateSession } from "@/app/actions/chat"
import { useRouter } from "next/navigation"

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
          aria-label={item.label}
          className={cn(
            "flex items-center justify-center rounded-lg transition-colors",
            active
              ? "bg-accent-subtle text-accent"
              : "text-fg-muted hover:bg-surface-2 hover:text-fg"
          )}
          style={{ width: "42px", height: "42px" }}
        >
          {item.icon}
        </Link>
      </TooltipTrigger>
      <TooltipContent side="right" sideOffset={16}>
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
  const router = useRouter()

  async function handleNewChat() {
    const session = await actionCreateSession({ collection: "tecpia" })
    router.push(`/chat/${session!.id}`)
  }

  return (
    <TooltipProvider delayDuration={100}>
      <nav
        aria-label="Navegación principal"
        className="flex flex-col items-center h-screen shrink-0 bg-surface border-r border-border"
        style={{ width: 56, padding: "16px 0 12px", gap: "4px" }}
      >
        {/* Brand */}
        <div style={{ marginBottom: "4px" }}>
          <div className="flex items-center justify-center rounded-lg bg-accent" style={{ width: "32px", height: "32px" }}>
            <span className="text-xs font-bold text-accent-fg select-none">S</span>
          </div>
        </div>

        {/* New chat button */}
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type="button"
              onClick={handleNewChat}
              aria-label="Nuevo chat"
              className="flex items-center justify-center rounded-lg text-fg-muted hover:bg-surface-2 hover:text-fg transition-colors"
              style={{ width: "42px", height: "42px" }}
            >
              <SquarePen size={16} aria-hidden />
            </button>
          </TooltipTrigger>
          <TooltipContent side="right" sideOffset={16}>
            Nuevo chat
          </TooltipContent>
        </Tooltip>

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
          <Tooltip>
            <TooltipTrigger asChild>
              <div className="text-fg-muted">
                <ThemeToggle />
              </div>
            </TooltipTrigger>
            <TooltipContent side="right" sideOffset={16}>
              Cambiar tema
            </TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                type="button"
                aria-label="Cerrar sesión"
                onClick={() => actionLogout()}
                className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-fg hover:bg-surface-2 transition-colors"
                style={{ width: "42px", height: "42px" }}
              >
                <LogOut size={15} aria-hidden />
              </button>
            </TooltipTrigger>
            <TooltipContent side="right" sideOffset={16}>
              Cerrar sesión
            </TooltipContent>
          </Tooltip>
        </div>
      </nav>
    </TooltipProvider>
  )
}
