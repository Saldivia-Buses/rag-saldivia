"use client";

import { useState, useEffect, useCallback } from "react";
import { api } from "@/lib/api/client";
import { wsManager } from "@/lib/ws/manager";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Separator } from "@/components/ui/separator";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";

// ─── Types ──────────────────────────────────────────────────────────────────

interface Suggestion {
  id: string;
  user_id: string;
  origin: string;
  body: string;
  is_read: boolean;
  created_at: string;
  updated_at: string;
  response_count: number;
}

interface SuggestionResponse {
  id: string;
  suggestion_id: string;
  user_id: string;
  body: string;
  created_at: string;
}

// ─── Page ───────────────────────────────────────────────────────────────────

export default function SugerenciasPage() {
  const [suggestions, setSuggestions] = useState<Suggestion[]>([]);
  const [selected, setSelected] = useState<Suggestion | null>(null);
  const [responses, setResponses] = useState<SuggestionResponse[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);

  // Fetch suggestions list
  const fetchSuggestions = useCallback(async () => {
    try {
      const data = await api.get<{ suggestions: Suggestion[] }>("/v1/erp/suggestions?page_size=50");
      setSuggestions(data.suggestions);
    } catch (err) {
      console.error("Failed to fetch suggestions:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  // Fetch unread count
  const fetchUnread = useCallback(async () => {
    try {
      const data = await api.get<{ unread: number }>("/v1/erp/suggestions/unread");
      setUnreadCount(data.unread);
    } catch {}
  }, []);

  // Fetch detail + responses
  const fetchDetail = useCallback(async (id: string) => {
    try {
      const data = await api.get<{ suggestion: Suggestion; responses: SuggestionResponse[] }>(
        `/v1/erp/suggestions/${id}`
      );
      setSelected(data.suggestion);
      setResponses(data.responses);
    } catch (err) {
      console.error("Failed to fetch suggestion detail:", err);
    }
  }, []);

  // Initial load
  useEffect(() => {
    fetchSuggestions();
    fetchUnread();
  }, [fetchSuggestions, fetchUnread]);

  // Real-time: listen for suggestion updates via WebSocket
  useEffect(() => {
    const handler = () => {
      fetchSuggestions();
      fetchUnread();
      if (selected) fetchDetail(selected.id);
    };
    const unsubscribe = wsManager.subscribe("erp_suggestions", handler);
    return unsubscribe;
  }, [selected, fetchSuggestions, fetchUnread, fetchDetail]);

  // Create suggestion
  const handleCreate = async (origin: string, body: string) => {
    await api.post("/v1/erp/suggestions", { origin, body });
    setCreateOpen(false);
    fetchSuggestions();
    fetchUnread();
  };

  // Respond to suggestion
  const handleRespond = async (suggestionId: string, body: string) => {
    await api.post(`/v1/erp/suggestions/${suggestionId}/respond`, { body });
    fetchDetail(suggestionId);
    fetchSuggestions();
    fetchUnread();
  };

  // Mark as read
  const handleMarkRead = async (id: string) => {
    await api.patch(`/v1/erp/suggestions/${id}/read`);
    fetchSuggestions();
    fetchUnread();
  };

  return (
    <div className="flex-1 flex flex-col gap-6 p-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold">Sugerencias</h1>
          {unreadCount > 0 && (
            <Badge variant="destructive">{unreadCount} sin leer</Badge>
          )}
        </div>
        <Button onClick={() => setCreateOpen(true)}>Nueva sugerencia</Button>
        <Dialog open={createOpen} onOpenChange={setCreateOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Nueva sugerencia</DialogTitle>
            </DialogHeader>
            <CreateForm onSubmit={handleCreate} />
          </DialogContent>
        </Dialog>
      </div>

      {/* Content: list + detail */}
      <div className="flex gap-6 flex-1 min-h-0">
        {/* List */}
        <div className="w-1/3 min-w-[320px]">
          <ScrollArea className="h-[calc(100vh-200px)]">
            <div className="flex flex-col gap-2 pr-4">
              {loading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <Skeleton key={i} className="h-24 w-full rounded-lg" />
                ))
              ) : suggestions.length === 0 ? (
                <p className="text-muted-foreground text-sm text-center py-8">
                  No hay sugerencias todavia
                </p>
              ) : (
                suggestions.map((s) => (
                  <SuggestionCard
                    key={s.id}
                    suggestion={s}
                    isSelected={selected?.id === s.id}
                    onClick={() => {
                      fetchDetail(s.id);
                      if (!s.is_read) handleMarkRead(s.id);
                    }}
                  />
                ))
              )}
            </div>
          </ScrollArea>
        </div>

        {/* Detail panel */}
        <div className="flex-1 min-w-0">
          {selected ? (
            <DetailPanel
              suggestion={selected}
              responses={responses}
              onRespond={handleRespond}
            />
          ) : (
            <div className="flex items-center justify-center h-full text-muted-foreground">
              Selecciona una sugerencia para ver el detalle
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── Components ─────────────────────────────────────────────────────────────

function SuggestionCard({
  suggestion: s,
  isSelected,
  onClick,
}: {
  suggestion: Suggestion;
  isSelected: boolean;
  onClick: () => void;
}) {
  return (
    <Card
      className={`p-4 cursor-pointer transition-colors hover:bg-accent/50 ${
        isSelected ? "border-primary bg-accent/30" : ""
      } ${!s.is_read ? "border-l-4 border-l-primary" : ""}`}
      onClick={onClick}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            {s.origin && (
              <Badge variant="outline" className="text-xs">{s.origin}</Badge>
            )}
            {!s.is_read && <Badge variant="default" className="text-xs">Nueva</Badge>}
          </div>
          <p className="text-sm line-clamp-2">{s.body}</p>
          <div className="flex items-center gap-2 mt-2 text-xs text-muted-foreground">
            <span>{new Date(s.created_at).toLocaleDateString("es-AR")}</span>
            {s.response_count > 0 && (
              <span>{s.response_count} respuesta{s.response_count > 1 ? "s" : ""}</span>
            )}
          </div>
        </div>
      </div>
    </Card>
  );
}

function DetailPanel({
  suggestion,
  responses,
  onRespond,
}: {
  suggestion: Suggestion;
  responses: SuggestionResponse[];
  onRespond: (id: string, body: string) => void;
}) {
  const [replyText, setReplyText] = useState("");
  const [sending, setSending] = useState(false);

  const handleSubmit = async () => {
    if (!replyText.trim()) return;
    setSending(true);
    try {
      await onRespond(suggestion.id, replyText);
      setReplyText("");
    } finally {
      setSending(false);
    }
  };

  return (
    <Card className="flex flex-col h-[calc(100vh-200px)]">
      {/* Header */}
      <div className="p-6 border-b">
        <div className="flex items-center gap-2 mb-2">
          {suggestion.origin && <Badge variant="outline">{suggestion.origin}</Badge>}
          <span className="text-xs text-muted-foreground">
            {new Date(suggestion.created_at).toLocaleDateString("es-AR", {
              year: "numeric", month: "long", day: "numeric", hour: "2-digit", minute: "2-digit"
            })}
          </span>
        </div>
        <p className="text-sm whitespace-pre-wrap">{suggestion.body}</p>
      </div>

      {/* Responses thread */}
      <ScrollArea className="flex-1 p-6">
        {responses.length === 0 ? (
          <p className="text-sm text-muted-foreground text-center py-4">
            Sin respuestas todavia
          </p>
        ) : (
          <div className="flex flex-col gap-4">
            {responses.map((r) => (
              <div key={r.id} className="flex gap-3">
                <Avatar className="h-8 w-8 mt-0.5">
                  <AvatarFallback className="text-xs">
                    {r.user_id.slice(0, 2).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-xs font-medium">{r.user_id.slice(0, 8)}</span>
                    <span className="text-xs text-muted-foreground">
                      {new Date(r.created_at).toLocaleDateString("es-AR", {
                        day: "numeric", month: "short", hour: "2-digit", minute: "2-digit"
                      })}
                    </span>
                  </div>
                  <p className="text-sm whitespace-pre-wrap">{r.body}</p>
                </div>
              </div>
            ))}
          </div>
        )}
      </ScrollArea>

      <Separator />

      {/* Reply input */}
      <div className="p-4 flex gap-2">
        <Textarea
          placeholder="Escribir respuesta..."
          value={replyText}
          onChange={(e) => setReplyText(e.target.value)}
          className="flex-1 min-h-[60px] max-h-[120px]"
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              handleSubmit();
            }
          }}
        />
        <Button
          onClick={handleSubmit}
          disabled={!replyText.trim() || sending}
          className="self-end"
        >
          {sending ? "..." : "Responder"}
        </Button>
      </div>
    </Card>
  );
}

function CreateForm({ onSubmit }: { onSubmit: (origin: string, body: string) => void }) {
  const [origin, setOrigin] = useState("");
  const [body, setBody] = useState("");
  const [sending, setSending] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!body.trim()) return;
    setSending(true);
    try {
      await onSubmit(origin, body);
    } finally {
      setSending(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
      <div>
        <Label htmlFor="origin">Area / Origen</Label>
        <Input
          id="origin"
          placeholder="Ej: Produccion, Administracion..."
          value={origin}
          onChange={(e) => setOrigin(e.target.value)}
        />
      </div>
      <div>
        <Label htmlFor="body">Sugerencia</Label>
        <Textarea
          id="body"
          placeholder="Describe tu sugerencia..."
          value={body}
          onChange={(e) => setBody(e.target.value)}
          className="min-h-[120px]"
          required
        />
      </div>
      <Button type="submit" disabled={!body.trim() || sending}>
        {sending ? "Enviando..." : "Enviar sugerencia"}
      </Button>
    </form>
  );
}
