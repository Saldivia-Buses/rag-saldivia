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

function usePersistedState(key: string, defaultValue: boolean): [boolean, () => void] {
  const [value, setValue] = useState(() => {
    if (typeof window === "undefined") return defaultValue
    const stored = localStorage.getItem(key)
    return stored !== null ? stored === "true" : defaultValue
  })

  const toggle = useCallback(() => {
    setValue(prev => {
      const next = !prev
      localStorage.setItem(key, String(next))
      return next
    })
  }, [key])

  return [value, toggle]
}

export function ChatLayout({ children }: { children: ReactNode }) {
  const [open, toggle] = usePersistedState("saldivia-sidebar-open", true)

  // Keyboard shortcut: Ctrl+Shift+S to toggle sidebar
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
      <div className="flex h-full">
        {children}
      </div>
    </SidebarContext.Provider>
  )
}
