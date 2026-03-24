"use client"

import { useState, useTransition } from "react"
import type { DbUser } from "@rag-saldivia/db"
import {
  actionUpdateProfile,
  actionUpdatePassword,
  actionUpdatePreferences,
} from "@/app/actions/settings"

type Tab = "perfil" | "contrasena" | "preferencias"

export function SettingsClient({ user }: { user: DbUser }) {
  const [tab, setTab] = useState<Tab>("perfil")
  const [isPending, startTransition] = useTransition()

  // Perfil
  const [name, setName] = useState(user.name)
  const [profileMsg, setProfileMsg] = useState<{ type: "ok" | "error"; text: string } | null>(null)

  // Contraseña
  const [currentPwd, setCurrentPwd] = useState("")
  const [newPwd, setNewPwd] = useState("")
  const [pwdMsg, setPwdMsg] = useState<{ type: "ok" | "error"; text: string } | null>(null)

  const preferences = user.preferences as Record<string, unknown>

  const TABS: Array<{ key: Tab; label: string }> = [
    { key: "perfil", label: "Perfil" },
    { key: "contrasena", label: "Contraseña" },
    { key: "preferencias", label: "Preferencias" },
  ]

  async function handleProfileSave(e: React.FormEvent) {
    e.preventDefault()
    setProfileMsg(null)
    startTransition(async () => {
      try {
        await actionUpdateProfile({ name })
        setProfileMsg({ type: "ok", text: "Perfil actualizado" })
      } catch (err) {
        setProfileMsg({ type: "error", text: String(err) })
      }
    })
  }

  async function handlePasswordSave(e: React.FormEvent) {
    e.preventDefault()
    setPwdMsg(null)
    if (newPwd.length < 8) {
      setPwdMsg({ type: "error", text: "La contraseña debe tener al menos 8 caracteres" })
      return
    }
    startTransition(async () => {
      try {
        await actionUpdatePassword(currentPwd, newPwd)
        setPwdMsg({ type: "ok", text: "Contraseña actualizada" })
        setCurrentPwd(""); setNewPwd("")
      } catch (err) {
        setPwdMsg({ type: "error", text: String(err) })
      }
    })
  }

  return (
    <div className="space-y-4">
      {/* Tabs */}
      <div className="flex gap-1 border-b" style={{ borderColor: "var(--border)" }}>
        {TABS.map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
            style={{
              borderColor: tab === t.key ? "var(--primary)" : "transparent",
              color: tab === t.key ? "var(--foreground)" : "var(--muted-foreground)",
            }}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Perfil */}
      {tab === "perfil" && (
        <form onSubmit={handleProfileSave} className="space-y-4">
          <div className="space-y-1">
            <label className="text-sm font-medium">Nombre</label>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full px-3 py-2 rounded-md border text-sm"
              style={{ borderColor: "var(--border)" }}
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">Email</label>
            <input
              value={user.email}
              disabled
              className="w-full px-3 py-2 rounded-md border text-sm opacity-60"
              style={{ borderColor: "var(--border)" }}
            />
            <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
              El email no se puede cambiar. Contactá al administrador.
            </p>
          </div>
          {profileMsg && (
            <p className="text-sm" style={{ color: profileMsg.type === "ok" ? "#16a34a" : "var(--destructive)" }}>
              {profileMsg.text}
            </p>
          )}
          <button
            type="submit"
            disabled={isPending}
            className="px-4 py-2 rounded-md text-sm font-medium disabled:opacity-50"
            style={{ background: "var(--primary)", color: "var(--primary-foreground)" }}
          >
            {isPending ? "Guardando..." : "Guardar cambios"}
          </button>
        </form>
      )}

      {/* Contraseña */}
      {tab === "contrasena" && (
        <form onSubmit={handlePasswordSave} className="space-y-4">
          <div className="space-y-1">
            <label className="text-sm font-medium">Contraseña actual</label>
            <input
              type="password"
              value={currentPwd}
              onChange={(e) => setCurrentPwd(e.target.value)}
              required
              className="w-full px-3 py-2 rounded-md border text-sm"
              style={{ borderColor: "var(--border)" }}
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">Nueva contraseña</label>
            <input
              type="password"
              value={newPwd}
              onChange={(e) => setNewPwd(e.target.value)}
              required
              minLength={8}
              className="w-full px-3 py-2 rounded-md border text-sm"
              style={{ borderColor: "var(--border)" }}
            />
          </div>
          {pwdMsg && (
            <p className="text-sm" style={{ color: pwdMsg.type === "ok" ? "#16a34a" : "var(--destructive)" }}>
              {pwdMsg.text}
            </p>
          )}
          <button
            type="submit"
            disabled={isPending || !currentPwd || !newPwd}
            className="px-4 py-2 rounded-md text-sm font-medium disabled:opacity-50"
            style={{ background: "var(--primary)", color: "var(--primary-foreground)" }}
          >
            {isPending ? "Actualizando..." : "Actualizar contraseña"}
          </button>
        </form>
      )}

      {/* Preferencias */}
      {tab === "preferencias" && (
        <div className="space-y-4">
          <PreferenceToggle
            label="Tema"
            description="Preferencia de tema visual"
            value={String(preferences["theme"] ?? "system")}
            options={[
              { value: "system", label: "Sistema" },
              { value: "light", label: "Claro" },
              { value: "dark", label: "Oscuro" },
            ]}
            onChange={(v) => actionUpdatePreferences({ theme: v })}
          />
          <PreferenceToggle
            label="Crossdoc por defecto"
            description="Usar el modo crossdoc al iniciar sesiones"
            value={String(preferences["crossdocEnabled"] ?? false)}
            options={[
              { value: "false", label: "Desactivado" },
              { value: "true", label: "Activado" },
            ]}
            onChange={(v) => actionUpdatePreferences({ crossdocEnabled: v === "true" })}
          />
        </div>
      )}
    </div>
  )
}

function PreferenceToggle({
  label,
  description,
  value,
  options,
  onChange,
}: {
  label: string
  description: string
  value: string
  options: Array<{ value: string; label: string }>
  onChange: (value: string) => void
}) {
  return (
    <div className="flex items-center justify-between py-3 border-b" style={{ borderColor: "var(--border)" }}>
      <div>
        <p className="text-sm font-medium">{label}</p>
        <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>{description}</p>
      </div>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="px-3 py-1.5 rounded-md border text-sm"
        style={{ borderColor: "var(--border)", background: "var(--background)" }}
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>{opt.label}</option>
        ))}
      </select>
    </div>
  )
}
