import chalk from "chalk"
import { api } from "../client.js"
import { out, makeTable, handleApiError } from "../output.js"
import { formatTimeline, reconstructFromEvents } from "@rag-saldivia/logger/blackbox"
import type { DbEvent } from "@rag-saldivia/db"

export async function auditLogCommand(opts: {
  limit?: number
  level?: string
  type?: string
}) {
  out.section("Audit Log")

  const result = await api.audit.list({
    limit: opts.limit ?? 50,
    level: opts.level,
    type: opts.type,
  })

  if (!result.ok) return handleApiError(result)

  const events = result.data as Array<{
    id: string
    ts: number
    level: string
    type: string
    source: string
    userId?: number
    payload: Record<string, unknown>
  }>

  if (events.length === 0) {
    out.info("No hay eventos registrados")
    return
  }

  const levelColor = (level: string) => {
    switch (level) {
      case "ERROR":
      case "FATAL": return chalk.red(level.padEnd(5))
      case "WARN": return chalk.yellow(level.padEnd(5))
      case "INFO": return chalk.green(level.padEnd(5))
      default: return chalk.dim(level.padEnd(5))
    }
  }

  const rows = events.map((e) => [
    chalk.dim(new Date(e.ts).toISOString().replace("T", " ").slice(0, 19)),
    levelColor(e.level),
    chalk.bold(e.type),
    e.userId ? chalk.dim(`user=${e.userId}`) : chalk.dim("—"),
    chalk.dim(JSON.stringify(e.payload).slice(0, 50)),
  ])

  console.log(makeTable(["Timestamp", "Nivel", "Tipo", "Usuario", "Payload"], rows))
  console.log(chalk.dim(`\n  Mostrando ${events.length} evento(s)\n`))
}

export async function auditReplayCommand(fromDate: string) {
  out.section(`Black Box Replay desde ${fromDate}`)

  const result = await api.audit.replay(fromDate)
  if (!result.ok) return handleApiError(result)

  const { timeline, stats } = result.data as {
    timeline: DbEvent[]
    stats: { totalEvents: number; errorCount: number }
  }

  const state = reconstructFromEvents(timeline as DbEvent[])
  console.log(formatTimeline(state))
}

export async function auditExportCommand() {
  const result = await api.audit.export()
  if (!result.ok) return handleApiError(result)

  console.log(JSON.stringify(result.data, null, 2))
}
