"use client"

import { Globe, Building2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { signIn } from "next-auth/react"

type Provider = "google" | "azure-ad"

const LABELS: Record<Provider, string> = {
  "google": "Continuar con Google",
  "azure-ad": "Continuar con Microsoft",
}

const ICONS: Record<Provider, React.ReactNode> = {
  "google": <Globe size={16} />,
  "azure-ad": <Building2 size={16} />,
}

type Props = {
  provider: Provider
}

export function SSOButton({ provider }: Props) {
  const clientId = provider === "google"
    ? process.env["NEXT_PUBLIC_GOOGLE_CLIENT_ID"]
    : process.env["NEXT_PUBLIC_AZURE_AD_CLIENT_ID"]

  // No mostrar el botón si el provider no está configurado
  if (!clientId && typeof window !== "undefined") return null

  return (
    <Button
      variant="outline"
      className="w-full gap-2"
      onClick={() => signIn(provider === "azure-ad" ? "microsoft-entra-id" : provider, { callbackUrl: "/chat" })}
    >
      {ICONS[provider]}
      {LABELS[provider]}
    </Button>
  )
}
