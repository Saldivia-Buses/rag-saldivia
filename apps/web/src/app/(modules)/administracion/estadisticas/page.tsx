"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtMoney } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { BarChart3Icon, TrendingUpIcon, FactoryIcon, UsersIcon, ShieldCheckIcon, DownloadIcon } from "lucide-react";
import { BarChart, Bar, LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid, PieChart, Pie, Cell, Legend } from "recharts";

const COLORS = ["#3b82f6", "#ef4444", "#22c55e", "#f59e0b", "#8b5cf6", "#06b6d4", "#ec4899", "#14b8a6"];

interface KPIs {
  month_revenue?: number;
  month_expense?: number;
  cash_balance?: number;
  accounts_receivable?: number;
  accounts_payable?: number;
  active_prod_orders?: number;
  pending_purchases?: number;
  open_quotations?: number;
  stock_below_min?: number;
  headcount?: number;
  absences_this_month?: number;
  open_nonconformities?: number;
  pending_work_orders?: number;
}

function useDateRange() {
  const now = new Date();
  const [from, setFrom] = useState(() => {
    const d = new Date(now.getFullYear() - 1, now.getMonth(), 1);
    return d.toISOString().slice(0, 10);
  });
  const [to, setTo] = useState(() => now.toISOString().slice(0, 10));
  return { from, to, setFrom, setTo };
}

function DateFilter({ from, to, setFrom, setTo }: ReturnType<typeof useDateRange>) {
  return (
    <div className="flex items-center gap-2 mb-4">
      <Input type="date" value={from} onChange={(e) => setFrom(e.target.value)} className="w-40" />
      <span className="text-muted-foreground">a</span>
      <Input type="date" value={to} onChange={(e) => setTo(e.target.value)} className="w-40" />
    </div>
  );
}

function ExportButtons({ report, from, to }: { report: string; from: string; to: string }) {
  const base = `/v1/erp/analytics/${report}?date_from=${from}&date_to=${to}`;
  return (
    <div className="flex gap-2">
      <a href={base + "&format=csv"} download>
        <Button variant="outline" size="sm"><DownloadIcon className="size-3.5 mr-1" />CSV</Button>
      </a>
      <a href={base + "&format=excel"} download>
        <Button variant="outline" size="sm"><DownloadIcon className="size-3.5 mr-1" />Excel</Button>
      </a>
    </div>
  );
}

function useAnalytics(domain: string, report: string, from: string, to: string, extra?: Record<string, string>) {
  return useQuery({
    queryKey: erpKeys.analytics(domain, report, { date_from: from, date_to: to, ...extra }),
    queryFn: () => {
      const params = new URLSearchParams({ date_from: from, date_to: to, ...extra });
      return api.get<{ rows: Record<string, any>[]; meta: { count: number } }>(`/v1/erp/analytics/${domain}/${report}?${params}`);
    },
    select: (d) => d.rows,
  });
}

function KPICard({ label, value, color }: { label: string; value: any; color?: string }) {
  const display = typeof value === "number" ? (value > 1000 ? fmtMoney(value) : String(value)) : "\u2014";
  return (
    <div className="rounded-xl border border-border/40 bg-card p-4">
      <p className="text-xs text-muted-foreground mb-1">{label}</p>
      <p className={`text-xl font-semibold ${color || ""}`}>{display}</p>
    </div>
  );
}

