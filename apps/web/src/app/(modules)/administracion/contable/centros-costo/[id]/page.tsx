"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate } from "@/lib/erp/format";
import type { CostCenter, JournalEntry } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default function CostCenterDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const centerQ = useQuery({
    queryKey: erpKeys.costCenter(id),
    queryFn: () => api.get<CostCenter>(`/v1/erp/accounting/cost-centers/${id}`),
    enabled: !!id,
  });

  const siblingsQ = useQuery({
    queryKey: [...erpKeys.all, "accounting", "cost-centers"] as const,
    queryFn: () =>
      api.get<{ cost_centers: CostCenter[] }>("/v1/erp/accounting/cost-centers"),
    select: (d) => d.cost_centers,
    enabled: !!id,
  });

  const entriesQ = useQuery({
    queryKey: [...erpKeys.all, "entries", { cost_center_id: id, page_size: "100" }] as const,
    queryFn: () =>
      api.get<{ entries: JournalEntry[] }>(
        `/v1/erp/accounting/entries?cost_center_id=${id}&page_size=100`,
      ),
    select: (d) => d.entries,
    enabled: !!id,
  });

  if (centerQ.error)
    return <ErrorState message="Error cargando centro de costo" onRetry={() => window.location.reload()} />;

  const center = centerQ.data;
  const siblings = siblingsQ.data ?? [];
  const entries = entriesQ.data ?? [];
  const parent = center?.id
    ? siblings.find((s) => s.id === (center as unknown as { parent_id?: string }).parent_id)
    : null;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-4xl px-6 py-8 sm:px-8">
        <Link
          href="/administracion/contable?tab=cost-centers"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a contable
        </Link>

        {centerQ.isLoading && <Skeleton className="h-48 w-full" />}

        {center && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">{center.name}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Código <span className="font-mono">{center.code}</span>
                  {parent ? ` · Padre ${parent.code} · ${parent.name}` : ""}
                </p>
              </div>
              <Badge variant={center.active ? "default" : "secondary"}>
                {center.active ? "Activo" : "Inactivo"}
              </Badge>
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">
              Asientos imputados ({entries.length})
            </h2>
            <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[110px]">Fecha</TableHead>
                    <TableHead className="w-[120px]">Número</TableHead>
                    <TableHead>Concepto</TableHead>
                    <TableHead className="w-[100px]">Tipo</TableHead>
                    <TableHead className="w-[110px]">Estado</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {entriesQ.isLoading && (
                    <TableRow>
                      <TableCell colSpan={5}>
                        <Skeleton className="h-24 w-full" />
                      </TableCell>
                    </TableRow>
                  )}
                  {!entriesQ.isLoading && entries.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} className="h-20 text-center text-sm text-muted-foreground">
                        Sin asientos imputados a este centro de costo.
                      </TableCell>
                    </TableRow>
                  )}
                  {entries.map((e) => (
                    <TableRow key={e.id}>
                      <TableCell className="font-mono text-xs">{fmtDate(e.date)}</TableCell>
                      <TableCell className="font-mono text-sm">
                        <Link href={`/administracion/contable/asientos/${e.id}`} className="hover:underline">
                          {e.number}
                        </Link>
                      </TableCell>
                      <TableCell className="max-w-[320px] truncate text-sm">{e.concept}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">{e.entry_type}</TableCell>
                      <TableCell>
                        <Badge variant={e.status === "posted" ? "default" : "outline"}>
                          {e.status}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
