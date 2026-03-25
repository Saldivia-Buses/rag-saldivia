import * as p from "@clack/prompts"
import chalk from "chalk"
import { api } from "../client.js"
import { out, makeTable, progressBar, handleApiError } from "../output.js"

export async function ingestStartCommand(opts: { collection?: string; path?: string }) {
  const collection = opts.collection ?? String(await p.text({
    message: "Colección destino",
    validate: (v) => !v ? "Requerido" : undefined,
  }))

  const filePath = opts.path ?? String(await p.text({
    message: "Ruta del documento o directorio",
    validate: (v) => !v ? "Requerido" : undefined,
  }))

  if (p.isCancel(collection) || p.isCancel(filePath)) {
    p.cancel("Cancelado")
    return
  }

  const spinner = p.spinner()
  spinner.start("Encolando para ingesta...")

  const result = await api.ingestion.start(collection, filePath)
  if (!result.ok) {
    spinner.stop(chalk.red("Error"))
    return handleApiError(result)
  }

  spinner.stop(chalk.green("Encolado"))
  out.ok(`Job de ingesta creado para ${chalk.bold(filePath)} → ${chalk.bold(collection)}`)
  out.info("Usá 'rag ingest status' para ver el progreso")
}

export async function ingestStatusCommand() {
  out.section("Estado de ingesta")

  const result = await api.ingestion.status()
  if (!result.ok) return handleApiError(result)

  // La API retorna { queue: [...], jobs: [...] } — combinamos ambas listas
  const data = result.data as { queue?: unknown[]; jobs?: unknown[] } | unknown[]
  const items: Array<{
    id: string; filePath?: string; filename?: string
    collection?: string; status?: string; state?: string
    retryCount?: number; progress?: number; createdAt: number
  }> = Array.isArray(data) ? data : [
    ...((data as { queue?: unknown[] }).queue ?? []),
    ...((data as { jobs?: unknown[] }).jobs ?? []),
  ] as never

  if (items.length === 0) {
    out.info("No hay jobs activos")
    return
  }

  const rows = items.map((item) => {
    const state = item.status ?? item.state ?? "unknown"
    const stateColor = state === "done" || state === "completed" ? chalk.green
      : state === "error" ? chalk.red
      : state === "locked" || state === "running" ? chalk.cyan
      : chalk.dim

    const filename = item.filename ?? item.filePath?.split("/").pop() ?? item.id.slice(0, 8)

    return [
      chalk.dim(item.id.slice(0, 8)),
      chalk.bold(filename.slice(0, 30)),
      item.collection ?? "—",
      progressBar(item.progress ?? 0),
      stateColor(state),
    ]
  })

  console.log(makeTable(["ID", "Archivo", "Colección", "Progreso", "Estado"], rows))
}

export async function ingestCancelCommand(jobId: string) {
  const confirmed = await p.confirm({
    message: `¿Cancelar el job ${jobId}?`,
  })

  if (!confirmed || p.isCancel(confirmed)) return

  const result = await api.ingestion.cancel(jobId)
  if (!result.ok) return handleApiError(result)

  out.ok(`Job ${jobId} cancelado`)
}
