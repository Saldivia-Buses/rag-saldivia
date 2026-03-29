"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { Plus } from "lucide-react"
import { Separator } from "@/components/ui/separator"
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip"

export function ChatPanel() {
  const pathname = usePathname()

  return (
    <TooltipProvider delayDuration={100}>
      <div
        className="flex flex-col h-full"
        style={{ background: "var(--sidebar-bg)" }}
      >
        <div
          className="px-3 py-3 flex items-center justify-between flex-shrink-0"
        >
          <span
            className="text-xs font-semibold uppercase tracking-wider"
            style={{ color: "var(--muted-foreground)" }}
          >
            Sesiones
          </span>
          <Tooltip>
            <TooltipTrigger asChild>
              <Link
                href="/chat"
                className="w-6 h-6 flex items-center justify-center rounded-md transition-colors hover:opacity-80"
                style={{ color: "var(--muted-foreground)" }}
              >
                <Plus size={14} />
              </Link>
            </TooltipTrigger>
            <TooltipContent side="right" sideOffset={8}>
              Nueva sesión
            </TooltipContent>
          </Tooltip>
        </div>
        <Separator />
        {/* La lista de sesiones se renderiza via SessionList.tsx en la página /chat */}
        {/* En Fase 1 se integra aquí directamente */}
        <div className="flex-1 overflow-y-auto" id="chat-sessions-slot">
          <div
            className="px-3 py-2 text-xs"
            style={{ color: "var(--muted-foreground)" }}
          >
            <Link
              href="/chat"
              className="block w-full px-2 py-1.5 rounded-md text-sm transition-colors"
              style={{
                background: pathname === "/chat" ? "var(--accent)" : "transparent",
                color: pathname === "/chat" ? "white" : "var(--foreground)",
                opacity: pathname === "/chat" ? 1 : 0.75,
              }}
            >
              Ver todas las sesiones
            </Link>
          </div>
        </div>
      </div>
    </TooltipProvider>
  )
}
