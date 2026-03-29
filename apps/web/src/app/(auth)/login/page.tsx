"use client"

import { Suspense, useState, useTransition } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { ThemeToggle } from "@/components/ui/theme-toggle"

const LoginSchema = z.object({
  email: z.string().email("Email inválido"),
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

function LoginContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const from = searchParams.get("from") ?? "/chat"
  const [serverError, setServerError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

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
          <div className="rounded-2xl border border-border bg-surface shadow-sm px-8 py-10 space-y-8">

            {/* Header */}
            <div className="text-center space-y-1.5">
              <div className="inline-flex h-10 w-10 items-center justify-center rounded-xl bg-accent mb-3">
                <span className="text-base font-bold text-accent-fg select-none">S</span>
              </div>
              <h1 className="text-xl font-semibold text-fg tracking-tight">
                Saldivia RAG
              </h1>
              <p className="text-sm text-fg-muted">
                Iniciá sesión para continuar
              </p>
            </div>

            {/* Form */}
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div className="space-y-1.5">
                <label htmlFor="email" className="text-sm font-medium text-fg block">
                  Email
                </label>
                <Input
                  id="email"
                  type="email"
                  placeholder="usuario@empresa.com"
                  autoComplete="email"
                  autoFocus
                  {...register("email")}
                />
                {errors.email && (
                  <p className="text-xs text-destructive">{errors.email.message}</p>
                )}
              </div>

              <div className="space-y-1.5">
                <label htmlFor="password" className="text-sm font-medium text-fg block">
                  Contraseña
                </label>
                <Input
                  id="password"
                  type="password"
                  placeholder="••••••••"
                  autoComplete="current-password"
                  {...register("password")}
                />
                {errors.password && (
                  <p className="text-xs text-destructive">{errors.password.message}</p>
                )}
              </div>

              {serverError && (
                <div className="px-3 py-2.5 rounded-lg bg-destructive-subtle border border-destructive/20 text-sm text-destructive">
                  {serverError}
                </div>
              )}

              <Button
                type="submit"
                disabled={isPending}
                className="w-full"
                size="default"
              >
                {isPending ? "Iniciando sesión..." : "Iniciar sesión"}
              </Button>
            </form>
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
