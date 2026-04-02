import { defineConfig } from "@playwright/test"

export default defineConfig({
  testDir: "./tests/visual",
  snapshotDir: "./tests/visual/snapshots",
  snapshotPathTemplate: "{snapshotDir}/{testFilePath}/{arg}{ext}",
  use: {
    baseURL: "http://localhost:6006",
  },
  // No emular colorScheme — el dark mode es class-based (next-themes attribute="class")
  webServer: {
    command: "bun run storybook",
    url: "http://localhost:6006",
    reuseExistingServer: true,
    timeout: 60_000,
  },
})
