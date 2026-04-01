"use client"

import { Suspense, useState, useTransition, useEffect } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { ThemeToggle } from "@/components/ui/theme-toggle"

const LoginSchema = z.object({
  email: z.string().min(1, "El email es requerido"),
  password: z.string().min(1, "La contraseña es requerida"),
})

type LoginInput = z.infer<typeof LoginSchema>

export default function LoginPage() {
  return (
    <Suspense>
      <LoginContent />
    </Suspense>
  )
}

const SSO_ERRORS: Record<string, string> = {
  no_account: "No se encontró una cuenta. Contactá al administrador.",
  inactive: "Tu cuenta está desactivada.",
  provider_error: "Error al conectar con el proveedor SSO.",
  invalid_state: "Sesión de SSO inválida. Intentá de nuevo.",
  expired_state: "La sesión de SSO expiró. Intentá de nuevo.",
  already_linked: "Esta cuenta ya está vinculada a otro proveedor SSO.",
  missing_params: "Respuesta incompleta del proveedor SSO.",
}

const SSO_ICONS: Record<string, React.ReactNode> = {
  google: <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4"/><path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/><path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/><path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/></svg>,
  microsoft: <svg viewBox="0 0 24 24" width="18" height="18"><rect x="1" y="1" width="10" height="10" fill="#F25022"/><rect x="13" y="1" width="10" height="10" fill="#7FBA00"/><rect x="1" y="13" width="10" height="10" fill="#00A4EF"/><rect x="13" y="13" width="10" height="10" fill="#FFB900"/></svg>,
  github: <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0 1 12 6.844a9.59 9.59 0 0 1 2.504.337c1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.02 10.02 0 0 0 22 12.017C22 6.484 17.522 2 12 2z"/></svg>,
}

type SsoProviderPublic = { id: number; name: string; type: string }

function LoginContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const from = searchParams.get("from") ?? "/chat"
  const ssoError = searchParams.get("sso_error")
  const [serverError, setServerError] = useState<string | null>(
    ssoError ? (SSO_ERRORS[ssoError] ?? "Error de SSO desconocido.") : null
  )
  const [isPending, startTransition] = useTransition()
  const [ssoProviders, setSsoProviders] = useState<SsoProviderPublic[]>([])

  useEffect(() => {
    fetch("/api/auth/sso/providers")
      .then((r) => r.json())
      .then((data) => {
        if (data.ok && Array.isArray(data.data)) setSsoProviders(data.data)
      })
      .catch(() => {})
  }, [])

  const { register, handleSubmit, formState: { errors } } = useForm<LoginInput>({
    resolver: zodResolver(LoginSchema),
    defaultValues: { email: "", password: "" },
    mode: "onBlur",
  })

  function onSubmit(data: LoginInput) {
    setServerError(null)
    startTransition(async () => {
      try {
        const res = await fetch("/api/auth/login", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(data),
        })

        const result = await res.json()

        if (!res.ok || !result.ok) {
          setServerError(result.error ?? "Error al iniciar sesión")
          return
        }

        router.push(from)
        router.refresh()
      } catch {
        setServerError("No se pudo conectar con el servidor")
      }
    })
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-bg">
      <main className="w-full max-w-sm">
        <div className="w-full">

          {/* Card */}
          <div className="flex flex-col gap-10 rounded-2xl border border-border bg-surface shadow-sm" style={{ padding: "3rem 2.5rem" }}>

            {/* Header */}
            <div className="flex flex-col items-center gap-2 text-center">
              <div className="inline-flex h-12 w-12 items-center justify-center rounded-xl bg-accent mb-2">
                <span className="text-lg font-bold text-accent-fg select-none">S</span>
              </div>
              <h1 className="text-2xl font-semibold text-fg tracking-tight">
                Saldivia RAG
              </h1>
              <p className="text-sm text-fg-muted">
                Iniciá sesión para continuar
              </p>
            </div>

            {/* Form */}
            <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-5">
              <div className="flex flex-col gap-2">
                <label htmlFor="email" className="text-sm font-medium text-fg">
                  Email
                </label>
                <Input
                  id="email"
                  type="email"
                  placeholder="usuario@empresa.com"
                  autoComplete="email"
                  autoFocus
                  className="h-11 text-base px-4 rounded-[10px]"
                  {...register("email")}
                />
                {errors.email && (
                  <p className="text-xs text-destructive mt-1">{errors.email.message}</p>
                )}
              </div>

              <div className="flex flex-col gap-2">
                <label htmlFor="password" className="text-sm font-medium text-fg">
                  Contraseña
                </label>
                <Input
                  id="password"
                  type="password"
                  placeholder="••••••••"
                  autoComplete="current-password"
                  className="h-11 text-base px-4 rounded-[10px]"
                  {...register("password")}
                />
                {errors.password && (
                  <p className="text-xs text-destructive mt-1">{errors.password.message}</p>
                )}
              </div>

              {serverError && (
                <div className="px-4 py-3 rounded-xl bg-destructive-subtle border border-destructive/20 text-sm text-destructive">
                  {serverError}
                </div>
              )}

              <Button
                type="submit"
                disabled={isPending}
                className="w-full h-11 text-base rounded-[10px] mt-1"
                size="default"
              >
                {isPending ? "Iniciando sesión..." : "Iniciar sesión"}
              </Button>
            </form>

            {/* SSO providers */}
            {ssoProviders.length > 0 && (
              <div className="flex flex-col gap-4">
                <div className="flex items-center gap-3">
                  <div className="h-px flex-1 bg-border" />
                  <span className="text-xs text-fg-subtle">o continuar con</span>
                  <div className="h-px flex-1 bg-border" />
                </div>
                <div className="flex flex-col gap-2">
                  {ssoProviders.map((p) => (
                    <a
                      key={p.id}
                      href={`/api/auth/sso/${p.type}`}
                      className="flex items-center justify-center gap-3 h-11 rounded-[10px] border border-border bg-bg text-sm font-medium text-fg hover:bg-surface-2 transition-colors"
                    >
                      {SSO_ICONS[p.type] ?? null}
                      Iniciar con {p.name}
                    </a>
                  ))}
                </div>
              </div>
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
      </main>
    </div>
  )
}
