import type { Metadata } from "next";
import { TooltipProvider } from "@/components/ui/tooltip";
import { ThemeProvider } from "@/lib/theme-provider";
import { fontVariables } from "@/lib/fonts";
import "./globals.css";

export const metadata: Metadata = {
  title: "SDA Framework",
  description: "SDA Framework — Plataforma empresarial",
};

// Inline script that runs before React hydration to prevent theme flash.
// Reads cached CSS variables from localStorage and applies them immediately.
const themeScript = `(function(){try{var v=localStorage.getItem("sda-theme-vars");if(v){var d=document.documentElement,o=JSON.parse(v);for(var k in o)d.style.setProperty("--"+k,o[k])}}catch(e){}})()`;

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
        <ThemeProvider>
          <TooltipProvider>{children}</TooltipProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
