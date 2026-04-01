import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render, fireEvent } from "@testing-library/react"
import { ChatLayout, SidebarProvider, useSidebar } from "../ChatLayout"

afterEach(cleanup)

describe("ChatLayout", () => {
  test("renders children", () => {
    const { getByText } = render(<ChatLayout><p>Chat content</p></ChatLayout>)
    expect(getByText("Chat content")).toBeTruthy()
  })

  test("renders as flex container", () => {
    const { container } = render(<ChatLayout><div /></ChatLayout>)
    expect(container.firstElementChild?.className).toContain("flex")
  })
})

describe("SidebarProvider", () => {
  test("provides default open=true", () => {
    function ReadSidebar() {
      const { open } = useSidebar()
      return <span>{open ? "open" : "closed"}</span>
    }
    const { getByText } = render(
      <SidebarProvider><ReadSidebar /></SidebarProvider>
    )
    expect(getByText("open")).toBeTruthy()
  })

  test("toggle changes open state", () => {
    function ToggleSidebar() {
      const { open, toggle } = useSidebar()
      return <button onClick={toggle}>{open ? "open" : "closed"}</button>
    }
    const { getByText } = render(
      <SidebarProvider><ToggleSidebar /></SidebarProvider>
    )
    expect(getByText("open")).toBeTruthy()
    fireEvent.click(getByText("open"))
    expect(getByText("closed")).toBeTruthy()
  })
})
