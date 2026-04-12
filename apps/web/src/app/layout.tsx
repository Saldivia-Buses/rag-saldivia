import type { Metadata } from "next";
import { Toaster } from "sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryProvider } from "@/lib/api/query-provider";
import { AuthInitializer } from "@/lib/auth/auth-initializer";
import { WsProvider } from "@/lib/ws/provider";
import { fontVariables, fontClassNames } from "@/lib/fonts";
import "./globals.css";

export const metadata: Metadata = {
  title: "SDA Framework",
  description: "SDA Framework — Plataforma empresarial",
};

// Inline script — applies dark mode class before React hydration to prevent flash.
const themeScript = `(function(){try{if(localStorage.getItem("sda-dark-mode")==="true")document.documentElement.classList.add("dark")}catch(e){}})()`;

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="es"
      className={`${fontVariables} h-full antialiased`}
      suppressHydrationWarning
    >
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeScript }} />
      </head>
      <body className="min-h-full flex flex-col font-sans">
        {/* Hidden preloader — forces browser to download all font files upfront */}
        <div aria-hidden="true" className={`${fontClassNames} absolute opacity-0 pointer-events-none h-0 overflow-hidden`}>.</div>
        <QueryProvider>
          <TooltipProvider>
            <AuthInitializer>
              <WsProvider>
                {children}
                <Toaster richColors closeButton position="bottom-right" />
              </WsProvider>
            </AuthInitializer>
          </TooltipProvider>
        </QueryProvider>
      </body>
    </html>
  );
}
