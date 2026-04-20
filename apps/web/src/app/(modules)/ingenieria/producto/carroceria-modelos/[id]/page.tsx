"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

interface CarroceriaModelDetail {
  id: string;
  tenant_id: string;
  code: string;
  model_code: string;
  description: string;
  abbreviation: string;
  double_deck: boolean;
  axle_weight_pct: string | null;
  productive_hours_per_station: string | null;
  active: boolean;
  tech_sheet_image: string | null;
  created_at: string;
  updated_at: string;
}

interface BOMRow {
  id: string;
  component_code: string;
  component_name: string;
  quantity: string;
  unit: string;
  notes: string | null;
}

export default function CarroceriaModelDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data: model, isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "manufacturing", "carroceria-models", id] as const,
    queryFn: () => api.get<CarroceriaModelDetail>(`/v1/erp/manufacturing/carroceria-models/${id}`),
    enabled: !!id,
  });

  const { data: bom = [] } = useQuery({
    queryKey: [...erpKeys.all, "manufacturing", "carroceria-models", id, "bom"] as const,
    queryFn: () =>
      api.get<{ bom: BOMRow[] }>(`/v1/erp/manufacturing/carroceria-models/${id}/bom`),
    select: (d) => d.bom ?? [],
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando modelo de carrocería" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <Link
          href="/ingenieria/producto"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a ingeniería de producto
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {model && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">{model.description}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  <span className="font-mono">{model.model_code}</span>
                  {model.abbreviation ? ` · ${model.abbreviation}` : ""}
                  {model.double_deck ? " · Doble piso" : ""}
                </p>
              </div>
              <Badge variant={model.active ? "default" : "secondary"}>
                {model.active ? "Activo" : "Inactivo"}
              </Badge>
            </div>

            <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Código interno</div>
                <div className="mt-1 font-mono text-sm">{model.code || "—"}</div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Peso por eje %</div>
                <div className="mt-1 font-mono text-sm">
                  {model.axle_weight_pct ? Number(model.axle_weight_pct).toFixed(2) : "—"}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Horas prod. por puesto</div>
                <div className="mt-1 font-mono text-sm">
                  {model.productive_hours_per_station ? Number(model.productive_hours_per_station).toFixed(2) : "—"}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Creado</div>
                <div className="mt-1 font-mono text-sm">{fmtDate(model.created_at)}</div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Actualizado</div>
                <div className="mt-1 font-mono text-sm">{fmtDate(model.updated_at)}</div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Ficha técnica</div>
                <div className="mt-1 font-mono text-sm">{model.tech_sheet_image || "—"}</div>
              </div>
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">
              Lista de materiales (BOM — {bom.length})
            </h2>
            <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[140px]">Cód. componente</TableHead>
                    <TableHead>Componente</TableHead>
                    <TableHead className="w-[110px] text-right">Cantidad</TableHead>
                    <TableHead className="w-[80px]">Unidad</TableHead>
                    <TableHead>Notas</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {bom.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} className="h-16 text-center text-sm text-muted-foreground">
                        Sin componentes definidos.
                      </TableCell>
                    </TableRow>
                  )}
                  {bom.map((b) => (
                    <TableRow key={b.id}>
                      <TableCell className="font-mono text-xs">{b.component_code}</TableCell>
                      <TableCell className="text-sm">{b.component_name}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{Number(b.quantity).toFixed(2)}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">{b.unit}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{b.notes || "—"}</TableCell>
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
