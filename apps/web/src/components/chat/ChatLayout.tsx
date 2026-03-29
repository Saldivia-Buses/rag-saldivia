"use client"

import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from "react"

type SidebarContextType = {
  open: boolean
  toggle: () => void
}

const SidebarContext = createContext<SidebarContextType>({ open: true, toggle: () => {} })

export function useSidebar() {
  return useContext(SidebarContext)
}

export function SidebarProvider({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(() => {
    if (typeof window === "undefined") return true
    const stored = localStorage.getItem("saldivia-sidebar-open")
    return stored !== null ? stored === "true" : true
  })

  const toggle = useCallback(() => {
    setOpen(prev => {
      const next = !prev
      localStorage.setItem("saldivia-sidebar-open", String(next))
      return next
    })
  }, [])

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
