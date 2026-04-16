"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtMoney, fmtDateShort } from "@/lib/erp/format";
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
import { GlobeIcon, PlusIcon } from "lucide-react";

interface PurchaseOrder { id: string; number: string; date: string; supplier_name: string; status: string; total: number; }
interface Supplier { id: string; name: string; }
interface POLine { description: string; quantity: string; unit_price: string; }

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  draft: { label: "Borrador", variant: "secondary" },
  approved: { label: "Aprobada", variant: "default" },
  partial: { label: "Parcial", variant: "outline" },
  received: { label: "Recibida", variant: "default" },
  cancelled: { label: "Cancelada", variant: "secondary" },
};

const IMP_PREFIX = "IMP-";

export default function ComexPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: orders = [], isLoading, error } = useQuery({
    queryKey: erpKeys.purchaseOrders(),
    queryFn: () => api.get<{ orders: PurchaseOrder[] }>("/v1/erp/purchasing/orders?page_size=50"),
    select: (d) => d.orders,
  });

  const approveMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/erp/purchasing/orders/${id}/approve`),
    onSuccess: () => {
      toast.success("Orden aprobada");
      queryClient.invalidateQueries({ queryKey: erpKeys.purchaseOrders() });
    },
    onError: permissionErrorToast,
  });

  const createMutation = useMutation({
    mutationFn: (data: { number: string; supplier_id: string; date: string; lines: POLine[] }) =>
      api.post("/v1/erp/purchasing/orders", data),
    onSuccess: () => {
      toast.success("OC de importación creada");
      queryClient.invalidateQueries({ queryKey: erpKeys.purchaseOrders() });
      setCreateOpen(false);
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando órdenes de compra" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  // Show all orders but highlight import ones (IMP- prefix)
  const importOrders = orders.filter((o) => o.number.startsWith(IMP_PREFIX));
  const displayOrders = importOrders.length > 0 ? importOrders : orders;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-2">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Importaciones</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              Órdenes de compra internacionales — {displayOrders.length} registros
            </p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />Nueva OC Import
          </Button>
        </div>

        {importOrders.length === 0 && orders.length > 0 && (
          <div className="mb-4 rounded-lg border border-border/40 bg-muted/30 px-4 py-3 flex items-start gap-2">
            <GlobeIcon className="size-4 text-muted-foreground mt-0.5 shrink-0" />
            <p className="text-sm text-muted-foreground">
              No hay órdenes de importación aún (prefijo <span className="font-mono">IMP-</span>). Mostrando todas las OC. Use{" "}
              <span className="font-semibold">Nueva OC Import</span> para crear una con prefijo de importación.
            </p>
          </div>
        )}

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-32">OC</TableHead>
                <TableHead className="w-28">Fecha</TableHead>
                <TableHead>Proveedor</TableHead>
                <TableHead className="text-right w-32">Total</TableHead>
                <TableHead className="w-28">Estado</TableHead>
                <TableHead className="w-28" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {displayOrders.map((o) => {
                const s = statusBadge[o.status] ?? statusBadge.draft;
                return (
                  <TableRow key={o.id}>
                    <TableCell className="font-mono text-sm">{o.number}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">{fmtDateShort(o.date)}</TableCell>
                    <TableCell className="text-sm">{o.supplier_name}</TableCell>
                    <TableCell className="text-right font-mono text-sm">{fmtMoney(o.total)}</TableCell>
                    <TableCell><Badge variant={s.variant}>{s.label}</Badge></TableCell>
                    <TableCell>
                      {o.status === "draft" && (
                        <Button size="sm" variant="outline" disabled={approveMutation.isPending}
                          onClick={() => approveMutation.mutate(o.id)}>
                          Aprobar
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                );
              })}
              {displayOrders.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                    Sin órdenes de importación.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader><DialogTitle>Nueva OC de importación</DialogTitle></DialogHeader>
          <CreateImportOrderForm
            onSubmit={(data) => createMutation.mutate(data)}
            isPending={createMutation.isPending}
            onClose={() => setCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateImportOrderForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (data: { number: string; supplier_id: string; date: string; lines: POLine[] }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [number, setNumber] = useState("IMP-");
  const [date, setDate] = useState(new Date().toISOString().split("T")[0]);
  const [supplierId, setSupplierId] = useState("");
  const [lines, setLines] = useState<POLine[]>([{ description: "", quantity: "1", unit_price: "" }]);

  const { data: suppliers = [] } = useQuery({
    queryKey: erpKeys.entities("supplier"),
    queryFn: () => api.get<{ entities: Supplier[] }>("/v1/erp/entities?type=supplier&page_size=200"),
    select: (d) => d.entities,
  });

  function addLine() {
    setLines((prev) => [...prev, { description: "", quantity: "1", unit_price: "" }]);
  }

  function updateLine(i: number, field: keyof POLine, value: string) {
    setLines((prev) => prev.map((l, idx) => idx === i ? { ...l, [field]: value } : l));
  }

  function removeLine(i: number) {
    setLines((prev) => prev.filter((_, idx) => idx !== i));
  }

  const canSubmit = !!supplierId && lines.length > 0 && lines.every((l) => l.description && l.unit_price);

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!canSubmit) return;
        onSubmit({ number, supplier_id: supplierId, date, lines });
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Número</Label>
          <Input value={number} onChange={(e) => setNumber(e.target.value)} placeholder="IMP-0001" />
        </div>
        <div className="space-y-2">
          <Label>Fecha</Label>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        </div>
      </div>

      <div className="space-y-2">
        <Label>Proveedor</Label>
        <Select value={supplierId} onValueChange={(v) => v && setSupplierId(v)}>
          <SelectTrigger><SelectValue placeholder="Seleccionar proveedor..." /></SelectTrigger>
          <SelectContent>
            {(suppliers as Supplier[]).map((s) => (
              <SelectItem key={s.id} value={s.id}>{s.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label>Líneas</Label>
          <Button type="button" size="sm" variant="outline" onClick={addLine}>
            <PlusIcon className="size-3.5 mr-1" />Agregar línea
          </Button>
        </div>
        <div className="space-y-2">
          {lines.map((line, i) => (
            <div key={i} className="grid grid-cols-12 gap-2 items-start">
              <div className="col-span-6">
                {i === 0 && <p className="text-xs text-muted-foreground mb-1">Descripción</p>}
                <Input value={line.description} onChange={(e) => updateLine(i, "description", e.target.value)} placeholder="Descripción del ítem" />
              </div>
              <div className="col-span-2">
                {i === 0 && <p className="text-xs text-muted-foreground mb-1">Cantidad</p>}
                <Input type="number" value={line.quantity} onChange={(e) => updateLine(i, "quantity", e.target.value)} placeholder="1" />
              </div>
              <div className="col-span-3">
                {i === 0 && <p className="text-xs text-muted-foreground mb-1">Precio unit.</p>}
                <Input type="number" value={line.unit_price} onChange={(e) => updateLine(i, "unit_price", e.target.value)} placeholder="0.00" />
              </div>
              <div className="col-span-1 flex items-end">
                {i === 0 && <div className="h-[18px]" />}
                <Button type="button" size="sm" variant="ghost" disabled={lines.length === 1}
                  onClick={() => removeLine(i)} className="px-2 text-muted-foreground hover:text-destructive">×</Button>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!canSubmit || isPending}>{isPending ? "Creando..." : "Crear"}</Button>
      </div>
    </form>
  );
}
