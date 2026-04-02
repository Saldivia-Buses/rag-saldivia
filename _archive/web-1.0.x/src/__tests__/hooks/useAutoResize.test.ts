import { describe, test, expect, afterEach } from "bun:test"
import { renderHook, cleanup } from "@testing-library/react"
import { useAutoResize } from "@/hooks/useAutoResize"
import { useRef } from "react"

afterEach(cleanup)

/**
 * Helper: creates a ref wrapper hook for useAutoResize tests.
 * We need to use useRef inside a hook context, so we wrap both
 * useRef and useAutoResize in a single hook call.
 */
function renderAutoResize(opts: {
  value: string
  maxHeight?: number
  scrollHeight?: number
  skipRef?: boolean
}) {
  const { value, maxHeight, scrollHeight = 100, skipRef = false } = opts

  return renderHook(
    ({ value: v, maxHeight: mh }) => {
      const ref = useRef<HTMLTextAreaElement>(null)

      // Create and assign a fake textarea to the ref on first render
      if (!skipRef && ref.current === null) {
        const ta = document.createElement("textarea")
        // Mock scrollHeight via defineProperty since it's read-only
        Object.defineProperty(ta, "scrollHeight", {
          get: () => scrollHeight,
          configurable: true,
        })
        // @ts-expect-error -- assigning to ref.current in test
        ref.current = ta
      }

      useAutoResize(skipRef ? undefined : ref, v, mh)

      return { ref }
    },
    { initialProps: { value, maxHeight } }
  )
}

describe("useAutoResize", () => {
  test("sets textarea height based on scrollHeight", () => {
    const { result } = renderAutoResize({
      value: "hello",
      scrollHeight: 80,
    })

    const ta = result.current.ref.current!
    expect(ta.style.height).toBe("80px")
  })

  test("respects maxHeight when scrollHeight exceeds it", () => {
    const { result } = renderAutoResize({
      value: "very long text",
      scrollHeight: 500,
      maxHeight: 200,
    })

    const ta = result.current.ref.current!
    expect(ta.style.height).toBe("200px")
  })

  test("uses scrollHeight when under maxHeight", () => {
    const { result } = renderAutoResize({
      value: "short",
      scrollHeight: 50,
      maxHeight: 200,
    })

    const ta = result.current.ref.current!
    expect(ta.style.height).toBe("50px")
  })

  test("handles undefined ref gracefully (no throw)", () => {
    // Should not throw when ref is undefined
    expect(() => {
      renderAutoResize({
        value: "text",
        skipRef: true,
      })
    }).not.toThrow()
  })

  test("re-runs when value changes", () => {
    let currentScrollHeight = 40

    const { result, rerender } = renderHook(
      ({ value, maxHeight }) => {
        const ref = useRef<HTMLTextAreaElement>(null)

        if (ref.current === null) {
          const ta = document.createElement("textarea")
          Object.defineProperty(ta, "scrollHeight", {
            get: () => currentScrollHeight,
            configurable: true,
          })
          // @ts-expect-error -- assigning to ref.current in test
          ref.current = ta
        }

        useAutoResize(ref, value, maxHeight)
        return { ref }
      },
      { initialProps: { value: "short", maxHeight: 300 } }
    )

    expect(result.current.ref.current!.style.height).toBe("40px")

    // Simulate content growth
    currentScrollHeight = 120
    rerender({ value: "much longer text that wraps to multiple lines", maxHeight: 300 })

    expect(result.current.ref.current!.style.height).toBe("120px")
  })

  test("uses default maxHeight of 200 when not provided", () => {
    const { result } = renderAutoResize({
      value: "text",
      scrollHeight: 300,
      // no maxHeight provided -> defaults to 200
    })

    const ta = result.current.ref.current!
    expect(ta.style.height).toBe("200px")
  })

  test("sets height to auto before calculating", () => {
    // This verifies the reset-then-measure pattern
    const heights: string[] = []

    renderHook(() => {
      const ref = useRef<HTMLTextAreaElement>(null)

      if (ref.current === null) {
        const ta = document.createElement("textarea")
        Object.defineProperty(ta, "scrollHeight", {
          get: () => 60,
          configurable: true,
        })
        // Intercept style.height assignments
        let internalHeight = ""
        Object.defineProperty(ta.style, "height", {
          get: () => internalHeight,
          set: (v: string) => {
            heights.push(v)
            internalHeight = v
          },
          configurable: true,
        })
        // @ts-expect-error -- assigning to ref.current in test
        ref.current = ta
      }

      useAutoResize(ref, "text", 200)
      return { ref }
    })

    // Should first set to "auto", then to the calculated height
    expect(heights).toContain("auto")
    expect(heights).toContain("60px")
    // "auto" should come before the calculated height
    const autoIdx = heights.indexOf("auto")
    const pxIdx = heights.indexOf("60px")
    expect(autoIdx).toBeLessThan(pxIdx)
  })
})
