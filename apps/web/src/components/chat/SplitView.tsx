"use client"

import { useState } from "react"
import { Columns2, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { ChatInterface } from "@/components/chat/ChatInterface"
import type { DbChatSession, DbChatMessage } from "@rag-saldivia/db"

type SessionWithMessages = DbChatSession & { messages?: DbChatMessage[] }

type Props = {
  primarySession: SessionWithMessages
  secondarySession: SessionWithMessages | null
  userId: number
  onOpenSplit?: () => void
}

export function SplitView({ primarySession, secondarySession, userId, onOpenSplit }: Props) {
  const [splitActive, setSplitActive] = useState(!!secondarySession)

  if (!splitActive || !secondarySession) {
    return (
      <div className="flex-1 flex flex-col min-h-0 relative">
        <ChatInterface session={primarySession} userId={userId} />
        <div className="absolute top-3 right-16">
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            title="Vista dividida"
            onClick={() => {
              setSplitActive(true)
              onOpenSplit?.()
            }}
          >
            <Columns2 size={15} />
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex-1 flex min-h-0 relative">
      <div className="flex-1 flex flex-col min-h-0 border-r" style={{ borderColor: "var(--border)" }}>
        <ChatInterface session={primarySession} userId={userId} />
      </div>
      <div className="flex-1 flex flex-col min-h-0 relative">
        <ChatInterface session={secondarySession} userId={userId} />
        <Button
          variant="ghost"
          size="icon"
          className="absolute top-3 right-3 h-7 w-7 z-10"
          title="Cerrar vista dividida"
          onClick={() => setSplitActive(false)}
        >
          <X size={13} />
        </Button>
      </div>
    </div>
  )
}
