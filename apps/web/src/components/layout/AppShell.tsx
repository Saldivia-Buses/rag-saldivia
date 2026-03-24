"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import type { CurrentUser } from "@/lib/auth/current-user"
import { MessageSquare, FolderOpen, Upload, Settings, Users, ShieldCheck, Database, FileText, BarChart3, LogOut } from "lucide-react"

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
  { href: "/audit", label: "Audit", icon: <FileText size={18} />, roles: ["admin", "area_manager"] },
  { href: "/settings", label: "Configuración", icon: <Settings size={18} /> },
]

const ADMIN_ITEMS: NavItem[] = [
  { href: "/admin/users", label: "Usuarios", icon: <Users size={18} /> },
  { href: "/admin/areas", label: "Áreas", icon: <Database size={18} /> },
  { href: "/admin/permissions", label: "Permisos", icon: <ShieldCheck size={18} /> },
  { href: "/admin/rag-config", label: "Config RAG", icon: <BarChart3 size={18} /> },
  { href: "/admin/system", label: "Sistema", icon: <BarChart3 size={18} /> },
]

function NavLink({ item, pathname }: { item: NavItem; pathname: string }) {
  const active = pathname.startsWith(item.href)
  return (
    <Link
      href={item.href}
      className={`flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors ${
        active
          ? "font-medium"
          : "opacity-70 hover:opacity-100"
      }`}
      style={{
        background: active ? "var(--accent)" : "transparent",
        color: active ? "var(--foreground)" : "var(--muted-foreground)",
      }}
    >
      {item.icon}
      {item.label}
    </Link>
  )
}

async function handleLogout() {
  await fetch("/api/auth/logout", { method: "DELETE" })
  window.location.href = "/login"
}

export function AppShell({
  user,
  children,
}: {
  user: CurrentUser
  children: React.ReactNode
}) {
  const pathname = usePathname()

  const visibleNav = NAV_ITEMS.filter(
    (item) => !item.roles || item.roles.includes(user.role)
  )

  return (
    <div className="flex h-screen overflow-hidden">
      {/* Sidebar */}
      <aside
        className="w-56 flex-shrink-0 flex flex-col border-r"
        style={{ borderColor: "var(--border)", background: "var(--background)" }}
      >
        {/* Logo */}
        <div className="px-4 py-4 border-b" style={{ borderColor: "var(--border)" }}>
          <span className="font-semibold text-sm">RAG Saldivia</span>
        </div>

        {/* Nav */}
        <nav className="flex-1 overflow-y-auto p-2 space-y-0.5">
          {visibleNav.map((item) => (
            <NavLink key={item.href} item={item} pathname={pathname} />
          ))}

          {user.role === "admin" && (
            <>
              <div
                className="mx-3 my-2 text-xs font-medium uppercase tracking-wider"
                style={{ color: "var(--muted-foreground)" }}
              >
                Admin
              </div>
              {ADMIN_ITEMS.map((item) => (
                <NavLink key={item.href} item={item} pathname={pathname} />
              ))}
            </>
          )}
        </nav>

        {/* User */}
        <div
          className="p-3 border-t space-y-1"
          style={{ borderColor: "var(--border)" }}
        >
          <div className="px-3 py-1">
            <p className="text-sm font-medium truncate">{user.name}</p>
            <p className="text-xs truncate" style={{ color: "var(--muted-foreground)" }}>
              {user.email}
            </p>
          </div>
          <button
            onClick={handleLogout}
            className="flex items-center gap-2 w-full px-3 py-2 rounded-md text-sm transition-colors hover:opacity-100 opacity-60"
          >
            <LogOut size={16} />
            Cerrar sesión
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-y-auto">{children}</main>
    </div>
  )
}
