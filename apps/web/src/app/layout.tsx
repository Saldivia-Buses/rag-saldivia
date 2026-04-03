import type { Metadata } from "next";
import { TooltipProvider } from "@/components/ui/tooltip";
import { ThemeProvider } from "@/lib/theme-provider";
import { QueryProvider } from "@/lib/api/query-provider";
import { AuthInitializer } from "@/lib/auth/auth-initializer";
import { fontVariables, fontClassNames } from "@/lib/fonts";
import { SearchCommand } from "@/components/search-command";
import "./globals.css";

export const metadata: Metadata = {
  title: "SDA Framework",
  description: "SDA Framework — Plataforma empresarial",
};

// Inline script that runs before React hydration to prevent theme flash.
// Reads cached CSS variables from localStorage and applies them immediately.
const themeScript = `(function(){try{var d=document.documentElement;if(localStorage.getItem("sda-dark-mode")==="true")d.classList.add("dark");var v=localStorage.getItem("sda-theme-vars");if(v){var o=JSON.parse(v);for(var k in o)d.style.setProperty("--"+k,o[k])}}catch(e){}})()`;

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
          <ThemeProvider>
            <TooltipProvider>
              <AuthInitializer>
                <SearchCommand />
                {children}
              </AuthInitializer>
            </TooltipProvider>
          </ThemeProvider>
        </QueryProvider>
      </body>
    </html>
  );
}
