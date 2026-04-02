import { describe, test, expect, afterEach } from "bun:test"
import { renderHook, act, cleanup } from "@testing-library/react"
import { useLocalStorage } from "@/hooks/useLocalStorage"

afterEach(() => {
  cleanup()
  localStorage.clear()
})

describe("useLocalStorage", () => {
  test("returns default value on initial render (SSR safety)", () => {
    const { result } = renderHook(() => useLocalStorage("key", "default"))
    // Before the mount effect fires, value should be the default
    expect(result.current[0]).toBe("default")
  })

  test("reads stored value from localStorage after mount", async () => {
    localStorage.setItem("test-key", JSON.stringify("stored-value"))

    const { result, rerender } = renderHook(() =>
      useLocalStorage("test-key", "default")
    )

    // Force effects to flush
    rerender()

    // After mount effect, should have read from localStorage
    expect(result.current[0]).toBe("stored-value")
  })

  test("set() updates both state and localStorage", () => {
    const { result } = renderHook(() => useLocalStorage("key", "initial"))

    act(() => {
      result.current[1]("updated")
    })

    expect(result.current[0]).toBe("updated")
    expect(JSON.parse(localStorage.getItem("key")!)).toBe("updated")
  })

  test("handles function updater: set(prev => !prev)", () => {
    const { result } = renderHook(() => useLocalStorage("bool-key", false))

    act(() => {
      result.current[1]((prev) => !prev)
    })

    expect(result.current[0]).toBe(true)
    expect(JSON.parse(localStorage.getItem("bool-key")!)).toBe(true)

    act(() => {
      result.current[1]((prev) => !prev)
    })

    expect(result.current[0]).toBe(false)
    expect(JSON.parse(localStorage.getItem("bool-key")!)).toBe(false)
  })

  test("JSON serialization for objects", () => {
    const obj = { name: "test", count: 42 }
    const { result } = renderHook(() =>
      useLocalStorage("obj-key", { name: "", count: 0 })
    )

    act(() => {
      result.current[1](obj)
    })

    expect(result.current[0]).toEqual(obj)
    expect(JSON.parse(localStorage.getItem("obj-key")!)).toEqual(obj)
  })

  test("JSON serialization for arrays", () => {
    const arr = [1, 2, 3]
    const { result } = renderHook(() =>
      useLocalStorage<number[]>("arr-key", [])
    )

    act(() => {
      result.current[1](arr)
    })

    expect(result.current[0]).toEqual(arr)
    expect(JSON.parse(localStorage.getItem("arr-key")!)).toEqual(arr)
  })

  test("ignores corrupted JSON in localStorage", () => {
    localStorage.setItem("bad-key", "not-valid-json{{{")

    const { result, rerender } = renderHook(() =>
      useLocalStorage("bad-key", "fallback")
    )
    rerender()

    // Should keep the default because JSON.parse failed
    expect(result.current[0]).toBe("fallback")
  })

  test("different keys are independent", () => {
    const { result: resultA } = renderHook(() =>
      useLocalStorage("key-a", "a")
    )
    const { result: resultB } = renderHook(() =>
      useLocalStorage("key-b", "b")
    )

    act(() => {
      resultA.current[1]("updated-a")
    })

    expect(resultA.current[0]).toBe("updated-a")
    expect(resultB.current[0]).toBe("b")
  })

  test("returns default when localStorage has null for the key", () => {
    // Key not set at all
    const { result, rerender } = renderHook(() =>
      useLocalStorage("missing", 99)
    )
    rerender()

    expect(result.current[0]).toBe(99)
  })
})
