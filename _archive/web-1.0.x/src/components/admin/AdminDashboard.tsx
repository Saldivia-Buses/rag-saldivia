/**
 * Admin dashboard — system overview with live presence and metrics.
 *
 * Shows: stat cards, online users with presence dots, role distribution,
 * recent activity metrics, and quick actions.
 *
 * Rendered by: app/(app)/admin/page.tsx (server component passes all data)
 */

"use client"

import { useState, useEffect } from "react"
import {
  Users, MessageSquare, Shield, FolderOpen,
  Activity, UserPlus, ShieldPlus, ArrowRight,
  Circle,
} from "lucide-react"
import { StatCard } from "@/components/ui/stat-card"
import { RoleBadge } from "./RoleBadge"
import Link from "next/link"

type RoleInfo = {
  id: number
  name: string
  color: string
  level: number
  userCount: number
}

type UserPresence = {
  id: number
  name: string
  email: string
  lastSeen: number | null
  active: boolean
}

type DashboardStats = {
  users: { total: number; active: number; inactive: number }
  sessions: number
  messages: number
  roles: RoleInfo[]
  usersPresence: UserPresence[]
}

const ONLINE_THRESHOLD = 2 * 60 * 1000 // 2 minutes

function isOnline(lastSeen: number | null) {
  if (!lastSeen) return false
  return Date.now() - lastSeen < ONLINE_THRESHOLD
}

function timeAgo(ts: number | null): string {
  if (!ts) return "nunca"
  const diff = Date.now() - ts
  if (diff < 60_000) return "ahora"
  if (diff < 3600_000) return `hace ${Math.floor(diff / 60_000)}m`
  if (diff < 86400_000) return `hace ${Math.floor(diff / 3600_000)}h`
  return `hace ${Math.floor(diff / 86400_000)}d`
}

