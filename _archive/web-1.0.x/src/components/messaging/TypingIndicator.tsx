/**
 * Typing indicator — "Fulano está escribiendo..." with animated dots.
 */
"use client"

export function TypingIndicator({
  typingUsers,
}: {
  typingUsers: Array<{ userId: number; displayName: string }>
}) {
  if (typingUsers.length === 0) return null

  const names = typingUsers.map((u) => u.displayName)
  let text: string
  if (names.length === 1) {
    text = `${names[0]} está escribiendo`
  } else if (names.length === 2) {
    text = `${names[0]} y ${names[1]} están escribiendo`
  } else {
    text = `${names[0]} y ${names.length - 1} más están escribiendo`
  }

  return (
    <div className="px-4 py-1.5 text-xs text-fg-subtle flex items-center gap-1.5">
      <span>{text}</span>
      <span className="inline-flex gap-0.5" aria-hidden>
        <span className="animate-bounce h-1 w-1 rounded-full bg-fg-subtle" style={{ animationDelay: "0ms" }} />
        <span className="animate-bounce h-1 w-1 rounded-full bg-fg-subtle" style={{ animationDelay: "150ms" }} />
        <span className="animate-bounce h-1 w-1 rounded-full bg-fg-subtle" style={{ animationDelay: "300ms" }} />
      </span>
    </div>
  )
}
