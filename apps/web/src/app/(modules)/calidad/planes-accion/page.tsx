"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { ActionPlan } from "@/lib/erp/types";
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

export default function PlanesAccionPage() {
  const [tab, setTab] = useState<StatusTab>("open");
  const queryParams: Record<string, string> = {};
  if (tabToStatus[tab]) queryParams.status = tabToStatus[tab];

  const { data: plans = [], isLoading, error } = useQuery({
    queryKey: erpKeys.actionPlans(queryParams),
    queryFn: () => {
      const qs = new URLSearchParams({ ...queryParams, page_size: "100" }).toString();
      return api.get<{ plans: ActionPlan[] }>(`/v1/erp/quality/action-plans?${qs}`);
    },
    select: (d) => d.plans,
  });

  if (error)
    return <ErrorState message="Error cargando planes de acción" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-7xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Planes de acción</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Acciones correctivas derivadas de no conformidades o auditorías. Muestra inicio planificado, fecha objetivo y estado.
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
                <TableHead>Descripción</TableHead>
                <TableHead className="w-[110px]">Inicio plan.</TableHead>
                <TableHead className="w-[110px]">Fecha obj.</TableHead>
                <TableHead className="w-[110px]">Cierre</TableHead>
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
              {!isLoading && plans.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-20 text-center text-sm text-muted-foreground">
                    Sin planes en esta vista.
                  </TableCell>
                </TableRow>
              )}
              {plans.map((p) => (
                <TableRow key={p.id}>
                  <TableCell className="max-w-[500px] whitespace-pre-wrap text-sm">{p.description || "—"}</TableCell>
                  <TableCell className="text-sm">
                    {p.planned_start ? fmtDateShort(p.planned_start) : "—"}
                  </TableCell>
                  <TableCell className="text-sm">
                    {p.target_date ? fmtDateShort(p.target_date) : "—"}
                  </TableCell>
                  <TableCell className="text-sm">
                    {p.closed_date ? fmtDateShort(p.closed_date) : <span className="text-muted-foreground">—</span>}
                  </TableCell>
                  <TableCell>
                    <Badge variant={statusVariant[p.status] ?? "outline"}>
                      {statusLabel[p.status] ?? p.status}
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
