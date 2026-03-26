"use client"

import { useState, useTransition } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { ThemeToggle } from "@/components/ui/theme-toggle"

export default function LoginPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const from = searchParams.get("from") ?? "/chat"

  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)

    startTransition(async () => {
      try {
        const res = await fetch("/api/auth/login", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email, password }),
        })

        const data = await res.json()

        if (!res.ok || !data.ok) {
          setError(data.error ?? "Error al iniciar sesión")
          return
        }

        router.push(from)
        router.refresh()
      } catch {
        setError("No se pudo conectar con el servidor")
      }
    })
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-bg">
        <div className="w-full max-w-sm">

          {/* Card */}
          <div className="rounded-2xl border border-border bg-surface shadow-sm px-8 py-10 space-y-8">

            {/* Header */}
            <div className="text-center space-y-1.5">
              <div className="inline-flex h-10 w-10 items-center justify-center rounded-xl bg-accent mb-3">
                <span className="text-base font-bold text-accent-fg select-none">R</span>
              </div>
              <h1 className="text-xl font-semibold text-fg tracking-tight">
                RAG Saldivia
              </h1>
              <p className="text-sm text-fg-muted">
                Iniciá sesión para continuar
              </p>
            </div>

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-1.5">
                <label htmlFor="email" className="text-sm font-medium text-fg block">
                  Email
                </label>
                <Input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="usuario@empresa.com"
                  required
                  autoComplete="email"
                  autoFocus
                />
              </div>

              <div className="space-y-1.5">
                <label htmlFor="password" className="text-sm font-medium text-fg block">
                  Contraseña
                </label>
                <Input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  required
                  autoComplete="current-password"
                />
              </div>

              {error && (
                <div className="px-3 py-2.5 rounded-lg bg-destructive-subtle border border-destructive/20 text-sm text-destructive">
                  {error}
                </div>
              )}

              <Button
                type="submit"
                disabled={isPending || !email || !password}
                className="w-full"
                size="default"
              >
                {isPending ? "Iniciando sesión..." : "Iniciar sesión"}
              </Button>
            </form>

            {/* SSO */}
            {(process.env["NEXT_PUBLIC_GOOGLE_CLIENT_ID"] || process.env["NEXT_PUBLIC_AZURE_AD_CLIENT_ID"]) && (
              <>
                <div className="flex items-center gap-3">
                  <div className="flex-1 h-px bg-border" />
                  <span className="text-xs text-fg-subtle">o</span>
                  <div className="flex-1 h-px bg-border" />
                </div>
                <SSOButtons />
              </>
            )}
          </div>

          {/* Footer */}
          <div className="flex items-center justify-between mt-6">
            <p className="text-xs text-fg-subtle">
              ¿Problemas para acceder? Contactá al administrador.
            </p>
            <ThemeToggle />
          </div>
        </div>
    </div>
  )
}

function SSOButtons() {
  const { SSOButton } = require("@/components/auth/SSOButton") as { SSOButton: (p: { provider: "google" | "azure-ad" }) => React.ReactNode }
  return (
    <div className="space-y-2">
      <SSOButton provider="google" />
      <SSOButton provider="azure-ad" />
    </div>
  )
}
