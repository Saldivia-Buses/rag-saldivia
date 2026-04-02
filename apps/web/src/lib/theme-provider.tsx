"use client";

import { createContext, useContext, useEffect, useState, useCallback } from "react";
import { themePresets, type ThemePreset } from "./theme-presets";
import { fontFamilyMap } from "./fonts";

const STORAGE_KEY = "sda-theme";
const DEFAULT_THEME = "modern-minimal";

interface ThemeContextValue {
  themeId: string;
  theme: ThemePreset;
  setTheme: (id: string) => void;
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

function resolveFontVar(fontValue: string): string {
  // fontValue is like "Plus Jakarta Sans, sans-serif"
  // Extract the font name and map to CSS variable
  const fontName = fontValue.split(",")[0].trim().replace(/"/g, "");
  const cssVar = fontFamilyMap[fontName];
  if (cssVar) {
    // Replace the font name with the CSS variable, keep fallbacks
    const fallbacks = fontValue.substring(fontValue.indexOf(",") + 1).trim();
    return `${cssVar}, ${fallbacks}`;
  }
  return fontValue;
}

// Defaults for variables that themes might not define
const defaults: Record<string, string> = {
  "font-sans": "Inter, sans-serif",
  "font-serif": "Georgia, serif",
  "font-mono": "monospace",
  "radius": "0.5rem",
};

// All CSS variable names that themes can set
const allThemeVars = [
  "background", "foreground", "card", "card-foreground", "popover", "popover-foreground",
  "primary", "primary-foreground", "secondary", "secondary-foreground",
  "muted", "muted-foreground", "accent", "accent-foreground",
  "destructive", "destructive-foreground", "border", "input", "ring",
  "chart-1", "chart-2", "chart-3", "chart-4", "chart-5",
  "sidebar", "sidebar-foreground", "sidebar-primary", "sidebar-primary-foreground",
  "sidebar-accent", "sidebar-accent-foreground", "sidebar-border", "sidebar-ring",
  "font-sans", "font-serif", "font-mono", "radius",
];

function applyTheme(preset: ThemePreset) {
  const root = document.documentElement;
  const isDark = root.classList.contains("dark");
  // Use dark vars but fall back to light for any missing keys (fonts, radius, etc.)
  const vars = isDark ? { ...preset.light, ...preset.dark } : preset.light;

  // Clear all theme vars first to prevent bleed from previous theme
  for (const key of allThemeVars) {
    root.style.removeProperty(`--${key}`);
  }

  // Apply theme vars, falling back to defaults for missing ones
  const resolved: Record<string, string> = {};
  for (const key of allThemeVars) {
    const value = vars[key] || defaults[key];
    if (!value) continue;
    const resolvedValue = key.startsWith("font-") ? resolveFontVar(value) : value;
    root.style.setProperty(`--${key}`, resolvedValue);
    resolved[key] = resolvedValue;
  }

  // Cache resolved vars for the inline script (prevents flash on navigation)
  try {
    localStorage.setItem("sda-theme-vars", JSON.stringify(resolved));
  } catch {}
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [themeId, setThemeId] = useState(DEFAULT_THEME);

  useEffect(() => {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved && themePresets[saved]) {
      setThemeId(saved);
      applyTheme(themePresets[saved]);
    } else {
      applyTheme(themePresets[DEFAULT_THEME]);
    }
  }, []);

  const setTheme = useCallback((id: string) => {
    if (!themePresets[id]) return;
    setThemeId(id);
    localStorage.setItem(STORAGE_KEY, id);
    applyTheme(themePresets[id]);
  }, []);

  // Re-apply theme when dark mode toggles
  useEffect(() => {
    const handler = () => {
      const current = themePresets[themeId] || themePresets[DEFAULT_THEME];
      applyTheme(current);
    };
    window.addEventListener("sda-mode-change", handler);
    return () => window.removeEventListener("sda-mode-change", handler);
  }, [themeId]);

  const theme = themePresets[themeId] || themePresets[DEFAULT_THEME];

  return (
    <ThemeContext.Provider value={{ themeId, theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error("useTheme must be used within ThemeProvider");
  return ctx;
}
