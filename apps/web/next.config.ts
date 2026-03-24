import type { NextConfig } from "next"
import { join } from "path"

const nextConfig: NextConfig = {
  // Forzar el root del proyecto para evitar que Next.js confunda el workspace root
  // con un package-lock.json externo (problema en WSL2 con filesystem montado)
  outputFileTracingRoot: join(__dirname, "../../"),
  turbopack: {
    root: __dirname,
  },
  // Transpile workspace packages
  transpilePackages: [
    "@rag-saldivia/shared",
    "@rag-saldivia/db",
    "@rag-saldivia/config",
    "@rag-saldivia/logger",
  ],

  // Logging detallado en dev
  logging: {
    fetches: {
      fullUrl: process.env["NODE_ENV"] === "development",
    },
  },

  // Headers de seguridad
  async headers() {
    return [
      {
        source: "/(.*)",
        headers: [
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "X-Frame-Options", value: "DENY" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
        ],
      },
    ]
  },
}

export default nextConfig
