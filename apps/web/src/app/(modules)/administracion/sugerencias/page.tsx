"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { ErrorState } from "@/components/erp/error-state";
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

interface Suggestion { id: string; user_id: string; origin: string; body: string; is_read: boolean; created_at: string; updated_at: string; response_count: number; }
interface SuggestionResponse { id: string; suggestion_id: string; user_id: string; body: string; created_at: string; }
interface SuggestionDetail { suggestion: Suggestion; responses: SuggestionResponse[]; }

const suggestionsKey = [...erpKeys.all, "suggestions"] as const;
const unreadKey = [...erpKeys.all, "suggestions", "unread"] as const;

export default function SugerenciasPage() {
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: suggestions = [], isLoading, error } = useQuery({
    queryKey: suggestionsKey,
    queryFn: () => api.get<{ suggestions: Suggestion[] }>("/v1/erp/suggestions?page_size=50"),
    select: (d) => d.suggestions,
  });

  const { data: unreadCount = 0 } = useQuery({
    queryKey: unreadKey,
    queryFn: () => api.get<{ unread: number }>("/v1/erp/suggestions/unread"),
    select: (d) => d.unread,
  });

  const { data: detail } = useQuery({
    queryKey: [...suggestionsKey, selectedId] as const,
    queryFn: () => api.get<SuggestionDetail>(`/v1/erp/suggestions/${selectedId}`),
    enabled: !!selectedId,
  });

  const createMutation = useMutation({
    mutationFn: (data: { origin: string; body: string }) => api.post("/v1/erp/suggestions", data),
    onSuccess: () => {
      toast.success("Sugerencia enviada");
      queryClient.invalidateQueries({ queryKey: suggestionsKey });
      queryClient.invalidateQueries({ queryKey: unreadKey });
      setCreateOpen(false);
    },
    onError: (err) => toast.error("Error al enviar sugerencia", { description: err instanceof Error ? err.message : undefined }),
  });

  const respondMutation = useMutation({
    mutationFn: ({ id, body }: { id: string; body: string }) => api.post(`/v1/erp/suggestions/${id}/respond`, { body }),
    onSuccess: () => {
      toast.success("Respuesta enviada");
      queryClient.invalidateQueries({ queryKey: suggestionsKey });
      queryClient.invalidateQueries({ queryKey: unreadKey });
      if (selectedId) queryClient.invalidateQueries({ queryKey: [...suggestionsKey, selectedId] });
    },
    onError: (err) => toast.error("Error al responder", { description: err instanceof Error ? err.message : undefined }),
  });

  const markReadMutation = useMutation({
    mutationFn: (id: string) => api.patch(`/v1/erp/suggestions/${id}/read`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: suggestionsKey });
      queryClient.invalidateQueries({ queryKey: unreadKey });
    },
  });

  if (error) return <ErrorState message="Error cargando sugerencias" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 flex flex-col gap-6 p-8">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold">Sugerencias</h1>
          {unreadCount > 0 && <Badge variant="destructive">{unreadCount} sin leer</Badge>}
        </div>
        <Button onClick={() => setCreateOpen(true)}>Nueva sugerencia</Button>
        <Dialog open={createOpen} onOpenChange={setCreateOpen}>
          <DialogContent>
            <DialogHeader><DialogTitle>Nueva sugerencia</DialogTitle></DialogHeader>
            <CreateForm onSubmit={(origin, body) => createMutation.mutate({ origin, body })} isPending={createMutation.isPending} />
          </DialogContent>
        </Dialog>
      </div>

      <div className="flex gap-6 flex-1 min-h-0">
        <div className="w-1/3 min-w-[320px]">
          <ScrollArea className="h-[calc(100vh-200px)]">
            <div className="flex flex-col gap-2 pr-4">
              {isLoading ? Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-24 w-full rounded-lg" />) :
                suggestions.length === 0 ? <p className="text-muted-foreground text-sm text-center py-8">No hay sugerencias todavía</p> :
                suggestions.map((s) => (
                  <SuggestionCard key={s.id} suggestion={s} isSelected={selectedId === s.id} onClick={() => {
                    setSelectedId(s.id);
                    if (!s.is_read) markReadMutation.mutate(s.id);
                  }} />
                ))}
            </div>
          </ScrollArea>
        </div>

        <div className="flex-1 min-w-0">
          {detail ? (
            <DetailPanel suggestion={detail.suggestion} responses={detail.responses} onRespond={(body) => respondMutation.mutate({ id: detail.suggestion.id, body })} isSending={respondMutation.isPending} />
          ) : (
            <div className="flex items-center justify-center h-full text-muted-foreground">Seleccioná una sugerencia para ver el detalle</div>
          )}
        </div>
      </div>
    </div>
  );
}

