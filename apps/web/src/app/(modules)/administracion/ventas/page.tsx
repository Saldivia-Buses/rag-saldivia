"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtMoney, fmtDateShort } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { FileTextIcon, ShoppingBagIcon } from "lucide-react";

interface Quotation { id: string; number: string; date: string; customer_name: string; status: string; total: number; }
interface Order { id: string; number: string; date: string; order_type: string; customer_name: string; status: string; total: number; }

const statusColors: Record<string, "default" | "secondary" | "outline"> = { draft: "secondary", sent: "outline", approved: "default", rejected: "secondary", expired: "secondary", pending: "secondary", in_progress: "outline", shipped: "outline", delivered: "default", cancelled: "secondary" };

export default function VentasPage() {
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

  if (error) return <ErrorState message="Error cargando ventas" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Ventas</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Cotizaciones, pedidos y listas de precios</p>
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
                  <TableHead className="w-28">Estado</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {quotations.map((q) => (
                    <TableRow key={q.id}>
                      <TableCell className="font-mono text-sm">{q.number}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{fmtDateShort(q.date)}</TableCell>
                      <TableCell className="text-sm">{q.customer_name}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(q.total)}</TableCell>
                      <TableCell><Badge variant={statusColors[q.status] || "secondary"}>{q.status}</Badge></TableCell>
                    </TableRow>
                  ))}
                  {quotations.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin cotizaciones.</TableCell></TableRow>}
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
                      <TableCell><Badge variant={statusColors[o.status] || "secondary"}>{o.status}</Badge></TableCell>
                    </TableRow>
                  ))}
                  {orders.length === 0 && <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin pedidos.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
