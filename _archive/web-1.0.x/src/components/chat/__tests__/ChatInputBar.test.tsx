import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup, fireEvent } from "@testing-library/react"
import { ChatInputBar } from "@/components/chat/ChatInputBar"
import type { ComponentProps } from "react"

afterEach(cleanup)

type Props = ComponentProps<typeof ChatInputBar>

const defaultProps: Props = {
  value: "",
  onChange: mock(() => {}),
  onKeyDown: mock(() => {}),
  onSend: mock(() => {}),
  placeholder: "Escriba su consulta...",
  collection: "contratos",
}

function renderBar(overrides: Partial<Props> = {}) {
  const props: Props = {
    ...defaultProps,
    onChange: mock(() => {}),
    onKeyDown: mock(() => {}),
    onSend: mock(() => {}),
    ...overrides,
  }
  return { ...render(<ChatInputBar {...props} />), props }
}

describe("<ChatInputBar />", () => {
  test("renders textarea with placeholder", () => {
    const { getByPlaceholderText } = renderBar()
    const textarea = getByPlaceholderText("Escriba su consulta...")
    expect(textarea).toBeInTheDocument()
    expect(textarea.tagName).toBe("TEXTAREA")
  })

  test("send button disabled when value is empty", () => {
    const { container } = renderBar({ value: "" })
    // Send button is the last button in the bar (the one with SendIcon svg)
    const buttons = container.querySelectorAll("button")
    const sendButton = buttons[buttons.length - 1]!
    expect(sendButton.disabled).toBe(true)
  })

  test("calls onSend when send button clicked with non-empty value", () => {
    const { container, props } = renderBar({ value: "Hola mundo" })
    const buttons = container.querySelectorAll("button")
    const sendButton = buttons[buttons.length - 1]!
    expect(sendButton.disabled).toBe(false)

    fireEvent.click(sendButton)

    expect(props.onSend).toHaveBeenCalledTimes(1)
  })

  test("shows stop button instead of send when isStreaming with onStop", () => {
    const onStop = mock(() => {})
    const { getByTitle } = renderBar({
      isStreaming: true,
      onStop,
      value: "test",
    })

    // Stop button has title "Detener"
    expect(getByTitle("Detener")).toBeInTheDocument()

    fireEvent.click(getByTitle("Detener"))
    expect(onStop).toHaveBeenCalledTimes(1)
  })

  test("renders collectionSlot when provided instead of static label", () => {
    const { getByText, queryByText } = renderBar({
      collectionSlot: <span>Custom Slot</span>,
    })

    expect(getByText("Custom Slot")).toBeInTheDocument()
    // Static collection label should not be rendered
    expect(queryByText("contratos")).toBeNull()
  })

  test("textarea disabled when disabled prop is true", () => {
    const { getByPlaceholderText } = renderBar({ disabled: true })
    const textarea = getByPlaceholderText("Escriba su consulta...")
    expect(textarea).toBeDisabled()
  })
})