function SuggestionCard({ suggestion: s, isSelected, onClick }: { suggestion: Suggestion; isSelected: boolean; onClick: () => void }) {
  return (
    <Card className={`p-4 cursor-pointer transition-colors hover:bg-accent/50 ${isSelected ? "border-primary bg-accent/30" : ""} ${!s.is_read ? "border-l-4 border-l-primary" : ""}`} onClick={onClick}>
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            {s.origin && <Badge variant="outline" className="text-xs">{s.origin}</Badge>}
            {!s.is_read && <Badge variant="default" className="text-xs">Nueva</Badge>}
          </div>
          <p className="text-sm line-clamp-2">{s.body}</p>
          <div className="flex items-center gap-2 mt-2 text-xs text-muted-foreground">
            <span>{new Date(s.created_at).toLocaleDateString("es-AR")}</span>
            {s.response_count > 0 && <span>{s.response_count} respuesta{s.response_count > 1 ? "s" : ""}</span>}
          </div>
        </div>
      </div>
    </Card>
  );
}

function DetailPanel({ suggestion, responses, onRespond, isSending }: { suggestion: Suggestion; responses: SuggestionResponse[]; onRespond: (body: string) => void; isSending: boolean }) {
  const [replyText, setReplyText] = useState("");

  const handleSubmit = () => {
    if (!replyText.trim()) return;
    onRespond(replyText);
    setReplyText("");
  };

  return (
    <Card className="flex flex-col h-[calc(100vh-200px)]">
      <div className="p-6 border-b">
        <div className="flex items-center gap-2 mb-2">
          {suggestion.origin && <Badge variant="outline">{suggestion.origin}</Badge>}
          <span className="text-xs text-muted-foreground">{new Date(suggestion.created_at).toLocaleDateString("es-AR", { year: "numeric", month: "long", day: "numeric", hour: "2-digit", minute: "2-digit" })}</span>
        </div>
        <p className="text-sm whitespace-pre-wrap">{suggestion.body}</p>
      </div>

      <ScrollArea className="flex-1 p-6">
        {responses.length === 0 ? <p className="text-sm text-muted-foreground text-center py-4">Sin respuestas todavía</p> : (
          <div className="flex flex-col gap-4">
            {responses.map((r) => (
              <div key={r.id} className="flex gap-3">
                <Avatar className="h-8 w-8 mt-0.5"><AvatarFallback className="text-xs">{r.user_id.slice(0, 2).toUpperCase()}</AvatarFallback></Avatar>
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-xs font-medium">{r.user_id.slice(0, 8)}</span>
                    <span className="text-xs text-muted-foreground">{new Date(r.created_at).toLocaleDateString("es-AR", { day: "numeric", month: "short", hour: "2-digit", minute: "2-digit" })}</span>
                  </div>
                  <p className="text-sm whitespace-pre-wrap">{r.body}</p>
                </div>
              </div>
            ))}
          </div>
        )}
      </ScrollArea>

      <Separator />
      <div className="p-4 flex gap-2">
        <Textarea placeholder="Escribir respuesta..." value={replyText} onChange={(e) => setReplyText(e.target.value)} className="flex-1 min-h-[60px] max-h-[120px]" onKeyDown={(e) => { if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); handleSubmit(); } }} />
        <Button onClick={handleSubmit} disabled={!replyText.trim() || isSending} className="self-end">{isSending ? "..." : "Responder"}</Button>
      </div>
    </Card>
  );
}

const TIPOS = [
  { value: "bug", label: "Bug — algo no funciona" },
  { value: "error", label: "Error — vi un mensaje raro" },
  { value: "sugerencia", label: "Sugerencia — mejora a la app" },
  { value: "comentario", label: "Comentario — feedback general" },
];

function CreateForm({ onSubmit, isPending }: { onSubmit: (origin: string, body: string) => void; isPending: boolean }) {
  const [origin, setOrigin] = useState("sugerencia");
  const [body, setBody] = useState("");
  return (
    <form onSubmit={(e) => { e.preventDefault(); if (body.trim()) onSubmit(origin, body); }} className="flex flex-col gap-4">
      <div>
        <Label htmlFor="origin">Tipo</Label>
        <select
          id="origin"
          value={origin}
          onChange={(e) => setOrigin(e.target.value)}
          className="border-input bg-background mt-1.5 h-9 w-full rounded-md border px-3 text-sm"
        >
          {TIPOS.map((t) => (
            <option key={t.value} value={t.value}>{t.label}</option>
          ))}
        </select>
      </div>
      <div>
        <Label htmlFor="body">Mensaje</Label>
        <Textarea
          id="body"
          placeholder="Contá qué pasó (incluí pasos para reproducirlo si es un bug, o detalles del cambio que pedís)..."
          value={body}
          onChange={(e) => setBody(e.target.value)}
          className="mt-1.5 min-h-[120px]"
        />
      </div>
      <Button type="submit" disabled={!body.trim() || isPending}>
        {isPending ? "Enviando..." : "Enviar"}
      </Button>
    </form>
  );
}
