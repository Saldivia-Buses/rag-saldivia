"use client";

import Link from "next/link";
import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import type { WorkOrder, MaintenanceAsset } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { PlusIcon } from "lucide-react";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" | "destructive" }> = {
  open: { label: "Abierta", variant: "destructive" },
  in_progress: { label: "En curso", variant: "outline" },
  completed: { label: "Completada", variant: "default" },
  cancelled: { label: "Cancelada", variant: "secondary" },
};

export default function CorrectivoPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: workOrders = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.workOrders("corrective")] as const,
    queryFn: () =>
      api.get<{ work_orders: WorkOrder[] }>("/v1/erp/maintenance/work-orders?work_type=corrective&page_size=100"),
    select: (d) => d.work_orders,
  });

  const { data: assets = [] } = useQuery({
    queryKey: erpKeys.maintenanceAssets(),
    queryFn: () => api.get<{ assets: MaintenanceAsset[] }>("/v1/erp/maintenance/assets"),
    select: (d) => d.assets,
  });

  const createMutation = useMutation({
    mutationFn: (data: { number: string; asset_id: string; description?: string; assigned_to?: string }) =>
      api.post("/v1/erp/maintenance/work-orders", { ...data, work_type: "corrective" }),
    onSuccess: () => {
      toast.success("Orden de trabajo creada");
      queryClient.invalidateQueries({ queryKey: erpKeys.workOrders("corrective") });
      setCreateOpen(false);
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando órdenes correctivas" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Mantenimiento Correctivo</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              {workOrders.length} órdenes — reparaciones y fallas no programadas
            </p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />Nueva OT
          </Button>
        </div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader><TableRow>
              <TableHead className="w-20">OT</TableHead>
              <TableHead className="w-28">Fecha</TableHead>
              <TableHead>Equipo</TableHead>
              <TableHead>Descripción</TableHead>
              <TableHead className="w-28">Estado</TableHead>
            </TableRow></TableHeader>
            <TableBody>
              {workOrders.map((wo) => {
                const s = statusBadge[wo.status] || statusBadge.open;
                return (
                  <TableRow key={wo.id}>
                    <TableCell className="font-mono text-sm">
                      <Link href={`/mantenimiento/ordenes-trabajo/${wo.id}`} className="hover:underline">
                        {wo.number}
                      </Link>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">{fmtDateShort(wo.date)}</TableCell>
                    <TableCell className="text-sm">{wo.asset_name || "\u2014"}</TableCell>
                    <TableCell className="text-sm truncate max-w-64 text-muted-foreground">
                      {wo.description || "\u2014"}
                    </TableCell>
                    <TableCell><Badge variant={s.variant}>{s.label}</Badge></TableCell>
                  </TableRow>
                );
              })}
              {workOrders.length === 0 && (
                <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                  Sin órdenes correctivas.
                </TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva orden correctiva</DialogTitle></DialogHeader>
          <CreateWorkOrderForm
            assets={assets}
            onSubmit={(data) => createMutation.mutate(data)}
            isPending={createMutation.isPending}
            onClose={() => setCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateWorkOrderForm({
  assets,
  onSubmit,
  isPending,
  onClose,
}: {
  assets: MaintenanceAsset[];
  onSubmit: (data: { number: string; asset_id: string; description?: string; assigned_to?: string }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [number, setNumber] = useState("");
  const [assetId, setAssetId] = useState("");
  const [description, setDescription] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!number || !assetId) return;
        onSubmit({ number, asset_id: assetId, description: description || undefined });
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Número de OT</Label>
          <Input value={number} onChange={(e) => setNumber(e.target.value)} placeholder="OT-001" />
        </div>
        <div className="space-y-2">
          <Label>Equipo</Label>
          <Select value={assetId} onValueChange={(v) => setAssetId(v ?? "")}>
            <SelectTrigger className="bg-card">
              <SelectValue placeholder="Seleccionar..." />
            </SelectTrigger>
            <SelectContent>
              {assets.map((a) => (
                <SelectItem key={a.id} value={a.id}>{a.code} — {a.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
      <div className="space-y-2">
        <Label>Descripción <span className="text-muted-foreground">(opcional)</span></Label>
        <Textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Describe la falla o tarea a realizar..."
          rows={3}
        />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!number || !assetId || isPending}>
          {isPending ? "Creando..." : "Crear"}
        </Button>
      </div>
    </form>
  );
}
