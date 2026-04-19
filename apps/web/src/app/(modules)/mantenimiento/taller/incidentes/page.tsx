"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { VehicleIncident } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";

type StatusTab = "open" | "in_progress" | "closed" | "all";

const tabToStatus: Record<StatusTab, string> = {
  open: "open",
  in_progress: "in_progress",
  closed: "closed",
  all: "",
};

const statusVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  open: "destructive",
  in_progress: "default",
  closed: "secondary",
};

const statusLabel: Record<string, string> = {
  open: "Abierto",
  in_progress: "En curso",
  closed: "Cerrado",
};

export default function IncidentesVehicularesPage() {
  const [tab, setTab] = useState<StatusTab>("open");
  const queryParams: Record<string, string> = {};
  if (tabToStatus[tab]) queryParams.status = tabToStatus[tab];

  const { data: incidents = [], isLoading, error } = useQuery({
    queryKey: erpKeys.vehicleIncidents(queryParams),
    queryFn: () => {
      const qs = new URLSearchParams({ ...queryParams, page_size: "100" }).toString();
      return api.get<{ incidents: VehicleIncident[] }>(`/v1/erp/workshop/incidents?${qs}`);
    },
    select: (d) => d.incidents,
  });

  if (error)
    return <ErrorState message="Error cargando incidentes" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-7xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Incidentes vehiculares</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Siniestros / fallas reportadas sobre el parque vehicular — tipo, ubicación, responsable, estado.
          </p>
        </div>

        <Tabs value={tab} onValueChange={(v) => setTab(v as StatusTab)} className="mb-4">
          <TabsList>
            <TabsTrigger value="open">Abiertos</TabsTrigger>
            <TabsTrigger value="in_progress">En curso</TabsTrigger>
            <TabsTrigger value="closed">Cerrados</TabsTrigger>
            <TabsTrigger value="all">Todos</TabsTrigger>
          </TabsList>
        </Tabs>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[110px]">Fecha</TableHead>
                <TableHead className="w-[140px]">Tipo</TableHead>
                <TableHead>Ubicación</TableHead>
                <TableHead className="w-[140px]">Responsable</TableHead>
                <TableHead className="w-[130px]">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={5}>
                    <Skeleton className="h-32 w-full" />
                  </TableCell>
                </TableRow>
              )}
              {!isLoading && incidents.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-20 text-center text-sm text-muted-foreground">
                    Sin incidentes en esta vista.
                  </TableCell>
                </TableRow>
              )}
              {incidents.map((i) => (
                <TableRow key={i.id}>
                  <TableCell className="text-sm">
                    {i.incident_date ? fmtDateShort(i.incident_date) : "—"}
                  </TableCell>
                  <TableCell className="text-xs">
                    {i.incident_type_name ?? <span className="text-muted-foreground">—</span>}
                  </TableCell>
                  <TableCell className="max-w-[280px] truncate text-sm">{i.location || "—"}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">{i.responsible || "—"}</TableCell>
                  <TableCell>
                    <Badge variant={statusVariant[i.status] ?? "outline"}>
                      {statusLabel[i.status] ?? i.status}
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
