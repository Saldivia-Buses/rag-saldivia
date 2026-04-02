import { describe, test, expect, afterEach, beforeEach, mock } from "bun:test"
import { renderHook, act, cleanup } from "@testing-library/react"
import { useCopyToClipboard } from "@/hooks/useCopyToClipboard"

// Mock navigator.clipboard
const writeTextMock = mock(() => Promise.resolve())

beforeEach(() => {
  Object.defineProperty(navigator, "clipboard", {
    value: { writeText: writeTextMock },
    writable: true,
    configurable: true,
  })
  writeTextMock.mockClear()
})

afterEach(cleanup)

describe("useCopyToClipboard", () => {
  test("copy() writes text to clipboard", async () => {
    const { result } = renderHook(() => useCopyToClipboard())

    await act(async () => {
      await result.current.copy("hello world")
    })

    expect(writeTextMock).toHaveBeenCalledWith("hello world")
  })

  test("copied is false initially", () => {
    const { result } = renderHook(() => useCopyToClipboard())
    expect(result.current.copied).toBe(false)
    expect(result.current.copiedKey).toBeNull()
  })

  test("copied is true after copy", async () => {
    const { result } = renderHook(() => useCopyToClipboard())

    await act(async () => {
      await result.current.copy("text")
    })

    expect(result.current.copied).toBe(true)
  })

  test("copiedKey defaults to 'copied' when no key provided", async () => {
    const { result } = renderHook(() => useCopyToClipboard())

    await act(async () => {
      await result.current.copy("text")
    })

    expect(result.current.copiedKey).toBe("copied")
  })

  test("copiedKey matches the key passed to copy()", async () => {
    const { result } = renderHook(() => useCopyToClipboard())

    await act(async () => {
      await result.current.copy("some text", "msg-123")
    })

    expect(result.current.copiedKey).toBe("msg-123")
  })

  test("resets after timeout (fake timers)", async () => {
    const originalSetTimeout = globalThis.setTimeout
    const originalClearTimeout = globalThis.clearTimeout

    // Collect timeout callbacks
    const timeouts: Array<{ cb: () => void; delay: number; id: number }> = []
    let nextId = 1

    // @ts-expect-error -- overriding for fake timers
    globalThis.setTimeout = (cb: () => void, delay: number) => {
      const id = nextId++
      timeouts.push({ cb, delay, id })
      return id
    }
    globalThis.clearTimeout = (id: number) => {
      const idx = timeouts.findIndex((t) => t.id === id)
      if (idx !== -1) timeouts.splice(idx, 1)
    }

    try {
      const { result } = renderHook(() => useCopyToClipboard(1000))

      await act(async () => {
        await result.current.copy("text")
      })

      expect(result.current.copied).toBe(true)

      // Find the timeout that was set by the hook (the one with delay 1000)
      const hookTimeout = timeouts.find((t) => t.delay === 1000)
      expect(hookTimeout).toBeDefined()

      // Fire the timeout callback
      act(() => {
        hookTimeout!.cb()
      })

      expect(result.current.copied).toBe(false)
      expect(result.current.copiedKey).toBeNull()
    } finally {
      globalThis.setTimeout = originalSetTimeout
      globalThis.clearTimeout = originalClearTimeout
    }
  })

  test("subsequent copies replace the copiedKey", async () => {
    const { result } = renderHook(() => useCopyToClipboard())

    await act(async () => {
      await result.current.copy("first", "key-1")
    })
    expect(result.current.copiedKey).toBe("key-1")

    await act(async () => {
      await result.current.copy("second", "key-2")
    })
    expect(result.current.copiedKey).toBe("key-2")
  })

  test("uses custom timeout value", async () => {
    const originalSetTimeout = globalThis.setTimeout
    const originalClearTimeout = globalThis.clearTimeout

    const timeouts: Array<{ cb: () => void; delay: number; id: number }> = []
    let nextId = 1

    // @ts-expect-error -- overriding for fake timers
    globalThis.setTimeout = (cb: () => void, delay: number) => {
      const id = nextId++
      timeouts.push({ cb, delay, id })
      return id
    }
    globalThis.clearTimeout = (_id: number) => {}

    try {
      const { result } = renderHook(() => useCopyToClipboard(5000))

      await act(async () => {
        await result.current.copy("text")
      })

      const hookTimeout = timeouts.find((t) => t.delay === 5000)
      expect(hookTimeout).toBeDefined()
    } finally {
      globalThis.setTimeout = originalSetTimeout
      globalThis.clearTimeout = originalClearTimeout
    }
  })
})
