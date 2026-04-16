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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PlusIcon, FileTextIcon, ShoppingBagIcon, CheckCircleIcon, TrashIcon } from "lucide-react";

interface Quotation { id: string; number: string; date: string; customer_name: string; status: string; total: number; }
interface Order { id: string; number: string; date: string; order_type: string; customer_name: string; status: string; total: number; }
interface Customer { id: string; name: string; }
interface QuotationLine { description: string; quantity: string; unit_price: string; }

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  draft: { label: "Borrador", variant: "secondary" },
  sent: { label: "Enviada", variant: "outline" },
  approved: { label: "Aprobada", variant: "default" },
  rejected: { label: "Rechazada", variant: "secondary" },
  expired: { label: "Vencida", variant: "secondary" },
  pending: { label: "Pendiente", variant: "secondary" },
  in_progress: { label: "En progreso", variant: "outline" },
  shipped: { label: "Enviado", variant: "outline" },
  delivered: { label: "Entregado", variant: "default" },
  cancelled: { label: "Cancelado", variant: "secondary" },
};

export default function VentasPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: quotations = [], isLoading, error } = useQuery({
    queryKey: erpKeys.quotations(),
    queryFn: () => api.get<{ quotations: Quotation[] }>("/v1/erp/sales/quotations?page_size=50"),
    select: (d) => d.quotations,
  });

  const { data: orders = [] } = useQuery({
    queryKey: [...erpKeys.all, "sales", "orders"] as const,
    queryFn: () => api.get<{ orders: Order[] }>("/v1/erp/sales/orders?page_size=50"),
    select: (d) => d.orders,
  });

  const { data: customers = [] } = useQuery({
    queryKey: erpKeys.entities("customer"),
    queryFn: () => api.get<{ entities: Customer[] }>("/v1/erp/entities?type=customer&page_size=200"),
    select: (d) => d.entities,
  });

  const createMutation = useMutation({
    mutationFn: (data: { number: string; date: string; customer_id: string; valid_days: number; lines: QuotationLine[] }) =>
      api.post("/v1/erp/sales/quotations", data),
    onSuccess: () => {
      toast.success("Cotización creada");
      queryClient.invalidateQueries({ queryKey: erpKeys.quotations() });
      setCreateOpen(false);
    },
    onError: permissionErrorToast,
  });

  const approveMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/erp/sales/quotations/${id}/approve`, {}),
    onSuccess: () => {
      toast.success("Cotización aprobada");
      queryClient.invalidateQueries({ queryKey: erpKeys.quotations() });
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando ventas" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Ventas</h1>
            <p className="text-sm text-muted-foreground mt-0.5">Cotizaciones, pedidos y listas de precios</p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}><PlusIcon className="size-4 mr-1.5" />Nueva cotización</Button>
        </div>

        <Tabs defaultValue="quotations">
          <TabsList className="mb-4">
            <TabsTrigger value="quotations"><FileTextIcon className="size-3.5 mr-1.5" />Cotizaciones</TabsTrigger>
            <TabsTrigger value="orders"><ShoppingBagIcon className="size-3.5 mr-1.5" />Pedidos</TabsTrigger>
          </TabsList>

          <TabsContent value="quotations">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-24">Nro</TableHead><TableHead className="w-28">Fecha</TableHead>
                  <TableHead>Cliente</TableHead><TableHead className="text-right w-28">Total</TableHead>
                  <TableHead className="w-28">Estado</TableHead><TableHead className="w-24"></TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {quotations.map((q) => (
                    <TableRow key={q.id}>
                      <TableCell className="font-mono text-sm">{q.number}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{fmtDateShort(q.date)}</TableCell>
                      <TableCell className="text-sm">{q.customer_name}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(q.total)}</TableCell>
                      <TableCell><Badge variant={(statusBadge[q.status] || statusBadge.draft).variant}>{(statusBadge[q.status] || statusBadge.draft).label}</Badge></TableCell>
                      <TableCell>
                        {(q.status === "draft" || q.status === "sent") && (
                          <Button
                            size="sm"
                            variant="outline"
                            className="h-7 px-2 text-xs"
                            disabled={approveMutation.isPending}
                            onClick={() => approveMutation.mutate(q.id)}
                          >
                            <CheckCircleIcon className="size-3 mr-1" />Aprobar
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                  {quotations.length === 0 && <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin cotizaciones.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          <TabsContent value="orders">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-24">Nro</TableHead><TableHead className="w-28">Fecha</TableHead>
                  <TableHead className="w-24">Tipo</TableHead><TableHead>Cliente</TableHead>
                  <TableHead className="text-right w-28">Total</TableHead><TableHead className="w-28">Estado</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {orders.map((o) => (
                    <TableRow key={o.id}>
                      <TableCell className="font-mono text-sm">{o.number}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{fmtDateShort(o.date)}</TableCell>
                      <TableCell><Badge variant="secondary">{o.order_type === "customer" ? "Cliente" : "Interno"}</Badge></TableCell>
                      <TableCell className="text-sm">{o.customer_name || "\u2014"}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(o.total)}</TableCell>
                      <TableCell><Badge variant={(statusBadge[o.status] || statusBadge.draft).variant}>{(statusBadge[o.status] || statusBadge.draft).label}</Badge></TableCell>
                    </TableRow>
                  ))}
                  {orders.length === 0 && <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin pedidos.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-lg">
          <DialogHeader><DialogTitle>Nueva cotización</DialogTitle></DialogHeader>
          <CreateQuotationForm
            customers={customers}
            onSubmit={(data) => createMutation.mutate(data)}
            isPending={createMutation.isPending}
            onClose={() => setCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateQuotationForm({
  customers,
  onSubmit,
  isPending,
  onClose,
}: {
  customers: Customer[];
  onSubmit: (data: { number: string; date: string; customer_id: string; valid_days: number; lines: QuotationLine[] }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().split("T")[0];
  const [number, setNumber] = useState("");
  const [date, setDate] = useState(today);
  const [customerId, setCustomerId] = useState("");
  const [validDays, setValidDays] = useState("30");
  const [lines, setLines] = useState<QuotationLine[]>([{ description: "", quantity: "1", unit_price: "0" }]);

  function addLine() {
    setLines((prev) => [...prev, { description: "", quantity: "1", unit_price: "0" }]);
  }

  function removeLine(idx: number) {
    setLines((prev) => prev.filter((_, i) => i !== idx));
  }

  function updateLine(idx: number, field: keyof QuotationLine, value: string) {
    setLines((prev) => prev.map((l, i) => i === idx ? { ...l, [field]: value } : l));
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!number || !date || !customerId) return;
    onSubmit({ number, date, customer_id: customerId, valid_days: parseInt(validDays) || 30, lines });
  }

  const canSubmit = number && date && customerId && lines.every((l) => l.description);

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2"><Label>Número</Label><Input value={number} onChange={(e) => setNumber(e.target.value)} placeholder="COT-001" /></div>
        <div className="space-y-2"><Label>Fecha</Label><Input type="date" value={date} onChange={(e) => setDate(e.target.value)} /></div>
      </div>
      <div className="space-y-2">
        <Label>Cliente</Label>
        <Select value={customerId} onValueChange={(v) => setCustomerId(v ?? "")}>
          <SelectTrigger><SelectValue placeholder="Seleccionar cliente..." /></SelectTrigger>
          <SelectContent>
            {customers.map((c) => <SelectItem key={c.id} value={c.id}>{c.name}</SelectItem>)}
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2"><Label>Días de validez</Label><Input type="number" min="1" value={validDays} onChange={(e) => setValidDays(e.target.value)} /></div>
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label>Líneas</Label>
          <Button type="button" size="sm" variant="ghost" onClick={addLine}><PlusIcon className="size-3.5 mr-1" />Agregar</Button>
        </div>
        <div className="space-y-2">
          {lines.map((line, idx) => (
            <div key={idx} className="grid grid-cols-[1fr_80px_90px_32px] gap-2 items-center">
              <Input placeholder="Descripción" value={line.description} onChange={(e) => updateLine(idx, "description", e.target.value)} />
              <Input placeholder="Cant." type="number" min="0" step="any" value={line.quantity} onChange={(e) => updateLine(idx, "quantity", e.target.value)} />
              <Input placeholder="Precio" type="number" min="0" step="any" value={line.unit_price} onChange={(e) => updateLine(idx, "unit_price", e.target.value)} />
              <Button type="button" size="icon" variant="ghost" className="size-8 text-muted-foreground" onClick={() => removeLine(idx)} disabled={lines.length === 1}>
                <TrashIcon className="size-3.5" />
              </Button>
            </div>
          ))}
        </div>
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!canSubmit || isPending}>{isPending ? "Creando..." : "Crear"}</Button>
      </div>
    </form>
  );
}
