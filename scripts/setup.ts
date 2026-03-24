#!/usr/bin/env bun
/**
 * RAG Saldivia — Setup Script (Fase 0)
 *
 * Onboarding cero-fricción: cualquiera clona el repo y con un comando tiene
 * el entorno listo para desarrollar o deployar.
 *
 * Usage:
 *   bun run setup           → setup completo
 *   bun run setup --check   → preflight check sin instalar nada
 *   bun run setup --reset   → limpia data/ y rehace migraciones + seed
 *   bun run setup --help    → muestra esta ayuda
 */

import { join } from "path"

// ── ANSI colors (sin dependencias externas) ────────────────────────────────
const c = {
  reset: "\x1b[0m",
  bold: (s: string) => `\x1b[1m${s}\x1b[0m`,
  dim: (s: string) => `\x1b[2m${s}\x1b[0m`,
  green: (s: string) => `\x1b[32m${s}\x1b[0m`,
  red: (s: string) => `\x1b[31m${s}\x1b[0m`,
  yellow: (s: string) => `\x1b[33m${s}\x1b[0m`,
  cyan: (s: string) => `\x1b[36m${s}\x1b[0m`,
  blue: (s: string) => `\x1b[34m${s}\x1b[0m`,
  magenta: (s: string) => `\x1b[35m${s}\x1b[0m`,
}

const icons = {
  ok: c.green("✓"),
  warn: c.yellow("⚠"),
  fail: c.red("✗"),
  info: c.cyan("ℹ"),
  arrow: c.dim("→"),
  bullet: c.dim("•"),
}

// ── Helpers ────────────────────────────────────────────────────────────────
const log = {
  ok: (msg: string, detail?: string) =>
    console.log(`  ${icons.ok} ${msg}${detail ? c.dim("  " + detail) : ""}`),
  warn: (msg: string, detail?: string) =>
    console.log(`  ${icons.warn} ${msg}${detail ? c.dim("\n      " + detail) : ""}`),
  fail: (msg: string, detail?: string) =>
    console.log(`  ${icons.fail} ${c.bold(msg)}${detail ? `\n      ${c.yellow("→")} ${detail}` : ""}`),
  info: (msg: string) => console.log(`  ${icons.info} ${c.dim(msg)}`),
  section: (msg: string) => console.log(`\n${c.bold(c.cyan(msg))}`),
  blank: () => console.log(""),
}

type StepResult = { ok: boolean; message: string; skipped?: boolean; suggestion?: string }

const ROOT = join(import.meta.dir, "..")

// ── Detección de plataforma ────────────────────────────────────────────────
const isWindows = process.platform === "win32"

// ── Ejecutar comando y capturar output ────────────────────────────────────
async function run(cmd: string[], silent = true): Promise<{ ok: boolean; stdout: string; stderr: string }> {
  try {
    const proc = Bun.spawn(cmd, {
      cwd: ROOT,
      stdout: "pipe",
      stderr: "pipe",
    })
    const [stdout, stderr, exitCode] = await Promise.all([
      new Response(proc.stdout).text(),
      new Response(proc.stderr).text(),
      proc.exited,
    ])
    return { ok: exitCode === 0, stdout: stdout.trim(), stderr: stderr.trim() }
  } catch {
    return { ok: false, stdout: "", stderr: "Command not found" }
  }
}

// ── Verificar si un puerto está en uso ────────────────────────────────────
async function isPortInUse(port: number): Promise<boolean> {
  try {
    const conn = await Bun.connect({
      hostname: "127.0.0.1",
      port,
      socket: { data() {}, open(socket) { socket.end() }, close() {}, error() {} },
    })
    conn.end()
    return true
  } catch {
    return false
  }
}

// ── PASOS DEL SETUP ────────────────────────────────────────────────────────

