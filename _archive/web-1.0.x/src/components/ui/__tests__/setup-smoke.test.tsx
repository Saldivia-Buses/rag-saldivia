import { test, expect } from "bun:test"
import { render, screen } from "@testing-library/react"

test("setup funciona: puede renderizar JSX con happy-dom", () => {
  render(<div data-testid="smoke">ok</div>)
  expect(screen.getByTestId("smoke")).toBeInTheDocument()
  expect(screen.getByTestId("smoke")).toHaveTextContent("ok")
})
