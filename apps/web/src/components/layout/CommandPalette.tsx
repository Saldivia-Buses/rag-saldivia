"use client"

import { useEffect, useRef, useState } from "react"
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
  Search,
  LayoutTemplate,
} from "lucide-react"
import type { CurrentUser } from "@/lib/auth/current-user"
import type { SearchResult } from "@rag-saldivia/db"

type Session = { id: string; title: string; collection: string }

const RESULT_ICONS: Record<string, React.ReactNode> = {
  session: <MessageSquare size={13} className="mr-2 opacity-60" />,
  message: <FileText size={13} className="mr-2 opacity-60" />,
  saved: <Bookmark size={13} className="mr-2 opacity-60" />,
  template: <LayoutTemplate size={13} className="mr-2 opacity-60" />,
}

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
  const [searchResults, setSearchResults] = useState<SearchResult[]>([])
  const [searching, setSearching] = useState(false)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (!open) { setQuery(""); setSearchResults([]); return }
    fetch("/api/chat/sessions")
      .then((r) => r.json())
      .then((d: { ok: boolean; data?: Session[] }) => {
        if (d.ok && d.data) setSessions(d.data.slice(0, 8))
      })
      .catch(() => {})
  }, [open])

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current)

    if (query.trim().length < 2) {
      setSearchResults([])
      setSearching(false)
      return
    }

    setSearching(true)
    debounceRef.current = setTimeout(() => {
      fetch(`/api/search?q=${encodeURIComponent(query)}`)
        .then((r) => r.json())
        .then((d: { ok: boolean; results?: SearchResult[] }) => {
          if (d.ok) setSearchResults(d.results ?? [])
        })
        .catch(() => {})
        .finally(() => setSearching(false))
    }, 300)

    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [query])

  function go(path: string) {
    router.push(path)
    onClose()
  }

  const filteredSessions = query
    ? sessions.filter((s) => s.title.toLowerCase().includes(query.toLowerCase()))
    : sessions

  const showSearch = query.trim().length >= 2

  return (
    <CommandDialog open={open} onOpenChange={(o) => !o && onClose()}>
      <CommandInput
        placeholder="Buscar sesiones, docs, templates, guardados..."
        value={query}
        onValueChange={setQuery}
      />
      <CommandList>
        <CommandEmpty>
          {searching ? "Buscando..." : "Sin resultados."}
        </CommandEmpty>

        {/* Resultados de búsqueda universal */}
        {showSearch && searchResults.length > 0 && (
          <>
            <CommandGroup heading={`Resultados para "${query}"`}>
              {searchResults.map((r) => (
                <CommandItem key={`${r.type}-${r.id}`} onSelect={() => go(r.href)}>
                  {RESULT_ICONS[r.type]}
                  <div className="flex flex-col min-w-0">
                    <span className="truncate text-sm">{r.title}</span>
                    {r.snippet && (
                      <span className="truncate text-xs opacity-50">{r.snippet.replace(/<[^>]+>/g, "")}</span>
                    )}
                  </div>
                  <span className="ml-auto text-xs opacity-40 shrink-0 capitalize">{r.type}</span>
                </CommandItem>
              ))}
            </CommandGroup>
            <CommandSeparator />
          </>
        )}

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

        {!showSearch && filteredSessions.length > 0 && (
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
