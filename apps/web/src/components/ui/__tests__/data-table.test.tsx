import { test, expect, describe, afterEach } from "bun:test"
import { render, cleanup, fireEvent } from "@testing-library/react"
import { DataTable } from "@/components/ui/data-table"
import type { ColumnDef } from "@tanstack/react-table"

afterEach(cleanup)

type User = { email: string; role: string; status: string }

const columns: ColumnDef<User>[] = [
  { accessorKey: "email",  header: "Email" },
  { accessorKey: "role",   header: "Rol" },
  { accessorKey: "status", header: "Estado" },
]

const data: User[] = [
  { email: "admin@test.com",   role: "Admin",   status: "Activo" },
  { email: "user1@test.com",   role: "Usuario", status: "Activo" },
  { email: "user2@test.com",   role: "Usuario", status: "Inactivo" },
]

describe("<DataTable />", () => {
  test("renderiza las columnas correctamente", () => {
    const { getByText } = render(<DataTable columns={columns} data={data} />)
    expect(getByText("Email")).toBeInTheDocument()
    expect(getByText("Rol")).toBeInTheDocument()
    expect(getByText("Estado")).toBeInTheDocument()
  })

  test("renderiza todas las filas de datos", () => {
    const { getByText } = render(<DataTable columns={columns} data={data} />)
    expect(getByText("admin@test.com")).toBeInTheDocument()
    expect(getByText("user1@test.com")).toBeInTheDocument()
    expect(getByText("user2@test.com")).toBeInTheDocument()
  })

  test("sin datos muestra 'Sin resultados'", () => {
    const { getByText } = render(<DataTable columns={columns} data={[]} />)
    expect(getByText("Sin resultados.")).toBeInTheDocument()
  })

  test("con searchKey filtra por texto", () => {
    const { getByPlaceholderText, queryByText } = render(
      <DataTable columns={columns} data={data} searchKey="email" searchPlaceholder="Buscar email..." />
    )
    const input = getByPlaceholderText("Buscar email...")
    fireEvent.change(input, { target: { value: "admin" } })
    expect(queryByText("user1@test.com")).toBeNull()
  })

  test("sin datos no muestra paginación", () => {
    const { queryByText } = render(<DataTable columns={columns} data={[]} />)
    expect(queryByText("Página")).toBeNull()
  })

  test("con pageSize pequeño muestra paginación", () => {
    const manyUsers = Array.from({ length: 15 }, (_, i) => ({
      email: `user${i}@test.com`, role: "Usuario", status: "Activo",
    }))
    const { getByText } = render(<DataTable columns={columns} data={manyUsers} pageSize={5} />)
    expect(getByText(/Página/)).toBeInTheDocument()
  })
})
