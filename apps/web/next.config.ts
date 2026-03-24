import type { NextConfig } from "next"

const nextConfig: NextConfig = {
  // No bundear los paquetes internos — Next.js los resuelve en runtime con Bun
  // Necesario para que bun:sqlite funcione (módulo nativo de Bun, no de Node.js/webpack)
  serverExternalPackages: [
    "@rag-saldivia/db",
    "@rag-saldivia/config",
    "@rag-saldivia/logger",
    "@rag-saldivia/shared",
    "bun:sqlite",
  ],
  // Forzar el root del proyecto para evitar que Next.js confunda el workspace root
  outputFileTracingRoot: __dirname,
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
