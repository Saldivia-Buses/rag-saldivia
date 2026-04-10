"use client";

import { useState, useEffect, useCallback } from "react";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";

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

  const [isDark, setIsDark] = useState(false);

  useEffect(() => {
    setIsDark(document.documentElement.classList.contains("dark"));
  }, []);

  const toggleDark = useCallback((checked: boolean) => {
    setIsDark(checked);
    document.documentElement.classList.toggle("dark", checked);
    localStorage.setItem("sda-dark-mode", String(checked));
  }, []);

  return (
    <div className="flex-1 overflow-y-auto p-8">
      <div className="mx-auto w-full max-w-2xl">
        <div className="mb-6">
          <h1 className="text-lg font-semibold">Configuración del sistema</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Preferencias generales de la plataforma.
          </p>
        </div>

        <div className="rounded-xl bg-card border border-border/40">
          <div className="p-6 space-y-6">
            {/* Language */}
            <div className="space-y-2">
              <Label htmlFor="language" className="text-sm">Idioma</Label>
              <Select value={language} onValueChange={setLanguage}>
                <SelectTrigger id="language" className="h-10">
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
              <Label htmlFor="timezone" className="text-sm">Zona horaria</Label>
              <Select value={timezone} onValueChange={setTimezone}>
                <SelectTrigger id="timezone" className="h-10">
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

            <div className="h-px bg-border/40" />

            {/* Dark mode */}
            <div className="flex items-center justify-between">
              <div>
                <Label className="text-sm">Modo oscuro</Label>
                <p className="text-[11px] text-muted-foreground mt-0.5">
                  Alternar entre modo claro y oscuro.
                </p>
              </div>
              <Switch
                checked={isDark}
                onCheckedChange={toggleDark}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
