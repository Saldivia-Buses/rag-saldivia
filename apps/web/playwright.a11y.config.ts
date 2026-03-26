import { defineConfig } from "@playwright/test"

export default defineConfig({
  testDir: "./tests/a11y",
  use: {
    baseURL: "http://localhost:3000",
  },
  webServer: {
    command: "MOCK_RAG=true bun run dev",
    url: "http://localhost:3000",
    reuseExistingServer: true,
    timeout: 30_000,
  },
})
