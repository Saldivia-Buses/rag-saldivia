"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { PaymentComplaint } from "@/lib/erp/types";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { AlertCircleIcon, CheckCircle2Icon, PlusIcon } from "lucide-react";

type StatusTab = "pending" | "done" | "all";

const tabToStatus: Record<StatusTab, string> = {
  pending: "0",
  done: "1",
  all: "-1",
};

export default function ReclamosPage() {
  const queryClient = useQueryClient();
  const [tab, setTab] = useState<StatusTab>("pending");
  const [createOpen, setCreateOpen] = useState(false);
  const [newDate, setNewDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [newCtacod, setNewCtacod] = useState("");
  const [newObservation, setNewObservation] = useState("");

  const statusParam = tabToStatus[tab];

  const { data: complaints = [], isLoading, error } = useQuery({
    queryKey: erpKeys.paymentComplaints({ status: statusParam }),
    queryFn: () =>
      api.get<{ complaints: PaymentComplaint[] }>(
        `/v1/erp/accounts/complaints?status=${statusParam}&limit=200`,
      ),
    select: (d) => d.complaints,
  });

  const createMutation = useMutation({
    mutationFn: (data: { date: string; entity_legacy_code: number; observation: string }) =>
      api.post<{ complaint: PaymentComplaint }>("/v1/erp/accounts/complaints", data),
    onSuccess: () => {
      toast.success("Reclamo creado");
      queryClient.invalidateQueries({ queryKey: erpKeys.paymentComplaints() });
      setCreateOpen(false);
      setNewCtacod("");
      setNewObservation("");
    },
    onError: permissionErrorToast,
  });

  const toggleMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: number }) =>
      api.patch(`/v1/erp/accounts/complaints/${id}/status`, { status }),
    onSuccess: (_d, vars) => {
      toast.success(vars.status === 1 ? "Marcado como cumplido" : "Marcado como pendiente");
      queryClient.invalidateQueries({ queryKey: erpKeys.paymentComplaints() });
    },
    onError: permissionErrorToast,
  });

  function submitCreate(e: React.FormEvent) {
    e.preventDefault();
    const code = parseInt(newCtacod, 10);
    if (!newCtacod || Number.isNaN(code) || code <= 0) {
      toast.error("Ingresá un código de proveedor válido");
      return;
    }
    if (!newObservation.trim()) {
      toast.error("La observación no puede estar vacía");
      return;
    }
    createMutation.mutate({
      date: newDate,
      entity_legacy_code: code,
      observation: newObservation.trim(),
    });
  }

  if (error)
    return <ErrorState message="Error cargando reclamos de pagos" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6 flex items-start justify-between gap-4">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Reclamos de pagos</h1>
            <p className="mt-0.5 text-sm text-muted-foreground">
              Reclamos recibidos de proveedores por pagos pendientes — marcar como "cumplido" cuando se resuelven.
            </p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <PlusIcon className="mr-1.5 size-3.5" />
            Nuevo reclamo
          </Button>
          <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Nuevo reclamo de pago</DialogTitle>
              </DialogHeader>
              <form onSubmit={submitCreate} className="grid gap-4 py-2">
                <div className="grid gap-1.5">
                  <Label htmlFor="rec-date">Fecha</Label>
                  <Input
                    id="rec-date"
                    type="date"
                    value={newDate}
                    onChange={(e) => setNewDate(e.target.value)}
                    required
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label htmlFor="rec-ctacod">Código de proveedor</Label>
                  <Input
                    id="rec-ctacod"
                    type="number"
                    inputMode="numeric"
                    value={newCtacod}
                    onChange={(e) => setNewCtacod(e.target.value)}
                    placeholder="ctacod (ej. 42)"
                    required
                  />
                  <p className="text-xs text-muted-foreground">
                    Código legacy del proveedor. Selector de entidad llega en un PR posterior.
                  </p>
                </div>
                <div className="grid gap-1.5">
                  <Label htmlFor="rec-obs">Observación</Label>
                  <Textarea
                    id="rec-obs"
                    value={newObservation}
                    onChange={(e) => setNewObservation(e.target.value)}
                    rows={4}
                    placeholder="Detalle del reclamo"
                    required
                  />
                </div>
                <DialogFooter>
                  <Button type="button" variant="outline" onClick={() => setCreateOpen(false)}>
                    Cancelar
                  </Button>
                  <Button type="submit" disabled={createMutation.isPending}>
                    {createMutation.isPending ? "Guardando…" : "Crear reclamo"}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        <Tabs value={tab} onValueChange={(v) => setTab(v as StatusTab)} className="mb-4">
          <TabsList>
            <TabsTrigger value="pending">
              <AlertCircleIcon className="mr-1.5 size-3.5" />
              Pendientes
            </TabsTrigger>
            <TabsTrigger value="done">
              <CheckCircle2Icon className="mr-1.5 size-3.5" />
              Cumplidos
            </TabsTrigger>
            <TabsTrigger value="all">Todos</TabsTrigger>
          </TabsList>
        </Tabs>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[110px]">Fecha</TableHead>
                <TableHead className="w-[130px]">Proveedor (cod.)</TableHead>
                <TableHead>Observación</TableHead>
                <TableHead className="w-[110px]">Estado</TableHead>
                <TableHead className="w-[110px]">Login</TableHead>
                <TableHead className="w-[120px] text-right">Acción</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={6}>
                    <Skeleton className="h-32 w-full" />
                  </TableCell>
                </TableRow>
              )}
              {!isLoading && complaints.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-20 text-center text-sm text-muted-foreground">
                    Sin reclamos en esta vista.
                  </TableCell>
                </TableRow>
              )}
              {complaints.map((c) => (
                <TableRow key={c.id} className={c.status_flag === 0 ? "bg-amber-50/40 dark:bg-amber-950/20" : ""}>
                  <TableCell className="text-sm">{c.complaint_date ? fmtDateShort(c.complaint_date) : "—"}</TableCell>
                  <TableCell className="text-sm font-mono">{c.entity_legacy_code || "—"}</TableCell>
                  <TableCell className="max-w-[420px] whitespace-pre-wrap text-sm">{c.observation || "—"}</TableCell>
                  <TableCell>
                    {c.status_flag === 1 ? (
                      <Badge variant="secondary" className="gap-1">
                        <CheckCircle2Icon className="size-3" />
                        Cumplido
                      </Badge>
                    ) : (
                      <Badge variant="destructive" className="gap-1">
                        <AlertCircleIcon className="size-3" />
                        Pendiente
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">{c.login || "—"}</TableCell>
                  <TableCell className="text-right">
                    <Button
                      size="sm"
                      variant={c.status_flag === 1 ? "outline" : "default"}
                      disabled={toggleMutation.isPending}
                      onClick={() =>
                        toggleMutation.mutate({ id: c.id, status: c.status_flag === 1 ? 0 : 1 })
                      }
                    >
                      {c.status_flag === 1 ? "Reabrir" : "Cumplir"}
                    </Button>
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
