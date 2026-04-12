"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { FactoryIcon, TruckIcon } from "lucide-react";

interface ProdOrder { id: string; number: string; date: string; product_code: string; product_name: string; quantity: number; status: string; priority: number; }
interface Unit { id: string; chassis_number: string; internal_number: string; model: string; status: string; engine_brand: string; patent: string; }

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  planned: { label: "Planificada", variant: "secondary" },
  in_progress: { label: "En producción", variant: "outline" },
  completed: { label: "Completada", variant: "default" },
  cancelled: { label: "Cancelada", variant: "secondary" },
  in_production: { label: "En producción", variant: "outline" },
  ready: { label: "Lista", variant: "default" },
  delivered: { label: "Entregada", variant: "default" },
};

export default function ProduccionPage() {
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

  if (error) return <ErrorState message="Error cargando producción" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Producción</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Órdenes de producción y unidades — {orders.length} órdenes, {units.length} unidades</p>
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
                  <TableHead className="w-20 text-center">Prior.</TableHead><TableHead className="w-32">Estado</TableHead>
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
                        <TableCell><Badge variant={s.variant}>{s.label}</Badge></TableCell>
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
    </div>
  );
}
