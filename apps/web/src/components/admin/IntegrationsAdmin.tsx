"use client"

import { useState, useEffect } from "react"
import { MessageSquare, Copy, Check } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"

const BASE_URL = typeof window !== "undefined" ? window.location.origin : ""

export function IntegrationsAdmin() {
  const [slackUrl, setSlackUrl] = useState("")
  const [teamsUrl, setTeamsUrl] = useState("")
  const [copied, setCopied] = useState<string | null>(null)

  useEffect(() => {
    setSlackUrl(`${window.location.origin}/api/slack`)
    setTeamsUrl(`${window.location.origin}/api/teams`)
  }, [])

  async function copy(text: string, key: string) {
    await navigator.clipboard.writeText(text)
    setCopied(key)
    setTimeout(() => setCopied(null), 2000)
  }

  const urlRowClass = "flex items-center gap-2 p-2 rounded-md text-sm font-mono"
  const urlStyle = { background: "var(--muted)", color: "var(--foreground)", wordBreak: "break-all" as const }

  return (
    <div className="space-y-8">
      {/* Slack */}
      <div className="p-4 rounded-xl border space-y-3" style={{ borderColor: "var(--border)" }}>
        <div className="flex items-center gap-2">
          <MessageSquare size={18} style={{ color: "#4A154B" }} />
          <h2 className="font-semibold">Slack</h2>
          <Badge variant="outline" className="text-xs">Bot</Badge>
        </div>
        <ol className="text-sm space-y-2" style={{ color: "var(--muted-foreground)" }}>
          <li>1. Crear una Slack App en <a href="https://api.slack.com/apps" target="_blank" rel="noreferrer" className="underline" style={{ color: "var(--accent)" }}>api.slack.com/apps</a></li>
          <li>2. Habilitar "Slash Commands" y usar la URL:</li>
        </ol>
        <div className={urlRowClass} style={urlStyle}>
          <span className="flex-1 break-all">{slackUrl}</span>
          <Button variant="ghost" size="icon" className="h-6 w-6 shrink-0" onClick={() => copy(slackUrl, "slack")}>
            {copied === "slack" ? <Check size={12} /> : <Copy size={12} />}
          </Button>
        </div>
        <ol className="text-sm space-y-1" style={{ color: "var(--muted-foreground)" }} start={3}>
          <li>3. Configurar env vars: <code className="text-xs px-1 rounded" style={{ background: "var(--muted)" }}>SLACK_BOT_TOKEN</code>, <code className="text-xs px-1 rounded" style={{ background: "var(--muted)" }}>SLACK_SIGNING_SECRET</code></li>
          <li>4. Mapear usuarios en "Mapeo de usuarios" abajo</li>
        </ol>
      </div>

      {/* Teams */}
      <div className="p-4 rounded-xl border space-y-3" style={{ borderColor: "var(--border)" }}>
        <div className="flex items-center gap-2">
          <MessageSquare size={18} style={{ color: "#464EB8" }} />
          <h2 className="font-semibold">Microsoft Teams</h2>
          <Badge variant="outline" className="text-xs">Bot Framework</Badge>
        </div>
        <ol className="text-sm space-y-2" style={{ color: "var(--muted-foreground)" }}>
          <li>1. Registrar un bot en <a href="https://dev.botframework.com" target="_blank" rel="noreferrer" className="underline" style={{ color: "var(--accent)" }}>dev.botframework.com</a></li>
          <li>2. Usar el Messaging Endpoint:</li>
        </ol>
        <div className={urlRowClass} style={urlStyle}>
          <span className="flex-1 break-all">{teamsUrl}</span>
          <Button variant="ghost" size="icon" className="h-6 w-6 shrink-0" onClick={() => copy(teamsUrl, "teams")}>
            {copied === "teams" ? <Check size={12} /> : <Copy size={12} />}
          </Button>
        </div>
        <ol className="text-sm space-y-1" style={{ color: "var(--muted-foreground)" }} start={3}>
          <li>3. Configurar env vars: <code className="text-xs px-1 rounded" style={{ background: "var(--muted)" }}>TEAMS_BOT_ID</code>, <code className="text-xs px-1 rounded" style={{ background: "var(--muted)" }}>TEAMS_BOT_PASSWORD</code></li>
        </ol>
      </div>

      <div className="p-4 rounded-xl border text-sm" style={{ borderColor: "var(--border)" }}>
        <p className="font-medium mb-2">Mapeo de usuarios</p>
        <p style={{ color: "var(--muted-foreground)" }}>
          El mapeo entre IDs de Slack/Teams y usuarios del sistema se configura via la tabla <code className="text-xs px-1 rounded" style={{ background: "var(--muted)" }}>bot_user_mappings</code> en la DB, o via CLI: <code className="text-xs px-1 rounded" style={{ background: "var(--muted)" }}>rag users list</code> para obtener los IDs del sistema.
        </p>
      </div>
    </div>
  )
}
