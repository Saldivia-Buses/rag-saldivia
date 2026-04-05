import { defineConfig } from "@playwright/test"

export default defineConfig({
  testDir: "./tests/a11y",
  use: {
    baseURL: "http://localhost:3000",
  },
  webServer: {
    command: "MOCK_RAG=true bun run dev:webpack",
    url: "http://localhost:3000",
    reuseExistingServer: true,
    timeout: 60_000,
    env: {
      MOCK_RAG: "true",
      NEXT_PUBLIC_DISABLE_REACT_SCAN: "true",
      NODE_ENV: "development",
      REDIS_URL: process.env["REDIS_URL"] ?? "redis://127.0.0.1:6379",
      DATABASE_PATH: process.env["DATABASE_PATH"] ?? "./data/app.db",
      JWT_SECRET: process.env["JWT_SECRET"] ?? "dev-a11y-jwt-secret",
      SYSTEM_API_KEY: process.env["SYSTEM_API_KEY"] ?? "dev-system-api-key",
    },
  },
})
