"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort, fmtMoney } from "@/lib/erp/format";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import type { BankImport } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { AlertCircleIcon, CheckCircle2Icon, XCircleIcon } from "lucide-react";

type StatusTab = "pending" | "done" | "cancelled" | "all";

const tabToProcessed: Record<StatusTab, string> = {
  pending: "0",
  done: "1",
  cancelled: "2",
  all: "-1",
};

export default function BankImportsPage() {
  const queryClient = useQueryClient();
  const [tab, setTab] = useState<StatusTab>("pending");
  const [accountFilter, setAccountFilter] = useState("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");

  const processedParam = tabToProcessed[tab];
  const queryParams: Record<string, string> = { processed: processedParam };
  if (accountFilter) queryParams.account = accountFilter;
  if (dateFrom) queryParams.date_from = dateFrom;
  if (dateTo) queryParams.date_to = dateTo;

  const { data: imports = [], isLoading, error } = useQuery({
    queryKey: erpKeys.bankImports(queryParams),
    queryFn: () => {
      const qs = new URLSearchParams({ ...queryParams, page_size: "100" }).toString();
      return api.get<{ imports: BankImport[] }>(`/v1/erp/treasury/imports?${qs}`);
    },
    select: (d) => d.imports,
  });

  const toggleMutation = useMutation({
    mutationFn: ({ id, processed }: { id: string; processed: number }) =>
      api.patch(`/v1/erp/treasury/imports/${id}`, { processed }),
    onSuccess: (_d, vars) => {
      toast.success(
        vars.processed === 1 ? "Marcado como procesado"
          : vars.processed === 0 ? "Reabierto"
          : "Marcado como anulado",
      );
      queryClient.invalidateQueries({ queryKey: erpKeys.bankImports() });
    },
    onError: permissionErrorToast,
  });

  if (error)
    return <ErrorState message="Error cargando importaciones bancarias" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-7xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Importaciones bancarias</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Movimientos bancarios importados desde extractos (CSV/XLS) pendientes de conciliación contra REG_MOVIMIENTOS.
          </p>
        </div>

        <div className="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-4">
          <div className="grid gap-1.5">
            <Label htmlFor="imp-account" className="text-xs">Cuenta (número)</Label>
            <Input
              id="imp-account"
              type="number"
              inputMode="numeric"
              value={accountFilter}
              onChange={(e) => setAccountFilter(e.target.value)}
              placeholder="Todas"
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="imp-from" className="text-xs">Desde</Label>
            <Input id="imp-from" type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="imp-to" className="text-xs">Hasta</Label>
            <Input id="imp-to" type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
          </div>
          <div className="flex items-end">
            <Button
              variant="outline"
              onClick={() => { setAccountFilter(""); setDateFrom(""); setDateTo(""); }}
            >
              Limpiar
            </Button>
          </div>
        </div>

        <Tabs value={tab} onValueChange={(v) => setTab(v as StatusTab)} className="mb-4">
          <TabsList>
            <TabsTrigger value="pending">
              <AlertCircleIcon className="mr-1.5 size-3.5" />
              Pendientes
            </TabsTrigger>
            <TabsTrigger value="done">
              <CheckCircle2Icon className="mr-1.5 size-3.5" />
              Procesados
            </TabsTrigger>
            <TabsTrigger value="cancelled">
              <XCircleIcon className="mr-1.5 size-3.5" />
              Anulados
            </TabsTrigger>
            <TabsTrigger value="all">Todos</TabsTrigger>
          </TabsList>
        </Tabs>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[90px]">Fecha</TableHead>
                <TableHead className="w-[90px]">Cuenta</TableHead>
                <TableHead>Concepto</TableHead>
                <TableHead className="w-[90px]">N°</TableHead>
                <TableHead className="w-[120px] text-right">Débito</TableHead>
                <TableHead className="w-[120px] text-right">Crédito</TableHead>
                <TableHead className="w-[120px] text-right">Saldo</TableHead>
                <TableHead className="w-[110px]">Estado</TableHead>
                <TableHead className="w-[160px] text-right">Acción</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={9}>
                    <Skeleton className="h-32 w-full" />
                  </TableCell>
                </TableRow>
              )}
              {!isLoading && imports.length === 0 && (
                <TableRow>
                  <TableCell colSpan={9} className="h-20 text-center text-sm text-muted-foreground">
                    Sin movimientos en esta vista.
                  </TableCell>
                </TableRow>
              )}
              {imports.map((imp) => (
                <TableRow
                  key={imp.id}
                  className={
                    imp.processed === 0
                      ? "bg-amber-50/40 dark:bg-amber-950/20"
                      : imp.processed === 2
                        ? "text-muted-foreground line-through"
                        : ""
                  }
                >
                  <TableCell className="text-sm">{imp.movement_date ? fmtDateShort(imp.movement_date) : "—"}</TableCell>
                  <TableCell className="font-mono text-sm">{imp.account_number || "—"}</TableCell>
                  <TableCell className="max-w-[320px] truncate text-sm">{imp.concept_name || "—"}</TableCell>
                  <TableCell className="font-mono text-sm">{imp.movement_no || "—"}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{fmtMoney(imp.debit)}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{fmtMoney(imp.credit)}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{fmtMoney(imp.balance)}</TableCell>
                  <TableCell>
                    {imp.processed === 1 ? (
                      <Badge variant="secondary" className="gap-1">
                        <CheckCircle2Icon className="size-3" />
                        Procesado
                      </Badge>
                    ) : imp.processed === 2 ? (
                      <Badge variant="outline" className="gap-1">
                        <XCircleIcon className="size-3" />
                        Anulado
                      </Badge>
                    ) : (
                      <Badge variant="destructive" className="gap-1">
                        <AlertCircleIcon className="size-3" />
                        Pendiente
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    {imp.processed === 0 ? (
                      <Button
                        size="sm"
                        disabled={toggleMutation.isPending}
                        onClick={() => toggleMutation.mutate({ id: imp.id, processed: 1 })}
                      >
                        Marcar procesado
                      </Button>
                    ) : (
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={toggleMutation.isPending}
                        onClick={() => toggleMutation.mutate({ id: imp.id, processed: 0 })}
                      >
                        Reabrir
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>

        <p className="mt-3 text-xs text-muted-foreground">
          El matching contra un movimiento específico de tesorería (REG_MOVIMIENTOS) queda pendiente de un selector dedicado
          — por ahora el toggle solo actualiza el estado. La carga masiva de extractos se hace desde Histrix (ver waiver
          bcs_importacion_auto_ins).
        </p>
      </div>
    </div>
  );
}
