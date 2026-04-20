import type { NextConfig } from "next";

const securityHeaders = [
  {
    key: "Content-Security-Policy",
    value: [
      "default-src 'self'",
      "script-src 'self' 'unsafe-inline' 'unsafe-eval'",
      "style-src 'self' 'unsafe-inline'",
      "connect-src *",
      "img-src 'self' data: blob: https://models.dev",
      "font-src 'self'",
      "frame-ancestors 'none'",
      "base-uri 'self'",
      "form-action 'self'",
    ].join("; "),
  },
  { key: "X-Content-Type-Options", value: "nosniff" },
  { key: "X-Frame-Options", value: "DENY" },
  { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
  {
    key: "Permissions-Policy",
    value: "camera=(), microphone=(), geolocation=()",
  },
];

const nextConfig: NextConfig = {
  output: "standalone",
  experimental: {
    authInterrupts: true,
  },
  async headers() {
    return [
      {
        source: "/(.*)",
        headers: securityHeaders,
      },
    ];
  },
  // Dev-only proxy: when NEXT_PUBLIC_API_URL is empty (set by `make dev-frontend`),
  // browser fetches go to localhost:3000. Rewrite /v1/erp/* to sda-erp:8013 and
  // /v1/* (everything else) to sda-app:8020 so relative API calls resolve in dev.
  async rewrites() {
    if (process.env.NODE_ENV !== "development") return [];
    return [
      { source: "/v1/erp/:path*", destination: "http://localhost:8013/v1/erp/:path*" },
      { source: "/v1/:path*", destination: "http://localhost:8020/v1/:path*" },
    ];
  },
};

export default nextConfig;
