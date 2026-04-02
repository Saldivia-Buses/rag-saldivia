"use client";

import { useTheme } from "@/lib/theme-provider";
import { themePresets } from "@/lib/theme-presets";
import { cn } from "@/lib/utils";
import { CheckIcon } from "lucide-react";

export function ThemeSelector() {
  const { themeId, setTheme } = useTheme();

  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-lg font-semibold">Tema</h3>
        <p className="text-sm text-muted-foreground">
          Elegí el tema visual de la plataforma.
        </p>
      </div>

      <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
        {Object.entries(themePresets).map(([id, preset]) => {
          const isActive = id === themeId;
          const primary = preset.light.primary || "#000";
          const bg = preset.light.background || "#fff";
          const accent = preset.light.accent || "#eee";

          return (
            <button
              key={id}
              onClick={() => setTheme(id)}
              className={cn(
                "flex items-center gap-2.5 rounded-md border px-3 py-2 text-left transition-all hover:border-ring",
                isActive
                  ? "border-ring ring-2 ring-ring/20"
                  : "border-border"
              )}
            >
              <div className="flex -space-x-1 shrink-0">
                <div
                  className="size-3.5 rounded-full border border-background"
                  style={{ backgroundColor: primary }}
                />
                <div
                  className="size-3.5 rounded-full border border-background"
                  style={{ backgroundColor: accent }}
                />
                <div
                  className="size-3.5 rounded-full border border-background"
                  style={{ backgroundColor: bg }}
                />
              </div>
              <span className="text-xs">
                {preset.label}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
