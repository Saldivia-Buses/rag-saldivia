"use client"

import { Lightbulb } from "lucide-react"

type Props = {
  questions: string[]
  onSelect: (question: string) => void
}

export function RelatedQuestions({ questions, onSelect }: Props) {
  if (!questions || questions.length === 0) return null

  return (
    <div className="mt-3 space-y-1.5">
      <p
        className="text-xs font-medium flex items-center gap-1"
        style={{ color: "var(--muted-foreground)" }}
      >
        <Lightbulb size={11} />
        Preguntas relacionadas
      </p>
      <div className="flex flex-wrap gap-1.5">
        {questions.map((q, i) => (
          <button
            key={i}
            onClick={() => onSelect(q)}
            className="px-2.5 py-1 rounded-full text-xs border transition-colors hover:opacity-80 text-left"
            style={{
              borderColor: "var(--border)",
              background: "var(--background)",
              color: "var(--foreground)",
            }}
          >
            {q}
          </button>
        ))}
      </div>
    </div>
  )
}
