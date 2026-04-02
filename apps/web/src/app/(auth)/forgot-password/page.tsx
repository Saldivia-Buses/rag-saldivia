"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const res = await fetch("/v1/auth/forgot-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });

      if (res.status === 429) {
        setError("Demasiados intentos. Intentá de nuevo más tarde.");
        return;
      }

      // Always show success to avoid email enumeration
      setSubmitted(true);
    } catch {
      setError("No se pudo conectar al servidor.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <section className="h-screen bg-background">
      <div className="flex h-full items-center justify-center">
        <div className="flex flex-col items-center gap-6 lg:justify-start">
          <div>
            <img
              src="/logo-placeholder.svg"
              alt="SDA Framework"
              title="SDA Framework"
              className="h-10 dark:invert"
            />
          </div>
          <h1 className="text-2xl font-semibold">Recuperar contraseña</h1>

          {submitted ? (
            <div className="flex w-full max-w-sm min-w-sm flex-col items-center gap-y-4 rounded-lg bg-muted px-6 py-12">
              <p className="text-sm text-center text-muted-foreground">
                Si existe una cuenta con ese correo, vas a recibir un enlace
                para restablecer tu contraseña.
              </p>
              <a href="/login">
                <Button variant="outline" className="mt-2">
                  Volver al inicio de sesión
                </Button>
              </a>
            </div>
          ) : (
            <form
              onSubmit={handleSubmit}
              className="flex w-full max-w-sm min-w-sm flex-col items-center gap-y-4 rounded-lg bg-muted px-6 py-12"
            >
              <p className="text-sm text-center text-muted-foreground">
                Ingresá tu correo electrónico y te enviaremos un enlace para
                restablecer tu contraseña.
              </p>

              {error && (
                <div className="w-full rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
                  {error}
                </div>
              )}

              <div className="flex w-full flex-col gap-2">
                <Label htmlFor="email">Correo electrónico</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="tu@email.com"
                  className="bg-background text-sm"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  autoComplete="email"
                />
              </div>

              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? "Enviando..." : "Enviar enlace"}
              </Button>

              <a
                href="/login"
                className="text-xs text-muted-foreground hover:text-primary hover:underline"
              >
                Volver al inicio de sesión
              </a>
            </form>
          )}
        </div>
      </div>
    </section>
  );
}
