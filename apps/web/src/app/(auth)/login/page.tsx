"use client"

import { useState, useTransition } from "react"
import { useRouter, useSearchParams } from "next/navigation"

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
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-sm space-y-8">
        {/* Header */}
        <div className="text-center space-y-2">
          <h1 className="text-2xl font-bold tracking-tight">RAG Saldivia</h1>
          <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
            Iniciá sesión para continuar
          </p>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1">
            <label htmlFor="email" className="text-sm font-medium block">
              Email
            </label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="usuario@empresa.com"
              required
              autoComplete="email"
              className="w-full px-3 py-2 rounded-md border text-sm outline-none transition-all focus:ring-2"
              style={{
                borderColor: "var(--border)",
                background: "var(--background)",
                "--tw-ring-color": "var(--ring)",
              } as React.CSSProperties}
            />
          </div>

          <div className="space-y-1">
            <label htmlFor="password" className="text-sm font-medium block">
              Contraseña
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              required
              autoComplete="current-password"
              className="w-full px-3 py-2 rounded-md border text-sm outline-none transition-all focus:ring-2"
              style={{
                borderColor: "var(--border)",
                background: "var(--background)",
                "--tw-ring-color": "var(--ring)",
              } as React.CSSProperties}
            />
          </div>

          {error && (
            <div
              className="px-3 py-2 rounded-md text-sm"
              style={{ background: "#fef2f2", color: "var(--destructive)", border: "1px solid #fecaca" }}
            >
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={isPending || !email || !password}
            className="w-full py-2 px-4 rounded-md text-sm font-medium transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
            style={{
              background: "var(--primary)",
              color: "var(--primary-foreground)",
            }}
          >
            {isPending ? "Iniciando sesión..." : "Iniciar sesión"}
          </button>
        </form>

        {/* Footer */}
        <p className="text-center text-xs" style={{ color: "var(--muted-foreground)" }}>
          ¿Problemas para acceder? Contactá al administrador del sistema.
        </p>
      </div>
    </div>
  )
}
