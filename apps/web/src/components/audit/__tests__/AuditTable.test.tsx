import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup, fireEvent } from "@testing-library/react"
import { AuditTable } from "@/components/audit/AuditTable"

afterEach(cleanup)

const events = [
  { id: 1, type: "auth.login", level: "INFO", userId: 1, payload: { ip: "127.0.0.1" }, ts: Date.now(), sessionId: null },
  { id: 2, type: "user.created", level: "INFO", userId: 1, payload: { email: "new@test.com" }, ts: Date.now(), sessionId: null },
  { id: 3, type: "system.error", level: "ERROR", userId: null, payload: { msg: "fail" }, ts: Date.now(), sessionId: null },
]

describe("<AuditTable />", () => {
  test("renderiza los tipos de eventos", () => {
    const { getByText } = render(<AuditTable events={events} isAdmin />)
    expect(getByText("auth.login")).toBeInTheDocument()
    expect(getByText("user.created")).toBeInTheDocument()
  })

  test("badge ERROR está presente en la tabla", () => {
    const { getAllByText } = render(<AuditTable events={events} isAdmin />)
    const errorBadges = getAllByText("ERROR")
    expect(errorBadges.length).toBeGreaterThan(0)
  })

  test("badge INFO está presente en la tabla", () => {
    const { getAllByText } = render(<AuditTable events={events} isAdmin />)
    const infoBadges = getAllByText("INFO")
    expect(infoBadges.length).toBeGreaterThan(0)
  })

  test("admin ve columna Usuario", () => {
    const { getByText } = render(<AuditTable events={events} isAdmin />)
    expect(getByText("Usuario")).toBeInTheDocument()
  })

  test("no admin no ve columna Usuario", () => {
    const { queryByText } = render(<AuditTable events={events} isAdmin={false} />)
    expect(queryByText("Usuario")).toBeNull()
  })

  test("filtro de búsqueda filtra por tipo", () => {
    const { getByPlaceholderText, queryByText } = render(<AuditTable events={events} isAdmin />)
    fireEvent.change(getByPlaceholderText(/Buscar/), { target: { value: "auth" } })
    expect(queryByText("user.created")).toBeNull()
  })

  test("sin eventos muestra mensaje vacío", () => {
    const { getByText } = render(<AuditTable events={[]} isAdmin />)
    expect(getByText("No hay eventos que coincidan")).toBeInTheDocument()
  })

  test("muestra conteo de eventos", () => {
    const { getByText } = render(<AuditTable events={events} isAdmin />)
    expect(getByText(/3 de 3/)).toBeInTheDocument()
  })
})
