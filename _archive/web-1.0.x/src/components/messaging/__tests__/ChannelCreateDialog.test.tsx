import { describe, test, expect, afterEach, mock } from "bun:test"
import { cleanup, render, fireEvent } from "@testing-library/react"
import { ChannelCreateDialog } from "../ChannelCreateDialog"

afterEach(cleanup)

describe("ChannelCreateDialog", () => {
  test("renders nothing when closed", () => {
    const { container } = render(<ChannelCreateDialog open={false} onClose={() => {}} />)
    expect(container.innerHTML).toBe("")
  })

  test("renders dialog when open", () => {
    const { getByRole } = render(<ChannelCreateDialog open={true} onClose={() => {}} />)
    expect(getByRole("heading", { level: 2 })).toBeTruthy()
  })

  test("renders type selector with public/private", () => {
    const { getByText } = render(<ChannelCreateDialog open={true} onClose={() => {}} />)
    expect(getByText("Público")).toBeTruthy()
    expect(getByText("Privado")).toBeTruthy()
  })

  test("renders name and description inputs", () => {
    const { getByPlaceholderText } = render(<ChannelCreateDialog open={true} onClose={() => {}} />)
    expect(getByPlaceholderText("ej. proyecto-flota")).toBeTruthy()
    expect(getByPlaceholderText("De qué se trata este canal...")).toBeTruthy()
  })

  test("create button disabled when name is empty", () => {
    const { getAllByText } = render(<ChannelCreateDialog open={true} onClose={() => {}} />)
    const buttons = getAllByText("Crear canal").filter((el) => el.tagName === "BUTTON")
    expect(buttons[0]!.closest("button")?.disabled).toBe(true)
  })

  test("calls onClose when cancel clicked", () => {
    const onClose = mock(() => {})
    const { getByText } = render(<ChannelCreateDialog open={true} onClose={onClose} />)
    fireEvent.click(getByText("Cancelar"))
    expect(onClose).toHaveBeenCalled()
  })
})
