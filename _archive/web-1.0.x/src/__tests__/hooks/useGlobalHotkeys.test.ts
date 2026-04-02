import { describe, test, expect, afterEach, mock } from "bun:test"
import { renderHook, cleanup } from "@testing-library/react"
import { useGlobalHotkeys } from "@/hooks/useGlobalHotkeys"

afterEach(cleanup)

describe("useGlobalHotkeys", () => {
  test("renders without error with no options", () => {
    const { result } = renderHook(() => useGlobalHotkeys())
    expect(result.current).toBeUndefined() // void hook
  })

  test("renders without error with callbacks", () => {
    const onOpenPalette = mock(() => {})
    const onNewChannel = mock(() => {})
    const { result } = renderHook(() =>
      useGlobalHotkeys({ onOpenPalette, onNewChannel })
    )
    expect(result.current).toBeUndefined()
  })

  test("accepts partial options", () => {
    const onOpenPalette = mock(() => {})
    expect(() => {
      renderHook(() => useGlobalHotkeys({ onOpenPalette }))
    }).not.toThrow()
  })
})
