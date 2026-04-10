"use client";

import { useState, useEffect } from "react";
import { Bell, ChevronRight, Mail, User } from "lucide-react";
import Link from "next/link";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/lib/auth/store";
import { useUpdateProfile } from "@/hooks/use-update-profile";

interface SettingsProfile4Props {
  className?: string;
}

const SettingsProfile4 = ({ className }: SettingsProfile4Props) => {
  const user = useAuthStore((s) => s.user);
  const updateProfile = useUpdateProfile();

  const [name, setName] = useState(user?.name ?? "");
  const [saveState, setSaveState] = useState<
    "idle" | "saving" | "saved" | "error"
  >("idle");

  useEffect(() => {
    if (user?.name) setName(user.name);
  }, [user?.name]);

  const initials = user?.name
    ?.split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase();

  const handleSave = () => {
    const trimmed = name.trim();
    if (!trimmed || trimmed === user?.name) return;

    setSaveState("saving");
    updateProfile.mutate(
      { name: trimmed },
      {
        onSuccess: () => {
          setSaveState("saved");
          setTimeout(() => setSaveState("idle"), 2000);
        },
        onError: () => {
          setSaveState("error");
          setTimeout(() => setSaveState("idle"), 3000);
        },
      },
    );
  };

  const buttonLabel = {
    idle: "Guardar cambios",
    saving: "Guardando...",
    saved: "Guardado",
    error: "Error al guardar",
  }[saveState];

  return (
    <section className={cn("", className)}>
      <div className="mx-auto w-full max-w-2xl">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-lg font-semibold">Mi cuenta</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Administrá tu información personal y preferencias.
          </p>
        </div>

        <div className="rounded-xl bg-card border border-border/40">
          {/* Personal info */}
          <div className="p-6">
            <h2 className="text-base font-medium">Información personal</h2>
            <p className="text-sm text-muted-foreground mt-0.5">
              Tu nombre y datos de acceso.
            </p>

            <div className="mt-6 flex flex-col gap-6">
              {/* Avatar */}
              <div className="flex items-center gap-4">
                <Avatar className="size-14 shrink-0">
                  <AvatarFallback className="text-lg font-medium bg-primary/10 text-primary">
                    {initials}
                  </AvatarFallback>
                </Avatar>
                <div>
                  <p className="font-medium">{user?.name ?? "—"}</p>
                  <p className="text-sm text-muted-foreground">
                    {user?.email ?? "—"}
                  </p>
                </div>
              </div>

              <div className="h-px bg-border/40" />

              {/* Name */}
              <div className="space-y-2">
                <Label htmlFor="name" className="text-sm">
                  Nombre
                </Label>
                <Input
                  id="name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Tu nombre completo"
                  className="h-10"
                />
              </div>

              {/* Email */}
              <div className="space-y-2">
                <Label htmlFor="email" className="text-sm">
                  Email
                </Label>
                <Input
                  id="email"
                  type="email"
                  value={user?.email ?? ""}
                  readOnly
                  className="h-10 bg-muted/50 cursor-not-allowed"
                />
                <p className="text-[11px] text-muted-foreground">
                  El email no se puede cambiar.
                </p>
              </div>
            </div>
          </div>

          {/* Notification preferences link */}
          <div className="px-6 pb-6">
            <Link
              href="/notifications"
              className="flex items-center gap-3 rounded-lg border border-border/40 px-4 py-3 text-sm transition-colors hover:bg-accent/30"
            >
              <Bell className="size-4 text-muted-foreground" />
              <div className="flex-1">
                <p className="font-medium text-sm">Preferencias de notificaciones</p>
                <p className="text-xs text-muted-foreground">
                  Configurá cómo recibís las notificaciones.
                </p>
              </div>
              <ChevronRight className="size-4 text-muted-foreground" />
            </Link>
          </div>

          {/* Footer */}
          <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-border/40">
            <Button
              onClick={handleSave}
              disabled={
                saveState === "saving" ||
                !name.trim() ||
                name.trim() === user?.name
              }
              loading={saveState === "saving"}
              variant={saveState === "error" ? "destructive" : "default"}
            >
              {buttonLabel}
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
};

export { SettingsProfile4 };
