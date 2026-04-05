import { describe, test, expect, afterEach, mock } from "bun:test"
import { cleanup, render, fireEvent } from "@testing-library/react"
import { ReactionPicker } from "../ReactionPicker"

afterEach(cleanup)

describe("ReactionPicker", () => {
  test("renders trigger button", () => {
    const { getByTitle } = render(<ReactionPicker onSelect={() => {}} />)
    expect(getByTitle("Reaccionar")).toBeTruthy()
  })

  test("opens emoji grid on click", () => {
    const { getByTitle, getByText } = render(<ReactionPicker onSelect={() => {}} />)
    fireEvent.click(getByTitle("Reaccionar"))
    expect(getByText("👍")).toBeTruthy()
  })

  test("calls onSelect with emoji", () => {
    const onSelect = mock(() => {})
    const { getByTitle, getByText } = render(<ReactionPicker onSelect={onSelect} />)
    fireEvent.click(getByTitle("Reaccionar"))
    fireEvent.click(getByText("🔥"))
    expect(onSelect).toHaveBeenCalledWith("🔥")
  })
})
