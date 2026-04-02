import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup, fireEvent } from "@testing-library/react"
import { CollectionSelector } from "@/components/chat/CollectionSelector"

afterEach(() => {
  cleanup()
  localStorage.clear()
})

const defaultProps = {
  defaultCollection: "contratos",
  availableCollections: ["contratos", "politicas", "manuales"],
  onCollectionsChange: mock(() => {}),
}

function renderSelector(overrides: Partial<typeof defaultProps> = {}) {
  const props = { ...defaultProps, onCollectionsChange: mock(() => {}), ...overrides }
  return { ...render(<CollectionSelector {...props} />), props }
}

describe("<CollectionSelector />", () => {
  test("renders with default collection label", () => {
    const { getByText } = renderSelector()
    expect(getByText("contratos")).toBeInTheDocument()
  })

  test("shows number of collections when multiple selected", () => {
    // Pre-seed localStorage with two collections
    localStorage.setItem(
      "rag-selected-collections",
      JSON.stringify(["contratos", "politicas"])
    )
    const { getByText } = renderSelector()
    // useLocalStorage reads from localStorage in a useEffect, so we need to
    // wait for the component to re-render after the effect fires.
    // happy-dom runs effects synchronously on render, so the value should
    // be available immediately.
    expect(getByText("2 colecciones")).toBeInTheDocument()
  })

  test("opens dropdown on click", () => {
    const { getByText, queryByText } = renderSelector()
    // Dropdown heading should not be visible before click
    expect(queryByText("Colecciones activas")).toBeNull()

    fireEvent.click(getByText("contratos"))

    expect(getByText("Colecciones activas")).toBeInTheDocument()
  })

  test("shows available collections in dropdown", () => {
    const { getByText, getAllByRole } = renderSelector()

    fireEvent.click(getByText("contratos"))

    // All three collections should appear as buttons inside the dropdown
    expect(getByText("politicas")).toBeInTheDocument()
    expect(getByText("manuales")).toBeInTheDocument()
    // "contratos" appears both in trigger label and in the dropdown list
    // There should be at least 4 buttons total: trigger + 3 collection toggles
    const buttons = getAllByRole("button")
    expect(buttons.length).toBeGreaterThanOrEqual(4)
  })

  test("toggle adds collection to selection", () => {
    const { getByText, props } = renderSelector()

    // Open dropdown
    fireEvent.click(getByText("contratos"))
    // Select "politicas"
    fireEvent.click(getByText("politicas"))

    expect(props.onCollectionsChange).toHaveBeenCalledWith(["contratos", "politicas"])
  })

  test("toggle removes collection from selection", () => {
    // Start with two selected
    localStorage.setItem(
      "rag-selected-collections",
      JSON.stringify(["contratos", "politicas"])
    )
    const { getByText, props } = renderSelector()

    // Open dropdown
    fireEvent.click(getByText("2 colecciones"))
    // Deselect "politicas"
    fireEvent.click(getByText("politicas"))

    expect(props.onCollectionsChange).toHaveBeenCalledWith(["contratos"])
  })

  test("cannot deselect last collection — falls back to default", () => {
    // Only "contratos" is selected (the default)
    const { getByText, getAllByText, props } = renderSelector()

    // Open dropdown
    fireEvent.click(getByText("contratos"))
    // After opening, "contratos" appears in both trigger and dropdown list.
    // The dropdown toggle button is the second match.
    const contratosElements = getAllByText("contratos")
    const dropdownToggle = contratosElements[contratosElements.length - 1]!
    fireEvent.click(dropdownToggle)

    // Should fall back to defaultCollection since we tried to deselect the only one
    expect(props.onCollectionsChange).toHaveBeenCalledWith(["contratos"])
  })

  test("calls onCollectionsChange when selection changes", () => {
    const onChange = mock(() => {})
    const { getByText } = renderSelector({ onCollectionsChange: onChange })

    fireEvent.click(getByText("contratos"))
    fireEvent.click(getByText("manuales"))

    expect(onChange).toHaveBeenCalledTimes(1)
    expect(onChange).toHaveBeenCalledWith(["contratos", "manuales"])
  })

  test("shows empty message when no collections available", () => {
    const { getByText } = renderSelector({ availableCollections: [] })

    fireEvent.click(getByText("contratos"))

    expect(getByText("Sin colecciones")).toBeInTheDocument()
  })

  test("disabled prop prevents opening dropdown", () => {
    const { getByRole, queryByText } = renderSelector({ disabled: true })
    const trigger = getByRole("button")

    expect(trigger).toBeDisabled()
    fireEvent.click(trigger)
    expect(queryByText("Colecciones activas")).toBeNull()
  })
})
