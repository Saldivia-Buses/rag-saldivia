"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";

/**
 * Login form — isolated from the main app.
 * No sidebar, no nav, no system UI. Just email, password, and the logo.
 * This is the only door to the system.
 */
export function LoginForm() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const apiBase = process.env.NEXT_PUBLIC_API_URL ?? "";

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const res = await fetch(`${apiBase}/v1/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ email, password }),
      });

      if (res.status === 401) {
        setError("Email o contrasena incorrectos");
        return;
      }
      if (res.status === 429) {
        setError("Demasiados intentos. Intenta de nuevo mas tarde.");
        return;
      }
      if (!res.ok) {
        setError("Error interno. Intenta de nuevo.");
        return;
      }

      // Login succeeded — refresh cookie is set by the backend.
      // Redirect to the main app (same tenant subdomain, different app).
      const host = window.location.hostname;
      const appUrl =
        host === "localhost" || host === "127.0.0.1"
          ? "http://localhost:3000"
          : `https://${host.replace(/^login\./, "")}`;

      window.location.href = appUrl;
    } catch {
      setError("No se pudo conectar al servidor.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="flex w-full max-w-sm flex-col gap-y-4 rounded-lg bg-muted px-6 py-12"
    >
      {error && (
        <div className="w-full rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

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

      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <label htmlFor="password" className="text-sm font-medium">
            Contrasena
          </label>
          <a
            href="/forgot-password"
            className="text-xs text-muted-foreground hover:text-primary hover:underline"
          >
            Olvide mi contrasena
          </a>
        </div>
        <input
          id="password"
          type="password"
          placeholder="••••••••"
          className="rounded-[var(--radius)] border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          autoComplete="current-password"
        />
      </div>

      <button
        type="submit"
        disabled={loading}
        className={cn(
          "rounded-[var(--radius)] bg-primary px-4 py-2 text-sm font-medium text-primary-foreground",
          "hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed",
        )}
      >
        {loading ? "Ingresando..." : "Iniciar sesion"}
      </button>
    </form>
  );
}