async function checkBun(): Promise<StepResult> {
  const MIN_BUN = [1, 0, 0]
  const res = await run(["bun", "--version"])
  if (!res.ok) {
    return {
      ok: false,
      message: "Bun no está instalado",
      suggestion: isWindows
        ? 'Instalá Bun: powershell -c "irm bun.sh/install.ps1 | iex"'
        : 'Instalá Bun: curl -fsSL https://bun.sh/install | bash',
    }
  }
  const version = res.stdout.replace(/^v/, "")
  const parts = version.split(".").map(Number)
  const tooOld = parts[0] < MIN_BUN[0]
  if (tooOld) {
    return {
      ok: false,
      message: `Bun ${version} instalado, se requiere >= ${MIN_BUN.join(".")}`,
      suggestion: "Actualizá Bun: bun upgrade",
    }
  }
  return { ok: true, message: `Bun ${version}` }
}

async function checkDocker(): Promise<StepResult> {
  const res = await run(["docker", "--version"])
  if (!res.ok) {
    return {
      ok: false,
      message: "Docker no está instalado o no está en el PATH",
      suggestion: "Instalá Docker Desktop: https://docs.docker.com/get-docker/\nEl sistema puede arrancar sin Docker usando MOCK_RAG=true para desarrollo de UI",
    }
  }
  const composeRes = await run(["docker", "compose", "version"])
  if (!composeRes.ok) {
    return {
      ok: false,
      message: "Docker está instalado pero 'docker compose' no está disponible",
      suggestion: "Actualizá Docker Desktop a una versión reciente (incluye Compose v2)",
    }
  }
  const version = res.stdout.replace("Docker version ", "").split(",")[0]
  return { ok: true, message: `Docker ${version}` }
}

async function checkPorts(): Promise<StepResult> {
  const requiredPorts = [
    { port: 3000, service: "Next.js (web)" },
    { port: 9000, service: "Gateway legacy" },
  ]
  const optionalPorts = [
    { port: 8081, service: "RAG Server" },
    { port: 19530, service: "Milvus" },
  ]

  const conflicts: string[] = []
  const warnings: string[] = []

  for (const { port, service } of requiredPorts) {
    if (await isPortInUse(port)) {
      conflicts.push(`${port} (${service})`)
    }
  }
  for (const { port, service } of optionalPorts) {
    if (await isPortInUse(port)) {
      warnings.push(`${port} (${service}) ya tiene algo corriendo — puede ser el sistema en producción`)
    }
  }

  if (warnings.length > 0) {
    log.info(`Puertos opcionales en uso: ${warnings.join(", ")}`)
  }

  if (conflicts.length > 0) {
    return {
      ok: false,
      message: `Puertos ocupados: ${conflicts.join(", ")}`,
      suggestion: isWindows
        ? `En PowerShell: Get-Process -Id (Get-NetTCPConnection -LocalPort ${conflicts[0]?.split(" ")[0]} -ErrorAction SilentlyContinue).OwningProcess`
        : `Buscá qué proceso usa el puerto: lsof -i :${conflicts[0]?.split(" ")[0]}`,
    }
  }

  return { ok: true, message: "Puertos 3000, 8081 disponibles" }
}

async function copyEnvFile(): Promise<StepResult> {
  const envLocal = join(ROOT, ".env.local")
  const envExample = join(ROOT, ".env.example")

  const localExists = await Bun.file(envLocal).exists()
  if (localExists) {
    return { ok: true, message: ".env.local ya existe", skipped: true }
  }

  const exampleExists = await Bun.file(envExample).exists()
  if (!exampleExists) {
    return {
      ok: false,
      message: ".env.example no encontrado",
      suggestion: "Verificá que el repositorio esté completo (git status)",
    }
  }

  const content = await Bun.file(envExample).text()
  await Bun.write(envLocal, content)

  return {
    ok: true,
    message: ".env.local creado desde .env.example",
    suggestion: "Editá .env.local y completá las variables marcadas como REQUIRED",
  }
}

