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
