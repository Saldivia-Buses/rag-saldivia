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

// ── Messaging mocks (Plan 29) ────────────────────────────────────────────

// Mock WebSocket client singleton — used by useMessaging, useTyping
mock.module("@/lib/ws/client", () => ({
  wsClient: {
    send: mock(() => {}),
    on: mock(() => mock(() => {})),  // returns unsubscribe fn
    connect: mock(() => {}),
    disconnect: mock(() => {}),
    get connected() { return false },
    get userId() { return null },
  },
}))

// Mock messaging server actions — used by 5+ messaging components
const mockAction = mock(async () => ({ data: true }))
mock.module("@/app/actions/messaging", () => ({
  actionSendMessage: mockAction,
  actionEditMessage: mockAction,
  actionDeleteMessage: mockAction,
  actionCreateChannel: mock(async () => ({ data: { id: 1 } })),
  actionJoinChannel: mockAction,
  actionLeaveChannel: mockAction,
  actionPinMessage: mockAction,
  actionUnpinMessage: mockAction,
  actionReactToMessage: mockAction,
  actionRemoveReaction: mockAction,
  actionMarkAsRead: mockAction,
}))

// Mock admin server actions
mock.module("@/app/actions/admin", () => ({
  actionCreateUser: mock(async () => ({ data: { id: 1 } })),
  actionUpdateUser: mock(async () => ({ data: true })),
  actionResetPassword: mock(async () => ({ data: true })),
  actionDeleteUser: mock(async () => ({ data: true })),
  actionListUsers: mock(async () => []),
}))

mock.module("@/app/actions/roles", () => ({
  actionCreateRole: mock(async () => ({ data: { id: 1 } })),
  actionUpdateRole: mock(async () => ({ data: true })),
  actionDeleteRole: mock(async () => ({ data: true })),
  actionListRoles: mock(async () => []),
  actionGetRolePermissions: mock(async () => ({ data: [] })),
  actionSetRolePermissions: mock(async () => ({ data: true })),
  actionSetUserRoles: mock(async () => ({ data: true })),
  actionListPermissions: mock(async () => ({ data: [] })),
}))

mock.module("@/app/actions/areas", () => ({
  actionCreateArea: mock(async () => ({ data: { id: 1 } })),
  actionUpdateArea: mock(async () => ({ data: true })),
  actionDeleteArea: mock(async () => ({ data: true })),
  actionSetAreaCollections: mock(async () => ({ data: true })),
  actionAddUserToArea: mock(async () => ({ data: true })),
  actionRemoveUserFromArea: mock(async () => ({ data: true })),
}))

// Mock auth + chat actions — used by NavRail, ChatInterface, SessionList
mock.module("@/app/actions/auth", () => ({
  actionLogout: mock(async () => ({ data: true })),
}))

mock.module("@/app/actions/chat", () => ({
  actionCreateSession: mock(async () => ({ data: { id: "new-session" } })),
  actionDeleteSession: mock(async () => ({ data: true })),
  actionRenameSession: mock(async () => ({ data: true })),
  actionAddMessage: mock(async () => ({ data: true })),
  actionAddFeedback: mock(async () => ({ data: true })),
  actionToggleSaved: mock(async () => ({ data: true })),
  actionForkSession: mock(async () => ({ data: { id: "forked" } })),
}))

// Mock AI SDK — used by ChatInterface
mock.module("@ai-sdk/react", () => ({
  useChat: () => ({
    messages: [],
    sendMessage: mock(async () => {}),
    status: "idle",
    error: null,
    stop: mock(() => {}),
  }),
}))

mock.module("ai", () => ({
  DefaultChatTransport: class { constructor() {} },
}))

// Mock error feedback — used by admin components
mock.module("@/lib/error-feedback", () => ({
  reportError: mock(async () => true),
  showErrorFeedback: mock(() => {}),
  registerFeedbackSetter: mock(() => {}),
}))

// Mock navigator.clipboard (happy-dom doesn't support it)
if (!navigator.clipboard) {
  Object.defineProperty(navigator, "clipboard", {
    value: {
      writeText: mock(async () => {}),
      readText: mock(async () => ""),
    },
    writable: true,
  })
}