async function installDependencies(): Promise<StepResult> {
  const pkgJson = join(ROOT, "package.json")
  const exists = await Bun.file(pkgJson).exists()

  if (!exists) {
    return {
      ok: true,
      message: "package.json raíz no encontrado — skipping (Fase 2 lo crea)",
      skipped: true,
    }
  }

  log.info("Instalando dependencias (bun install)...")
  const res = await run(["bun", "install"], false)
  if (!res.ok) {
    return {
      ok: false,
      message: "bun install falló",
      suggestion: `Error: ${res.stderr.slice(0, 200)}`,
    }
  }

  return { ok: true, message: "Dependencias instaladas" }
}

async function runMigrations(): Promise<StepResult> {
  const dbPkg = join(ROOT, "packages", "db", "package.json")
  const exists = await Bun.file(dbPkg).exists()

  if (!exists) {
    return {
      ok: true,
      message: "packages/db no encontrado — skipping (Fase 2 lo crea)",
      skipped: true,
    }
  }

  log.info("Corriendo migraciones de DB (bun run db:migrate)...")
  const res = await run(["bun", "run", "db:migrate"])
  if (!res.ok) {
    return {
      ok: false,
      message: "Migraciones fallaron",
      suggestion: `Error: ${res.stderr.slice(0, 200)}\nVerificá que DATABASE_PATH en .env.local tenga permisos de escritura`,
    }
  }

  return { ok: true, message: "Migraciones aplicadas" }
}

async function seedDatabase(): Promise<StepResult> {
  const dbPkg = join(ROOT, "packages", "db", "package.json")
  const exists = await Bun.file(dbPkg).exists()

  if (!exists) {
    return {
      ok: true,
      message: "packages/db no encontrado — skipping",
      skipped: true,
    }
  }

  log.info("Creando datos de desarrollo (bun run db:seed)...")
  const res = await run(["bun", "run", "db:seed"])
  if (!res.ok) {
    return {
      ok: false,
      message: "Seed falló",
      suggestion: res.stderr.slice(0, 200),
    }
  }

  return { ok: true, message: "Seed completado (usuario admin@localhost creado)" }
}

async function validateEnvVars(): Promise<StepResult> {
  const envLocal = join(ROOT, ".env.local")
  const exists = await Bun.file(envLocal).exists()
  if (!exists) {
    return { ok: true, message: ".env.local no existe — skipping", skipped: true }
  }

  const content = await Bun.file(envLocal).text()

  // Variables requeridas en producción — warn si están vacías o tienen placeholder
  const requiredInProd = ["JWT_SECRET", "SYSTEM_API_KEY"]
  const missing: string[] = []

  for (const varName of requiredInProd) {
    const line = content.split("\n").find((l) => l.startsWith(`${varName}=`))
    const value = line?.split("=").slice(1).join("=").trim()
    if (!value || value.includes("CHANGE_ME") || value.includes("YOUR_") || value === "") {
      missing.push(varName)
    }
  }

  if (missing.length > 0) {
    return {
      ok: true, // No falla el setup, solo avisa
      message: `Variables con placeholder en .env.local: ${missing.join(", ")}`,
      suggestion: "Completá estos valores antes de deployar a producción.\nPara desarrollo local los valores placeholder son suficientes.",
    }
  }

  return { ok: true, message: "Variables de entorno OK" }
}

// ── MODO RESET ─────────────────────────────────────────────────────────────
async function resetData(): Promise<void> {
  log.section("Reset — limpiando data/")
  const dataDir = join(ROOT, "data")
  const exists = await Bun.file(join(dataDir, "app.db")).exists()

  if (!exists) {
    log.info("data/app.db no existe, nada que resetear")
    return
  }

  await Bun.file(join(dataDir, "app.db")).delete?.()
  log.ok("data/app.db eliminado")
  log.info("Corriendo migraciones fresh...")
  await runMigrations()
  await seedDatabase()
}

