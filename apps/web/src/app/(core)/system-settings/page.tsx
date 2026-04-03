"use client";

import { useState, useEffect } from "react";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { ThemeSelector } from "@/components/theme-selector";
import { DarkModeToggle } from "@/components/dark-mode-toggle";

function useLocalStorage(key: string, fallback: string) {
  const [value, setValue] = useState(fallback);

  useEffect(() => {
    const stored = localStorage.getItem(key);
    if (stored) setValue(stored);
  }, [key]);

  const set = (v: string | null) => {
    if (v === null) return;
    setValue(v);
    localStorage.setItem(key, v);
  };

  return [value, set] as const;
}

export default function SystemSettingsPage() {
  const [language, setLanguage] = useLocalStorage("sda-language", "es");
  const [timezone, setTimezone] = useLocalStorage(
    "sda-timezone",
    "America/Argentina/Buenos_Aires",
  );

  return (
    <div className="flex-1 overflow-y-auto px-10 py-8">
      <div className="mx-auto w-full max-w-4xl">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold">Configuración del sistema</h1>
          <p className="text-muted-foreground">
            Preferencias generales de la plataforma.
          </p>
        </div>

        <div className="rounded-xl border bg-card shadow-sm">
          <div className="p-6 space-y-6">
            {/* Language */}
            <div className="space-y-2">
              <Label htmlFor="language">Idioma</Label>
              <Select value={language} onValueChange={setLanguage}>
                <SelectTrigger id="language">
                  <SelectValue placeholder="Seleccionar idioma" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="en">English</SelectItem>
                  <SelectItem value="es">Español</SelectItem>
                  <SelectItem value="fr">Français</SelectItem>
                  <SelectItem value="de">Deutsch</SelectItem>
                  <SelectItem value="ja">日本語</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Timezone */}
            <div className="space-y-2">
              <Label htmlFor="timezone">Zona horaria</Label>
              <Select value={timezone} onValueChange={setTimezone}>
                <SelectTrigger id="timezone">
                  <SelectValue placeholder="Seleccionar zona horaria" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="America/Argentina/Buenos_Aires">
                    Argentina (ART)
                  </SelectItem>
                  <SelectItem value="America/New_York">
                    Eastern Time (ET)
                  </SelectItem>
                  <SelectItem value="America/Chicago">
                    Central Time (CT)
                  </SelectItem>
                  <SelectItem value="America/Los_Angeles">
                    Pacific Time (PT)
                  </SelectItem>
                  <SelectItem value="Europe/London">
                    Greenwich Mean Time (GMT)
                  </SelectItem>
                  <SelectItem value="Europe/Paris">
                    Central European Time (CET)
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Dark mode */}
            <div className="flex items-center justify-between">
              <div>
                <Label>Modo oscuro</Label>
                <p className="text-xs text-muted-foreground">
                  Alternar entre modo claro y oscuro.
                </p>
              </div>
              <DarkModeToggle />
            </div>
          </div>
        </div>

        <Separator className="my-8" />

        {/* Theme selector */}
        <ThemeSelector />
      </div>
    </div>
  );
}