export function AdminDashboard({ stats }: { stats: DashboardStats }) {
  // Force re-render every 30s so presence dots update
  const [, setTick] = useState(0)
  useEffect(() => {
    const id = setInterval(() => setTick((t) => t + 1), 30_000)
    return () => clearInterval(id)
  }, [])

  const onlineCount = stats.usersPresence.filter((u) => isOnline(u.lastSeen)).length

  return (
    <div>
      {/* Stats grid */}
      <div className="grid grid-cols-2 lg:grid-cols-4" style={{ gap: "12px", marginBottom: "20px" }}>
        <StatCard label="Usuarios" value={stats.users.total} icon={Users} />
        <StatCard label="Online ahora" value={onlineCount} icon={Activity} />
        <StatCard label="Sesiones" value={stats.sessions} icon={MessageSquare} />
        <StatCard label="Mensajes" value={stats.messages} icon={MessageSquare} />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2" style={{ gap: "16px", marginBottom: "20px" }}>
        {/* Online users */}
        <div
          className="rounded-xl bg-surface"
          style={{ padding: "20px", border: "1px solid var(--border)" }}
        >
          <div className="flex items-center justify-between" style={{ marginBottom: "16px" }}>
            <h2 className="text-sm font-semibold text-fg flex items-center" style={{ gap: "8px" }}>
              <Activity size={14} className="text-accent" />
              Usuarios
            </h2>
            <Link href="/admin/users" className="text-xs text-accent hover:underline flex items-center" style={{ gap: "4px" }}>
              Ver todos <ArrowRight size={12} />
            </Link>
          </div>

          {(() => {
            const MAX_VISIBLE = 9
            const sorted = [...stats.usersPresence].sort((a, b) => {
              const aOn = isOnline(a.lastSeen)
              const bOn = isOnline(b.lastSeen)
              if (aOn !== bOn) return aOn ? -1 : 1
              return (b.lastSeen ?? 0) - (a.lastSeen ?? 0)
            })
            const visible = sorted.slice(0, MAX_VISIBLE)
            const remaining = sorted.length - MAX_VISIBLE

            return (
              <div className="flex flex-col" style={{ gap: "2px" }}>
                {visible.map((u) => {
                  const online = isOnline(u.lastSeen)
                  return (
                    <div
                      key={u.id}
                      className="flex items-center justify-between rounded-lg transition-colors hover:bg-surface-2"
                      style={{ padding: "7px 10px" }}
                    >
                      <div className="flex items-center" style={{ gap: "10px" }}>
                        <div className="relative shrink-0">
                          <Circle size={8} fill={online ? "var(--success)" : "var(--border)"} stroke="none" />
                          {online && (
                            <Circle size={8} fill="var(--success)" stroke="none" className="absolute inset-0 animate-ping" style={{ opacity: 0.4, animationDuration: "2s" }} />
                          )}
                        </div>
                        <div className="truncate">
                          <span className="text-sm text-fg">{u.name}</span>
                        </div>
                      </div>
                      <span className={`text-xs shrink-0 ${online ? "text-success" : "text-fg-subtle"}`}>
                        {online ? "online" : timeAgo(u.lastSeen)}
                      </span>
                    </div>
                  )
                })}
                {remaining > 0 && (
                  <Link
                    href="/admin/users"
                    className="text-xs text-fg-subtle hover:text-accent text-center transition-colors"
                    style={{ padding: "6px 10px" }}
                  >
                    +{remaining} usuario{remaining !== 1 ? "s" : ""} más
                  </Link>
                )}
              </div>
            )
          })()}
        </div>

        {/* Role distribution */}
        <div
          className="rounded-xl bg-surface"
          style={{ padding: "20px", border: "1px solid var(--border)" }}
        >
          <div className="flex items-center justify-between" style={{ marginBottom: "16px" }}>
            <h2 className="text-sm font-semibold text-fg flex items-center" style={{ gap: "8px" }}>
              <Shield size={14} className="text-accent" />
              Distribución de roles
            </h2>
            <Link href="/admin/roles" className="text-xs text-accent hover:underline flex items-center" style={{ gap: "4px" }}>
              Gestionar <ArrowRight size={12} />
            </Link>
          </div>

          <div className="flex flex-col" style={{ gap: "12px" }}>
            {stats.roles
              .sort((a, b) => b.level - a.level)
              .map((role) => {
                const pct = stats.users.total > 0 ? (role.userCount / stats.users.total) * 100 : 0
                return (
                  <div key={role.id}>
                    <div className="flex items-center justify-between" style={{ marginBottom: "6px" }}>
                      <RoleBadge name={role.name} color={role.color} />
                      <span className="text-xs text-fg-muted tabular-nums">
                        {role.userCount} usuario{role.userCount !== 1 ? "s" : ""}
                      </span>
                    </div>
                    {/* Bar */}
                    <div
                      className="rounded-full overflow-hidden"
                      style={{ height: "6px", background: "var(--surface-2)" }}
                    >
                      <div
                        className="h-full rounded-full transition-all duration-500"
                        style={{
                          width: `${Math.max(pct, 2)}%`,
                          background: role.color,
                          opacity: 0.7,
                        }}
                      />
                    </div>
                  </div>
                )
              })}
          </div>
        </div>
      </div>

      {/* Quick actions */}
      <div className="grid grid-cols-1 md:grid-cols-3" style={{ gap: "12px" }}>
        <Link
          href="/admin/users"
          className="flex items-center rounded-xl bg-surface hover:shadow-sm transition-all"
          style={{ padding: "16px", gap: "12px", border: "1px solid var(--border)" }}
        >
          <div
            className="flex items-center justify-center rounded-lg shrink-0"
            style={{ width: "40px", height: "40px", background: "color-mix(in srgb, var(--accent) 10%, transparent)" }}
          >
            <UserPlus size={18} className="text-accent" />
          </div>
          <div>
            <div className="text-sm font-medium text-fg">Gestionar usuarios</div>
            <div className="text-xs text-fg-subtle">Crear, editar y asignar roles</div>
          </div>
        </Link>

        <Link
          href="/admin/roles"
          className="flex items-center rounded-xl bg-surface hover:shadow-sm transition-all"
          style={{ padding: "16px", gap: "12px", border: "1px solid var(--border)" }}
        >
          <div
            className="flex items-center justify-center rounded-lg shrink-0"
            style={{ width: "40px", height: "40px", background: "color-mix(in srgb, var(--accent) 10%, transparent)" }}
          >
            <ShieldPlus size={18} className="text-accent" />
          </div>
          <div>
            <div className="text-sm font-medium text-fg">Gestionar roles</div>
            <div className="text-xs text-fg-subtle">Roles, permisos y matrix</div>
          </div>
        </Link>

        <Link
          href="/collections"
          className="flex items-center rounded-xl bg-surface hover:shadow-sm transition-all"
          style={{ padding: "16px", gap: "12px", border: "1px solid var(--border)" }}
        >
          <div
            className="flex items-center justify-center rounded-lg shrink-0"
            style={{ width: "40px", height: "40px", background: "color-mix(in srgb, var(--accent) 10%, transparent)" }}
          >
            <FolderOpen size={18} className="text-accent" />
          </div>
          <div>
            <div className="text-sm font-medium text-fg">Colecciones</div>
            <div className="text-xs text-fg-subtle">Ver y gestionar documentos</div>
          </div>
        </Link>
      </div>
    </div>
  )
}