// ── PRINT SUMMARY ──────────────────────────────────────────────────────────
function printSummary(results: Array<StepResult & { name: string }>, checkOnly: boolean): void {
  log.section("Resumen")

  const width = 50
  const divider = c.dim("─".repeat(width))

  console.log(divider)
  for (const r of results) {
    if (r.skipped) {
      console.log(`  ${c.dim("–")} ${c.dim(r.name.padEnd(28))} ${c.dim(r.message)}`)
    } else if (r.ok) {
      console.log(`  ${icons.ok} ${r.name.padEnd(28)} ${c.dim(r.message)}`)
    } else {
      console.log(`  ${icons.fail} ${c.bold(r.name.padEnd(28))} ${c.yellow(r.message)}`)
      if (r.suggestion) {
        for (const line of r.suggestion.split("\n")) {
          console.log(`      ${icons.arrow} ${c.yellow(line)}`)
        }
      }
    }
  }
  console.log(divider)

  const failed = results.filter((r) => !r.ok && !r.skipped)
  const warned = results.filter((r) => r.ok && r.suggestion && !r.skipped)

  if (failed.length > 0) {
    log.blank()
    console.log(`  ${icons.fail} ${c.bold(`${failed.length} problema(s) encontrado(s). Resolvelos antes de continuar.`)}`)
    log.blank()
    process.exit(1)
  }

  log.blank()
  if (checkOnly) {
    console.log(`  ${icons.ok} ${c.bold("Preflight check OK — el entorno está listo.")}`)
  } else {
    console.log(`  ${icons.ok} ${c.bold("Setup completo.")}`)
    log.blank()
    console.log(`  ${c.cyan("Próximos pasos:")}`)
    console.log(`    ${icons.bullet} ${c.dim("Editá")} .env.local ${c.dim("con tus credenciales")}`)
    console.log(`    ${icons.bullet} ${c.dim("Arrancá el servidor:")} ${c.cyan("bun run dev")}`)
    console.log(`    ${icons.bullet} ${c.dim("Abrí")} ${c.cyan("http://localhost:3000")}`)
    if (warned.length > 0) {
      console.log(`    ${icons.bullet} ${c.dim("Revisá los")} ${icons.warn} ${c.dim("warnings de arriba")}`)
    }
  }
  log.blank()
}

// ── BANNER ─────────────────────────────────────────────────────────────────
function printBanner(mode: string): void {
  console.log("")
  console.log(c.bold(c.cyan("  RAG Saldivia")))
  console.log(c.dim(`  Setup — ${mode}`))
  console.log("")
}

// ── AYUDA ──────────────────────────────────────────────────────────────────
function printHelp(): void {
  console.log(`
${c.bold("RAG Saldivia Setup")}

${c.bold("Uso:")}
  bun run setup              Setup completo (instalar + migrar + seed)
  bun run setup --check      Solo verifica prerequisitos, sin instalar nada
  bun run setup --reset      Limpia la DB y rehace migraciones + seed
  bun run setup --help       Muestra esta ayuda

${c.bold("Qué hace:")}
  ${icons.bullet} Verifica Bun, Docker y puertos disponibles
  ${icons.bullet} Crea .env.local desde .env.example (si no existe)
  ${icons.bullet} Instala dependencias con bun install
  ${icons.bullet} Corre migraciones de base de datos
  ${icons.bullet} Crea datos de desarrollo (usuario admin@localhost)
  ${icons.bullet} Muestra resumen de estado

${c.bold("Variables de entorno importantes:")}
  JWT_SECRET          Clave para firmar tokens JWT (requerida en prod)
  SYSTEM_API_KEY      Clave interna para CLI y service-to-service
  RAG_SERVER_URL      URL del servidor RAG de NVIDIA (default: http://localhost:8081)
  DATABASE_PATH       Ruta al archivo SQLite (default: ./data/app.db)
  MOCK_RAG            true/false — modo sin RAG real, para desarrollo de UI
  NGC_API_KEY         API key de NVIDIA (requerida para el blueprint)
`)
}

