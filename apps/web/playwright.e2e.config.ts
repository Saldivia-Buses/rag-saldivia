import { defineConfig } from "@playwright/test"

export default defineConfig({
  testDir: "./tests/e2e-playwright",
  use: {
    baseURL: "http://localhost:3000",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  timeout: 30_000,
  retries: 1,
  webServer: {
    command: "MOCK_RAG=true bun run dev:webpack",
    url: "http://localhost:3000",
    reuseExistingServer: true,
    timeout: 120_000,
    env: {
      MOCK_RAG: "true",
      NODE_ENV: process.env["NODE_ENV"] ?? "development",
      NEXT_PUBLIC_DISABLE_REACT_SCAN: "true",
      REDIS_URL: process.env["REDIS_URL"] ?? "redis://127.0.0.1:6379",
      JWT_SECRET: process.env["JWT_SECRET"] ?? "dev-e2e-jwt-secret-change-me",
      DATABASE_PATH: process.env["DATABASE_PATH"] ?? "./data/app.db",
      SYSTEM_API_KEY: process.env["SYSTEM_API_KEY"] ?? "dev-system-api-key",
    },
  },
})
