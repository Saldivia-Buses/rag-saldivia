"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

interface Suggestion {
  id: string;
  title: string;
  content: string;
  category: string;
  status: string;
  priority: string;
  created_at: string;
}

const priorityBadge: Record<string, { label: string; variant: "default" | "secondary" | "destructive" | "outline" }> = {
  low: { label: "Baja", variant: "secondary" },
  medium: { label: "Media", variant: "outline" },
  high: { label: "Alta", variant: "default" },
  critical: { label: "Crítica", variant: "destructive" },
};

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  open: { label: "Abierto", variant: "secondary" },
  in_progress: { label: "En proceso", variant: "outline" },
  resolved: { label: "Resuelto", variant: "default" },
  closed: { label: "Cerrado", variant: "secondary" },
};

const suggestionsKey = [...erpKeys.all, "suggestions"] as const;

export default function PostventaPage() {
  const { data: suggestions = [], isLoading, error } = useQuery({
    queryKey: suggestionsKey,
    queryFn: () => api.get<{ suggestions: Suggestion[] }>("/v1/erp/suggestions?page_size=50"),
    select: (d) => d.suggestions,
  });

  if (error) return <ErrorState message="Error cargando reclamos" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Postventa</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Sugerencias y reclamos de clientes — {suggestions.length} registros
          </p>
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Título</TableHead>
                <TableHead className="w-36">Categoría</TableHead>
                <TableHead className="w-28">Prioridad</TableHead>
                <TableHead className="w-28">Estado</TableHead>
                <TableHead className="w-28">Fecha</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {suggestions.map((s) => {
                const prio = priorityBadge[s.priority] ?? { label: s.priority, variant: "secondary" as const };
                const stat = statusBadge[s.status] ?? { label: s.status, variant: "secondary" as const };
                return (
                  <TableRow key={s.id}>
                    <TableCell className="text-sm">
                      <p className="font-medium truncate max-w-xs">{s.title}</p>
                      {s.content && (
                        <p className="text-xs text-muted-foreground truncate max-w-xs mt-0.5">{s.content}</p>
                      )}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">{s.category || "—"}</TableCell>
                    <TableCell><Badge variant={prio.variant}>{prio.label}</Badge></TableCell>
                    <TableCell><Badge variant={stat.variant}>{stat.label}</Badge></TableCell>
                    <TableCell className="text-sm text-muted-foreground">{fmtDateShort(s.created_at)}</TableCell>
                  </TableRow>
                );
              })}
              {suggestions.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                    Sin reclamos registrados.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
