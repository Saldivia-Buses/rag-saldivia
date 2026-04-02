"use client"

import {
  Users, Trash2, Pencil, User, Crown, ShieldCheck, Eye,
  Lock, Star, Zap, Briefcase, BookOpen, Headphones,
  Wrench, Globe, Bell, Heart, Layers, Database,
} from "lucide-react"
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip"

const ICON_MAP: Record<string, React.ComponentType<{ size?: number; strokeWidth?: number }>> = {
  user: User, shield: Crown, "user-cog": ShieldCheck, eye: Eye,
  lock: Lock, star: Star, zap: Zap, briefcase: Briefcase,
  book: BookOpen, headphones: Headphones, wrench: Wrench,
  globe: Globe, bell: Bell, heart: Heart, layers: Layers, database: Database,
}

function RoleIcon({ icon, size = 20 }: { icon: string; size?: number }) {
  const IconComponent = ICON_MAP[icon] ?? User
  return <IconComponent size={size} strokeWidth={1.5} />
}

export type RoleWithCount = {
  id: number
  name: string
  description: string
  level: number
  color: string
  icon: string
  isSystem: boolean
  userCount: number
}

export function RoleCard({
  role,
  isSelected,
  onSelect,
  onEdit,
  onDelete,
}: {
  role: RoleWithCount
  isSelected: boolean
  onSelect: () => void
  onEdit: () => void
  onDelete: () => void
}) {
  return (
    <button
      type="button"
      onClick={onSelect}
      className="relative text-left rounded-xl transition-all"
      style={{
        padding: 0,
        overflow: "hidden",
        border: isSelected ? `2px solid ${role.color}` : "1px solid var(--border)",
        boxShadow: isSelected ? `0 0 0 1px color-mix(in srgb, ${role.color} 20%, transparent)` : "none",
      }}
    >
      {/* Colored top bar */}
      <div style={{ height: "4px", background: role.color }} />

      <div style={{ padding: "18px 18px 16px" }}>
        {/* Icon + name */}
        <div className="flex items-start justify-between" style={{ marginBottom: "8px" }}>
          <div className="flex items-center" style={{ gap: "12px" }}>
            <div
              className="flex items-center justify-center rounded-lg shrink-0"
              style={{ width: "40px", height: "40px", color: role.color, background: `color-mix(in srgb, ${role.color} 10%, transparent)` }}
            >
              <RoleIcon icon={role.icon} size={22} />
            </div>
            <div>
              <div className="font-semibold text-fg" style={{ fontSize: "14px" }}>{role.name}</div>
              <div className="text-xs text-fg-subtle">Nivel {role.level}</div>
            </div>
          </div>
          <div className="flex items-center shrink-0" style={{ gap: "4px" }}>
            {role.isSystem && (
              <span
                className="text-[10px] font-medium text-fg-subtle rounded-full"
                style={{ padding: "2px 8px", background: "var(--surface-2)" }}
              >
                sistema
              </span>
            )}
            <Tooltip>
              <TooltipTrigger asChild>
                <div
                  role="button"
                  onClick={(e) => { e.stopPropagation(); onEdit() }}
                  className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-accent hover:bg-accent/10 transition-colors"
                  style={{ width: "28px", height: "28px" }}
                >
                  <Pencil size={13} />
                </div>
              </TooltipTrigger>
              <TooltipContent side="bottom" sideOffset={4}>Editar rol</TooltipContent>
            </Tooltip>
            {!role.isSystem && (
              <Tooltip>
                <TooltipTrigger asChild>
                  <div
                    role="button"
                    onClick={(e) => { e.stopPropagation(); onDelete() }}
                    className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-destructive hover:bg-destructive/10 transition-colors"
                    style={{ width: "28px", height: "28px" }}
                  >
                    <Trash2 size={13} />
                  </div>
                </TooltipTrigger>
                <TooltipContent side="bottom" sideOffset={4}>Eliminar rol</TooltipContent>
              </Tooltip>
            )}
          </div>
        </div>

        {/* Description */}
        <p className="text-xs text-fg-muted" style={{ marginBottom: "14px", lineHeight: "1.5" }}>
          {role.description || "Sin descripción"}
        </p>

        {/* User count */}
        <div
          className="flex items-center text-xs text-fg-subtle rounded-lg"
          style={{ gap: "5px", padding: "6px 10px", background: "var(--surface-2)" }}
        >
          <Users size={12} />
          {role.userCount} usuario{role.userCount !== 1 ? "s" : ""}
        </div>
      </div>

      {/* Selected indicator at bottom (custom roles only) */}
      {isSelected && !role.isSystem && (
        <div className="absolute bottom-0 left-0 right-0" style={{ height: "2px", background: role.color }} />
      )}
    </button>
  )
}
