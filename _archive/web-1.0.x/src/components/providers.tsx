"use client"

import { ThemeProvider } from "next-themes"
import { ErrorFeedbackMount } from "@/components/ui/error-feedback"

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider
      attribute="class"
      defaultTheme="light"
      enableSystem={false}
      disableTransitionOnChange
      storageKey="rag-theme"
    >
      {children}
      <ErrorFeedbackMount />
    </ThemeProvider>
  )
}
