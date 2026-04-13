"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { PlusIcon, FlaskConicalIcon } from "lucide-react";

interface ProdOrder {
  id: string; number: string; date: string;
  product_code: string; product_name: string;
  quantity: number; status: string; priority: number;
}

type ProdStatus = "planned" | "in_progress" | "completed" | "cancelled";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  planned: { label: "Planificada", variant: "secondary" },
  in_progress: { label: "En producción", variant: "outline" },
  completed: { label: "Completada", variant: "default" },
  cancelled: { label: "Cancelada", variant: "secondary" },
};

const statusOptions: { value: ProdStatus; label: string }[] = [
  { value: "planned", label: "Planificada" },
  { value: "in_progress", label: "En producción" },
  { value: "completed", label: "Completada" },
  { value: "cancelled", label: "Cancelada" },
];

export default function DesarrolloPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: orders = [], isLoading, error } = useQuery({
    queryKey: erpKeys.productionOrders(),
    queryFn: () => api.get<{ orders: ProdOrder[] }>("/v1/erp/production/orders?page_size=50"),
    select: (d) => d.orders,
  });

  const createMutation = useMutation({
    mutationFn: (data: { number: string; date: string; product_code: string; product_name: string; quantity: number; priority: number }) =>
      api.post("/v1/erp/production/orders", data),
    onSuccess: () => {
      toast.success("Proyecto creado");
      queryClient.invalidateQueries({ queryKey: erpKeys.productionOrders() });
      setCreateOpen(false);
    },
    onError: permissionErrorToast,
  });

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: ProdStatus }) =>
      api.patch(`/v1/erp/production/orders/${id}/status`, { status }),
    onSuccess: () => {
      toast.success("Estado actualizado");
      queryClient.invalidateQueries({ queryKey: erpKeys.productionOrders() });
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando proyectos" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Desarrollo</h1>
            <p className="text-sm text-muted-foreground mt-0.5">Proyectos I+D, prototipos y ensayos — {orders.length} proyectos</p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}><PlusIcon className="size-4 mr-1.5" />Nuevo proyecto</Button>
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-24">Número</TableHead>
                <TableHead className="w-28">Fecha</TableHead>
                <TableHead>Proyecto</TableHead>
                <TableHead className="w-20 text-center">Prior.</TableHead>
                <TableHead className="w-44">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {orders.map((o) => {
                const s = statusBadge[o.status] ?? statusBadge.planned;
                return (
                  <TableRow key={o.id}>
                    <TableCell className="font-mono text-sm">{o.number}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">{fmtDateShort(o.date)}</TableCell>
                    <TableCell>
                      <span className="font-mono text-xs text-muted-foreground">{o.product_code}</span>{" "}
                      <span className="text-sm">{o.product_name}</span>
                    </TableCell>
                    <TableCell className="text-center text-sm">{o.priority}</TableCell>
                    <TableCell>
                      <Select
                        value={o.status}
                        onValueChange={(v) => updateStatusMutation.mutate({ id: o.id, status: v as ProdStatus })}
                      >
                        <SelectTrigger className="h-7 text-xs w-36">
                          <Badge variant={s.variant} className="text-xs">{s.label}</Badge>
                        </SelectTrigger>
                        <SelectContent>
                          {statusOptions.map((opt) => (
                            <SelectItem key={opt.value} value={opt.value} className="text-xs">{opt.label}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </TableCell>
                  </TableRow>
                );
              })}
              {orders.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                    <div className="flex flex-col items-center gap-2">
                      <FlaskConicalIcon className="size-8 text-muted-foreground/40" />
                      <span>Sin proyectos. Creá el primero.</span>
                    </div>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nuevo proyecto de desarrollo</DialogTitle></DialogHeader>
          <CreateProjectForm
            onSubmit={(data) => createMutation.mutate(data)}
            isPending={createMutation.isPending}
            onClose={() => setCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateProjectForm({ onSubmit, isPending, onClose }: {
  onSubmit: (data: { number: string; date: string; product_code: string; product_name: string; quantity: number; priority: number }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().split("T")[0];
  const [number, setNumber] = useState("");
  const [date, setDate] = useState(today);
  const [productCode, setProductCode] = useState("");
  const [productName, setProductName] = useState("");

  const canSubmit = number && date && productCode && productName;

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (canSubmit) onSubmit({ number, date, product_code: productCode, product_name: productName, quantity: 1, priority: 5 });
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2"><Label>Número</Label><Input value={number} onChange={(e) => setNumber(e.target.value)} placeholder="ID-001" /></div>
        <div className="space-y-2"><Label>Fecha</Label><Input type="date" value={date} onChange={(e) => setDate(e.target.value)} /></div>
      </div>
      <div className="space-y-2"><Label>Código</Label><Input value={productCode} onChange={(e) => setProductCode(e.target.value)} placeholder="Ej: PROTO-001" /></div>
      <div className="space-y-2"><Label>Nombre del proyecto</Label><Input value={productName} onChange={(e) => setProductName(e.target.value)} placeholder="Ej: Carrocería eléctrica" /></div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!canSubmit || isPending}>{isPending ? "Creando..." : "Crear"}</Button>
      </div>
    </form>
  );
}
