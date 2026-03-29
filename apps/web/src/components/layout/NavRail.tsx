/**
 * Vertical navigation rail — the 64px-wide strip on the left edge of the app.
 *
 * Always visible on authenticated pages. Contains (top to bottom):
 *   - Sidebar toggle (only on /chat/* routes) or brand logo
 *   - New chat button
 *   - Separator
 *   - Nav links: Chat, Collections, Settings (with active state highlighting)
 *   - (spacer)
 *   - Theme toggle (light/dark)
 *   - Logout button
 *
 * All icons have tooltips (right-aligned) for discoverability.
 * Uses `usePathname` for active-state detection via prefix matching.
 *
 * Rendered by: AppShellChrome (client component)
 * Depends on: useSidebar (ChatLayout), server actions (logout, createSession)
 */
"use client"

import Link from "next/link"
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
  ShieldCheck,
} from "lucide-react"
import { cn } from "@/lib/utils"
import type { CurrentUser } from "@/lib/auth/current-user"
import { actionLogout } from "@/app/actions/auth"
import { actionCreateSession } from "@/app/actions/chat"
import { useSidebar } from "@/components/chat/ChatLayout"
import { useRouter, usePathname } from "next/navigation"
import { PanelLeft, PanelLeftClose } from "lucide-react"

type Changelog = { version: string; entries: { version: string; content: string }[] }

const BTN = "flex items-center justify-center rounded-xl transition-colors"
const BTN_SIZE = { width: "44px", height: "44px" } as const

type NavItem = {
  href: string
  label: string
  icon: React.ReactNode
}

const NAV_ITEMS: NavItem[] = [
  { href: "/chat",        label: "Chat",          icon: <MessageSquare size={20} /> },
  { href: "/collections", label: "Colecciones",   icon: <FolderOpen size={20} /> },
  { href: "/settings",    label: "Configuración", icon: <Settings size={20} /> },
]

/** Admin-only nav items — shown below the separator when user.role === "admin" */
const ADMIN_NAV_ITEMS: NavItem[] = [
  { href: "/admin", label: "Admin",      icon: <ShieldCheck size={20} /> },
]

function NavIcon({ item, active }: { item: NavItem; active: boolean }) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Link
          href={item.href}
          aria-label={item.label}
          className={cn(
            BTN,
            active
              ? "bg-accent-subtle text-accent"
              : "text-fg-muted hover:bg-surface-2 hover:text-fg"
          )}
          style={BTN_SIZE}
        >
          {item.icon}
        </Link>
      </TooltipTrigger>
      <TooltipContent side="right" sideOffset={12}>
        {item.label}
      </TooltipContent>
    </Tooltip>
  )
}

export function NavRail({
  user,
  changelog: _changelog,
}: {
  user: CurrentUser
  changelog: Changelog
}) {
  const isAdmin = user.role === "admin"
  const pathname = usePathname()
  const router = useRouter()
  const { open: sidebarOpen, toggle: toggleSidebar } = useSidebar()
  const isOnChat = pathname.startsWith("/chat")

  async function handleNewChat() {
    const session = await actionCreateSession({ collection: "tecpia" })
    router.push(`/chat/${session!.id}`)
  }

  return (
    <TooltipProvider delayDuration={100}>
      <nav
        aria-label="Navegación principal"
        className="flex flex-col items-center h-screen shrink-0 bg-surface border-r border-border"
        style={{ width: 64, padding: "12px 0", gap: "4px" }}
      >
        {/* Brand — always visible */}
        <div style={{ marginBottom: "2px" }}>
          <div
            className="flex items-center justify-center rounded-xl bg-accent"
            style={{ width: "38px", height: "38px" }}
          >
            <span className="text-sm font-bold text-accent-fg select-none">S</span>
          </div>
        </div>

        {/* Sidebar toggle — only on chat pages, below brand */}
        {isOnChat && (
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                type="button"
                onClick={toggleSidebar}
                aria-label={sidebarOpen ? "Ocultar chats" : "Mostrar chats"}
                className={`${BTN} ${sidebarOpen ? "text-fg" : "text-fg-muted"} hover:bg-surface-2 hover:text-fg`}
                style={BTN_SIZE}
              >
                {sidebarOpen ? <PanelLeftClose size={20} /> : <PanelLeft size={20} />}
              </button>
            </TooltipTrigger>
            <TooltipContent side="right" sideOffset={12}>
              {sidebarOpen ? "Ocultar chats" : "Mostrar chats"} (Ctrl+Shift+S)
            </TooltipContent>
          </Tooltip>
        )}

        {/* New chat */}
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type="button"
              onClick={handleNewChat}
              aria-label="Nuevo chat"
              className={`${BTN} text-fg-muted hover:bg-surface-2 hover:text-fg`}
              style={BTN_SIZE}
            >
              <SquarePen size={20} aria-hidden />
            </button>
          </TooltipTrigger>
          <TooltipContent side="right" sideOffset={12}>
            Nuevo chat
          </TooltipContent>
        </Tooltip>

        {/* Separator */}
        <div className="w-6 h-px bg-border" style={{ margin: "4px 0" }} />

        {/* Nav items */}
        <div className="flex flex-col items-center gap-1 flex-1">
          {NAV_ITEMS.map((item) => (
            <NavIcon
              key={item.href}
              item={item}
              active={pathname.startsWith(item.href)}
            />
          ))}

          {/* Admin-only section */}
          {isAdmin && (
            <>
              <div className="w-6 h-px bg-border" style={{ margin: "4px 0" }} />
              {ADMIN_NAV_ITEMS.map((item) => (
                <NavIcon
                  key={item.href}
                  item={item}
                  active={pathname.startsWith(item.href)}
                />
              ))}
            </>
          )}
        </div>

        {/* Bottom */}
        <div className="flex flex-col items-center gap-1 shrink-0">
          <Tooltip>
            <TooltipTrigger asChild>
              <div className="text-fg-muted">
                <ThemeToggle />
              </div>
            </TooltipTrigger>
            <TooltipContent side="right" sideOffset={12}>
              Cambiar tema
            </TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                type="button"
                aria-label="Cerrar sesión"
                onClick={() => actionLogout()}
                className={`${BTN} text-fg-subtle hover:text-fg hover:bg-surface-2`}
                style={BTN_SIZE}
              >
                <LogOut size={18} aria-hidden />
              </button>
            </TooltipTrigger>
            <TooltipContent side="right" sideOffset={12}>
              Cerrar sesión
            </TooltipContent>
          </Tooltip>
        </div>
      </nav>
    </TooltipProvider>
  )
}
