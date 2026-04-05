import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render, fireEvent } from "@testing-library/react"
import { DirectMessageDialog } from "../DirectMessageDialog"

afterEach(cleanup)

const USERS = [
  { id: 1, name: "Admin", email: "admin@test.com" },
  { id: 2, name: "María", email: "maria@test.com" },
  { id: 3, name: "Juan", email: "juan@test.com" },
]

describe("DirectMessageDialog", () => {
  test("renders nothing when closed", () => {
    const { container } = render(
      <DirectMessageDialog open={false} onClose={() => {}} users={USERS} currentUserId={1} />
    )
    expect(container.innerHTML).toBe("")
  })

  test("renders dialog when open", () => {
    const { getByText } = render(
      <DirectMessageDialog open={true} onClose={() => {}} users={USERS} currentUserId={1} />
    )
    expect(getByText("Mensaje directo")).toBeTruthy()
  })

  test("lists other users (not current user)", () => {
    const { getByText, queryByText } = render(
      <DirectMessageDialog open={true} onClose={() => {}} users={USERS} currentUserId={1} />
    )
    expect(getByText("María")).toBeTruthy()
    expect(getByText("Juan")).toBeTruthy()
    // Current user should not appear in the list
    expect(queryByText("admin@test.com")).toBeNull()
  })

  test("search filters users", () => {
    const { getByPlaceholderText, getByText, queryByText } = render(
      <DirectMessageDialog open={true} onClose={() => {}} users={USERS} currentUserId={1} />
    )
    fireEvent.change(getByPlaceholderText("Buscar usuarios..."), { target: { value: "mar" } })
    expect(getByText("María")).toBeTruthy()
    expect(queryByText("Juan")).toBeNull()
  })

  test("shows selected count", () => {
    const { getByText } = render(
      <DirectMessageDialog open={true} onClose={() => {}} users={USERS} currentUserId={1} />
    )
    expect(getByText("0 seleccionados")).toBeTruthy()
    fireEvent.click(getByText("María"))
    expect(getByText("1 seleccionado")).toBeTruthy()
  })

  test("start button disabled when no selection", () => {
    const { getByText } = render(
      <DirectMessageDialog open={true} onClose={() => {}} users={USERS} currentUserId={1} />
    )
    const btn = getByText("Iniciar chat")
    expect(btn.closest("button")?.disabled).toBe(true)
  })
})
