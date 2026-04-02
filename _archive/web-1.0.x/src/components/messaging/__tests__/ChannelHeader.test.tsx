import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { ChannelHeader } from "../ChannelHeader"

afterEach(cleanup)

describe("ChannelHeader", () => {
  test("renders channel name", () => {
    const { getByText } = render(
      <ChannelHeader channel={{ id: "1", type: "public", name: "general", description: null, topic: null }} memberCount={5} />
    )
    expect(getByText("general")).toBeTruthy()
  })

  test("renders fallback for DM without name", () => {
    const { getByText } = render(
      <ChannelHeader channel={{ id: "2", type: "dm", name: null, description: null, topic: null }} memberCount={2} />
    )
    expect(getByText("Mensaje directo")).toBeTruthy()
  })

  test("renders topic when present", () => {
    const { getByText } = render(
      <ChannelHeader channel={{ id: "1", type: "public", name: "dev", description: null, topic: "Desarrollo" }} memberCount={3} />
    )
    expect(getByText("Desarrollo")).toBeTruthy()
  })

  test("renders member count", () => {
    const { getByText } = render(
      <ChannelHeader channel={{ id: "1", type: "public", name: "team", description: null, topic: null }} memberCount={12} />
    )
    expect(getByText("12")).toBeTruthy()
  })
})
