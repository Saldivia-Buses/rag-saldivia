import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render, fireEvent } from "@testing-library/react"
import { SourcesPanel } from "../SourcesPanel"

afterEach(cleanup)

const SOURCES = [
  { document: "Manual de flota", content: "Procedimiento de mantenimiento...", score: 0.92 },
  { document: "Política RR.HH.", content: "Licencias y vacaciones...", score: 0.78 },
]

describe("SourcesPanel", () => {
  test("renders nothing when no sources", () => {
    const { container } = render(<SourcesPanel sources={[]} />)
    expect(container.innerHTML).toBe("")
  })

  test("renders source count", () => {
    const { getByText } = render(<SourcesPanel sources={SOURCES} />)
    expect(getByText("2 fuentes")).toBeTruthy()
  })

  test("renders singular for 1 source", () => {
    const { getByText } = render(<SourcesPanel sources={[SOURCES[0]!]} />)
    expect(getByText("1 fuente")).toBeTruthy()
  })

  test("renders document names", () => {
    const { getByText } = render(<SourcesPanel sources={SOURCES} />)
    expect(getByText("Manual de flota")).toBeTruthy()
    expect(getByText("Política RR.HH.")).toBeTruthy()
  })

  test("renders score as percentage", () => {
    const { getByText } = render(<SourcesPanel sources={SOURCES} />)
    expect(getByText("92%")).toBeTruthy()
    expect(getByText("78%")).toBeTruthy()
  })

  test("renders content snippets", () => {
    const { getByText } = render(<SourcesPanel sources={SOURCES} />)
    expect(getByText(/Procedimiento de mantenimiento/)).toBeTruthy()
  })

  test("collapses on toggle click", () => {
    const { getByText, queryByText } = render(<SourcesPanel sources={SOURCES} />)
    fireEvent.click(getByText("2 fuentes"))
    // After collapse, document names should be hidden
    expect(queryByText("Manual de flota")).toBeNull()
  })
})
