/**
 * Voice input — speech-to-text for message composer (es-AR).
 * Adapted from _archive/components/chat/VoiceInput.tsx.
 * Graceful fallback: renders nothing if browser doesn't support SpeechRecognition.
 */
"use client"

import { useEffect, useRef, useState } from "react"
import { Mic, MicOff } from "lucide-react"

type SpeechRecognitionResult = { transcript: string }
type SpeechRecognitionResultList = SpeechRecognitionResult[][]
type SpeechRecognitionEventLocal = { results: SpeechRecognitionResultList }
type SpeechRecognitionInstance = {
  lang: string
  interimResults: boolean
  continuous: boolean
  start: () => void
  stop: () => void
  onresult: ((event: SpeechRecognitionEventLocal) => void) | null
  onend: (() => void) | null
  onerror: (() => void) | null
}
type SpeechRecognitionConstructor = new () => SpeechRecognitionInstance
type WindowWithSpeech = Window & {
  SpeechRecognition?: SpeechRecognitionConstructor
  webkitSpeechRecognition?: SpeechRecognitionConstructor
}

function isSpeechSupported(): boolean {
  if (typeof window === "undefined") return false
  return "SpeechRecognition" in window || "webkitSpeechRecognition" in window
}

export function VoiceInput({
  onTranscript,
  disabled,
}: {
  onTranscript: (text: string) => void
  disabled?: boolean
}) {
  const [supported, setSupported] = useState(false)
  const [listening, setListening] = useState(false)
  const recognitionRef = useRef<SpeechRecognitionInstance | null>(null)

  useEffect(() => {
    setSupported(isSpeechSupported())
  }, [])

  if (!supported) return null

  function toggle() {
    if (listening) {
      recognitionRef.current?.stop()
      setListening(false)
      return
    }

    const w = window as WindowWithSpeech
    const SpeechRecognitionAPI = w.SpeechRecognition ?? w.webkitSpeechRecognition
    if (!SpeechRecognitionAPI) return

    const recognition = new SpeechRecognitionAPI()
    recognition.lang = "es-AR"
    recognition.interimResults = true
    recognition.continuous = false

    recognition.onresult = (event: SpeechRecognitionEventLocal) => {
      const transcript = event.results
        .flat()
        .map((r) => r.transcript)
        .join("")
      onTranscript(transcript)
    }

    recognition.onend = () => setListening(false)
    recognition.onerror = () => setListening(false)

    recognition.start()
    recognitionRef.current = recognition
    setListening(true)
  }

  return (
    <button
      type="button"
      onClick={toggle}
      disabled={disabled}
      title={listening ? "Detener grabación" : "Dictar mensaje"}
      className="flex items-center justify-center rounded-lg transition-colors"
      style={{
        width: "34px",
        height: "34px",
        color: listening ? "var(--destructive)" : "var(--fg-subtle)",
      }}
    >
      {listening ? <MicOff size={16} /> : <Mic size={16} />}
    </button>
  )
}
