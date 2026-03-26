import { expect } from "bun:test"
import * as matchers from "@testing-library/jest-dom/matchers"
import { mock } from "bun:test"

expect.extend(matchers)

// Mock next/navigation
mock.module("next/navigation", () => ({
  useRouter: () => ({
    push: mock(() => {}),
    replace: mock(() => {}),
    back: mock(() => {}),
    prefetch: mock(() => {}),
    refresh: mock(() => {}),
  }),
  usePathname: () => "/",
  useSearchParams: () => new URLSearchParams(),
  useParams: () => ({}),
}))

// Mock next/font/google
mock.module("next/font/google", () => ({
  Instrument_Sans: () => ({
    className: "mock-font",
    variable: "--font-instrument-sans",
  }),
}))

// Mock next-themes
mock.module("next-themes", () => ({
  useTheme: () => ({ theme: "light", setTheme: mock(() => {}), resolvedTheme: "light" }),
  ThemeProvider: ({ children }: { children: React.ReactNode }) => children,
}))

// Mock next/dynamic
mock.module("next/dynamic", () => ({
  default: (_fn: unknown) => () => null,
}))
