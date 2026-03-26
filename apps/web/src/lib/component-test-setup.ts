import { GlobalRegistrator } from "@happy-dom/global-registrator"
import { expect } from "bun:test"
import * as matchers from "@testing-library/jest-dom/matchers"
import { mock } from "bun:test"

// Activar happy-dom para tests de componentes React
GlobalRegistrator.register()

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
  redirect: mock((_url: string) => { throw new Error(`NEXT_REDIRECT: ${_url}`) }),
  permanentRedirect: mock((_url: string) => { throw new Error(`NEXT_REDIRECT: ${_url}`) }),
  notFound: mock(() => { throw new Error("NEXT_NOT_FOUND") }),
  headers: mock(() => new Headers()),
  cookies: mock(() => ({ get: () => null, set: () => {}, delete: () => {} })),
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

// Mock next/image (evita errores con img optimization)
mock.module("next/image", () => ({
  default: ({ src, alt, ...props }: { src: string; alt: string; [key: string]: unknown }) =>
    Object.assign(document.createElement("img"), { src, alt, ...props }),
}))
