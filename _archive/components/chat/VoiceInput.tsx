"use client"

import { useEffect, useRef, useState } from "react"
import { Mic, MicOff } from "lucide-react"
import { Button } from "@/components/ui/button"

// Web Speech API no tiene tipos en el DOM estándar de TypeScript
// Se declaran localmente para evitar errores de type-check
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

type Props = {
  onTranscript: (text: string) => void
  disabled?: boolean
}

/** Fallback graceful: si el browser no soporta SpeechRecognition, no renderiza nada */
function isSpeechSupported(): boolean {
  if (typeof window === "undefined") return false
  return "SpeechRecognition" in window || "webkitSpeechRecognition" in window
}

export function VoiceInput({ onTranscript, disabled }: Props) {
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
    <Button
      type="button"
      variant="ghost"
      size="icon"
      onClick={toggle}
      disabled={disabled}
      title={listening ? "Detener grabación" : "Dictar mensaje"}
      className="h-10 w-10"
      style={listening ? { color: "var(--destructive)" } : {}}
    >
      {listening ? <MicOff size={16} /> : <Mic size={16} />}
    </Button>
  )
}
