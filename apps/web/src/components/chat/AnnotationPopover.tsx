"use client"

import { useEffect, useRef, useState } from "react"
import { Bookmark, MessageSquare, HelpCircle, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { actionSaveAnnotation } from "@/app/actions/chat"

type Props = {
  sessionId: string
  messageId?: number
  onAskAbout?: (text: string) => void
  children: React.ReactNode
}

export function AnnotationPopover({ sessionId, messageId, onAskAbout, children }: Props) {
  const [popover, setPopover] = useState<{ x: number; y: number; text: string } | null>(null)
  const [noteInput, setNoteInput] = useState("")
  const [saved, setSaved] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleMouseUp() {
      const selection = window.getSelection()
      if (!selection || selection.isCollapsed) {
        setPopover(null)
        return
      }
      const text = selection.toString().trim()
      if (!text || text.length < 3) return

      // Verificar que la selección está dentro de este contenedor
      const range = selection.getRangeAt(0)
      if (!containerRef.current?.contains(range.commonAncestorContainer)) return

      const rect = range.getBoundingClientRect()
      const containerRect = containerRef.current?.getBoundingClientRect()
      if (!containerRect) return

      setPopover({
        x: rect.left - containerRect.left + rect.width / 2,
        y: rect.top - containerRect.top - 8,
        text,
      })
      setSaved(false)
      setNoteInput("")
    }

    document.addEventListener("mouseup", handleMouseUp)
    return () => document.removeEventListener("mouseup", handleMouseUp)
  }, [])

  async function handleSave() {
    if (!popover) return
    await actionSaveAnnotation({
      sessionId,
      messageId,
      selectedText: popover.text,
      note: noteInput || undefined,
    })
    setSaved(true)
    setTimeout(() => setPopover(null), 1200)
  }

  function handleAsk() {
    if (!popover) return
    onAskAbout?.(popover.text)
    setPopover(null)
    window.getSelection()?.removeAllRanges()
  }

  return (
    <div ref={containerRef} className="relative">
      {children}
      {popover && (
        <div
          className="absolute z-50 rounded-lg border shadow-lg p-2 space-y-1.5"
          style={{
            left: popover.x,
            top: popover.y,
            transform: "translate(-50%, -100%)",
            background: "var(--background)",
            borderColor: "var(--border)",
            minWidth: 180,
          }}
        >
          {saved ? (
            <p className="text-xs px-1 text-center" style={{ color: "var(--accent)" }}>
              ✓ Guardado
            </p>
          ) : (
            <>
              <div className="flex gap-1">
                <Button variant="ghost" size="sm" className="h-7 text-xs flex-1 gap-1" onClick={handleSave}>
                  <Bookmark size={11} /> Guardar
                </Button>
                {onAskAbout && (
                  <Button variant="ghost" size="sm" className="h-7 text-xs flex-1 gap-1" onClick={handleAsk}>
                    <HelpCircle size={11} /> Preguntar
                  </Button>
                )}
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setPopover(null)}>
                  <X size={11} />
                </Button>
              </div>
              <div className="flex gap-1">
                <input
                  value={noteInput}
                  onChange={(e) => setNoteInput(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleSave()}
                  placeholder="Nota (opcional)"
                  className="flex-1 text-xs px-2 py-1 rounded border outline-none"
                  style={{ borderColor: "var(--border)", background: "var(--muted)", color: "var(--foreground)" }}
                />
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleSave}>
                  <MessageSquare size={11} />
                </Button>
              </div>
            </>
          )}
        </div>
      )}
    </div>
  )
}
