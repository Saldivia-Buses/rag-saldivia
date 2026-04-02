/**
 * Chat sidebar context and layout wrapper.
 *
 * Provides `SidebarProvider` (context for open/closed state) and `ChatLayout`
 * (the flex container that holds SessionList + chat content side by side).
 *
 * Sidebar state is persisted in localStorage (`saldivia-sidebar-open`) so it
 * survives page reloads and navigation. A keyboard shortcut (Ctrl+Shift+S)
 * toggles the sidebar from anywhere within the provider.
 *
 * Consumed by: NavRail (toggle button), SessionList (reads `open` state),
 *              AppShellChrome (wraps entire app in SidebarProvider)
 * Depends on: React context only — no external deps
 */
"use client"

import { createContext, useContext, useEffect, useCallback, type ReactNode } from "react"
import { useLocalStorage } from "@/hooks/useLocalStorage"

type SidebarContextType = {
  open: boolean
  toggle: () => void
}

const SidebarContext = createContext<SidebarContextType>({ open: true, toggle: () => {} })

export function useSidebar() {
  return useContext(SidebarContext)
}

export function SidebarProvider({ children }: { children: ReactNode }) {
  const [open, setOpen] = useLocalStorage("saldivia-sidebar-open", true)

  const toggle = useCallback(() => {
    setOpen((prev) => !prev)
  }, [setOpen])

  // Keyboard shortcut: Ctrl+Shift+S
  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.ctrlKey && e.shiftKey && e.key === "S") {
        e.preventDefault()
        toggle()
      }
    }
    window.addEventListener("keydown", onKeyDown)
    return () => window.removeEventListener("keydown", onKeyDown)
  }, [toggle])

  return (
    <SidebarContext.Provider value={{ open, toggle }}>
      {children}
    </SidebarContext.Provider>
  )
}

export function ChatLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex h-full">
      {children}
    </div>
  )
}