// ── MAIN ───────────────────────────────────────────────────────────────────
async function main(): Promise<void> {
  const args = process.argv.slice(2)

  if (args.includes("--help") || args.includes("-h")) {
    printHelp()
    process.exit(0)
  }

  const checkOnly = args.includes("--check")
  const reset = args.includes("--reset")

  if (reset) {
    printBanner("Reset")
    await resetData()
    return
  }

  printBanner(checkOnly ? "Preflight Check" : "Setup Completo")

  // ── Paso 1: Prerequisitos ────────────────────────────────────────────────
  log.section("Prerequisitos")

  const bunResult = await checkBun()
  if (bunResult.ok) log.ok(bunResult.message)
  else log.fail(bunResult.message, bunResult.suggestion)

  const dockerResult = await checkDocker()
  if (dockerResult.ok) log.ok(dockerResult.message)
  else log.warn(dockerResult.message, dockerResult.suggestion)

  const portsResult = await checkPorts()
  if (portsResult.ok) log.ok(portsResult.message)
  else log.fail(portsResult.message, portsResult.suggestion)

  if (checkOnly) {
    printSummary(
      [
        { name: "Bun", ...bunResult },
        { name: "Docker", ...dockerResult },
        { name: "Puertos", ...portsResult },
      ],
      true
    )
    return
  }

  // ── Paso 2: Configuración ────────────────────────────────────────────────
  log.section("Configuración")

  const envResult = await copyEnvFile()
  if (envResult.ok && !envResult.skipped) {
    log.ok(envResult.message)
    if (envResult.suggestion) log.warn(envResult.suggestion)
  } else if (envResult.skipped) {
    log.ok(envResult.message)
  } else {
    log.fail(envResult.message, envResult.suggestion)
  }

  const envVarsResult = await validateEnvVars()
  if (envVarsResult.ok && envVarsResult.suggestion) {
    log.warn(envVarsResult.message, envVarsResult.suggestion)
  } else if (envVarsResult.ok) {
    log.ok(envVarsResult.message)
  } else {
    log.fail(envVarsResult.message, envVarsResult.suggestion)
  }

  // ── Paso 3: Dependencias ─────────────────────────────────────────────────
  log.section("Dependencias")

  const depsResult = await installDependencies()
  if (depsResult.skipped) {
    log.info(depsResult.message)
  } else if (depsResult.ok) {
    log.ok(depsResult.message)
  } else {
    log.fail(depsResult.message, depsResult.suggestion)
  }

  // ── Paso 4: Base de datos ────────────────────────────────────────────────
  log.section("Base de datos")

  const migrationsResult = await runMigrations()
  if (migrationsResult.skipped) {
    log.info(migrationsResult.message)
  } else if (migrationsResult.ok) {
    log.ok(migrationsResult.message)
  } else {
    log.fail(migrationsResult.message, migrationsResult.suggestion)
  }

  const seedResult = await seedDatabase()
  if (seedResult.skipped) {
    log.info(seedResult.message)
  } else if (seedResult.ok) {
    log.ok(seedResult.message)
  } else {
    log.fail(seedResult.message, seedResult.suggestion)
  }

  // ── Summary ──────────────────────────────────────────────────────────────
  printSummary(
    [
      { name: "Bun", ...bunResult },
      { name: "Docker", ...dockerResult },
      { name: "Puertos", ...portsResult },
      { name: ".env.local", ...envResult },
      { name: "Variables de entorno", ...envVarsResult },
      { name: "Dependencias", ...depsResult },
      { name: "Migraciones DB", ...migrationsResult },
      { name: "Seed", ...seedResult },
    ],
    false
  )
}

main().catch((err) => {
  console.error(`\n  ${icons.fail} ${c.bold("Error inesperado en el setup:")}`)
  console.error(`      ${c.yellow(String(err))}`)
  console.error(`\n  Si el error persiste, abrí un issue en:`)
  console.error(`  ${c.cyan("https://github.com/Camionerou/rag-saldivia/issues")}`)
  process.exit(1)
})
