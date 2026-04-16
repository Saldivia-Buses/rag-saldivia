"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";
import { api } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ArrowLeft, RotateCcw, Trash2 } from "lucide-react";

interface DeadEventDetail {
  id: string;
  original_subject: string;
  original_stream: string;
  consumer_name: string;
  tenant_id: string | null;
  event_type: string | null;
  delivery_count: number;
  last_error: string;
  dead_at: string;
  envelope: Record<string, unknown>;
  replay_count: number;
  last_replayed_at: string | null;
  dropped_at: string | null;
}

export default function DLQDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const qc = useQueryClient();

  const { data: event, isLoading, error } = useQuery<DeadEventDetail>({
    queryKey: ["dlq", params.id],
    queryFn: () => api.get<DeadEventDetail>(`/v1/admin/dlq/${params.id}`),
  });

  const replayMutation = useMutation({
    mutationFn: () => api.post(`/v1/admin/dlq/${params.id}/replay`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["dlq"] }),
  });

  const dropMutation = useMutation({
    mutationFn: () => api.post(`/v1/admin/dlq/${params.id}/drop`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["dlq"] });
      router.push("/admin/dlq");
    },
  });

  if (isLoading) return <p className="p-6 text-muted-foreground">Cargando...</p>;
  if (error) return <p className="p-6 text-destructive">Error: {String(error)}</p>;
  if (!event) return null;

  return (
    <div className="space-y-6 p-6 max-w-4xl">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="sm" onClick={() => router.push("/admin/dlq")}>
          <ArrowLeft className="h-4 w-4 mr-1" /> Volver
        </Button>
        <h1 className="text-lg font-semibold">Dead Event</h1>
        {event.dropped_at && <Badge variant="destructive">Descartado</Badge>}
      </div>

      <div className="grid grid-cols-2 gap-4">
        <Card>
          <CardHeader><CardTitle className="text-sm">Metadata</CardTitle></CardHeader>
          <CardContent className="space-y-2 text-sm">
            <Row label="ID" value={event.id} />
            <Row label="Subject" value={event.original_subject} />
            <Row label="Stream" value={event.original_stream} />
            <Row label="Consumer" value={event.consumer_name} />
            <Row label="Tenant" value={event.tenant_id ?? "—"} />
            <Row label="Tipo" value={event.event_type ?? "—"} />
            <Row label="Entregas" value={String(event.delivery_count)} />
            <Row label="Muerto" value={new Date(event.dead_at).toLocaleString("es-AR")} />
            <Row label="Replays" value={String(event.replay_count)} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle className="text-sm">Error</CardTitle></CardHeader>
          <CardContent>
            <pre className="text-xs bg-muted p-3 rounded-md whitespace-pre-wrap break-words">
              {event.last_error || "Sin error registrado"}
            </pre>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader><CardTitle className="text-sm">Envelope completo</CardTitle></CardHeader>
        <CardContent>
          <pre className="text-xs bg-muted p-3 rounded-md overflow-x-auto max-h-96">
            {JSON.stringify(event.envelope, null, 2)}
          </pre>
        </CardContent>
      </Card>

      {!event.dropped_at && (
        <div className="flex gap-3">
          <Button
            onClick={() => replayMutation.mutate()}
            disabled={replayMutation.isPending}
          >
            <RotateCcw className="h-4 w-4 mr-2" />
            {event.replay_count >= 3
              ? `Replay (${event.replay_count} previos)`
              : "Replay"}
          </Button>
          <Button
            variant="destructive"
            onClick={() => dropMutation.mutate()}
            disabled={dropMutation.isPending}
          >
            <Trash2 className="h-4 w-4 mr-2" /> Descartar
          </Button>
        </div>
      )}
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-mono">{value}</span>
    </div>
  );
}
