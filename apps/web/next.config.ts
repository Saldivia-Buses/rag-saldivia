import type { NextConfig } from "next"

const nextConfig: NextConfig = {
  // Transpilar los paquetes workspace (TypeScript → JS via webpack)
  transpilePackages: [
    "@rag-saldivia/shared",
    "@rag-saldivia/db",
    "@rag-saldivia/config",
    "@rag-saldivia/logger",
  ],

  // bcrypt-ts es puro TS/JS, no necesita ser externo
  // @libsql/client funciona en Node.js sin compilación nativa

  // Forzar el root del proyecto para evitar que Next.js confunda el workspace root
  outputFileTracingRoot: __dirname,

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
