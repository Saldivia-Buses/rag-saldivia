import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "SDA Framework — Iniciar sesion",
  description: "Accede a tu cuenta de SDA Framework",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="es" className="h-full antialiased">
      <body className="min-h-full flex flex-col bg-background text-foreground">
        {children}
      </body>
    </html>
  );
}
