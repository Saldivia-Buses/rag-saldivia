import type { NextConfig } from "next"

const nextConfig: NextConfig = {
  // Transpilar los paquetes workspace (TypeScript → JS via webpack)
  transpilePackages: [
    "@rag-saldivia/shared",
    "@rag-saldivia/db",
    "@rag-saldivia/config",
    "@rag-saldivia/logger",
  ],

  // Excluir del bundling: dejar que Node.js los resuelva en runtime
  serverExternalPackages: [
    "@libsql/client",
    "@libsql/isomorphic-fetch",
    "@libsql/isomorphic-ws",
    "drizzle-orm",
  ],

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
