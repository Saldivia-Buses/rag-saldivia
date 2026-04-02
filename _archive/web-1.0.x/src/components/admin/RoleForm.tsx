"use client"

import {
  User, Crown, ShieldCheck, Eye, Pencil, Lock, Star, Zap, Briefcase,
  BookOpen, Headphones, Wrench, Globe, Bell, Heart, Layers, Database,
} from "lucide-react"
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip"

export const ICON_OPTIONS = [
  { value: "user", label: "Usuario", Icon: User },
  { value: "shield", label: "Corona", Icon: Crown },
  { value: "user-cog", label: "Escudo", Icon: ShieldCheck },
  { value: "eye", label: "Ojo", Icon: Eye },
  { value: "pencil", label: "Lápiz", Icon: Pencil },
  { value: "lock", label: "Candado", Icon: Lock },
  { value: "star", label: "Estrella", Icon: Star },
  { value: "zap", label: "Rayo", Icon: Zap },
  { value: "briefcase", label: "Maletín", Icon: Briefcase },
  { value: "book", label: "Libro", Icon: BookOpen },
  { value: "headphones", label: "Soporte", Icon: Headphones },
  { value: "wrench", label: "Herramienta", Icon: Wrench },
  { value: "globe", label: "Globo", Icon: Globe },
  { value: "bell", label: "Campana", Icon: Bell },
  { value: "heart", label: "Corazón", Icon: Heart },
  { value: "layers", label: "Capas", Icon: Layers },
  { value: "database", label: "Base de datos", Icon: Database },
] as const

export const COLOR_SWATCHES = [
  "#2563eb", "#7c3aed", "#db2777", "#dc2626", "#ea580c", "#d97706",
  "#16a34a", "#0d9488", "#0284c7", "#6e6c69", "#1e1e1e", "#8b5cf6",
] as const

export type RoleFormData = {
  name: string
  description: string
  level: string
  color: string
  icon: string
}

export const EMPTY_FORM: RoleFormData = { name: "", description: "", level: "20", color: "#6e6c69", icon: "user" }

function RoleIcon({ icon, size = 16 }: { icon: string; size?: number }) {
  const found = ICON_OPTIONS.find((o) => o.value === icon)
  const IconComponent = found?.Icon ?? User
  return <IconComponent size={size} strokeWidth={1.5} />
}

