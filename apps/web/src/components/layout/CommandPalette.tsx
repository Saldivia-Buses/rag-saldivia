"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { useTheme } from "next-themes"
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command"
import {
  MessageSquare,
  FolderOpen,
  Upload,
  Settings,
  Sun,
  Moon,
  Maximize2,
  Bookmark,
  FileText,
  Users,
} from "lucide-react"
import type { CurrentUser } from "@/lib/auth/current-user"

type Session = { id: string; title: string; collection: string }

type Props = {
  open: boolean
  onClose: () => void
  user: CurrentUser
  onToggleZen?: () => void
}

export function CommandPalette({ open, onClose, user, onToggleZen }: Props) {
  const router = useRouter()
  const { theme, setTheme } = useTheme()
  const [sessions, setSessions] = useState<Session[]>([])
  const [query, setQuery] = useState("")

  useEffect(() => {
    if (!open) return
    fetch("/api/chat/sessions")
      .then((r) => r.json())
      .then((d: { ok: boolean; data?: Session[] }) => {
        if (d.ok && d.data) setSessions(d.data.slice(0, 10))
      })
      .catch(() => {})
  }, [open])

  function go(path: string) {
    router.push(path)
    onClose()
  }

  const filteredSessions = query
    ? sessions.filter((s) => s.title.toLowerCase().includes(query.toLowerCase()))
    : sessions

  return (
    <CommandDialog open={open} onOpenChange={(o) => !o && onClose()}>
      <CommandInput
        placeholder="Buscar acciones, sesiones..."
        value={query}
        onValueChange={setQuery}
      />
      <CommandList>
        <CommandEmpty>Sin resultados.</CommandEmpty>

        <CommandGroup heading="Navegar">
          <CommandItem onSelect={() => go("/chat")}>
            <MessageSquare size={14} className="mr-2" />
            Nueva sesión de chat
          </CommandItem>
          <CommandItem onSelect={() => go("/collections")}>
            <FolderOpen size={14} className="mr-2" />
            Colecciones
          </CommandItem>
          <CommandItem onSelect={() => go("/upload")}>
            <Upload size={14} className="mr-2" />
            Subir documentos
          </CommandItem>
          <CommandItem onSelect={() => go("/saved")}>
            <Bookmark size={14} className="mr-2" />
            Respuestas guardadas
          </CommandItem>
          <CommandItem onSelect={() => go("/audit")}>
            <FileText size={14} className="mr-2" />
            Audit log
          </CommandItem>
          <CommandItem onSelect={() => go("/settings")}>
            <Settings size={14} className="mr-2" />
            Configuración
          </CommandItem>
          {user.role === "admin" && (
            <CommandItem onSelect={() => go("/admin/users")}>
              <Users size={14} className="mr-2" />
              Admin — Usuarios
            </CommandItem>
          )}
        </CommandGroup>

        <CommandSeparator />

        <CommandGroup heading="Apariencia">
          <CommandItem onSelect={() => { setTheme(theme === "dark" ? "light" : "dark"); onClose() }}>
            {theme === "dark" ? <Sun size={14} className="mr-2" /> : <Moon size={14} className="mr-2" />}
            {theme === "dark" ? "Cambiar a light mode" : "Cambiar a dark mode"}
          </CommandItem>
          {onToggleZen && (
            <CommandItem onSelect={() => { onToggleZen(); onClose() }}>
              <Maximize2 size={14} className="mr-2" />
              Modo Zen (Cmd+Shift+Z)
            </CommandItem>
          )}
        </CommandGroup>

        {filteredSessions.length > 0 && (
          <>
            <CommandSeparator />
            <CommandGroup heading="Sesiones recientes">
              {filteredSessions.map((s) => (
                <CommandItem key={s.id} onSelect={() => go(`/chat/${s.id}`)}>
                  <MessageSquare size={14} className="mr-2 opacity-50" />
                  <span className="truncate">{s.title}</span>
                  <span className="ml-auto text-xs opacity-40">{s.collection}</span>
                </CommandItem>
              ))}
            </CommandGroup>
          </>
        )}
      </CommandList>
    </CommandDialog>
  )
}
