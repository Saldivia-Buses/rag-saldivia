"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtMoney, fmtDateShort } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ShoppingCartIcon, PackageCheckIcon, ClipboardCheckIcon } from "lucide-react";
import type { QCInspection } from "@/lib/erp/types";

interface PurchaseOrder { id: string; number: string; date: string; supplier_name: string; status: string; total: number; }
interface Receipt { id: string; order_number: string; date: string; number: string; user_id: string; }
interface OrderLine { article_code: string; article_name: string; quantity: number; unit_price: number; received_qty: number; }
interface OrderDetail { order: PurchaseOrder; lines: OrderLine[]; }

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  draft: { label: "Borrador", variant: "secondary" },
  approved: { label: "Aprobada", variant: "default" },
  partial: { label: "Parcial", variant: "outline" },
  received: { label: "Recibida", variant: "default" },
  cancelled: { label: "Cancelada", variant: "secondary" },
};

export default function ComprasPage() {
  const [selectedId, setSelectedId] = useState<string | null>(null);

  const { data: orders = [], isLoading, error } = useQuery({
    queryKey: erpKeys.purchaseOrders(),
    queryFn: () => api.get<{ orders: PurchaseOrder[] }>("/v1/erp/purchasing/orders?page_size=50"),
    select: (d) => d.orders,
  });

  const { data: receipts = [] } = useQuery({
    queryKey: [...erpKeys.all, "purchasing", "receipts"] as const,
    queryFn: () => api.get<{ receipts: Receipt[] }>("/v1/erp/purchasing/receipts?page_size=50"),
    select: (d) => d.receipts,
  });

  const { data: inspections = [] } = useQuery({
    queryKey: [...erpKeys.all, "purchasing", "inspections"] as const,
    queryFn: () => api.get<{ inspections: QCInspection[] }>("/v1/erp/purchasing/inspections?page_size=50"),
    select: (d) => d.inspections,
  });

  const { data: selected } = useQuery({
    queryKey: [...erpKeys.all, "purchasing", "orders", selectedId] as const,
    queryFn: () => api.get<OrderDetail>(`/v1/erp/purchasing/orders/${selectedId}`),
    enabled: !!selectedId,
  });

  if (error) return <ErrorState message="Error cargando compras" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Compras</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Órdenes de compra y recepciones — {orders.length} órdenes</p>
        </div>

        <Tabs defaultValue="orders">
          <TabsList className="mb-4">
            <TabsTrigger value="orders"><ShoppingCartIcon className="size-3.5 mr-1.5" />Órdenes</TabsTrigger>
            <TabsTrigger value="receipts"><PackageCheckIcon className="size-3.5 mr-1.5" />Recepciones</TabsTrigger>
            <TabsTrigger value="inspections"><ClipboardCheckIcon className="size-3.5 mr-1.5" />Inspecciones QC</TabsTrigger>
          </TabsList>

          <TabsContent value="orders">
            <div className="flex gap-6">
              <div className="flex-1 min-w-0 rounded-xl border border-border/40 bg-card overflow-hidden">
                <Table>
                  <TableHeader><TableRow>
                    <TableHead className="w-24">OC</TableHead><TableHead className="w-28">Fecha</TableHead>
                    <TableHead>Proveedor</TableHead><TableHead className="text-right w-28">Total</TableHead>
                    <TableHead className="w-28">Estado</TableHead>
                  </TableRow></TableHeader>
                  <TableBody>
                    {orders.map((o) => {
                      const s = statusBadge[o.status] || statusBadge.draft;
                      return (
                        <TableRow key={o.id} className="cursor-pointer" onClick={() => setSelectedId(o.id)}>
                          <TableCell className="font-mono text-sm">{o.number}</TableCell>
                          <TableCell className="text-sm text-muted-foreground">{fmtDateShort(o.date)}</TableCell>
                          <TableCell className="text-sm">{o.supplier_name}</TableCell>
                          <TableCell className="text-right font-mono text-sm">{fmtMoney(o.total)}</TableCell>
                          <TableCell><Badge variant={s.variant}>{s.label}</Badge></TableCell>
                        </TableRow>
                      );
                    })}
                    {orders.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin órdenes de compra.</TableCell></TableRow>}
                  </TableBody>
                </Table>
              </div>

              {selected && (
                <div className="w-80 shrink-0 rounded-xl border border-border/40 bg-card p-5">
                  <h3 className="font-semibold mb-1">OC {selected.order.number}</h3>
                  <p className="text-sm text-muted-foreground mb-4">{selected.order.supplier_name}</p>
                  <div className="space-y-2">
                    {selected.lines.map((l, i) => (
                      <div key={i} className="text-sm border-b border-border/20 pb-2">
                        <div className="flex justify-between">
                          <span className="font-mono text-xs text-muted-foreground">{l.article_code}</span>
                          <span className="font-mono">{fmtMoney(l.quantity * l.unit_price)}</span>
                        </div>
                        <p>{l.article_name}</p>
                        <p className="text-xs text-muted-foreground">{l.quantity} x {fmtMoney(l.unit_price)} — recibido: {l.received_qty}</p>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </TabsContent>

          <TabsContent value="receipts">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-24">Recepción</TableHead><TableHead className="w-28">Fecha</TableHead>
                  <TableHead>OC</TableHead><TableHead>Usuario</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {receipts.map((r) => (
                    <TableRow key={r.id}>
                      <TableCell className="font-mono text-sm">{r.number}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{fmtDateShort(r.date)}</TableCell>
                      <TableCell className="font-mono text-sm">{r.order_number}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{r.user_id}</TableCell>
                    </TableRow>
                  ))}
                  {receipts.length === 0 && <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">Sin recepciones.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          <TabsContent value="inspections">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead>Artículo</TableHead><TableHead className="w-24 text-right">Cantidad</TableHead>
                  <TableHead className="w-24 text-right">Aceptado</TableHead><TableHead className="w-24 text-right">Rechazado</TableHead>
                  <TableHead className="w-28">Estado</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {inspections.map((insp) => (
                    <TableRow key={insp.id}>
                      <TableCell className="text-sm">{insp.article_name}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{insp.quantity}</TableCell>
                      <TableCell className="text-right font-mono text-sm text-green-600">{insp.accepted_qty}</TableCell>
                      <TableCell className="text-right font-mono text-sm text-red-500">{insp.rejected_qty}</TableCell>
                      <TableCell><Badge variant={insp.status === "completed" ? "default" : "secondary"}>{insp.status === "completed" ? "Completada" : "Pendiente"}</Badge></TableCell>
                    </TableRow>
                  ))}
                  {inspections.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin inspecciones.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
