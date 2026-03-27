"use client"

import { useState } from "react"
import { LayoutTemplate, ChevronDown } from "lucide-react"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Separator } from "@/components/ui/separator"

type Template = {
  id: number
  title: string
  prompt: string
  focusMode: string
}

type Props = {
  templates: Template[]
  onSelect: (prompt: string, focusMode: string) => void
  disabled?: boolean
}

export function PromptTemplates({ templates, onSelect, disabled }: Props) {
  const [open, setOpen] = useState(false)

  if (templates.length === 0) return null

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          disabled={disabled}
          className="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs border transition-colors hover:opacity-80 disabled:opacity-40"
          style={{ borderColor: "var(--border)", color: "var(--muted-foreground)" }}
        >
          <LayoutTemplate size={11} />
          Templates
          <ChevronDown size={10} />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-64 p-2">
        <p className="text-xs font-medium px-1 pb-1" style={{ color: "var(--muted-foreground)" }}>
          Templates de query
        </p>
        <Separator className="mb-1" />
        {templates.map((t) => (
          <button
            key={t.id}
            onClick={() => {
              onSelect(t.prompt, t.focusMode)
              setOpen(false)
            }}
            className="w-full text-left px-2 py-2 rounded-md text-sm transition-colors hover:opacity-80 space-y-0.5"
            style={{ background: "transparent" }}
          >
            <p className="font-medium text-sm truncate" style={{ color: "var(--foreground)" }}>
              {t.title}
            </p>
            <p className="text-xs truncate" style={{ color: "var(--muted-foreground)" }}>
              {t.prompt}
            </p>
          </button>
        ))}
      </PopoverContent>
    </Popover>
  )
}
