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

  const jobs = result.data as Array<{
    id: string
    filename: string
    collection: string
    state: string
    progress: number
    tier: string
    createdAt: number
  }>

  if (jobs.length === 0) {
    out.info("No hay jobs activos")
    return
  }

  const rows = jobs.map((job) => {
    const stateColor = job.state === "done" ? chalk.green
      : job.state === "error" ? chalk.red
      : job.state === "running" ? chalk.cyan
      : chalk.dim

    return [
      chalk.dim(job.id.slice(0, 8)),
      chalk.bold(job.filename.slice(0, 30)),
      job.collection,
      progressBar(job.progress),
      stateColor(job.state),
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
