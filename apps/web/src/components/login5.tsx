"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/lib/auth/store";
import { ApiError } from "@/lib/api/client";
import { LogIn } from "lucide-react";

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
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const login = useAuthStore((s) => s.login);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await login(email, password);
      window.location.href = "/inicio";
    } catch (err) {
      if (err instanceof ApiError) {
        switch (err.status) {
          case 401:
            setError("Email o contraseña incorrectos");
            break;
          case 429:
            setError("Demasiados intentos. Intenta de nuevo más tarde.");
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
    <section
      className={cn(
        "h-screen bg-background",
        "dark:bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] dark:from-card dark:to-background",
        className,
      )}
    >
      <div className="flex h-full items-center justify-center px-4">
        <div className="flex w-full max-w-sm flex-col items-center gap-8">
          {/* Logo + branding */}
          <div className="flex flex-col items-center gap-3">
            <div className="flex size-12 items-center justify-center rounded-xl bg-primary shadow-lg shadow-primary/20">
              <img
                src={logo.src}
                alt={logo.alt}
                title={logo.title}
                className={cn("size-7 invert dark:invert-0", logo.className)}
              />
            </div>
            <div className="text-center">
              <h1 className="text-xl font-semibold tracking-tight">{heading}</h1>
              <p className="text-sm text-muted-foreground mt-1">
                Plataforma empresarial
              </p>
            </div>
          </div>

          {/* Form card */}
          <form
            onSubmit={handleSubmit}
            className={cn(
              "flex w-full flex-col gap-5 rounded-xl p-6",
              "bg-card/80 backdrop-blur-xl border border-border/50",
              "shadow-xl shadow-black/5 dark:shadow-black/20",
            )}
          >
            {/* Error */}
            {error && (
              <div className="rounded-lg bg-destructive/10 border border-destructive/20 px-4 py-3 text-sm text-destructive">
                {error}
              </div>
            )}

            {/* Email */}
            <div className="flex flex-col gap-2">
              <Label htmlFor="email" className="text-sm font-medium">
                Correo electrónico
              </Label>
              <Input
                id="email"
                type="email"
                placeholder="tu@email.com"
                className="h-10 bg-background/50"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoComplete="email"
                autoFocus
              />
            </div>

            {/* Password */}
            <div className="flex flex-col gap-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="password" className="text-sm font-medium">
                  Contraseña
                </Label>
                <a
                  href="/forgot-password"
                  className="text-xs text-muted-foreground hover:text-primary transition-colors"
                >
                  Olvidé mi contraseña
                </a>
              </div>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                className="h-10 bg-background/50"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoComplete="current-password"
              />
            </div>

            {/* Submit */}
            <Button
              type="submit"
              className="h-10 w-full mt-1"
              disabled={loading}
              loading={loading}
              leadingIcon={!loading ? LogIn : undefined}
            >
              Iniciar sesión
            </Button>
          </form>

          {/* Footer */}
          <p className="text-xs text-muted-foreground/60">
            SDA Framework · v2.0
          </p>
        </div>
      </div>
    </section>
  );
};

export { Login5 };
