"use client";

import { useEffect, useState, useCallback, useRef, useMemo } from "react";
import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import {
  Bell,
  LayoutDashboard,
  MessageSquare,
  SearchIcon,
  Settings,
  SlidersHorizontal,
} from "lucide-react";
import { MODULE_REGISTRY } from "@/lib/modules/registry";
import { useEnabledModules } from "@/lib/modules/hooks";

const corePages = [
  { label: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
  { label: "Chat", href: "/chat", icon: MessageSquare },
  { label: "Notificaciones", href: "/notifications", icon: Bell },
  { label: "Mi cuenta", href: "/settings", icon: Settings },
  { label: "Configuración", href: "/system-settings", icon: SlidersHorizontal },
];

let globalOpen: ((v: boolean) => void) | null = null;

export function SearchCommand() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const router = useRouter();
  const { data: enabledModules } = useEnabledModules();

  const pages = useMemo(() => {
    const modulePages = (enabledModules ?? [])
      .map((m) => MODULE_REGISTRY[m.id])
      .filter(Boolean)
      .sort((a, b) => a.nav.position - b.nav.position)
      .map((manifest) => ({
        label: manifest.nav.label,
        href: manifest.nav.path,
        icon: manifest.nav.icon,
      }));

    return [...corePages, ...modulePages];
  }, [enabledModules]);

  useEffect(() => {
    globalOpen = setOpen;
    return () => { globalOpen = null; };
  }, []);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setOpen((prev) => !prev);
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  useEffect(() => {
    if (open) {
      setQuery("");
      setSelected(0);
      setTimeout(() => inputRef.current?.focus(), 0);
    }
  }, [open]);

  const filtered = pages.filter((p) =>
    p.label.toLowerCase().includes(query.toLowerCase())
  );

  const navigate = useCallback(
    (href: string) => {
      setOpen(false);
      router.push(href);
    },
    [router]
  );

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setSelected((prev) => Math.min(prev + 1, filtered.length - 1));
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setSelected((prev) => Math.max(prev - 1, 0));
      } else if (e.key === "Enter" && filtered[selected]) {
        navigate(filtered[selected].href);
      } else if (e.key === "Escape") {
        setOpen(false);
      }
    },
    [filtered, selected, navigate]
  );

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[20vh]">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50"
        onClick={() => setOpen(false)}
      />

      {/* Dialog */}
      <div className="relative w-full max-w-md rounded-xl border bg-popover shadow-lg animate-in fade-in-0 zoom-in-95">
        {/* Input */}
        <div className="flex items-center gap-2 border-b px-3 py-2.5">
          <SearchIcon className="size-4 text-muted-foreground" />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setSelected(0);
            }}
            onKeyDown={handleKeyDown}
            placeholder="Buscar páginas..."
            className="flex-1 bg-transparent text-sm outline-none placeholder:text-muted-foreground"
          />
          <kbd className="hidden sm:inline-flex h-5 items-center rounded border bg-muted px-1.5 text-[10px] text-muted-foreground">
            ESC
          </kbd>
        </div>

        {/* Results */}
        <div className="max-h-64 overflow-y-auto p-1.5">
          {filtered.length === 0 && (
            <p className="py-4 text-center text-sm text-muted-foreground">
              Sin resultados.
            </p>
          )}
          {filtered.map((page, i) => {
            const Icon = page.icon;
            return (
              <button
                key={page.href}
                onClick={() => navigate(page.href)}
                className={cn(
                  "flex w-full items-center gap-2.5 rounded-md px-2.5 py-2 text-sm transition-colors",
                  i === selected
                    ? "bg-accent text-accent-foreground"
                    : "text-popover-foreground hover:bg-muted"
                )}
              >
                <Icon className="size-4 shrink-0" />
                {page.label}
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
}

export function useSearchCommand() {
  return {
    open: () => globalOpen?.(true),
  };
}
