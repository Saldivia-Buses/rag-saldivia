/**
 * MentionSuggestions — dropdown that appears when typing "@" in the composer.
 * Filters channel members by name and inserts @mention on selection.
 */
"use client"

import { useState, useEffect, useRef } from "react"
import { cn } from "@/lib/utils"

type MemberInfo = { id: number; name: string; email: string }

export function MentionSuggestions({
  query,
  members,
  onSelect,
  onClose,
  visible,
}: {
  query: string
  members: MemberInfo[]
  onSelect: (member: MemberInfo) => void
  onClose: () => void
  visible: boolean
}) {
  const [selectedIndex, setSelectedIndex] = useState(0)
  const listRef = useRef<HTMLDivElement>(null)

  const filtered = query
    ? members.filter((m) =>
        m.name.toLowerCase().includes(query.toLowerCase()) ||
        m.email.toLowerCase().includes(query.toLowerCase())
      )
    : members

  // Reset selection when results change
  useEffect(() => {
    setSelectedIndex(0)
  }, [query])

  if (!visible || filtered.length === 0) return null

  return (
    <>
      <div className="fixed inset-0 z-40" onClick={onClose} />
      <div
        ref={listRef}
        className="absolute bottom-full left-0 mb-1 z-50 w-64 bg-surface border border-border rounded-lg shadow-lg overflow-hidden"
      >
        <div className="py-1 max-h-48 overflow-y-auto">
          {filtered.slice(0, 8).map((member, i) => (
            <button
              key={member.id}
              type="button"
              onClick={() => onSelect(member)}
              onMouseEnter={() => setSelectedIndex(i)}
              className={cn(
                "w-full text-left px-3 py-1.5 text-sm flex items-center gap-2 transition-colors",
                i === selectedIndex ? "bg-accent-subtle text-accent" : "text-fg hover:bg-surface-2",
              )}
            >
              <div className="h-6 w-6 rounded-full bg-accent-subtle text-accent flex items-center justify-center text-xs font-semibold shrink-0">
                {member.name.charAt(0).toUpperCase()}
              </div>
              <div className="flex-1 min-w-0">
                <span className="font-medium truncate block">{member.name}</span>
              </div>
            </button>
          ))}
        </div>
      </div>
    </>
  )
}

/** Parse message content and render @mentions with blue highlight. */
export function renderWithMentions(content: string, members: { id: number; name: string }[]): React.ReactNode[] {
  // Match @Name patterns
  const memberNames = members.map((m) => m.name).sort((a, b) => b.length - a.length)
  if (memberNames.length === 0) return [content]

  const pattern = new RegExp(`@(${memberNames.map(escapeRegex).join("|")})`, "g")
  const parts: React.ReactNode[] = []
  let lastIndex = 0
  let match: RegExpExecArray | null

  while ((match = pattern.exec(content)) !== null) {
    if (match.index > lastIndex) {
      parts.push(content.slice(lastIndex, match.index))
    }
    parts.push(
      <span key={match.index} className="text-accent font-medium bg-accent-subtle/50 rounded px-0.5">
        @{match[1]}
      </span>
    )
    lastIndex = pattern.lastIndex
  }

  if (lastIndex < content.length) {
    parts.push(content.slice(lastIndex))
  }

  return parts.length > 0 ? parts : [content]
}

function escapeRegex(str: string) {
  return str.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")
}
