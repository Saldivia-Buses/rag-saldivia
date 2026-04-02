"use client"

import { useState, useTransition } from "react"
import { KeyRound, Wand2, X, Copy } from "lucide-react"
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip"
import { actionResetPassword } from "@/app/actions/admin"

function generatePassword() {
  const chars = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
  let pw = ""
  for (let i = 0; i < 12; i++) pw += chars[Math.floor(Math.random() * chars.length)]
  return pw
}

export function PasswordResetCell({
  userId,
  userName,
  onSuccess,
}: {
  userId: number
  userName: string
  onSuccess: (msg: string) => void
}) {
  const [mode, setMode] = useState<"idle" | "editing" | "generated">("idle")
  const [newPassword, setNewPassword] = useState("")
  const [generatedPw, setGeneratedPw] = useState("")
  const [resetError, setResetError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  function handleReset() {
    if (!newPassword) { setResetError("Ingresá una contraseña"); return }
    if (newPassword.length < 6) { setResetError("Mínimo 6 caracteres"); return }
    setResetError(null)
    startTransition(async () => {
      try {
        await actionResetPassword({ userId, newPassword })
        setGeneratedPw(newPassword)
        setMode("generated")
        setNewPassword("")
        onSuccess(`Contraseña de ${userName} reseteada`)
      } catch {
        setResetError("Error al resetear contraseña")
      }
    })
  }

  function handleAutoReset() {
    const pw = generatePassword()
    setNewPassword(pw)
    setMode("editing")
  }

  if (mode === "editing") {
    return (
      <div className="flex items-center" style={{ gap: "4px" }}>
        <div>
          <div className="flex items-center" style={{ gap: "3px" }}>
            <input
              value={newPassword}
              onChange={(e) => { setNewPassword(e.target.value); setResetError(null) }}
              placeholder="Nueva contraseña" type="text"
              className={`text-xs rounded bg-bg text-fg outline-none focus:ring-1 focus:ring-accent transition-shadow font-mono ${resetError ? "ring-1 ring-destructive" : ""}`}
              style={{ padding: "4px 8px", width: "140px", border: "1px solid var(--border)" }}
              autoFocus
              onKeyDown={(e) => {
                if (e.key === "Enter") handleReset()
                if (e.key === "Escape") { setMode("idle"); setNewPassword(""); setResetError(null) }
              }}
            />
            <Tooltip>
              <TooltipTrigger asChild>
                <button type="button" onClick={() => setNewPassword(generatePassword())}
                  className="flex items-center justify-center rounded text-fg-subtle hover:text-accent transition-colors"
                  style={{ width: "24px", height: "24px" }}>
                  <Wand2 size={12} />
                </button>
              </TooltipTrigger>
              <TooltipContent side="bottom" sideOffset={4}>Generar</TooltipContent>
            </Tooltip>
          </div>
          {resetError && <div className="text-[10px] text-destructive" style={{ marginTop: "2px" }}>{resetError}</div>}
        </div>
        <button onClick={handleReset} disabled={isPending} className="text-xs text-accent hover:underline font-medium">OK</button>
        <button onClick={() => { setMode("idle"); setNewPassword(""); setResetError(null) }} className="text-xs text-fg-subtle hover:text-fg">
          <X size={12} />
        </button>
      </div>
    )
  }

  if (mode === "generated") {
    return (
      <div className="flex items-center" style={{ gap: "4px" }}>
        <code className="text-xs font-mono text-fg bg-surface-2 rounded select-all" style={{ padding: "3px 8px" }}>
          {generatedPw}
        </code>
        <Tooltip>
          <TooltipTrigger asChild>
            <button type="button" onClick={() => { navigator.clipboard.writeText(generatedPw); onSuccess("Contraseña copiada") }}
              className="flex items-center justify-center rounded text-fg-subtle hover:text-accent transition-colors"
              style={{ width: "24px", height: "24px" }}>
              <Copy size={12} />
            </button>
          </TooltipTrigger>
          <TooltipContent side="bottom" sideOffset={4}>Copiar</TooltipContent>
        </Tooltip>
        <button onClick={() => setMode("idle")} className="text-xs text-fg-subtle hover:text-fg"><X size={12} /></button>
      </div>
    )
  }

  // idle mode
  return (
    <div className="flex items-center" style={{ gap: "2px" }}>
      <Tooltip>
        <TooltipTrigger asChild>
          <button onClick={handleAutoReset}
            className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-accent hover:bg-accent/10 transition-colors"
            style={{ width: "32px", height: "32px" }}>
            <Wand2 size={14} />
          </button>
        </TooltipTrigger>
        <TooltipContent side="bottom" sideOffset={4}>Generar contraseña</TooltipContent>
      </Tooltip>
      <Tooltip>
        <TooltipTrigger asChild>
          <button onClick={() => { setMode("editing"); setResetError(null) }}
            className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-fg hover:bg-surface-2 transition-colors"
            style={{ width: "32px", height: "32px" }}>
            <KeyRound size={14} />
          </button>
        </TooltipTrigger>
        <TooltipContent side="bottom" sideOffset={4}>Elegir contraseña</TooltipContent>
      </Tooltip>
    </div>
  )
}
