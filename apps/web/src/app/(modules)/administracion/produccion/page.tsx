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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PlusIcon, FactoryIcon, TruckIcon } from "lucide-react";

interface ProdOrder { id: string; number: string; date: string; product_code: string; product_name: string; quantity: number; status: string; priority: number; }
interface Unit { id: string; chassis_number: string; internal_number: string; model: string; status: string; engine_brand: string; patent: string; }

type ProdStatus = "planned" | "in_progress" | "completed" | "cancelled";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  planned: { label: "Planificada", variant: "secondary" },
  in_progress: { label: "En producción", variant: "outline" },
  completed: { label: "Completada", variant: "default" },
  cancelled: { label: "Cancelada", variant: "secondary" },
  in_production: { label: "En producción", variant: "outline" },
  ready: { label: "Lista", variant: "default" },
  delivered: { label: "Entregada", variant: "default" },
};

const statusOptions: { value: ProdStatus; label: string }[] = [
  { value: "planned", label: "Planificada" },
  { value: "in_progress", label: "En producción" },
  { value: "completed", label: "Completada" },
  { value: "cancelled", label: "Cancelada" },
];

export default function ProduccionPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: orders = [], isLoading, error } = useQuery({
    queryKey: erpKeys.productionOrders(),
    queryFn: () => api.get<{ orders: ProdOrder[] }>("/v1/erp/production/orders?page_size=50"),
    select: (d) => d.orders,
  });

  const { data: units = [] } = useQuery({
    queryKey: [...erpKeys.all, "production", "units"] as const,
    queryFn: () => api.get<{ units: Unit[] }>("/v1/erp/production/units?page_size=50"),
    select: (d) => d.units,
  });

  const createMutation = useMutation({
    mutationFn: (data: { number: string; date: string; product_id?: string | null; product_code: string; product_name: string; quantity: number; priority: number }) =>
      api.post("/v1/erp/production/orders", data),
    onSuccess: () => {
      toast.success("Orden de producción creada");
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

  if (error) return <ErrorState message="Error cargando producción" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Producción</h1>
            <p className="text-sm text-muted-foreground mt-0.5">Órdenes de producción y unidades — {orders.length} órdenes, {units.length} unidades</p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}><PlusIcon className="size-4 mr-1.5" />Nueva OP</Button>
        </div>

        <Tabs defaultValue="orders">
          <TabsList className="mb-4">
            <TabsTrigger value="orders"><FactoryIcon className="size-3.5 mr-1.5" />Órdenes</TabsTrigger>
            <TabsTrigger value="units"><TruckIcon className="size-3.5 mr-1.5" />Unidades</TabsTrigger>
          </TabsList>

          <TabsContent value="orders">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-24">OP</TableHead><TableHead className="w-28">Fecha</TableHead>
                  <TableHead>Producto</TableHead><TableHead className="text-right w-20">Cant.</TableHead>
                  <TableHead className="w-20 text-center">Prior.</TableHead><TableHead className="w-44">Estado</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {orders.map((o) => {
                    const s = statusBadge[o.status] || statusBadge.planned;
                    return (
                      <TableRow key={o.id}>
                        <TableCell className="font-mono text-sm">{o.number}</TableCell>
                        <TableCell className="text-sm text-muted-foreground">{fmtDateShort(o.date)}</TableCell>
                        <TableCell><span className="font-mono text-xs text-muted-foreground">{o.product_code}</span> {o.product_name}</TableCell>
                        <TableCell className="text-right font-mono text-sm">{o.quantity}</TableCell>
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
                  {orders.length === 0 && <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin órdenes de producción.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          <TabsContent value="units">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-36">Chasis</TableHead><TableHead className="w-20">Interno</TableHead>
                  <TableHead>Modelo</TableHead><TableHead className="w-28">Motor</TableHead>
                  <TableHead className="w-28">Patente</TableHead><TableHead className="w-32">Estado</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {units.map((u) => {
                    const s = statusBadge[u.status] || statusBadge.in_production;
                    return (
                      <TableRow key={u.id}>
                        <TableCell className="font-mono text-sm">{u.chassis_number}</TableCell>
                        <TableCell className="text-sm">{u.internal_number || "\u2014"}</TableCell>
                        <TableCell className="text-sm">{u.model || "\u2014"}</TableCell>
                        <TableCell className="text-sm text-muted-foreground">{u.engine_brand || "\u2014"}</TableCell>
                        <TableCell className="font-mono text-sm">{u.patent || "\u2014"}</TableCell>
                        <TableCell><Badge variant={s.variant}>{s.label}</Badge></TableCell>
                      </TableRow>
                    );
                  })}
                  {units.length === 0 && <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin unidades.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva orden de producción</DialogTitle></DialogHeader>
          <CreateProductionOrderForm
            onSubmit={(data) => createMutation.mutate(data)}
            isPending={createMutation.isPending}
            onClose={() => setCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateProductionOrderForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (data: { number: string; date: string; product_id?: string | null; product_code: string; product_name: string; quantity: number; priority: number }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().split("T")[0];
  const [number, setNumber] = useState("");
  const [date, setDate] = useState(today);
  const [productCode, setProductCode] = useState("");
  const [productName, setProductName] = useState("");
  const [quantity, setQuantity] = useState("1");
  const [priority, setPriority] = useState("5");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!number || !date || !productCode || !productName) return;
    onSubmit({
      number,
      date,
      product_id: null,
      product_code: productCode,
      product_name: productName,
      quantity: parseInt(quantity) || 1,
      priority: Math.min(10, Math.max(1, parseInt(priority) || 5)),
    });
  }

  const canSubmit = number && date && productCode && productName;

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2"><Label>Número</Label><Input value={number} onChange={(e) => setNumber(e.target.value)} placeholder="OP-001" /></div>
        <div className="space-y-2"><Label>Fecha</Label><Input type="date" value={date} onChange={(e) => setDate(e.target.value)} /></div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2"><Label>Código producto</Label><Input value={productCode} onChange={(e) => setProductCode(e.target.value)} placeholder="BUS-123" /></div>
        <div className="space-y-2"><Label>Nombre producto</Label><Input value={productName} onChange={(e) => setProductName(e.target.value)} placeholder="Carrocería urbana" /></div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2"><Label>Cantidad</Label><Input type="number" min="1" value={quantity} onChange={(e) => setQuantity(e.target.value)} /></div>
        <div className="space-y-2">
          <Label>Prioridad (1-10)</Label>
          <Input type="number" min="1" max="10" value={priority} onChange={(e) => setPriority(e.target.value)} />
        </div>
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!canSubmit || isPending}>{isPending ? "Creando..." : "Crear"}</Button>
      </div>
    </form>
  );
}
