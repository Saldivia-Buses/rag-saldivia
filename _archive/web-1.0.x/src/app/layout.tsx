import type { Metadata } from "next"
import { Instrument_Sans } from "next/font/google"
import { Providers } from "@/components/providers"
import { Toaster } from "@/components/ui/sonner"
import { ReactScanProvider } from "@/components/dev/ReactScanProvider"
import { NuqsAdapter } from "nuqs/adapters/next/app"
import "./globals.css"

const instrumentSans = Instrument_Sans({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
  style: ["normal", "italic"],
  variable: "--font-instrument-sans",
  display: "swap",
})

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
    <html lang="es" suppressHydrationWarning className={instrumentSans.variable}>
      <body className="min-h-screen antialiased" style={{ background: "var(--background)", color: "var(--foreground)" }}>
        <ReactScanProvider />
        <NuqsAdapter>
          <Providers>
            {children}
            <Toaster />
          </Providers>
        </NuqsAdapter>
      </body>
    </html>
  )
}
