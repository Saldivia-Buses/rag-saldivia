"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { ControlledDocument } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";

type StatusTab = "all" | "draft" | "approved" | "obsolete";

const tabToStatus: Record<StatusTab, string> = {
  all: "",
  draft: "draft",
  approved: "approved",
  obsolete: "obsolete",
};

const statusVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  draft: "outline",
  approved: "secondary",
  obsolete: "destructive",
};

const statusLabel: Record<string, string> = {
  draft: "Borrador",
  approved: "Aprobado",
  obsolete: "Obsoleto",
};

export default function DocumentosControladosPage() {
  const [tab, setTab] = useState<StatusTab>("approved");
  const queryParams: Record<string, string> = {};
  if (tabToStatus[tab]) queryParams.status = tabToStatus[tab];

  const { data: docs = [], isLoading, error } = useQuery({
    queryKey: erpKeys.controlledDocuments(queryParams),
    queryFn: () => {
      const qs = new URLSearchParams({ ...queryParams, page_size: "100" }).toString();
      return api.get<{ documents: ControlledDocument[] }>(`/v1/erp/quality/documents?${qs}`);
    },
    select: (d) => d.documents,
  });

  if (error)
    return <ErrorState message="Error cargando documentos controlados" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Documentos controlados</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Registro de documentos del SGC — código, título, revisión y estado de aprobación.
          </p>
        </div>

        <Tabs value={tab} onValueChange={(v) => setTab(v as StatusTab)} className="mb-4">
          <TabsList>
            <TabsTrigger value="approved">Aprobados</TabsTrigger>
            <TabsTrigger value="draft">Borradores</TabsTrigger>
            <TabsTrigger value="obsolete">Obsoletos</TabsTrigger>
            <TabsTrigger value="all">Todos</TabsTrigger>
          </TabsList>
        </Tabs>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[130px]">Código</TableHead>
                <TableHead>Título</TableHead>
                <TableHead className="w-[90px] text-right">Rev.</TableHead>
                <TableHead className="w-[120px]">Aprobado</TableHead>
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
              {!isLoading && docs.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-20 text-center text-sm text-muted-foreground">
                    Sin documentos en esta vista.
                  </TableCell>
                </TableRow>
              )}
              {docs.map((d) => (
                <TableRow key={d.id}>
                  <TableCell className="font-mono text-sm">{d.code || "—"}</TableCell>
                  <TableCell className="max-w-[500px] truncate text-sm">{d.title || "—"}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{d.revision || 0}</TableCell>
                  <TableCell className="text-sm">
                    {d.approved_at ? fmtDateShort(d.approved_at) : <span className="text-muted-foreground">—</span>}
                  </TableCell>
                  <TableCell>
                    <Badge variant={statusVariant[d.status] ?? "outline"}>
                      {statusLabel[d.status] ?? d.status}
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
