"use client"

import { useState } from "react"
import { Share2, Copy, Check, Trash2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { actionCreateShare, actionRevokeShare } from "@/app/actions/share"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"

type Props = {
  sessionId: string
}

export function ShareDialog({ sessionId }: Props) {
  const [shareUrl, setShareUrl] = useState<string | null>(null)
  const [shareId, setShareId] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [copied, setCopied] = useState(false)
  const [open, setOpen] = useState(false)

  async function handleCreate() {
    setLoading(true)
    try {
      const result = await actionCreateShare(sessionId, 7)
      if (result) {
        setShareUrl(result.url)
        setShareId(result.share.id)
      }
    } finally {
      setLoading(false)
    }
  }

  async function handleCopy() {
    if (!shareUrl) return
    await navigator.clipboard.writeText(shareUrl)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  async function handleRevoke() {
    if (!shareId) return
    await actionRevokeShare(shareId)
    setShareUrl(null)
    setShareId(null)
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="ghost" size="icon" className="h-8 w-8" title="Compartir sesión">
          <Share2 size={15} />
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Compartir sesión</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div
            className="p-3 rounded-md text-sm"
            style={{ background: "var(--muted)", color: "var(--muted-foreground)" }}
          >
            ⚠️ El link es público. No compartas sesiones con información sensible.
            Expira en 7 días.
          </div>

          {!shareUrl ? (
            <Button onClick={handleCreate} disabled={loading} className="w-full">
              {loading ? "Generando..." : "Generar link de compartir"}
            </Button>
          ) : (
            <div className="space-y-2">
              <div
                className="flex items-center gap-2 p-2 rounded-md border text-sm"
                style={{ borderColor: "var(--border)" }}
              >
                <span className="flex-1 truncate text-xs">{shareUrl}</span>
                <Button variant="ghost" size="icon" className="h-7 w-7 shrink-0" onClick={handleCopy}>
                  {copied ? <Check size={13} /> : <Copy size={13} />}
                </Button>
              </div>
              <Button
                variant="outline"
                size="sm"
                className="w-full gap-1.5"
                onClick={handleRevoke}
                style={{ color: "var(--destructive)", borderColor: "var(--destructive)" }}
              >
                <Trash2 size={13} /> Revocar link
              </Button>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
