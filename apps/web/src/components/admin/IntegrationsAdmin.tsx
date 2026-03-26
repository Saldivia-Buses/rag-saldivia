"use client"

import { useState, useEffect } from "react"
import { MessageSquare, Copy, Check } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"

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
    setCopied(key); setTimeout(() => setCopied(null), 2000)
  }

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-lg font-semibold text-fg">Integraciones</h1>
        <p className="text-sm text-fg-muted mt-0.5">Conectá RAG Saldivia con tu stack de comunicación</p>
      </div>

      {/* Slack */}
      <div className="rounded-xl border border-border bg-surface p-5 space-y-4">
        <div className="flex items-center gap-2">
          <MessageSquare size={18} className="text-fg-muted" />
          <h2 className="font-semibold text-fg">Slack</h2>
          <Badge variant="outline">Bot</Badge>
        </div>
        <ol className="text-sm text-fg-muted space-y-2 list-decimal list-inside">
          <li>Crear una Slack App en <a href="https://api.slack.com/apps" target="_blank" rel="noreferrer" className="text-accent underline underline-offset-2">api.slack.com/apps</a></li>
          <li>Habilitar "Slash Commands" y usar la URL:</li>
        </ol>
        <div className="flex items-center gap-2 p-2.5 rounded-lg bg-surface-2 border border-border">
          <span className="flex-1 text-sm font-mono text-fg break-all">{slackUrl}</span>
          <Button variant="ghost" size="icon" className="h-7 w-7 shrink-0" onClick={() => copy(slackUrl, "slack")}>
            {copied === "slack" ? <Check size={12} /> : <Copy size={12} />}
          </Button>
        </div>
        <ol className="text-sm text-fg-muted space-y-1 list-decimal list-inside" start={3}>
          <li>Configurar: <code className="text-xs px-1 py-0.5 rounded bg-surface-2 font-mono">SLACK_BOT_TOKEN</code>, <code className="text-xs px-1 py-0.5 rounded bg-surface-2 font-mono">SLACK_SIGNING_SECRET</code></li>
          <li>Mapear usuarios con <code className="text-xs px-1 py-0.5 rounded bg-surface-2 font-mono">rag users list</code></li>
        </ol>
      </div>

      {/* Teams */}
      <div className="rounded-xl border border-border bg-surface p-5 space-y-4">
        <div className="flex items-center gap-2">
          <MessageSquare size={18} className="text-fg-muted" />
          <h2 className="font-semibold text-fg">Microsoft Teams</h2>
          <Badge variant="outline">Bot Framework</Badge>
        </div>
        <ol className="text-sm text-fg-muted space-y-2 list-decimal list-inside">
          <li>Registrar un bot en <a href="https://dev.botframework.com" target="_blank" rel="noreferrer" className="text-accent underline underline-offset-2">dev.botframework.com</a></li>
          <li>Usar el Messaging Endpoint:</li>
        </ol>
        <div className="flex items-center gap-2 p-2.5 rounded-lg bg-surface-2 border border-border">
          <span className="flex-1 text-sm font-mono text-fg break-all">{teamsUrl}</span>
          <Button variant="ghost" size="icon" className="h-7 w-7 shrink-0" onClick={() => copy(teamsUrl, "teams")}>
            {copied === "teams" ? <Check size={12} /> : <Copy size={12} />}
          </Button>
        </div>
        <ol className="text-sm text-fg-muted space-y-1 list-decimal list-inside" start={3}>
          <li>Configurar: <code className="text-xs px-1 py-0.5 rounded bg-surface-2 font-mono">TEAMS_BOT_ID</code>, <code className="text-xs px-1 py-0.5 rounded bg-surface-2 font-mono">TEAMS_BOT_PASSWORD</code></li>
        </ol>
      </div>

      {/* Mapeo */}
      <div className="rounded-xl border border-border bg-surface p-5">
        <p className="font-semibold text-fg mb-2">Mapeo de usuarios</p>
        <p className="text-sm text-fg-muted">
          Configurá el mapeo entre IDs de Slack/Teams y usuarios del sistema via la tabla{" "}
          <code className="text-xs px-1 py-0.5 rounded bg-surface-2 font-mono">bot_user_mappings</code> en la DB, o via CLI:{" "}
          <code className="text-xs px-1 py-0.5 rounded bg-surface-2 font-mono">rag users list</code>
        </p>
      </div>
    </div>
  )
}
