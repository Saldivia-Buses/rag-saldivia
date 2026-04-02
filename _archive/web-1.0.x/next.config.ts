import path from "path"
import type { NextConfig } from "next"
import withBundleAnalyzerFactory from "@next/bundle-analyzer"

const withBundleAnalyzer = withBundleAnalyzerFactory({
  enabled: process.env["ANALYZE"] === "true",
})

const nextConfig: NextConfig = {
  // Plan 26: standalone output (-300-500MB RAM in prod), gzip compression
  output: "standalone",
  compress: true,

  // React Compiler — auto-memoization, no manual useMemo/useCallback needed
  reactCompiler: true,

  experimental: {
    // Plan 26: improved tree-shaking for icon library
    optimizePackageImports: ["lucide-react"],
  },

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
    "@libsql/core",
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

  // Forzar el root al monorepo para que Turbopack resuelva packages/
  outputFileTracingRoot: path.resolve(__dirname, "../.."),

  // Logging detallado en dev
  logging: {
    fetches: {
      fullUrl: process.env["NODE_ENV"] === "development",
    },
  },

  // Next.js 16: Turbopack es el bundler default.
  // turbopack.root y outputFileTracingRoot deben apuntar al mismo valor.
  turbopack: {
    root: path.resolve(__dirname, "../.."),
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
          // Plan 26: security headers 6/6 (industry standard)
          { key: "Strict-Transport-Security", value: "max-age=63072000; includeSubDomains" },
          { key: "Permissions-Policy", value: "camera=(), microphone=(), geolocation=()" },
          {
            key: "Content-Security-Policy",
            value: [
              "default-src 'self'",
              // unsafe-inline required by Next.js (inline scripts). unsafe-eval removed (only needed in dev HMR).
              // Nonce-based CSP requires custom middleware — planned for future hardening.
              "script-src 'self' 'unsafe-inline'",
              "style-src 'self' 'unsafe-inline'",
              "img-src 'self' data: blob:",
              "font-src 'self'",
              "connect-src 'self' ws: wss:",
              "frame-ancestors 'none'",
            ].join("; "),
          },
        ],
      },
    ]
  },
}

export default withBundleAnalyzer(nextConfig)
