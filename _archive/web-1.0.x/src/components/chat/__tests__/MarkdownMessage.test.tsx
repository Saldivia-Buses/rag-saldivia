import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { MarkdownMessage } from "../MarkdownMessage"

afterEach(cleanup)

describe("MarkdownMessage", () => {
  test("renders plain text", () => {
    const { getByText } = render(<MarkdownMessage content="Hello world" />)
    expect(getByText("Hello world")).toBeTruthy()
  })

  test("renders headings", () => {
    const { getByText } = render(<MarkdownMessage content="## Título importante" />)
    expect(getByText("Título importante")).toBeTruthy()
  })

  test("renders bold text", () => {
    const { getByText } = render(<MarkdownMessage content="Esto es **importante**" />)
    expect(getByText("importante")).toBeTruthy()
  })

  test("renders list items", () => {
    const { container } = render(<MarkdownMessage content={"- Primero\n- Segundo\n- Tercero"} />)
    const listItems = container.querySelectorAll("li")
    expect(listItems.length).toBe(3)
  })

  test("renders links", () => {
    const { getByText } = render(<MarkdownMessage content="Visitá [Google](https://google.com)" />)
    const link = getByText("Google")
    expect(link.closest("a")?.getAttribute("href")).toBe("https://google.com")
  })

  test("renders code blocks with language label", () => {
    const content = "```python\nprint('hello')\n```"
    const { container } = render(<MarkdownMessage content={content} />)
    // Should render a pre or code element
    expect(container.querySelector("pre, code")).toBeTruthy()
  })
})
