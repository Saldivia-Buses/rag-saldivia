import type { NextConfig } from "next"

const nextConfig: NextConfig = {
  // Transpilar los paquetes workspace (TypeScript → JS via webpack)
  transpilePackages: [
    "@rag-saldivia/shared",
    "@rag-saldivia/db",
    "@rag-saldivia/config",
    "@rag-saldivia/logger",
  ],

  // Excluir del bundling todos los paquetes de la cadena SQLite nativa
  // drizzle-orm NO va acá — debe bundlearse con el schema para evitar conflictos de instancias
  serverExternalPackages: [
    "@libsql/client",
    "@libsql/isomorphic-fetch",
    "@libsql/isomorphic-ws",
    "@libsql/hrana-client",
    "libsql",
  ],

  webpack: (config, { isServer }) => {
    if (isServer) {
      // Asegurar que libsql y su cadena de dependencias nativas no se bundleen
      const existingExternals = Array.isArray(config.externals)
        ? config.externals
        : config.externals
          ? [config.externals]
          : []

      config.externals = [
        ...existingExternals,
        ({ request }: { request?: string }, callback: (err?: Error, result?: string) => void) => {
          if (
            request &&
            (request.startsWith("libsql") ||
              request.startsWith("@libsql/") ||
              request.endsWith(".node"))
          ) {
            callback(undefined, `commonjs ${request}`)
          } else {
            callback()
          }
        },
      ]
    }
    return config
  },

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
