/**
 * Helpers de output para la CLI.
 * Colores, tablas, spinners y formateo.
 */

import chalk from "chalk"
import Table from "cli-table3"
import { getSuggestion } from "@rag-saldivia/logger/suggestions"

// ── Iconos y colores ───────────────────────────────────────────────────────
export const icons = {
  ok: chalk.green("✓"),
  warn: chalk.yellow("⚠"),
  fail: chalk.red("✗"),
  info: chalk.cyan("ℹ"),
  arrow: chalk.dim("→"),
  bullet: chalk.dim("•"),
}

// ── Log helpers ────────────────────────────────────────────────────────────
export const out = {
  ok: (msg: string) => console.log(`  ${icons.ok} ${msg}`),
  warn: (msg: string) => console.log(`  ${icons.warn} ${chalk.yellow(msg)}`),
  error: (msg: string, error?: string) => {
    console.error(`  ${icons.fail} ${chalk.bold(msg)}`)
    if (error) {
      const suggestion = getSuggestion(error)
      if (suggestion) {
        console.error(
          suggestion.split("\n").map((l) => `      ${chalk.yellow(l)}`).join("\n")
        )
      } else {
        console.error(`      ${chalk.dim(error)}`)
      }
    }
  },
  info: (msg: string) => console.log(`  ${icons.info} ${chalk.dim(msg)}`),
  section: (msg: string) => console.log(`\n${chalk.bold(chalk.cyan(msg))}`),
  blank: () => console.log(""),
}

// ── Tablas ─────────────────────────────────────────────────────────────────

export function makeTable(headers: string[], rows: string[][]): string {
  const table = new Table({
    head: headers.map((h) => chalk.bold(h)),
    style: { head: [], border: ["dim"] },
    chars: {
      top: "─", "top-mid": "┬", "top-left": "┌", "top-right": "┐",
      bottom: "─", "bottom-mid": "┴", "bottom-left": "└", "bottom-right": "┘",
      left: "│", "left-mid": "├", mid: "─", "mid-mid": "┼",
      right: "│", "right-mid": "┤", middle: "│",
    },
  })

  for (const row of rows) {
    table.push(row)
  }

  return table.toString()
}

// ── Progress bar ───────────────────────────────────────────────────────────

export function progressBar(percent: number, width = 20): string {
  const filled = Math.round((percent / 100) * width)
  const empty = width - filled
  return chalk.green("█".repeat(filled)) + chalk.dim("░".repeat(empty))
}

// ── Status semáforo ────────────────────────────────────────────────────────

export function statusBadge(ok: boolean, latency?: number): string {
  if (ok) {
    const lat = latency !== undefined ? chalk.dim(` (${latency}ms)`) : ""
    return `${icons.ok}${lat}`
  }
  return icons.fail
}

// ── Banner ─────────────────────────────────────────────────────────────────

export function banner(subtitle?: string) {
  console.log("")
  console.log(chalk.bold(chalk.cyan("  RAG Saldivia")))
  if (subtitle) console.log(chalk.dim(`  ${subtitle}`))
  console.log("")
}

// ── Error handler ──────────────────────────────────────────────────────────

export function handleApiError(result: { error: string; suggestion?: string }) {
  out.error(result.error)
  if (result.suggestion) {
    console.error(
      result.suggestion.split("\n").map((l) => `      ${chalk.yellow(icons.arrow)} ${l}`).join("\n")
    )
  }
  process.exit(1)
}
