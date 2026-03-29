"use client"

import { createContext, useContext, useState, useEffect, type ReactNode } from "react"

type SidebarContextType = {
  open: boolean
  toggle: () => void
}

const SidebarContext = createContext<SidebarContextType>({ open: true, toggle: () => {} })

export function useSidebar() {
  return useContext(SidebarContext)
}

export function ChatLayout({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(true)

  // Keyboard shortcut: Ctrl+Shift+S to toggle sidebar
  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.ctrlKey && e.shiftKey && e.key === "S") {
        e.preventDefault()
        setOpen(o => !o)
      }
    }
    window.addEventListener("keydown", onKeyDown)
    return () => window.removeEventListener("keydown", onKeyDown)
  }, [])

  return (
    <SidebarContext.Provider value={{ open, toggle: () => setOpen(o => !o) }}>
      <div className="flex h-full">
        {children}
      </div>
    </SidebarContext.Provider>
  )
}
