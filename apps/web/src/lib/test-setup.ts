import { expect } from "bun:test"
import * as matchers from "@testing-library/jest-dom/matchers"
import { mock } from "bun:test"
import IORedisMock from "ioredis-mock"

expect.extend(matchers)

// Asegurar REDIS_URL para que getRedisClient no lance en tests
process.env["REDIS_URL"] ??= "redis://localhost:6379"

// Mockear ioredis con ioredis-mock para tests que usan extractClaims u otras
// funciones que internamente llaman a getRedisClient()
mock.module("ioredis", () => ({ default: IORedisMock }))

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
