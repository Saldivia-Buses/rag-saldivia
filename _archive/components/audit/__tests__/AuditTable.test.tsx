import { test, expect, describe, afterEach } from "bun:test"
import { render, cleanup, fireEvent } from "@testing-library/react"
import { NuqsTestingAdapter } from "nuqs/adapters/testing"
import { AuditTable } from "@/components/audit/AuditTable"

afterEach(cleanup)

const events = [
  { id: 1, type: "auth.login", level: "INFO", userId: 1, payload: { ip: "127.0.0.1" }, ts: Date.now(), sessionId: null },
  { id: 2, type: "user.created", level: "INFO", userId: 1, payload: { email: "new@test.com" }, ts: Date.now(), sessionId: null },
  { id: 3, type: "system.error", level: "ERROR", userId: null, payload: { msg: "fail" }, ts: Date.now(), sessionId: null },
]

function renderTable(props: { events: typeof events; isAdmin: boolean }) {
  return render(
    <NuqsTestingAdapter>
      <AuditTable {...props} />
    </NuqsTestingAdapter>
  )
}

describe("<AuditTable />", () => {
  test("renderiza los tipos de eventos", () => {
    const { getByText } = renderTable({ events, isAdmin: true })
    expect(getByText("auth.login")).toBeInTheDocument()
    expect(getByText("user.created")).toBeInTheDocument()
  })

  test("badge ERROR está presente en la tabla", () => {
    const { getAllByText } = renderTable({ events, isAdmin: true })
    const errorBadges = getAllByText("ERROR")
    expect(errorBadges.length).toBeGreaterThan(0)
  })

  test("badge INFO está presente en la tabla", () => {
    const { getAllByText } = renderTable({ events, isAdmin: true })
    const infoBadges = getAllByText("INFO")
    expect(infoBadges.length).toBeGreaterThan(0)
  })

  test("admin ve columna Usuario", () => {
    const { getByText } = renderTable({ events, isAdmin: true })
    expect(getByText("Usuario")).toBeInTheDocument()
  })

  test("no admin no ve columna Usuario", () => {
    const { queryByText } = renderTable({ events, isAdmin: false })
    expect(queryByText("Usuario")).toBeNull()
  })

  test("filtro de búsqueda filtra por tipo", () => {
    const { getByPlaceholderText, queryByText } = renderTable({ events, isAdmin: true })
    fireEvent.change(getByPlaceholderText(/Buscar/), { target: { value: "auth" } })
    expect(queryByText("user.created")).toBeNull()
  })

  test("sin eventos muestra mensaje vacío", () => {
    const { getByText } = renderTable({ events: [], isAdmin: true })
    expect(getByText("No hay eventos que coincidan")).toBeInTheDocument()
  })

  test("muestra conteo de eventos", () => {
    const { getByText } = renderTable({ events, isAdmin: true })
    expect(getByText(/3 de 3/)).toBeInTheDocument()
  })
})
