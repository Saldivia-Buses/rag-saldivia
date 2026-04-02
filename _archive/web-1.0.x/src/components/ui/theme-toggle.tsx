"use client"

import { useTheme } from "next-themes"
import { useEffect, useState } from "react"
import { Moon, Sun } from "lucide-react"

export function ThemeToggle() {
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = useState(false)

  useEffect(() => setMounted(true), [])

  if (!mounted) {
    return (
      <button
        className="flex items-center justify-center rounded-xl"
        style={{ width: "44px", height: "44px" }}
        aria-label="Tema"
      />
    )
  }

  return (
    <button
      onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
      title={theme === "dark" ? "Cambiar a light mode" : "Cambiar a dark mode"}
      aria-label={theme === "dark" ? "Cambiar a light mode" : "Cambiar a dark mode"}
      className="flex items-center justify-center rounded-xl text-fg-muted hover:text-fg hover:bg-surface-2 transition-colors"
      style={{ width: "44px", height: "44px" }}
    >
      {theme === "dark" ? <Sun size={20} /> : <Moon size={20} />}
    </button>
  )
}
