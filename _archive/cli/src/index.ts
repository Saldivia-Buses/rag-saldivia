#!/usr/bin/env bun
/**
 * RAG Saldivia CLI
 *
 * Interfaz de administración del sistema desde la terminal.
 * Habla con el servidor Next.js (apps/web) via API REST.
 *
 * Uso:
 *   rag <comando> [opciones]
 *   rag --help
 *   rag          (modo interactivo REPL)
 *
 * Para instalar globalmente:
 *   bun link (en el directorio apps/cli)
 */

import { Command } from "commander"
import * as p from "@clack/prompts"
import chalk from "chalk"
import { banner, out } from "./output.js"
import { statusCommand } from "./commands/status.js"
import { usersListCommand, usersCreateCommand, usersDeleteCommand } from "./commands/users.js"
import { collectionsListCommand, collectionsCreateCommand, collectionsDeleteCommand } from "./commands/collections.js"
import { ingestStartCommand, ingestStatusCommand, ingestCancelCommand } from "./commands/ingest.js"
import { auditLogCommand, auditReplayCommand, auditExportCommand } from "./commands/audit.js"
import { configGetCommand, configSetCommand, configResetCommand } from "./commands/config.js"
import { api } from "./client.js"

const program = new Command()

// ── Configuración global ───────────────────────────────────────────────────

program
  .name("rag")
  .description("RAG Saldivia — CLI de administración")
  .version("0.1.0")
  .option("--server <url>", "URL del servidor (default: http://localhost:3000)")
  .hook("preAction", (cmd) => {
    const serverOpt = cmd.opts()["server"] as string | undefined
    if (serverOpt) {
      process.env["RAG_WEB_URL"] = serverOpt
    }
  })

// ── rag status ─────────────────────────────────────────────────────────────

program
  .command("status")
  .description("Health check de todos los servicios con latencias")
  .action(statusCommand)

// ── rag users ──────────────────────────────────────────────────────────────

const users = program.command("users").description("Gestión de usuarios")

users
  .command("list")
  .description("Lista todos los usuarios con roles y áreas")
  .action(usersListCommand)

users
  .command("create")
  .description("Crear nuevo usuario (wizard interactivo)")
  .action(usersCreateCommand)

users
  .command("delete <id>")
  .description("Eliminar usuario por ID")
  .action((id) => usersDeleteCommand(parseInt(id)))

// ── rag collections ────────────────────────────────────────────────────────

const collections = program.command("collections").description("Gestión de colecciones")

collections
  .command("list")
  .description("Lista colecciones disponibles")
  .action(collectionsListCommand)

collections
  .command("create [name]")
  .description("Crear nueva colección")
  .action(collectionsCreateCommand)

collections
  .command("delete [name]")
  .description("Eliminar colección y todos sus documentos")
  .action(collectionsDeleteCommand)

// ── rag ingest ─────────────────────────────────────────────────────────────

const ingest = program.command("ingest").description("Gestión de ingesta de documentos")

ingest
  .command("start")
  .description("Iniciar ingesta de documentos")
  .option("-c, --collection <name>", "Colección destino")
  .option("-p, --path <path>", "Ruta del documento o directorio")
  .action((opts) => ingestStartCommand(opts))

ingest
  .command("status")
  .description("Ver estado de jobs activos con progreso")
  .action(ingestStatusCommand)

ingest
  .command("cancel <jobId>")
  .description("Cancelar un job de ingesta")
  .action(ingestCancelCommand)

// ── rag config ─────────────────────────────────────────────────────────────

const config = program.command("config").description("Configuración del RAG")

config
  .command("get [key]")
  .description("Mostrar configuración actual (o un parámetro específico)")
  .action(configGetCommand)

config
  .command("set <key> <value>")
  .description("Cambiar un parámetro de configuración")
  .action(configSetCommand)

config
  .command("reset")
  .description("Resetear configuración a valores por defecto")
  .action(configResetCommand)

// ── rag audit ──────────────────────────────────────────────────────────────

const audit = program.command("audit").description("Audit log y black box")

