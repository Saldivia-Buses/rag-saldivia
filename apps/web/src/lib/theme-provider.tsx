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

function applyTheme(preset: ThemePreset) {
  const root = document.documentElement;
  const isDark = root.classList.contains("dark");
  const vars = isDark ? preset.dark : preset.light;

  const resolved: Record<string, string> = {};
  for (const [key, value] of Object.entries(vars)) {
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
