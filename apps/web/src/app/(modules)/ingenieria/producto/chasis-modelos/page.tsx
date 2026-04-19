"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { ChassisModel } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function ChassisModelsPage() {
  const { data: models = [], isLoading, error } = useQuery({
    queryKey: erpKeys.chassisModels(),
    queryFn: () => api.get<{ chassis_models: ChassisModel[] }>("/v1/erp/manufacturing/chassis-models"),
    select: (d) => d.chassis_models,
  });

  if (error)
    return <ErrorState message="Error cargando modelos de chasis" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Modelos de chasis</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">Catálogo de modelos por marca — tracción, ubicación del motor, código.</p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[140px]">Código</TableHead>
                <TableHead>Descripción</TableHead>
                <TableHead className="w-[120px]">Tracción</TableHead>
                <TableHead className="w-[160px]">Motor</TableHead>
                <TableHead className="w-[110px]">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow><TableCell colSpan={5}><Skeleton className="h-32 w-full" /></TableCell></TableRow>
              )}
              {!isLoading && models.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-20 text-center text-sm text-muted-foreground">
                    Sin modelos registrados.
                  </TableCell>
                </TableRow>
              )}
              {models.map((m) => (
                <TableRow key={m.id}>
                  <TableCell className="font-mono text-sm">{m.model_code}</TableCell>
                  <TableCell className="text-sm">{m.description}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">{m.traction || "—"}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">{m.engine_location || "—"}</TableCell>
                  <TableCell>
                    <Badge variant={m.active ? "secondary" : "outline"}>
                      {m.active ? "Activo" : "Inactivo"}
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