audit
  .command("log")
  .description("Ver eventos del sistema")
  .option("-n, --limit <n>", "Número de eventos a mostrar", "50")
  .option("-l, --level <level>", "Filtrar por nivel (INFO, WARN, ERROR...)")
  .option("-t, --type <type>", "Filtrar por tipo de evento")
  .action((opts) =>
    auditLogCommand({
      limit: opts.limit ? parseInt(opts.limit) : undefined,
      level: opts.level,
      type: opts.type,
    })
  )

audit
  .command("replay <date>")
  .description("Reconstruir estado del sistema desde una fecha (YYYY-MM-DD)")
  .action(auditReplayCommand)

audit
  .command("export")
  .description("Exportar todos los eventos como JSON")
  .action(auditExportCommand)

// ── rag db ─────────────────────────────────────────────────────────────────

const db = program.command("db").description("Administración de la base de datos")

db
  .command("migrate")
  .description("Correr migraciones pendientes")
  .action(async () => {
    const spinner = p.spinner()
    spinner.start("Corriendo migraciones...")
    const r = await api.db.migrate()
    r.ok ? spinner.stop(chalk.green("Migraciones aplicadas")) : spinner.stop(chalk.red(r.error))
  })

db
  .command("seed")
  .description("Crear datos de desarrollo (usuario admin@localhost)")
  .action(async () => {
    const spinner = p.spinner()
    spinner.start("Corriendo seed...")
    const r = await api.db.seed()
    r.ok ? spinner.stop(chalk.green("Seed completado")) : spinner.stop(chalk.red(r.error))
  })

db
  .command("reset")
  .description("Limpiar la DB y rehacer migraciones + seed")
  .action(async () => {
    const confirmed = await p.confirm({
      message: "¿BORRAR TODA LA DB y rehacer desde cero?",
      initialValue: false,
    })
    if (!confirmed || p.isCancel(confirmed)) { out.info("Cancelado"); return }
    const spinner = p.spinner()
    spinner.start("Reseteando DB...")
    const r = await api.db.reset()
    r.ok ? spinner.stop(chalk.green("DB reseteada")) : spinner.stop(chalk.red(r.error))
  })

// ── rag setup ──────────────────────────────────────────────────────────────

program
  .command("setup")
  .description("Setup completo del sistema (equivalente a bun run setup)")
  .action(async () => {
    banner("Setup")
    const proc = Bun.spawn(["bun", "scripts/setup.ts"], {
      cwd: process.cwd(),
      stdout: "inherit",
      stderr: "inherit",
    })
    await proc.exited
  })

// ── Modo REPL interactivo (sin argumentos) ─────────────────────────────────

async function interactiveMode() {
  banner("Modo interactivo")

  const COMMANDS = [
    { value: "status", label: "rag status — Estado del sistema" },
    { value: "users list", label: "rag users list — Lista usuarios" },
    { value: "users create", label: "rag users create — Crear usuario" },
    { value: "collections list", label: "rag collections list — Colecciones" },
    { value: "ingest status", label: "rag ingest status — Estado de ingesta" },
    { value: "audit log", label: "rag audit log — Ver eventos" },
    { value: "config get", label: "rag config get — Configuración RAG" },
    { value: "exit", label: "Salir" },
  ]

  while (true) {
    const choice = await p.select({
      message: "¿Qué querés hacer?",
      options: COMMANDS,
    })

    if (p.isCancel(choice) || choice === "exit") {
      p.outro("¡Hasta luego!")
      process.exit(0)
    }

    // Re-ejecutar con el comando elegido
    const args = (choice as string).split(" ")
    await program.parseAsync(["node", "rag", ...args])
    console.log("")
  }
}

// ── Main ───────────────────────────────────────────────────────────────────

const args = process.argv.slice(2)

if (args.length === 0) {
  // Sin argumentos → modo REPL interactivo
  interactiveMode().catch((err) => {
    out.error("Error inesperado", String(err))
    process.exit(1)
  })
} else {
  program.parseAsync(process.argv).catch((err) => {
    out.error("Error", String(err))
    process.exit(1)
  })
}
