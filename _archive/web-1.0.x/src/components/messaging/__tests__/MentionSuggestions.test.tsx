import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { MentionSuggestions } from "../MentionSuggestions"

afterEach(cleanup)

const MEMBERS = [
  { id: 1, name: "Admin", email: "admin@test.com" },
  { id: 2, name: "María", email: "maria@test.com" },
]

describe("MentionSuggestions", () => {
  test("renders nothing when not visible", () => {
    const { container } = render(
      <MentionSuggestions query="" members={MEMBERS} onSelect={() => {}} onClose={() => {}} visible={false} />
    )
    expect(container.innerHTML).toBe("")
  })

  test("renders members when visible", () => {
    const { getByText } = render(
      <MentionSuggestions query="" members={MEMBERS} onSelect={() => {}} onClose={() => {}} visible={true} />
    )
    expect(getByText("Admin")).toBeTruthy()
    expect(getByText("María")).toBeTruthy()
  })

  test("filters by query", () => {
    const { getByText, queryByText } = render(
      <MentionSuggestions query="mar" members={MEMBERS} onSelect={() => {}} onClose={() => {}} visible={true} />
    )
    expect(getByText("María")).toBeTruthy()
    expect(queryByText("Admin")).toBeNull()
  })

  test("renders nothing when no matches", () => {
    const { container } = render(
      <MentionSuggestions query="zzz" members={MEMBERS} onSelect={() => {}} onClose={() => {}} visible={true} />
    )
    expect(container.innerHTML).toBe("")
  })
})
