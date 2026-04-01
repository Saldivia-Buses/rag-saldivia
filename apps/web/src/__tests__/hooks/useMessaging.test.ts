import { describe, test, expect, afterEach } from "bun:test"
import { renderHook, cleanup, act } from "@testing-library/react"
import { useMessaging } from "@/hooks/useMessaging"
import { wsClient } from "@/lib/ws/client"

afterEach(cleanup)

describe("useMessaging", () => {
  test("starts disconnected with empty messages", () => {
    const { result } = renderHook(() => useMessaging(null))
    expect(result.current.connected).toBe(false)
    expect(result.current.messages).toEqual([])
  })

  test("calls wsClient.connect when token provided", () => {
    renderHook(() => useMessaging("test-token"))
    expect(wsClient.connect).toHaveBeenCalledWith("test-token")
  })

  test("calls wsClient.disconnect when token is null", () => {
    renderHook(() => useMessaging(null))
    expect(wsClient.disconnect).toHaveBeenCalled()
  })

  test("subscribes to ws messages via wsClient.on", () => {
    renderHook(() => useMessaging("token"))
    expect(wsClient.on).toHaveBeenCalled()
  })

  test("subscribe sends ws message", () => {
    const { result } = renderHook(() => useMessaging("token"))
    act(() => { result.current.subscribe("ch-1") })
    expect(wsClient.send).toHaveBeenCalledWith({ type: "subscribe", channelId: "ch-1" })
  })

  test("unsubscribe sends ws message", () => {
    const { result } = renderHook(() => useMessaging("token"))
    act(() => {
      result.current.subscribe("ch-1")
      result.current.unsubscribe("ch-1")
    })
    expect(wsClient.send).toHaveBeenCalledWith({ type: "unsubscribe", channelId: "ch-1" })
  })

  test("clearMessages resets to empty array", () => {
    const { result } = renderHook(() => useMessaging("token"))
    act(() => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any -- mock msg
      result.current.setInitialMessages([{ id: "1", content: "test" }] as any)
    })
    expect(result.current.messages).toHaveLength(1)
    act(() => { result.current.clearMessages() })
    expect(result.current.messages).toEqual([])
  })

  test("setInitialMessages sets messages", () => {
    const { result } = renderHook(() => useMessaging("token"))
    // eslint-disable-next-line @typescript-eslint/no-explicit-any -- mock msgs
    const msgs = [{ id: "1", content: "hello" }, { id: "2", content: "world" }] as any[]
    act(() => { result.current.setInitialMessages(msgs) })
    expect(result.current.messages).toHaveLength(2)
  })
})
