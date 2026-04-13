"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ExternalLinkIcon, MessageSquareIcon } from "lucide-react";
import Link from "next/link";

interface Suggestion {
  id: string;
  user_id: string;
  origin: string;
  body: string;
  is_read: boolean;
  created_at: string;
  response_count: number;
}

const originLabel: Record<string, string> = {
  web: "Web",
  email: "Email",
  whatsapp: "WhatsApp",
  internal: "Interno",
};

const statusBadge = (s: Suggestion) => {
  if (s.response_count > 0) return { label: "Respondido", variant: "default" as const };
  if (!s.is_read) return { label: "Nuevo", variant: "destructive" as const };
  return { label: "En revisión", variant: "secondary" as const };
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
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Postventa</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              Reclamos y observaciones — {suggestions.length} registros
            </p>
          </div>
          <Link href="/administracion/sugerencias">
            <Button size="sm" variant="outline">
              <ExternalLinkIcon className="size-3.5 mr-1.5" />Ver detalle completo
            </Button>
          </Link>
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-32">ID</TableHead>
                <TableHead>Contenido</TableHead>
                <TableHead className="w-28">Origen</TableHead>
                <TableHead className="w-28">Estado</TableHead>
                <TableHead className="w-28">Fecha</TableHead>
                <TableHead className="w-24" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {suggestions.map((s) => {
                const badge = statusBadge(s);
                return (
                  <TableRow key={s.id}>
                    <TableCell className="font-mono text-xs text-muted-foreground">{s.id.slice(0, 8)}</TableCell>
                    <TableCell className="text-sm max-w-xs">
                      <p className="truncate">{s.body}</p>
                      {s.response_count > 0 && (
                        <p className="text-xs text-muted-foreground mt-0.5">
                          <MessageSquareIcon className="size-3 inline mr-1" />{s.response_count} respuesta{s.response_count !== 1 ? "s" : ""}
                        </p>
                      )}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">{originLabel[s.origin] ?? s.origin}</TableCell>
                    <TableCell><Badge variant={badge.variant}>{badge.label}</Badge></TableCell>
                    <TableCell className="text-sm text-muted-foreground">{fmtDateShort(s.created_at)}</TableCell>
                    <TableCell>
                      <Link href="/administracion/sugerencias">
                        <Button size="sm" variant="ghost">
                          <ExternalLinkIcon className="size-3.5" />
                        </Button>
                      </Link>
                    </TableCell>
                  </TableRow>
                );
              })}
              {suggestions.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
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
