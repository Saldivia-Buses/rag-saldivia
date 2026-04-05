/**
 * Command palette for messaging — Cmd+K to search channels, messages, users.
 * Adapted from _archive/components/layout/CommandPalette.tsx.
 */
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
  Hash,
  MessageSquare,
  FolderOpen,
  Settings,
  Sun,
  Moon,
  ShieldCheck,
  Search,
  MessagesSquare,
} from "lucide-react"

type ChannelInfo = { id: string; name: string | null; type: string }

type Props = {
  open: boolean
  onClose: () => void
  channels: ChannelInfo[]
  isAdmin: boolean
}

type SearchResult = {
  id: string
  channelId: string
  userId: number
  content: string
  createdAt: number
}

export function CommandPalette({ open, onClose, channels, isAdmin }: Props) {
  const router = useRouter()
  const { theme, setTheme } = useTheme()
  const [query, setQuery] = useState("")
  const [searchResults, setSearchResults] = useState<SearchResult[]>([])
  const [searching, setSearching] = useState(false)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (!open) {
      setQuery("")
      setSearchResults([])
    }
  }, [open])

  // Debounced FTS5 search
  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current)

    if (query.trim().length < 2) {
      setSearchResults([])
      setSearching(false)
      return
    }

    setSearching(true)
    debounceRef.current = setTimeout(() => {
      fetch(`/api/messaging/search?q=${encodeURIComponent(query)}`)
        .then((r) => r.json())
        .then((d: { ok: boolean; data?: SearchResult[] }) => {
          if (d.ok) setSearchResults(d.data ?? [])
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

  const filteredChannels = query
    ? channels.filter((c) => c.name?.toLowerCase().includes(query.toLowerCase()))
    : channels.slice(0, 8)

  const showSearch = query.trim().length >= 2

  return (
    <CommandDialog open={open} onOpenChange={(o) => !o && onClose()}>
      <CommandInput
        placeholder="Buscar canales, mensajes..."
        value={query}
        onValueChange={setQuery}
      />
      <CommandList>
        <CommandEmpty>
          {searching ? "Buscando..." : "Sin resultados."}
        </CommandEmpty>

        {/* FTS5 message search results */}
        {showSearch && searchResults.length > 0 && (
          <>
            <CommandGroup heading={`Mensajes con "${query}"`}>
              {searchResults.slice(0, 10).map((r) => (
                <CommandItem key={r.id} onSelect={() => go(`/messaging/${r.channelId}`)}>
                  <Search size={13} className="mr-2 opacity-60" />
                  <span className="truncate text-sm">{r.content.slice(0, 80)}</span>
                </CommandItem>
              ))}
            </CommandGroup>
            <CommandSeparator />
          </>
        )}

        {/* Channels */}
        {filteredChannels.length > 0 && (
          <CommandGroup heading="Canales">
            {filteredChannels.map((c) => (
              <CommandItem key={c.id} onSelect={() => go(`/messaging/${c.id}`)}>
                <Hash size={14} className="mr-2 opacity-60" />
                <span className="truncate">{c.name ?? "Canal"}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        <CommandSeparator />

        <CommandGroup heading="Navegar">
          <CommandItem onSelect={() => go("/chat")}>
            <MessageSquare size={14} className="mr-2" />
            Chat RAG
          </CommandItem>
          <CommandItem onSelect={() => go("/messaging")}>
            <MessagesSquare size={14} className="mr-2" />
            Mensajería
          </CommandItem>
          <CommandItem onSelect={() => go("/collections")}>
            <FolderOpen size={14} className="mr-2" />
            Colecciones
          </CommandItem>
          <CommandItem onSelect={() => go("/settings")}>
            <Settings size={14} className="mr-2" />
            Configuración
          </CommandItem>
          {isAdmin && (
            <CommandItem onSelect={() => go("/admin")}>
              <ShieldCheck size={14} className="mr-2" />
              Admin
            </CommandItem>
          )}
        </CommandGroup>

        <CommandSeparator />

        <CommandGroup heading="Apariencia">
          <CommandItem onSelect={() => { setTheme(theme === "dark" ? "light" : "dark"); onClose() }}>
            {theme === "dark" ? <Sun size={14} className="mr-2" /> : <Moon size={14} className="mr-2" />}
            {theme === "dark" ? "Cambiar a light mode" : "Cambiar a dark mode"}
          </CommandItem>
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  )
}
