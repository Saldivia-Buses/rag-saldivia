"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate, fmtNumber } from "@/lib/erp/format";
import type { MaintenanceAsset, MaintenancePlan } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const typeLabel: Record<string, string> = {
  vehicle: "Vehículo",
  machine: "Máquina",
  tool: "Herramienta",
  facility: "Instalación",
};

export default function EquipoDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data: assets = [] } = useQuery({
    queryKey: erpKeys.maintenanceAssets(),
    queryFn: () => api.get<{ assets: MaintenanceAsset[] }>("/v1/erp/maintenance/assets"),
    select: (d) => d.assets,
  });

  const { data: plans = [], isLoading, error } = useQuery({
    queryKey: erpKeys.maintenancePlans(id),
    queryFn: () =>
      api.get<{ plans: MaintenancePlan[] }>(`/v1/erp/maintenance/assets/${id}/plans`),
    select: (d) => d.plans,
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando planes del equipo" onRetry={() => window.location.reload()} />;

  const asset = assets.find((a) => a.id === id);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <Link
          href="/mantenimiento/equipos"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a equipos
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        <div className="mb-6 flex items-baseline justify-between gap-4">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">
              {asset?.name ?? "Equipo"}{" "}
              {asset ? <span className="font-mono text-sm text-muted-foreground">{asset.code}</span> : null}
            </h1>
            {asset && (
              <p className="mt-0.5 text-sm text-muted-foreground">
                {typeLabel[asset.asset_type] ?? asset.asset_type}
                {asset.location ? ` · ${asset.location}` : ""}
              </p>
            )}
          </div>
          {asset && (
            <Badge variant={asset.active ? "default" : "secondary"}>
              {asset.active ? "Activo" : "Inactivo"}
            </Badge>
          )}
        </div>

        <h2 className="mb-3 text-sm font-medium text-muted-foreground">
          Planes de mantenimiento ({plans.length})
        </h2>
        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Plan</TableHead>
                <TableHead className="w-[110px] text-right">Cada (días)</TableHead>
                <TableHead className="w-[110px] text-right">Cada (km)</TableHead>
                <TableHead className="w-[110px] text-right">Cada (hs)</TableHead>
                <TableHead className="w-[130px]">Último</TableHead>
                <TableHead className="w-[130px]">Próximo</TableHead>
                <TableHead className="w-[90px] text-center">Activo</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {plans.length === 0 && !isLoading && (
                <TableRow>
                  <TableCell colSpan={7} className="h-20 text-center text-sm text-muted-foreground">
                    Sin planes definidos.
                  </TableCell>
                </TableRow>
              )}
              {plans.map((p) => (
                <TableRow key={p.id}>
                  <TableCell className="text-sm">{p.name}</TableCell>
                  <TableCell className="text-right font-mono text-xs">{fmtNumber(p.frequency_days)}</TableCell>
                  <TableCell className="text-right font-mono text-xs">{fmtNumber(p.frequency_km)}</TableCell>
                  <TableCell className="text-right font-mono text-xs">{fmtNumber(p.frequency_hours)}</TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">{fmtDate(p.last_done)}</TableCell>
                  <TableCell className="font-mono text-xs">{fmtDate(p.next_due)}</TableCell>
                  <TableCell className="text-center">
                    <Badge variant={p.active ? "default" : "secondary"}>
                      {p.active ? "Sí" : "No"}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
