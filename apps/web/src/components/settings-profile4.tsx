"use client";

import { useState, useEffect } from "react";
import { Bell, ChevronRight, Mail, User } from "lucide-react";
import Link from "next/link";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
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

  // Sync name when user data loads/changes
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
    <section className={cn("py-16", className)}>
      <div className="container">
        {/* Header */}
        <div className="mb-6">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <span>Configuración</span>
            <ChevronRight className="size-4" />
            <span className="text-foreground">Mi cuenta</span>
          </div>
          <h1 className="mt-2 text-2xl font-semibold">Mi cuenta</h1>
          <p className="text-muted-foreground">
            Administra tu información personal y preferencias.
          </p>
        </div>

        <div className="mx-auto w-full max-w-2xl">
          <div className="rounded-xl bg-card">
            {/* Personal info */}
            <div className="p-6">
              <h2 className="text-lg font-semibold">Información personal</h2>
              <p className="text-sm text-muted-foreground">
                Tu nombre y datos de acceso.
              </p>

              <div className="mt-6 flex flex-col gap-6">
                {/* Avatar */}
                <div className="flex items-center gap-4">
                  <Avatar className="size-16 shrink-0">
                    <AvatarFallback className="text-xl">
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

                <div className="h-4" />

                {/* Name (editable) */}
                <div className="space-y-2">
                  <Label htmlFor="name">
                    <User className="size-4" />
                    Nombre
                  </Label>
                  <Input
                    id="name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="Tu nombre completo"
                  />
                </div>

                {/* Email (readonly) */}
                <div className="space-y-2">
                  <Label htmlFor="email">
                    <Mail className="size-4" />
                    Email
                  </Label>
                  <Input
                    id="email"
                    type="email"
                    value={user?.email ?? ""}
                    readOnly
                    className="bg-muted cursor-not-allowed"
                  />
                  <p className="text-xs text-muted-foreground">
                    El email no se puede cambiar.
                  </p>
                </div>
              </div>
            </div>

            <div className="h-4" />

            {/* Notification preferences link */}
            <div className="p-6">
              <Link
                href="/notifications"
                className="flex items-center gap-3 rounded-lg px-4 py-3 text-sm transition-colors hover:bg-muted"
              >
                <Bell className="size-4 text-muted-foreground" />
                <div className="flex-1">
                  <p className="font-medium">Preferencias de notificaciones</p>
                  <p className="text-muted-foreground">
                    Configura como recibis las notificaciones.
                  </p>
                </div>
                <ChevronRight className="size-4 text-muted-foreground" />
              </Link>
            </div>

            {/* Footer */}
            <div className="flex items-center justify-end gap-3 px-6 py-4">
              <Button
                onClick={handleSave}
                disabled={
                  saveState === "saving" ||
                  !name.trim() ||
                  name.trim() === user?.name
                }
                variant={saveState === "error" ? "destructive" : "default"}
              >
                {buttonLabel}
              </Button>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export { SettingsProfile4 };
