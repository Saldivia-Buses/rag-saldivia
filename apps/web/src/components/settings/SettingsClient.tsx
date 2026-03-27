"use client"

import { useState, useTransition } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import type { DbUser } from "@rag-saldivia/db"
import {
  actionUpdateProfile,
  actionUpdatePassword,
  actionUpdatePreferences,
} from "@/app/actions/settings"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

type Tab = "perfil" | "contrasena" | "preferencias"

const ProfileSchema = z.object({
  name: z.string().min(2, "El nombre debe tener al menos 2 caracteres"),
})

const PasswordSchema = z.object({
  currentPassword: z.string().min(1, "La contraseña actual es requerida"),
  newPassword: z.string().min(8, "La nueva contraseña debe tener al menos 8 caracteres"),
})

type ProfileInput = z.infer<typeof ProfileSchema>
type PasswordInput = z.infer<typeof PasswordSchema>

export function SettingsClient({ user }: { user: DbUser }) {
  const [tab, setTab] = useState<Tab>("perfil")
  const [isPending, startTransition] = useTransition()
  const [profileMsg, setProfileMsg] = useState<{ type: "ok" | "error"; text: string } | null>(null)
  const [pwdMsg, setPwdMsg] = useState<{ type: "ok" | "error"; text: string } | null>(null)

  const profileForm = useForm<ProfileInput>({
    resolver: zodResolver(ProfileSchema),
    defaultValues: { name: user.name },
  })

  const passwordForm = useForm<PasswordInput>({
    resolver: zodResolver(PasswordSchema),
    defaultValues: { currentPassword: "", newPassword: "" },
  })

  const preferences = user.preferences as Record<string, unknown>

  const TABS: Array<{ key: Tab; label: string }> = [
    { key: "perfil", label: "Perfil" },
    { key: "contrasena", label: "Contraseña" },
    { key: "preferencias", label: "Preferencias" },
  ]

  function handleProfileSave(data: ProfileInput) {
    setProfileMsg(null)
    startTransition(async () => {
      try {
        await actionUpdateProfile(data)
        setProfileMsg({ type: "ok", text: "Perfil actualizado" })
      } catch (err) {
        setProfileMsg({ type: "error", text: String(err) })
      }
    })
  }

  function handlePasswordSave(data: PasswordInput) {
    setPwdMsg(null)
    startTransition(async () => {
      try {
        await actionUpdatePassword(data.currentPassword, data.newPassword)
        setPwdMsg({ type: "ok", text: "Contraseña actualizada" })
        passwordForm.reset()
      } catch (err) {
        setPwdMsg({ type: "error", text: String(err) })
      }
    })
  }

  return (
    <div className="p-6 max-w-xl space-y-6">
      <div>
        <h1 className="text-lg font-semibold text-fg">Configuración</h1>
        <p className="text-sm text-fg-muted mt-0.5">Gestioná tu perfil y preferencias</p>
      </div>

      {/* Tabs */}
      <div className="flex gap-0 border-b border-border">
        {TABS.map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px ${
              tab === t.key
                ? "border-accent text-fg"
                : "border-transparent text-fg-muted hover:text-fg"
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Perfil */}
      {tab === "perfil" && (
        <form onSubmit={profileForm.handleSubmit(handleProfileSave)} className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-fg">Nombre</label>
            <Input {...profileForm.register("name")} />
            {profileForm.formState.errors.name && (
              <p className="text-xs text-destructive">{profileForm.formState.errors.name.message}</p>
            )}
          </div>
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-fg">Email</label>
            <Input value={user.email} disabled />
            <p className="text-xs text-fg-subtle">El email no se puede cambiar. Contactá al administrador.</p>
          </div>
          {profileMsg && (
            <p className={`text-sm ${profileMsg.type === "ok" ? "text-success" : "text-destructive"}`}>
              {profileMsg.text}
            </p>
          )}
          <Button type="submit" disabled={isPending} size="sm">
            {isPending ? "Guardando..." : "Guardar cambios"}
          </Button>
        </form>
      )}

      {/* Contraseña */}
      {tab === "contrasena" && (
        <form onSubmit={passwordForm.handleSubmit(handlePasswordSave)} className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-fg">Contraseña actual</label>
            <Input type="password" {...passwordForm.register("currentPassword")} />
            {passwordForm.formState.errors.currentPassword && (
              <p className="text-xs text-destructive">{passwordForm.formState.errors.currentPassword.message}</p>
            )}
          </div>
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-fg">Nueva contraseña</label>
            <Input type="password" {...passwordForm.register("newPassword")} />
            {passwordForm.formState.errors.newPassword && (
              <p className="text-xs text-destructive">{passwordForm.formState.errors.newPassword.message}</p>
            )}
          </div>
          {pwdMsg && (
            <p className={`text-sm ${pwdMsg.type === "ok" ? "text-success" : "text-destructive"}`}>
              {pwdMsg.text}
            </p>
          )}
          <Button type="submit" disabled={isPending || !passwordForm.formState.isDirty} size="sm">
            {isPending ? "Actualizando..." : "Actualizar contraseña"}
          </Button>
        </form>
      )}

      {/* Preferencias */}
      {tab === "preferencias" && (
        <div className="space-y-1">
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
  label, description, value, options, onChange,
}: {
  label: string
  description: string
  value: string
  options: Array<{ value: string; label: string }>
  onChange: (value: string) => void
}) {
  return (
    <div className="flex items-center justify-between py-4 border-b border-border">
      <div>
        <p className="text-sm font-medium text-fg">{label}</p>
        <p className="text-xs text-fg-muted">{description}</p>
      </div>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-8 rounded-md border border-border bg-bg px-2 text-sm text-fg focus:outline-none focus:ring-1 focus:ring-ring"
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>{opt.label}</option>
        ))}
      </select>
    </div>
  )
}
