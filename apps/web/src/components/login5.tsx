"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/lib/auth/store";
import { ApiError } from "@/lib/api/client";

interface Login5Props {
  heading?: string;
  logo?: {
    url?: string;
    src: string;
    alt: string;
    title?: string;
    className?: string;
  };
  className?: string;
}

const Login5 = ({
  heading = "Iniciar sesión",
  logo = {
    src: "/logo-placeholder.svg",
    alt: "SDA Framework",
    title: "SDA Framework",
  },
  className,
}: Login5Props) => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [rememberMe, setRememberMe] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const login = useAuthStore((s) => s.login);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await login(email, password);
      // Redirect to dashboard
      window.location.href = "/dashboard";
    } catch (err) {
      if (err instanceof ApiError) {
        switch (err.status) {
          case 401:
            setError("Email o contrasena incorrectos");
            break;
          case 429:
            setError("Demasiados intentos. Intenta de nuevo mas tarde.");
            break;
          default:
            setError("Error interno. Intenta de nuevo.");
        }
      } else {
        setError("No se pudo conectar al servidor.");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <section className={cn("h-screen bg-background", className)}>
      <div className="flex h-full items-center justify-center">
        <div className="flex flex-col items-center gap-6 lg:justify-start">
          {/* Logo */}
          <div>
            <img
              src={logo.src}
              alt={logo.alt}
              title={logo.title}
              className={cn("h-10 dark:invert", logo.className)}
            />
          </div>
          {heading && <h1 className="text-2xl font-semibold">{heading}</h1>}
          <form
            onSubmit={handleSubmit}
            className="flex w-full max-w-sm min-w-sm flex-col items-center gap-y-4 rounded-lg bg-muted px-6 py-12"
          >
            {/* Error message */}
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
            <div className="flex w-full flex-col gap-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="password">Contraseña</Label>
                <a
                  href="/forgot-password"
                  className="text-xs text-muted-foreground hover:text-primary hover:underline"
                >
                  Olvidé mi contraseña
                </a>
              </div>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                className="bg-background text-sm"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoComplete="current-password"
              />
            </div>

            {/* Remember me */}
            <label className="flex w-full items-center gap-2 text-sm text-muted-foreground cursor-pointer">
              <input
                type="checkbox"
                checked={rememberMe}
                onChange={(e) => setRememberMe(e.target.checked)}
                className="rounded"
              />
              Recordarme por 30 días
            </label>

            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "Ingresando..." : "Iniciar sesión"}
            </Button>

            {/* Social login divider */}
            <div className="flex w-full items-center gap-4">
              <div className="h-px flex-1 bg-muted-foreground/20" />
              <span className="text-xs text-muted-foreground">o continuar con</span>
              <div className="h-px flex-1 bg-muted-foreground/20" />
            </div>

            <div className="flex w-full flex-col gap-2">
              <Button type="button" className="w-full" variant="outline">
                <img
                  src="https://deifkwefumgah.cloudfront.net/shadcnblocks/block/logos/google-icon.svg"
                  className="size-5"
                  alt="Google"
                />
                Google
              </Button>
              <Button type="button" className="w-full" variant="outline">
                <img
                  src="https://deifkwefumgah.cloudfront.net/shadcnblocks/block/logos/facebook-icon.svg"
                  className="size-5"
                  alt="Facebook"
                />
                Facebook
              </Button>
              <Button type="button" className="w-full" variant="outline">
                <img
                  src="https://deifkwefumgah.cloudfront.net/shadcnblocks/block/logos/github-icon.svg"
                  className="size-5"
                  alt="GitHub"
                />
                GitHub
              </Button>
            </div>
          </form>
        </div>
      </div>
    </section>
  );
};

export { Login5 };
