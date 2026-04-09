"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, SearchIcon, CalendarDaysIcon, TrendingUpIcon, GaugeIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type SemanaPCP = {
  id: string;
  semana: string;
  fechaDesde: string;
  fechaHasta: string;
  planificadas: number;
  reales: number;
  eficiencia: number;
  cuelloBotella: string;
  utilizacionCapacidad: number;
};

const fmtDate = (d: string) => new Date(d).toLocaleDateString("es-AR", { day: "2-digit", month: "short" });

const MOCK: SemanaPCP[] = [
  { id: "1", semana: "S15-2026", fechaDesde: "2026-04-06", fechaHasta: "2026-04-12", planificadas: 3, reales: 3, eficiencia: 100, cuelloBotella: "—", utilizacionCapacidad: 88 },
  { id: "2", semana: "S14-2026", fechaDesde: "2026-03-30", fechaHasta: "2026-04-05", planificadas: 4, reales: 3, eficiencia: 75, cuelloBotella: "Pintura", utilizacionCapacidad: 92 },
  { id: "3", semana: "S13-2026", fechaDesde: "2026-03-23", fechaHasta: "2026-03-29", planificadas: 3, reales: 3, eficiencia: 100, cuelloBotella: "—", utilizacionCapacidad: 85 },
  { id: "4", semana: "S12-2026", fechaDesde: "2026-03-16", fechaHasta: "2026-03-22", planificadas: 4, reales: 2, eficiencia: 50, cuelloBotella: "Soldadura", utilizacionCapacidad: 95 },
  { id: "5", semana: "S11-2026", fechaDesde: "2026-03-09", fechaHasta: "2026-03-15", planificadas: 3, reales: 4, eficiencia: 133, cuelloBotella: "—", utilizacionCapacidad: 78 },
  { id: "6", semana: "S10-2026", fechaDesde: "2026-03-02", fechaHasta: "2026-03-08", planificadas: 3, reales: 3, eficiencia: 100, cuelloBotella: "—", utilizacionCapacidad: 82 },
  { id: "7", semana: "S09-2026", fechaDesde: "2026-02-23", fechaHasta: "2026-03-01", planificadas: 4, reales: 3, eficiencia: 75, cuelloBotella: "Ensamble", utilizacionCapacidad: 91 },
  { id: "8", semana: "S08-2026", fechaDesde: "2026-02-16", fechaHasta: "2026-02-22", planificadas: 3, reales: 1, eficiencia: 33, cuelloBotella: "Soldadura", utilizacionCapacidad: 97 },
];

const columns: ColumnDef<SemanaPCP>[] = [
  {
    accessorKey: "semana",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Semana <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <div><p className="text-sm font-medium font-mono">{row.original.semana}</p><p className="text-xs text-muted-foreground">{fmtDate(row.original.fechaDesde)} — {fmtDate(row.original.fechaHasta)}</p></div>,
  },
  {
    accessorKey: "planificadas",
    header: () => <div className="text-center">Plan.</div>,
    cell: ({ row }) => <p className="text-sm font-mono text-center">{row.original.planificadas}</p>,
  },
  {
    accessorKey: "reales",
    header: () => <div className="text-center">Real</div>,
    cell: ({ row }) => {
      const diff = row.original.reales - row.original.planificadas;
      const color = diff < 0 ? "text-red-500" : diff > 0 ? "text-green-500" : "";
      return <p className={`text-sm font-mono text-center font-medium ${color}`}>{row.original.reales}</p>;
    },
  },
  {
    accessorKey: "eficiencia",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Eficiencia <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => {
      const e = row.original.eficiencia;
      const color: "green" | "amber" | "red" = e >= 90 ? "green" : e >= 70 ? "amber" : "red";
      return <Badge variant="outline" color={color}>{e}%</Badge>;
    },
  },
  {
    accessorKey: "cuelloBotella",
    header: "Cuello de Botella",
    cell: ({ row }) => {
      const cb = row.original.cuelloBotella;
      if (cb === "—") return <p className="text-sm text-muted-foreground">—</p>;
      return <Badge variant="outline" color="red">{cb}</Badge>;
    },
  },
  {
    accessorKey: "utilizacionCapacidad",
    header: "Capacidad",
    cell: ({ row }) => (
      <div className="flex items-center gap-2 min-w-[120px]">
        <Progress value={row.original.utilizacionCapacidad} className="flex-1" />
        <span className="text-xs font-mono text-muted-foreground w-8 text-right">{row.original.utilizacionCapacidad}%</span>
      </div>
    ),
  },
];

export default function ProduccionPCPPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  const eficienciaPromedio = Math.round(MOCK.reduce((a, s) => a + s.eficiencia, 0) / MOCK.length);
  const capacidadPromedio = Math.round(MOCK.reduce((a, s) => a + s.utilizacionCapacidad, 0) / MOCK.length);
  const semanasConCuello = MOCK.filter((s) => s.cuelloBotella !== "—").length;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Planificacion y Control de Produccion</h1><p className="text-sm text-muted-foreground mt-0.5">Analisis semanal de eficiencia y capacidad</p></div>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Eficiencia promedio", value: `${eficienciaPromedio}%`, icon: TrendingUpIcon, color: eficienciaPromedio >= 80 ? "text-green-500" : "text-amber-500" },
            { label: "Uso de capacidad prom.", value: `${capacidadPromedio}%`, icon: GaugeIcon, color: "text-blue-500" },
            { label: "Semanas con cuello", value: semanasConCuello, icon: CalendarDaysIcon, color: semanasConCuello > 2 ? "text-red-500" : "text-amber-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2"><s.icon className="size-4 text-muted-foreground" /><p className="text-xs text-muted-foreground">{s.label}</p></div>
              <p className={`text-xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar por semana..." className="pl-9 bg-card" value={(table.getColumn("semana")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("semana")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>{table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron semanas.</TableCell></TableRow>)}</TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
