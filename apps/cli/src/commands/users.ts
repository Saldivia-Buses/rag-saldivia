import * as p from "@clack/prompts"
import chalk from "chalk"
import { api } from "../client.js"
import { out, makeTable, handleApiError } from "../output.js"

export async function usersListCommand() {
  out.section("Usuarios")
  const result = await api.users.list()
  if (!result.ok) return handleApiError(result)

  const users = result.data as Array<{
    id: number
    name: string
    email: string
    role: string
    active: boolean
    userAreas?: Array<{ area: { name: string } }>
  }>

  if (users.length === 0) {
    out.info("No hay usuarios registrados")
    return
  }

  const rows = users.map((u) => [
    String(u.id),
    chalk.bold(u.name),
    u.email,
    u.role,
    u.userAreas?.map((ua) => ua.area.name).join(", ") || "—",
    u.active ? chalk.green("activo") : chalk.red("inactivo"),
  ])

  console.log(makeTable(["ID", "Nombre", "Email", "Rol", "Áreas", "Estado"], rows))
  console.log(chalk.dim(`\n  Total: ${users.length} usuario(s)\n`))
}

export async function usersCreateCommand() {
  out.section("Crear usuario")

  const answers = await p.group({
    name: () => p.text({ message: "Nombre completo", validate: (v) => !v ? "Requerido" : undefined }),
    email: () => p.text({ message: "Email", validate: (v) => !v.includes("@") ? "Email inválido" : undefined }),
    password: () => p.password({ message: "Contraseña (mín. 8 caracteres)", validate: (v) => v.length < 8 ? "Mínimo 8 caracteres" : undefined }),
    role: () => p.select({
      message: "Rol",
      options: [
        { value: "user", label: "Usuario" },
        { value: "area_manager", label: "Gerente de área" },
        { value: "admin", label: "Administrador" },
      ],
    }),
  }, {
    onCancel: () => {
      p.cancel("Cancelado")
      process.exit(0)
    },
  })

  const spinner = p.spinner()
  spinner.start("Creando usuario...")

  const result = await api.users.create(answers)
  if (!result.ok) {
    spinner.stop(chalk.red("Error"))
    return handleApiError(result)
  }

  spinner.stop(chalk.green("Usuario creado"))
  out.ok(`Usuario ${chalk.bold((answers as { email: string }).email)} creado exitosamente`)
}

export async function usersDeleteCommand(id: number) {
  const confirmed = await p.confirm({
    message: `¿Eliminar usuario ID ${id}? Esta acción no se puede deshacer.`,
    initialValue: false,
  })

  if (!confirmed || p.isCancel(confirmed)) {
    out.info("Cancelado")
    return
  }

  const result = await api.users.delete(id)
  if (!result.ok) return handleApiError(result)

  out.ok(`Usuario ${id} eliminado`)
}
