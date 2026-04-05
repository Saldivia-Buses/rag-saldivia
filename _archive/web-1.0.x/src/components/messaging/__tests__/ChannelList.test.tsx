import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { ChannelList } from "../ChannelList"

afterEach(cleanup)

const CHANNELS = [
  { id: "ch-1", type: "public", name: "general", description: null, topic: null, channelMembers: [{ userId: 1 }] },
  { id: "ch-2", type: "private", name: "admins", description: null, topic: null, channelMembers: [{ userId: 1 }] },
  { id: "ch-3", type: "dm", name: null, description: null, topic: null, channelMembers: [{ userId: 1 }, { userId: 2 }] },
]

describe("ChannelList", () => {
  test("renders channel names", () => {
    const { getByText } = render(
      <ChannelList channels={CHANNELS} unreadCounts={{}} userId={1} />
    )
    expect(getByText("general")).toBeTruthy()
    expect(getByText("admins")).toBeTruthy()
  })

  test("renders section labels", () => {
    const { getByText } = render(
      <ChannelList channels={CHANNELS} unreadCounts={{}} userId={1} />
    )
    expect(getByText("Canales")).toBeTruthy()
    expect(getByText("Canales privados")).toBeTruthy()
    expect(getByText("Mensajes directos")).toBeTruthy()
  })

  test("shows empty state when no channels", () => {
    const { getByText } = render(
      <ChannelList channels={[]} unreadCounts={{}} userId={1} />
    )
    expect(getByText("No hay canales todavía")).toBeTruthy()
  })

  test("renders DM fallback name", () => {
    const { getByText } = render(
      <ChannelList channels={[CHANNELS[2]!]} unreadCounts={{}} userId={1} />
    )
    expect(getByText("Mensaje directo")).toBeTruthy()
  })

  test("renders create channel button", () => {
    const { getByTitle } = render(
      <ChannelList channels={CHANNELS} unreadCounts={{}} userId={1} />
    )
    expect(getByTitle("Crear canal")).toBeTruthy()
  })

  test("shows unread badge when channel has unreads", () => {
    const { getByText } = render(
      <ChannelList channels={CHANNELS} unreadCounts={{ "ch-1": 3 }} userId={1} />
    )
    expect(getByText("3")).toBeTruthy()
  })

  test("renders DM button when allUsers provided", () => {
    const { getByTitle } = render(
      <ChannelList
        channels={CHANNELS} unreadCounts={{}} userId={1}
        allUsers={[{ id: 1, name: "Admin", email: "a@test.com" }]}
      />
    )
    expect(getByTitle("Mensaje directo")).toBeTruthy()
  })
})
