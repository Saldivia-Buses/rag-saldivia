"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, ClipboardListIcon, TruckIcon, FactoryIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type OrdenProduccion = {
  id: string;
  numero: string;
  modelo: string;
  cliente: string;
  cantidad: number;
  fechaInicio: string;
  fechaEntrega: string;
  estado: "planificada" | "en_produccion" | "completada" | "entregada";
  progreso: number;
};

const fmtDate = (d: string) => new Date(d).toLocaleDateString("es-AR", { day: "2-digit", month: "short", year: "numeric" });

const MOCK: OrdenProduccion[] = [
  { id: "1", numero: "OP-2026-042", modelo: "Saldivia Gran Turismo 420", cliente: "Vía Bariloche S.A.", cantidad: 4, fechaInicio: "2026-03-01", fechaEntrega: "2026-06-15", progreso: 62, estado: "en_produccion" },
  { id: "2", numero: "OP-2026-041", modelo: "Saldivia Urbano 320", cliente: "Municipalidad de Córdoba", cantidad: 12, fechaInicio: "2026-02-15", fechaEntrega: "2026-07-30", progreso: 38, estado: "en_produccion" },
  { id: "3", numero: "OP-2026-040", modelo: "Saldivia Interurbano 380", cliente: "Empresa San José S.R.L.", cantidad: 3, fechaInicio: "2026-01-10", fechaEntrega: "2026-04-20", progreso: 95, estado: "en_produccion" },
  { id: "4", numero: "OP-2026-039", modelo: "Saldivia Doble Piso 500DD", cliente: "Andesmar S.A.", cantidad: 6, fechaInicio: "2026-04-01", fechaEntrega: "2026-09-30", progreso: 8, estado: "planificada" },
  { id: "5", numero: "OP-2026-038", modelo: "Saldivia Urbano 320", cliente: "Transporte Automotor La Estrella", cantidad: 8, fechaInicio: "2025-11-01", fechaEntrega: "2026-03-15", progreso: 100, estado: "completada" },
  { id: "6", numero: "OP-2026-037", modelo: "Saldivia Gran Turismo 420", cliente: "Flecha Bus S.A.", cantidad: 2, fechaInicio: "2025-10-01", fechaEntrega: "2026-02-28", progreso: 100, estado: "entregada" },
  { id: "7", numero: "OP-2026-036", modelo: "Saldivia Minibus 180", cliente: "Municipalidad de Neuquén", cantidad: 5, fechaInicio: "2026-03-15", fechaEntrega: "2026-06-30", progreso: 22, estado: "en_produccion" },
];

const estadoBadge: Record<OrdenProduccion["estado"], { label: string; color: "gray" | "blue" | "green" | "teal" }> = {
  planificada: { label: "Planificada", color: "gray" },
  en_produccion: { label: "En Producción", color: "blue" },
  completada: { label: "Completada", color: "green" },
  entregada: { label: "Entregada", color: "teal" },
};

const columns: ColumnDef<OrdenProduccion>[] = [
  {
    accessorKey: "numero",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>OP <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <p className="text-sm font-medium font-mono">{row.original.numero}</p>,
  },
  {
    accessorKey: "modelo",
    header: "Modelo",
    cell: ({ row }) => <div><p className="text-sm font-medium">{row.original.modelo}</p><p className="text-xs text-muted-foreground">{row.original.cliente}</p></div>,
  },
  {
    accessorKey: "cantidad",
    header: () => <div className="text-center">Cant.</div>,
    cell: ({ row }) => <p className="text-sm font-medium text-center font-mono">{row.original.cantidad}</p>,
  },
  {
    accessorKey: "fechaInicio",
    header: "Inicio",
    cell: ({ row }) => <p className="text-sm text-muted-foreground">{fmtDate(row.original.fechaInicio)}</p>,
  },
  {
    accessorKey: "fechaEntrega",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Entrega <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => {
      const d = new Date(row.original.fechaEntrega);
      const days = Math.ceil((d.getTime() - Date.now()) / 86400000);
      const color = row.original.estado === "entregada" || row.original.estado === "completada" ? "" : days <= 0 ? "text-red-500" : days <= 14 ? "text-amber-500" : "";
      return <p className={`text-sm ${color}`}>{fmtDate(row.original.fechaEntrega)}</p>;
    },
  },
  {
    accessorKey: "progreso",
    header: "Progreso",
    cell: ({ row }) => (
      <div className="flex items-center gap-2 min-w-[120px]">
        <Progress value={row.original.progreso} className="flex-1" />
        <span className="text-xs font-mono text-muted-foreground w-8 text-right">{row.original.progreso}%</span>
      </div>
    ),
  },
  {
    accessorKey: "estado",
    header: "Estado",
    cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; },
  },
];

export default function ProduccionOrdenesPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  const enProduccion = MOCK.filter((o) => o.estado === "en_produccion").length;
  const unidadesTotales = MOCK.reduce((a, o) => a + o.cantidad, 0);
  const completadas = MOCK.filter((o) => o.estado === "completada" || o.estado === "entregada").length;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Ordenes de Produccion</h1><p className="text-sm text-muted-foreground mt-0.5">Planificacion y seguimiento de ordenes de fabricacion</p></div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Nueva OP</Button>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "En produccion", value: enProduccion, icon: FactoryIcon, color: "text-blue-500" },
            { label: "Unidades totales", value: unidadesTotales, icon: TruckIcon, color: "text-amber-500" },
            { label: "Completadas / Entregadas", value: completadas, icon: ClipboardListIcon, color: "text-green-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2"><s.icon className="size-4 text-muted-foreground" /><p className="text-xs text-muted-foreground">{s.label}</p></div>
              <p className={`text-xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar por modelo o cliente..." className="pl-9 bg-card" value={(table.getColumn("modelo")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("modelo")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>{table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron ordenes.</TableCell></TableRow>)}</TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
