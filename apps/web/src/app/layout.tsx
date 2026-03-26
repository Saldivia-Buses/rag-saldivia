import type { Metadata } from "next"
import dynamic from "next/dynamic"
import { Providers } from "@/components/providers"
import { Toaster } from "@/components/ui/sonner"
import "./globals.css"

const ReactScanInit =
  process.env.NODE_ENV === "development"
    ? dynamic(() => import("@/components/dev/ReactScan"), { ssr: false })
    : null

export const metadata: Metadata = {
  title: "RAG Saldivia",
  description: "Sistema RAG empresarial con autenticación y RBAC",
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="es" suppressHydrationWarning>
      <body className="min-h-screen antialiased" style={{ background: "var(--background)", color: "var(--foreground)" }}>
        {process.env.NODE_ENV === "development" && ReactScanInit && <ReactScanInit />}
        <Providers>
          {children}
          <Toaster />
        </Providers>
      </body>
    </html>
  )
}
