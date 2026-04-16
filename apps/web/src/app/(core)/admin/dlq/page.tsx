"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { AlertTriangle, RotateCcw, Trash2 } from "lucide-react";
import Link from "next/link";

interface DeadEvent {
  id: string;
  original_subject: string;
  original_stream: string;
  consumer_name: string;
  tenant_id: string | null;
  event_type: string | null;
  delivery_count: number;
  last_error: string;
  dead_at: string;
  replay_count: number;
  last_replayed_at: string | null;
  dropped_at: string | null;
}

function formatAgo(dateStr: string): string {
  const d = new Date(dateStr);
  const diff = Date.now() - d.getTime();
  const min = Math.floor(diff / 60000);
  if (min < 60) return `${min}m`;
  const h = Math.floor(min / 60);
  if (h < 24) return `${h}h`;
  return `${Math.floor(h / 24)}d`;
}

export default function DLQPage() {
  const [consumer, setConsumer] = useState("");
  const [tenant, setTenant] = useState("");
  const qc = useQueryClient();

  const { data, isLoading, error } = useQuery({
    queryKey: ["dlq", consumer, tenant],
    queryFn: () => {
      const params = new URLSearchParams();
      if (consumer) params.set("consumer", consumer);
      if (tenant) params.set("tenant", tenant);
      params.set("limit", "100");
      return api.get(`/v1/admin/dlq?${params}`).then((r) => r.json());
    },
  });

  const replayMutation = useMutation({
    mutationFn: (id: string) =>
      api.post(`/v1/admin/dlq/${id}/replay`).then((r) => r.json()),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["dlq"] }),
  });

  const dropMutation = useMutation({
    mutationFn: (id: string) =>
      api.post(`/v1/admin/dlq/${id}/drop`).then((r) => r.json()),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["dlq"] }),
  });

  const events: DeadEvent[] = data?.events ?? [];

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center gap-2">
        <AlertTriangle className="h-5 w-5 text-destructive" />
        <h1 className="text-xl font-semibold">Dead Letter Queue</h1>
        <Badge variant="outline">{events.length}</Badge>
      </div>

      <div className="flex gap-3">
        <Input
          placeholder="Filtrar por consumer..."
          value={consumer}
          onChange={(e) => setConsumer(e.target.value)}
          className="max-w-xs"
        />
        <Input
          placeholder="Filtrar por tenant..."
          value={tenant}
          onChange={(e) => setTenant(e.target.value)}
          className="max-w-xs"
        />
      </div>

      {error && (
        <p className="text-sm text-destructive">
          Error cargando DLQ: {String(error)}
        </p>
      )}

      {isLoading ? (
        <p className="text-sm text-muted-foreground">Cargando...</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Tipo</TableHead>
              <TableHead>Consumer</TableHead>
              <TableHead>Tenant</TableHead>
              <TableHead>Error</TableHead>
              <TableHead>Entregas</TableHead>
              <TableHead>Hace</TableHead>
              <TableHead>Replays</TableHead>
              <TableHead className="text-right">Acciones</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {events.map((e) => (
              <TableRow key={e.id}>
                <TableCell>
                  <Link
                    href={`/admin/dlq/${e.id}`}
                    className="font-mono text-xs hover:underline"
                  >
                    {e.event_type ?? e.original_subject}
                  </Link>
                </TableCell>
                <TableCell className="text-xs">{e.consumer_name}</TableCell>
                <TableCell className="text-xs">
                  {e.tenant_id ?? "—"}
                </TableCell>
                <TableCell
                  className="max-w-[200px] truncate text-xs text-muted-foreground"
                  title={e.last_error}
                >
                  {e.last_error}
                </TableCell>
                <TableCell>{e.delivery_count}</TableCell>
                <TableCell className="text-xs">
                  {formatAgo(e.dead_at)}
                </TableCell>
                <TableCell>{e.replay_count}</TableCell>
                <TableCell className="text-right space-x-1">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => replayMutation.mutate(e.id)}
                    disabled={replayMutation.isPending}
                    title="Re-ejecutar contra estado actual"
                  >
                    <RotateCcw className="h-3 w-3" />
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => dropMutation.mutate(e.id)}
                    disabled={dropMutation.isPending}
                    title="Descartar definitivamente"
                  >
                    <Trash2 className="h-3 w-3 text-destructive" />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
            {events.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                  No hay eventos muertos
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      )}
    </div>
  );
}