export function RoleForm({
  form,
  onChange,
  onSubmit,
  onCancel,
  isEditing,
  isPending,
}: {
  form: RoleFormData
  onChange: (updater: (prev: RoleFormData) => RoleFormData) => void
  onSubmit: () => void
  onCancel: () => void
  isEditing: boolean
  isPending: boolean
}) {
  return (
    <div
      className="rounded-xl bg-surface"
      style={{ padding: "20px", marginBottom: "20px", border: `1px solid ${isEditing ? form.color : "var(--border)"}` }}
    >
      <h3 className="text-sm font-semibold text-fg" style={{ marginBottom: "16px" }}>
        {isEditing ? `Editar rol: ${form.name}` : "Crear rol personalizado"}
      </h3>
      <div className="grid grid-cols-2" style={{ gap: "12px" }}>
        <input
          value={form.name}
          onChange={(e) => onChange((f) => ({ ...f, name: e.target.value }))}
          placeholder="Nombre del rol"
          className="rounded-lg bg-bg text-fg text-sm outline-none focus:ring-1 focus:ring-accent transition-shadow"
          style={{ padding: "9px 12px", border: "1px solid var(--border)" }}
        />
        <input
          value={form.description}
          onChange={(e) => onChange((f) => ({ ...f, description: e.target.value }))}
          placeholder="Descripción (opcional)"
          className="rounded-lg bg-bg text-fg text-sm outline-none focus:ring-1 focus:ring-accent transition-shadow"
          style={{ padding: "9px 12px", border: "1px solid var(--border)" }}
        />
        <input
          value={form.level}
          onChange={(e) => onChange((f) => ({ ...f, level: e.target.value }))}
          placeholder="Nivel de prioridad (1-99)"
          type="number" min={1} max={99}
          className="rounded-lg bg-bg text-fg text-sm outline-none focus:ring-1 focus:ring-accent transition-shadow"
          style={{ padding: "9px 12px", border: "1px solid var(--border)" }}
        />
        <div className="flex items-center flex-wrap" style={{ gap: "8px" }}>
          <label className="text-sm text-fg-muted shrink-0">Color:</label>
          <div className="flex items-center" style={{ gap: "4px" }}>
            {COLOR_SWATCHES.map((c) => (
              <button
                key={c} type="button"
                onClick={() => onChange((f) => ({ ...f, color: c }))}
                className="rounded-full shrink-0"
                style={{
                  width: "20px", height: "20px", background: c,
                  outline: form.color === c ? "2px solid var(--fg)" : "none", outlineOffset: "1px",
                }}
              />
            ))}
            <input
              type="color" value={form.color}
              onChange={(e) => onChange((f) => ({ ...f, color: e.target.value }))}
              className="rounded cursor-pointer shrink-0"
              style={{ width: "20px", height: "20px", border: "1px solid var(--border)", padding: 0 }}
              title="Color personalizado"
            />
          </div>
          <div
            className="flex items-center rounded-lg shrink-0"
            style={{
              gap: "8px", padding: "5px 12px 5px 8px", color: form.color,
              background: `color-mix(in srgb, ${form.color} 10%, transparent)`,
              border: `1px solid color-mix(in srgb, ${form.color} 20%, transparent)`,
              borderRadius: "8px",
            }}
          >
            <RoleIcon icon={form.icon} />
            <span className="text-xs font-medium">{form.name || "Preview"}</span>
          </div>
        </div>
      </div>

      {/* Icon picker */}
      <div style={{ marginTop: "14px" }}>
        <label className="text-xs font-medium text-fg-muted" style={{ display: "block", marginBottom: "8px" }}>
          Ícono
        </label>
        <div className="flex flex-wrap" style={{ gap: "6px" }}>
          {ICON_OPTIONS.map((opt) => {
            const isSelected = form.icon === opt.value
            return (
              <Tooltip key={opt.value}>
                <TooltipTrigger asChild>
                  <button
                    type="button"
                    onClick={() => onChange((f) => ({ ...f, icon: opt.value }))}
                    className="flex items-center justify-center rounded-lg transition-all"
                    style={{
                      width: "36px", height: "36px",
                      color: isSelected ? form.color : "var(--fg-subtle)",
                      background: isSelected ? `color-mix(in srgb, ${form.color} 12%, transparent)` : "transparent",
                      border: isSelected ? `2px solid ${form.color}` : "1px solid var(--border)",
                    }}
                  >
                    <opt.Icon size={16} strokeWidth={1.5} />
                  </button>
                </TooltipTrigger>
                <TooltipContent side="bottom" sideOffset={4}>{opt.label}</TooltipContent>
              </Tooltip>
            )
          })}
        </div>
      </div>

      <div className="flex justify-end" style={{ gap: "8px", marginTop: "16px" }}>
        <button
          onClick={onCancel}
          className="text-sm text-fg-muted hover:text-fg rounded-lg transition-colors"
          style={{ padding: "8px 16px", border: "1px solid var(--border)" }}
        >
          Cancelar
        </button>
        <button
          onClick={onSubmit}
          disabled={isPending || !form.name.trim()}
          className="text-sm font-medium rounded-lg bg-accent text-accent-fg disabled:opacity-30 transition-opacity"
          style={{ padding: "8px 16px" }}
        >
          {isPending ? "Guardando..." : isEditing ? "Guardar" : "Crear"}
        </button>
      </div>
    </div>
  )
}
