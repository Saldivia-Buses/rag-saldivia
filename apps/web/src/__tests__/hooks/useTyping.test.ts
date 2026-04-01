import { describe, test, expect, afterEach, beforeEach } from "bun:test"
import { renderHook, cleanup, act } from "@testing-library/react"
import { useTyping } from "@/hooks/useTyping"
import { wsClient } from "@/lib/ws/client"

afterEach(cleanup)

// Clear mock call counts between tests
beforeEach(() => {
  const send = wsClient.send as ReturnType<typeof import("bun:test").mock>
  send.mockClear?.()
})

describe("useTyping", () => {
  test("starts with empty typingUsers", () => {
    const { result } = renderHook(() => useTyping("ch-1"))
    expect(result.current.typingUsers).toEqual([])
  })

  test("returns noop functions when channelId is null", () => {
    const { result } = renderHook(() => useTyping(null))
    expect(result.current.typingUsers).toEqual([])
    // Should not throw
    act(() => { result.current.startTyping() })
    act(() => { result.current.stopTyping() })
    expect(wsClient.send).not.toHaveBeenCalled()
  })

  test("startTyping sends ws message", () => {
    const { result } = renderHook(() => useTyping("ch-1"))
    act(() => { result.current.startTyping() })
    expect(wsClient.send).toHaveBeenCalledWith({ type: "typing_start", channelId: "ch-1" })
  })

  test("stopTyping sends ws message", () => {
    const { result } = renderHook(() => useTyping("ch-1"))
    act(() => { result.current.stopTyping() })
    expect(wsClient.send).toHaveBeenCalledWith({ type: "typing_stop", channelId: "ch-1" })
  })

  test("startTyping debounces within 1 second", () => {
    const { result } = renderHook(() => useTyping("ch-1"))
    act(() => {
      result.current.startTyping()
      result.current.startTyping()
      result.current.startTyping()
    })
    // Only first call should go through (debounce 1s)
    const send = wsClient.send as ReturnType<typeof import("bun:test").mock>
    const typingCalls = send.mock.calls.filter(
      (c) => (c[0] as Record<string, unknown>)?.type === "typing_start"
    )
    expect(typingCalls.length).toBe(1)
  })

  test("subscribes to ws on mount with channelId", () => {
    renderHook(() => useTyping("ch-1"))
    expect(wsClient.on).toHaveBeenCalled()
  })
})
