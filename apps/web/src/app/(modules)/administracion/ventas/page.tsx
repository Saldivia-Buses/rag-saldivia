"use client";

import { useState, useEffect, useCallback } from "react";
import { api } from "@/lib/api/client";
import { wsManager } from "@/lib/ws/manager";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { FileTextIcon, ShoppingBagIcon, TagIcon } from "lucide-react";

interface Quotation { id: string; number: string; date: string; customer_name: string; status: string; total: number; }
interface Order { id: string; number: string; date: string; order_type: string; customer_name: string; status: string; total: number; }

const fmtMoney = (n: number) => n === 0 ? "—" : new Intl.NumberFormat("es-AR", { style: "currency", currency: "ARS", maximumFractionDigits: 0 }).format(n);
const fmtDate = (s: string) => new Date(s).toLocaleDateString("es-AR", { day: "2-digit", month: "short" });
const statusColors: Record<string, "default" | "secondary" | "outline"> = { draft: "secondary", sent: "outline", approved: "default", rejected: "secondary", expired: "secondary", pending: "secondary", in_progress: "outline", shipped: "outline", delivered: "default", cancelled: "secondary" };

export default function VentasPage() {
  const [quotations, setQuotations] = useState<Quotation[]>([]);
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);

  const fetch = useCallback(async () => {
    try {
      const [q, o] = await Promise.all([
        api.get<{ quotations: Quotation[] }>("/v1/erp/sales/quotations?page_size=50"),
        api.get<{ orders: Order[] }>("/v1/erp/sales/orders?page_size=50"),
      ]);
      setQuotations(q.quotations); setOrders(o.orders);
    } catch (err) { console.error(err); } finally { setLoading(false); }
  }, []);

  useEffect(() => { fetch(); }, [fetch]);
  useEffect(() => { const unsub = wsManager.subscribe("erp_sales", fetch); return unsub; }, [fetch]);

  if (loading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

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
                  <TableHead className="w-24">Nro</TableHead>
                  <TableHead className="w-28">Fecha</TableHead>
                  <TableHead>Cliente</TableHead>
                  <TableHead className="text-right w-28">Total</TableHead>
                  <TableHead className="w-28">Estado</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {quotations.map((q) => (
                    <TableRow key={q.id}>
                      <TableCell className="font-mono text-sm">{q.number}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{fmtDate(q.date)}</TableCell>
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
                  <TableHead className="w-24">Nro</TableHead>
                  <TableHead className="w-28">Fecha</TableHead>
                  <TableHead className="w-24">Tipo</TableHead>
                  <TableHead>Cliente</TableHead>
                  <TableHead className="text-right w-28">Total</TableHead>
                  <TableHead className="w-28">Estado</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {orders.map((o) => (
                    <TableRow key={o.id}>
                      <TableCell className="font-mono text-sm">{o.number}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{fmtDate(o.date)}</TableCell>
                      <TableCell><Badge variant="secondary">{o.order_type === "customer" ? "Cliente" : "Interno"}</Badge></TableCell>
                      <TableCell className="text-sm">{o.customer_name || "—"}</TableCell>
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
