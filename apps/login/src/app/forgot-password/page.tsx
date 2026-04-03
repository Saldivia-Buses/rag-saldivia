"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [sent, setSent] = useState(false);
  const [loading, setLoading] = useState(false);

  const apiBase = process.env.NEXT_PUBLIC_API_URL ?? "";

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      await fetch(`${apiBase}/v1/auth/forgot-password`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });
    } catch {
      // Always show success to prevent email enumeration
    }

    setSent(true);
    setLoading(false);
  };

  if (sent) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <div className="flex flex-col items-center gap-4 text-center max-w-sm">
          <h1 className="text-xl font-semibold">Revisa tu correo</h1>
          <p className="text-sm text-muted-foreground">
            Si existe una cuenta con ese email, vas a recibir instrucciones para
            restablecer tu contrasena.
          </p>
          <a
            href="/login"
            className="text-sm text-primary hover:underline mt-2"
          >
            Volver a iniciar sesion
          </a>
        </div>
      </main>
    );
  }

  return (
    <main className="flex min-h-screen items-center justify-center">
      <div className="flex flex-col items-center gap-6">
        <h1 className="text-xl font-semibold">Restablecer contrasena</h1>
        <p className="text-sm text-muted-foreground text-center max-w-sm">
          Ingresa tu correo electronico y te enviaremos instrucciones.
        </p>

        <form
          onSubmit={handleSubmit}
          className="flex w-full max-w-sm flex-col gap-y-4 rounded-lg bg-muted px-6 py-8"
        >
          <div className="flex flex-col gap-2">
            <label htmlFor="email" className="text-sm font-medium">
              Correo electronico
            </label>
            <input
              id="email"
              type="email"
              placeholder="tu@email.com"
              className="rounded-[var(--radius)] border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              autoComplete="email"
              autoFocus
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className={cn(
              "rounded-[var(--radius)] bg-primary px-4 py-2 text-sm font-medium text-primary-foreground",
              "hover:bg-primary/90 disabled:opacity-50",
            )}
          >
            {loading ? "Enviando..." : "Enviar instrucciones"}
          </button>

          <a
            href="/login"
            className="text-xs text-center text-muted-foreground hover:text-primary hover:underline"
          >
            Volver a iniciar sesion
          </a>
        </form>
      </div>
    </main>
  );
}