export default function EstadisticasPage() {
  const dates = useDateRange();

  const { data: kpisData, isLoading, error } = useQuery({
    queryKey: erpKeys.dashboardKPIs(),
    queryFn: () => api.get<{ kpis: KPIs }>("/v1/erp/analytics/dashboard/kpis"),
    select: (d) => d.kpis,
  });

  if (error) return <ErrorState message="Error cargando estadísticas" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  const kpis = kpisData ?? {};

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-7xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Estadísticas</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Dashboard, reportes y exportación</p>
        </div>

        <Tabs defaultValue="dashboard">
          <TabsList className="mb-4">
            <TabsTrigger value="dashboard"><BarChart3Icon className="size-3.5 mr-1.5" />Dashboard</TabsTrigger>
            <TabsTrigger value="financiero"><TrendingUpIcon className="size-3.5 mr-1.5" />Financiero</TabsTrigger>
            <TabsTrigger value="operativo"><FactoryIcon className="size-3.5 mr-1.5" />Operativo</TabsTrigger>
            <TabsTrigger value="rrhh"><UsersIcon className="size-3.5 mr-1.5" />RRHH</TabsTrigger>
            <TabsTrigger value="calidad"><ShieldCheckIcon className="size-3.5 mr-1.5" />Calidad</TabsTrigger>
          </TabsList>

          <TabsContent value="dashboard">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6">
              <KPICard label="Facturado este mes" value={kpis.month_revenue} color="text-green-500" />
              <KPICard label="Gastos este mes" value={kpis.month_expense} color="text-red-500" />
              <KPICard label="Saldo disponible" value={kpis.cash_balance} color="text-blue-500" />
              <KPICard label="Cuentas por cobrar" value={kpis.accounts_receivable} />
            </div>
            <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
              <KPICard label="Cuentas por pagar" value={kpis.accounts_payable} />
              <KPICard label="OPs activas" value={kpis.active_prod_orders} />
              <KPICard label="OC pendientes" value={kpis.pending_purchases} />
              <KPICard label="Cotiz. abiertas" value={kpis.open_quotations} />
              <KPICard label="Stock bajo mínimo" value={kpis.stock_below_min} color={kpis.stock_below_min ? "text-red-500" : ""} />
            </div>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mt-3">
              <KPICard label="Empleados" value={kpis.headcount} />
              <KPICard label="Ausencias este mes" value={kpis.absences_this_month} />
              <KPICard label="NC abiertas" value={kpis.open_nonconformities} color={kpis.open_nonconformities ? "text-amber-500" : ""} />
              <KPICard label="OT pendientes" value={kpis.pending_work_orders} />
            </div>
          </TabsContent>

          <TabsContent value="financiero">
            <DateFilter {...dates} />
            <FinancieroTab from={dates.from} to={dates.to} />
          </TabsContent>

          <TabsContent value="operativo">
            <DateFilter {...dates} />
            <OperativoTab from={dates.from} to={dates.to} />
          </TabsContent>

          <TabsContent value="rrhh">
            <DateFilter {...dates} />
            <RRHHTab from={dates.from} to={dates.to} />
          </TabsContent>

          <TabsContent value="calidad">
            <DateFilter {...dates} />
            <CalidadTab from={dates.from} to={dates.to} />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

function FinancieroTab({ from, to }: { from: string; to: string }) {
  const { data: incomeExpense = [] } = useAnalytics("accounting", "income-expense", from, to);
  const { data: cashFlow = [] } = useAnalytics("treasury", "cash-flow", from, to);
  const { data: paymentMethods = [] } = useAnalytics("treasury", "payment-methods", from, to);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium">Ingresos vs Egresos</h3>
        <ExportButtons report="accounting/income-expense" from={from} to={to} />
      </div>
      <div className="rounded-xl border border-border/40 bg-card p-4 h-72">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={incomeExpense}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-border/40" />
            <XAxis dataKey="period" className="text-xs" />
            <YAxis className="text-xs" />
            <Tooltip />
            <Bar dataKey="income" name="Ingresos" fill="#22c55e" radius={[4, 4, 0, 0]} />
            <Bar dataKey="expense" name="Egresos" fill="#ef4444" radius={[4, 4, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <h3 className="text-sm font-medium">Flujo de Caja</h3>
      <div className="rounded-xl border border-border/40 bg-card p-4 h-72">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={cashFlow}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-border/40" />
            <XAxis dataKey="period" className="text-xs" />
            <YAxis className="text-xs" />
            <Tooltip />
            <Line type="monotone" dataKey="inflow" name="Ingresos" stroke="#22c55e" strokeWidth={2} />
            <Line type="monotone" dataKey="outflow" name="Egresos" stroke="#ef4444" strokeWidth={2} />
          </LineChart>
        </ResponsiveContainer>
      </div>

      {paymentMethods.length > 0 && (
        <>
          <h3 className="text-sm font-medium">Medios de Pago</h3>
          <div className="rounded-xl border border-border/40 bg-card p-4 h-72">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie data={paymentMethods} dataKey="total" nameKey="method" cx="50%" cy="50%" outerRadius={100} label>
                  {paymentMethods.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
                </Pie>
                <Tooltip />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </>
      )}
    </div>
  );
}

function OperativoTab({ from, to }: { from: string; to: string }) {
  const { data: prodByMonth = [] } = useAnalytics("production", "output-by-month", from, to);
  const { data: stockRotation = [] } = useAnalytics("stock", "rotation", from, to);
  const { data: belowMin = [] } = useAnalytics("stock", "below-minimum", from, to);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium">Producción Mensual</h3>
        <ExportButtons report="production/output-by-month" from={from} to={to} />
      </div>
      <div className="rounded-xl border border-border/40 bg-card p-4 h-72">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={prodByMonth}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-border/40" />
            <XAxis dataKey="period" className="text-xs" />
            <YAxis className="text-xs" />
            <Tooltip />
            <Bar dataKey="total_orders" name="Total" fill="#3b82f6" radius={[4, 4, 0, 0]} />
            <Bar dataKey="completed" name="Completadas" fill="#22c55e" radius={[4, 4, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>

      {belowMin.length > 0 && (
        <>
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium text-red-500">Artículos Bajo Mínimo ({belowMin.length})</h3>
            <ExportButtons report="stock/below-minimum" from={from} to={to} />
          </div>
          <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
            <Table>
              <TableHeader><TableRow>
                <TableHead className="w-28">Código</TableHead><TableHead>Nombre</TableHead>
                <TableHead className="text-right">Mínimo</TableHead><TableHead className="text-right">Actual</TableHead>
                <TableHead className="text-right">Déficit</TableHead>
              </TableRow></TableHeader>
              <TableBody>
                {belowMin.slice(0, 20).map((item: any, i: number) => (
                  <TableRow key={i}>
                    <TableCell className="font-mono text-sm">{item.code}</TableCell>
                    <TableCell className="text-sm">{item.name}</TableCell>
                    <TableCell className="text-right font-mono text-sm">{item.min_stock}</TableCell>
                    <TableCell className="text-right font-mono text-sm text-red-500">{item.current_stock}</TableCell>
                    <TableCell className="text-right font-mono text-sm font-medium text-red-500">{item.deficit}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </>
      )}
    </div>
  );
}

function RRHHTab({ from, to }: { from: string; to: string }) {
  const { data: headcount = [] } = useAnalytics("hr", "headcount", from, to);
  const { data: absences = [] } = useAnalytics("hr", "absences-by-month", from, to);

  return (
    <div className="space-y-6">
      <h3 className="text-sm font-medium">Dotación por Departamento</h3>
      <div className="rounded-xl border border-border/40 bg-card p-4 h-72">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={headcount} layout="vertical">
            <CartesianGrid strokeDasharray="3 3" className="stroke-border/40" />
            <XAxis type="number" className="text-xs" />
            <YAxis type="category" dataKey="department" className="text-xs" width={150} />
            <Tooltip />
            <Bar dataKey="headcount" name="Empleados" fill="#3b82f6" radius={[0, 4, 4, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium">Ausentismo Mensual</h3>
        <ExportButtons report="hr/absences-by-month" from={from} to={to} />
      </div>
      <div className="rounded-xl border border-border/40 bg-card p-4 h-72">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={absences}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-border/40" />
            <XAxis dataKey="period" className="text-xs" />
            <YAxis className="text-xs" />
            <Tooltip />
            <Bar dataKey="count" name="Ausencias" fill="#f59e0b" radius={[4, 4, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

function CalidadTab({ from, to }: { from: string; to: string }) {
  const { data: ncByType = [] } = useAnalytics("quality", "nonconformities-by-type", from, to);
  const { data: woCompletion = [] } = useAnalytics("maintenance", "completion-rate", from, to);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium">No Conformidades por Severidad</h3>
        <ExportButtons report="quality/nonconformities-by-type" from={from} to={to} />
      </div>
      <div className="rounded-xl border border-border/40 bg-card p-4 h-72">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={ncByType}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-border/40" />
            <XAxis dataKey="severity" className="text-xs" />
            <YAxis className="text-xs" />
            <Tooltip />
            <Bar dataKey="count" name="Cantidad" fill="#ef4444" radius={[4, 4, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium">Completitud OT por Mes</h3>
        <ExportButtons report="maintenance/completion-rate" from={from} to={to} />
      </div>
      <div className="rounded-xl border border-border/40 bg-card p-4 h-72">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={woCompletion}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-border/40" />
            <XAxis dataKey="period" className="text-xs" />
            <YAxis className="text-xs" />
            <Tooltip />
            <Bar dataKey="total" name="Total" fill="#94a3b8" radius={[4, 4, 0, 0]} />
            <Bar dataKey="completed" name="Completadas" fill="#22c55e" radius={[4, 4, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
