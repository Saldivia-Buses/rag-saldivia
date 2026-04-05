import { test, expect, describe, afterEach } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import {
  Table, TableBody, TableCell, TableHead,
  TableHeader, TableRow, TableCaption,
} from "@/components/ui/table"

afterEach(cleanup)

describe("<Table />", () => {
  test("renderiza datos en celdas", () => {
    const { getByText } = render(
      <Table>
        <TableHeader><TableRow><TableHead>Email</TableHead></TableRow></TableHeader>
        <TableBody>
          <TableRow><TableCell>admin@test.com</TableCell></TableRow>
          <TableRow><TableCell>user@test.com</TableCell></TableRow>
        </TableBody>
      </Table>
    )
    expect(getByText("admin@test.com")).toBeInTheDocument()
    expect(getByText("user@test.com")).toBeInTheDocument()
  })

  test("TableHeader tiene bg-surface", () => {
    const { container } = render(
      <Table>
        <TableHeader><TableRow><TableHead>Col</TableHead></TableRow></TableHeader>
        <TableBody />
      </Table>
    )
    expect(container.querySelector("thead")?.className).toContain("bg-surface")
  })

  test("TableRow tiene hover:bg-surface", () => {
    const { container } = render(
      <Table>
        <TableHeader><TableRow><TableHead>X</TableHead></TableRow></TableHeader>
        <TableBody><TableRow><TableCell>dato</TableCell></TableRow></TableBody>
      </Table>
    )
    const rows = container.querySelectorAll("tbody tr")
    expect(rows[0]?.className).toContain("hover:bg-surface")
  })

  test("TableCaption renderiza texto descriptivo", () => {
    const { getByText } = render(
      <Table>
        <TableCaption>Lista de usuarios</TableCaption>
        <TableBody />
      </Table>
    )
    expect(getByText("Lista de usuarios")).toBeInTheDocument()
  })

  test("TableHead tiene clase uppercase", () => {
    const { container } = render(
      <Table>
        <TableHeader><TableRow><TableHead>Email</TableHead></TableRow></TableHeader>
        <TableBody />
      </Table>
    )
    expect(container.querySelector("th")?.className).toContain("uppercase")
  })
})
