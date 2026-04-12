"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort, fmtNumber } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ClipboardListIcon, FuelIcon } from "lucide-react";

interface WorkOrder { id: string; number: string; date: string; asset_code: string; asset_name: string; work_type: string; status: string; priority: string; }
interface FuelLog { id: string; asset_code: string; asset_name: string; date: string; liters: number; km_reading: number; cost: number; }

const statusColor: Record<string, "default" | "secondary" | "outline"> = { open: "secondary", in_progress: "outline", completed: "default", cancelled: "secondary" };
const prioColor: Record<string, string> = { low: "text-muted-foreground", normal: "", high: "text-amber-500", urgent: "text-red-500 font-medium" };
const typeLabel: Record<string, string> = { preventive: "Preventivo", corrective: "Correctivo", inspection: "Inspección" };

export default function MantenimientoPage() {
  const { data: workOrders = [], isLoading, error } = useQuery({
    queryKey: erpKeys.workOrders(),
    queryFn: () => api.get<{ work_orders: WorkOrder[] }>("/v1/erp/maintenance/work-orders?page_size=50"),
    select: (d) => d.work_orders,
  });

  const { data: fuelLogs = [] } = useQuery({
    queryKey: [...erpKeys.all, "maintenance", "fuel-logs"] as const,
    queryFn: () => api.get<{ fuel_logs: FuelLog[] }>("/v1/erp/maintenance/fuel-logs?page_size=50"),
    select: (d) => d.fuel_logs,
  });

  if (error) return <ErrorState message="Error cargando mantenimiento" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Mantenimiento</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Órdenes de trabajo y combustible</p>
        </div>
        <Tabs defaultValue="work-orders">
          <TabsList className="mb-4">
            <TabsTrigger value="work-orders"><ClipboardListIcon className="size-3.5 mr-1.5" />Órdenes de Trabajo</TabsTrigger>
            <TabsTrigger value="fuel"><FuelIcon className="size-3.5 mr-1.5" />Combustible</TabsTrigger>
          </TabsList>
          <TabsContent value="work-orders">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead className="w-20">OT</TableHead><TableHead className="w-28">Fecha</TableHead><TableHead>Equipo</TableHead><TableHead className="w-28">Tipo</TableHead><TableHead className="w-20">Prior.</TableHead><TableHead className="w-28">Estado</TableHead></TableRow></TableHeader>
                <TableBody>{workOrders.map((wo) => (<TableRow key={wo.id}><TableCell className="font-mono text-sm">{wo.number}</TableCell><TableCell className="text-sm text-muted-foreground">{fmtDateShort(wo.date)}</TableCell><TableCell><span className="font-mono text-xs text-muted-foreground">{wo.asset_code}</span> {wo.asset_name}</TableCell><TableCell><Badge variant="secondary">{typeLabel[wo.work_type] || wo.work_type}</Badge></TableCell><TableCell className={`text-sm ${prioColor[wo.priority] || ""}`}>{wo.priority}</TableCell><TableCell><Badge variant={statusColor[wo.status] || "secondary"}>{wo.status}</Badge></TableCell></TableRow>))}
                {workOrders.length === 0 && <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin órdenes de trabajo.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
          <TabsContent value="fuel">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead className="w-28">Fecha</TableHead><TableHead>Equipo</TableHead><TableHead className="text-right w-20">Litros</TableHead><TableHead className="text-right w-24">Km</TableHead><TableHead className="text-right w-28">Costo</TableHead></TableRow></TableHeader>
                <TableBody>{fuelLogs.map((fl) => (<TableRow key={fl.id}><TableCell className="text-sm text-muted-foreground">{fmtDateShort(fl.date)}</TableCell><TableCell><span className="font-mono text-xs text-muted-foreground">{fl.asset_code}</span> {fl.asset_name}</TableCell><TableCell className="text-right font-mono text-sm">{fmtNumber(fl.liters)}</TableCell><TableCell className="text-right font-mono text-sm">{fl.km_reading || "\u2014"}</TableCell><TableCell className="text-right font-mono text-sm">{fl.cost ? `$${fmtNumber(fl.cost)}` : "\u2014"}</TableCell></TableRow>))}
                {fuelLogs.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin registros de combustible.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
