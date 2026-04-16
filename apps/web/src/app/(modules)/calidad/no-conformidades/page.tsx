"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const sevColor: Record<string, "default" | "secondary" | "destructive"> = { minor: "secondary", major: "default", critical: "destructive" };

export default function CalidadNCPage() {
  const { data: ncs = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "quality", "nc"] as const,
    queryFn: () => api.get<{ nonconformities: any[] }>("/v1/erp/quality/nc?page_size=50"),
    select: (d) => d.nonconformities,
  });

  if (error) return <ErrorState message="Error cargando no conformidades" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">No Conformidades</h1>
          <p className="text-sm text-muted-foreground mt-0.5">{ncs.length} registros</p>
        </div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader><TableRow>
              <TableHead className="w-20">NC</TableHead><TableHead className="w-28">Fecha</TableHead>
              <TableHead>Descripción</TableHead><TableHead className="w-20">Sev.</TableHead>
              <TableHead className="w-28">Estado</TableHead>
            </TableRow></TableHeader>
            <TableBody>
              {ncs.map((nc: any) => (
                <TableRow key={nc.id}>
                  <TableCell className="font-mono text-sm">{nc.number}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{fmtDateShort(nc.date)}</TableCell>
                  <TableCell className="text-sm truncate max-w-64">{nc.description}</TableCell>
                  <TableCell><Badge variant={sevColor[nc.severity] || "secondary"}>{nc.severity}</Badge></TableCell>
                  <TableCell><Badge variant={nc.status === "closed" ? "default" : "secondary"}>{nc.status}</Badge></TableCell>
                </TableRow>
              ))}
              {ncs.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin no conformidades.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
