import { defineConfig } from "drizzle-kit"

export default defineConfig({
  schema: "./src/schema",
  out: "./drizzle",
  dialect: "sqlite",
  dbCredentials: {
    url: process.env["DATABASE_PATH"] ?? "./data/app.db",
  },
})
