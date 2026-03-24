import * as p from "@clack/prompts"
import chalk from "chalk"
import { api } from "../client.js"
import { out, makeTable, handleApiError } from "../output.js"

export async function configGetCommand() {
  out.section("Configuración RAG")

  const result = await api.config.get()
  if (!result.ok) return handleApiError(result)

  const config = result.data as Record<string, unknown>

  const rows = Object.entries(config).map(([k, v]) => [
    chalk.bold(k),
    String(v),
  ])

  console.log(makeTable(["Parámetro", "Valor"], rows))
}

export async function configSetCommand(key: string, value: string) {
  // Intentar parsear como número o booleano
  let parsed: unknown = value
  if (value === "true") parsed = true
  else if (value === "false") parsed = false
  else if (!isNaN(Number(value))) parsed = Number(value)

  const result = await api.config.set(key, parsed)
  if (!result.ok) return handleApiError(result)

  out.ok(`${chalk.bold(key)} = ${chalk.cyan(String(parsed))}`)
}

export async function configResetCommand() {
  const confirmed = await p.confirm({
    message: "¿Resetear la configuración RAG a valores por defecto?",
    initialValue: false,
  })

  if (!confirmed || p.isCancel(confirmed)) {
    out.info("Cancelado")
    return
  }

  const result = await api.config.reset()
  if (!result.ok) return handleApiError(result)

  out.ok("Configuración reseteada a valores por defecto")
}
