import * as p from "@clack/prompts"
import chalk from "chalk"
import { api } from "../client.js"
import { out, makeTable, handleApiError } from "../output.js"

export async function collectionsListCommand() {
  out.section("Colecciones")

  const result = await api.collections.list()
  if (!result.ok) return handleApiError(result)

  const collections = result.data as string[]

  if (collections.length === 0) {
    out.info("No hay colecciones disponibles")
    return
  }

  const rows = collections.map((name) => [chalk.bold(name), "—", "—"])
  console.log(makeTable(["Nombre", "Documentos", "Última ingesta"], rows))
  console.log(chalk.dim(`\n  Total: ${collections.length} colección(es)\n`))
}

export async function collectionsCreateCommand(name?: string) {
  const collectionName = name ?? String(await p.text({
    message: "Nombre de la colección",
    validate: (v) => !v ? "Requerido" : undefined,
  }))

  if (p.isCancel(collectionName)) {
    p.cancel("Cancelado")
    return
  }

  const spinner = p.spinner()
  spinner.start("Creando colección...")

  const result = await api.collections.create(collectionName)
  if (!result.ok) {
    spinner.stop(chalk.red("Error"))
    return handleApiError(result)
  }

  spinner.stop(chalk.green("Colección creada"))
  out.ok(`Colección ${chalk.bold(collectionName)} creada`)
}

export async function collectionsDeleteCommand(name?: string) {
  const collectionName = name ?? String(await p.text({
    message: "Nombre de la colección a eliminar",
  }))

  const confirmed = await p.confirm({
    message: `¿Eliminar la colección '${collectionName}' y todos sus documentos?`,
    initialValue: false,
  })

  if (!confirmed || p.isCancel(confirmed)) {
    out.info("Cancelado")
    return
  }

  const result = await api.collections.delete(collectionName)
  if (!result.ok) return handleApiError(result)

  out.ok(`Colección ${chalk.bold(collectionName)} eliminada`)
}
