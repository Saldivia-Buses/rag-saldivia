import { describe, test, expect, afterEach, mock } from "bun:test"
import { cleanup, render, fireEvent } from "@testing-library/react"
import { MessageComposer } from "../MessageComposer"

afterEach(cleanup)

const MEMBERS = [{ id: 1, name: "Admin", email: "admin@test.com" }]

describe("MessageComposer", () => {
  test("renders textarea", () => {
    const { getByPlaceholderText } = render(
      <MessageComposer channelId="ch-1" currentUserId={1} members={MEMBERS} />
    )
    expect(getByPlaceholderText("Escribí un mensaje...")).toBeTruthy()
  })

  test("renders send button", () => {
    const { getByTitle } = render(
      <MessageComposer channelId="ch-1" currentUserId={1} members={MEMBERS} />
    )
    expect(getByTitle("Enviar (Enter)")).toBeTruthy()
  })

  test("send button disabled when empty", () => {
    const { getByTitle } = render(
      <MessageComposer channelId="ch-1" currentUserId={1} members={MEMBERS} />
    )
    expect(getByTitle("Enviar (Enter)").closest("button")?.disabled).toBe(true)
  })

  test("calls onOptimisticMessage on send", () => {
    const onOptimistic = mock(() => {})
    const { getByPlaceholderText, getByTitle } = render(
      <MessageComposer channelId="ch-1" currentUserId={1} members={MEMBERS} onOptimisticMessage={onOptimistic} />
    )
    fireEvent.change(getByPlaceholderText("Escribí un mensaje..."), { target: { value: "Hello" } })
    fireEvent.click(getByTitle("Enviar (Enter)"))
    expect(onOptimistic).toHaveBeenCalled()
  })

  test("clears input after send", () => {
    const { getByPlaceholderText, getByTitle } = render(
      <MessageComposer channelId="ch-1" currentUserId={1} members={MEMBERS} onOptimisticMessage={() => {}} />
    )
    const textarea = getByPlaceholderText("Escribí un mensaje...") as HTMLTextAreaElement
    fireEvent.change(textarea, { target: { value: "Test" } })
    fireEvent.click(getByTitle("Enviar (Enter)"))
    expect(textarea.value).toBe("")
  })

  test("shows reply indicator when replyTo is set", () => {
    const { getByText } = render(
      <MessageComposer
        channelId="ch-1" currentUserId={1} members={MEMBERS}
        replyTo={{ id: "m-1", userId: 1, content: "Original" }}
      />
    )
    expect(getByText("Admin")).toBeTruthy()
    expect(getByText(/Respondiendo a/)).toBeTruthy()
  })
})
