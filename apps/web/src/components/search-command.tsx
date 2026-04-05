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
  { label: "Inicio", href: "/inicio", icon: LayoutDashboard },
  { label: "Chat", href: "/chat", icon: MessageSquare },
  { label: "Notificaciones", href: "/notifications", icon: Bell },
  { label: "Mi cuenta", href: "/settings", icon: Settings },
  { label: "Configuración", href: "/system-settings", icon: SlidersHorizontal },
];

/**
 * Inline expanding search — sits fixed at viewport center in the header row.
 * Collapsed: small trigger bar. Expanded: wider input + results dropdown.
 */
export function HeaderSearch() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
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

  // ⌘K shortcut
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

  // Focus input when opening
  useEffect(() => {
    if (open) {
      setQuery("");
      setSelected(0);
      setTimeout(() => inputRef.current?.focus(), 50);
    }
  }, [open]);

  // Close on click outside
  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  const filtered = pages.filter((p) =>
    p.label.toLowerCase().includes(query.toLowerCase()),
  );

  const navigate = useCallback(
    (href: string) => {
      setOpen(false);
      router.push(href);
    },
    [router],
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
    [filtered, selected, navigate],
  );

  return (
    <>
      {/* Backdrop */}
      {open && (
        <div className="fixed inset-0 z-40 bg-black/40" />
      )}

      {/* Search — fixed at viewport center, header row */}
      <div
        ref={containerRef}
        className={cn(
          "fixed top-0 left-1/2 -translate-x-1/2 z-50 flex flex-col items-center",
          "h-14 justify-center",
          open && "h-auto",
        )}
      >
        <div
          className={cn(
            "flex flex-col transition-all duration-200 ease-out mt-1.5",
            open
              ? "w-[28rem] rounded-xl bg-popover shadow-2xl"
              : "w-64 rounded-lg bg-background shadow-md",
          )}
        >
          {/* Input row */}
          <div
            className={cn(
              "flex items-center gap-2 px-3 transition-all",
              open ? "py-2.5" : "py-1.5",
            )}
          >
            <SearchIcon className="size-3.5 text-muted-foreground shrink-0" />
            {open ? (
              <input
                ref={inputRef}
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value);
                  setSelected(0);
                }}
                onKeyDown={handleKeyDown}
                placeholder="Busca ayuda, páginas y más..."
                className="flex-1 bg-transparent text-sm outline-none placeholder:text-muted-foreground"
              />
            ) : (
              <button
                onClick={() => setOpen(true)}
                className="flex flex-1 items-center justify-between text-sm text-muted-foreground bg-transparent"
              >
                <span>Buscar en SDA...</span>
                <kbd className="hidden sm:inline-flex h-5 items-center gap-0.5 rounded bg-background/30 px-1.5 text-[10px] font-medium">
                  ⌘K
                </kbd>
              </button>
            )}
          </div>

          {/* Expanded: results */}
          {open && (
            <div className="max-h-72 overflow-y-auto px-1.5 pb-1.5">
              {filtered.length === 0 && (
                <p className="py-6 text-center text-sm text-muted-foreground">
                  Sin resultados.
                </p>
              )}
              {filtered.length > 0 && (
                <p className="px-2.5 pb-1.5 pt-1 text-xs text-muted-foreground">
                  Páginas
                </p>
              )}
              {filtered.map((page, i) => {
                const Icon = page.icon;
                return (
                  <button
                    key={page.href}
                    onClick={() => navigate(page.href)}
                    className={cn(
                      "flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm transition-colors",
                      i === selected
                        ? "bg-accent text-accent-foreground"
                        : "text-popover-foreground hover:bg-muted",
                    )}
                  >
                    <Icon className="size-4 shrink-0" />
                    {page.label}
                  </button>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </>
  );
}
