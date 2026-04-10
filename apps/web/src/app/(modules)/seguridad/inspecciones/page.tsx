"use client";

import {
  type ColumnDef,
  type ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  type SortingState,
  useReactTable,
} from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, CalendarIcon, MapPinIcon, ClipboardCheckIcon, AlertTriangleIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Inspeccion = {
  id: string;
  tipo: "rutinaria" | "especial" | "post_incidente";
  area: string;
  inspector: string;
  fecha: string;
  hallazgos: number;
  criticos: number;
  estado: "programada" | "en_curso" | "completada" | "vencida";
};

const MOCK: Inspeccion[] = [
  { id: "INS-041", tipo: "rutinaria", area: "Línea de soldadura", inspector: "Marcos Vidal", fecha: "2026-04-15", hallazgos: 0, criticos: 0, estado: "programada" },
  { id: "INS-040", tipo: "rutinaria", area: "Depósito de pinturas", inspector: "Marcos Vidal", fecha: "2026-04-08", hallazgos: 3, criticos: 1, estado: "completada" },
  { id: "INS-039", tipo: "especial", area: "Planta de ensamble", inspector: "Paula Ríos", fecha: "2026-04-05", hallazgos: 5, criticos: 2, estado: "completada" },
  { id: "INS-038", tipo: "rutinaria", area: "Oficinas administrativas", inspector: "Marcos Vidal", fecha: "2026-04-01", hallazgos: 1, criticos: 0, estado: "completada" },
  { id: "INS-037", tipo: "post_incidente", area: "Sector de corte", inspector: "Paula Ríos", fecha: "2026-03-29", hallazgos: 4, criticos: 3, estado: "completada" },
  { id: "INS-036", tipo: "rutinaria", area: "Línea de pintura", inspector: "Marcos Vidal", fecha: "2026-03-25", hallazgos: 2, criticos: 0, estado: "completada" },
];

const estadoBadge: Record<Inspeccion["estado"], { label: string; color: "amber" | "blue" | "green" | "red" }> = {
  programada: { label: "Programada", color: "amber" },
  en_curso: { label: "En curso", color: "blue" },
  completada: { label: "Completada", color: "green" },
  vencida: { label: "Vencida", color: "red" },
};

const tipoLabel: Record<Inspeccion["tipo"], string> = { rutinaria: "Rutinaria", especial: "Especial", post_incidente: "Post-incidente" };

const columns: ColumnDef<Inspeccion>[] = [
  {
    accessorKey: "id",
    header: ({ column }) => (
      <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>
        Inspección <ArrowUpDownIcon className="ml-1 size-3.5" />
      </Button>
    ),
    cell: ({ row }) => (
      <div>
        <p className="text-sm font-medium font-mono">{row.original.id}</p>
        <p className="text-xs text-muted-foreground">{tipoLabel[row.original.tipo]}</p>
      </div>
    ),
  },
  {
    accessorKey: "area",
    header: "Área",
    cell: ({ row }) => (
      <div className="flex items-center gap-2 text-sm">
        <MapPinIcon className="size-3.5 text-muted-foreground" />
        {row.original.area}
      </div>
    ),
  },
  {
    accessorKey: "fecha",
    header: ({ column }) => (
      <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>
        Fecha <ArrowUpDownIcon className="ml-1 size-3.5" />
      </Button>
    ),
    cell: ({ row }) => (
      <div className="flex items-center gap-2 text-sm">
        <CalendarIcon className="size-3.5 text-muted-foreground" />
        {new Date(row.original.fecha).toLocaleDateString("es-AR", { day: "2-digit", month: "short" })}
      </div>
    ),
  },
  { accessorKey: "inspector", header: "Inspector", cell: ({ row }) => <span className="text-sm">{row.original.inspector}</span> },
  {
    accessorKey: "hallazgos",
    header: "Hallazgos",
    cell: ({ row }) => {
      const { hallazgos, criticos } = row.original;
      if (hallazgos === 0) return <span className="text-sm text-muted-foreground">—</span>;
      return (
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">{hallazgos}</span>
          {criticos > 0 && <Badge variant="outline" color="red" className="text-[10px] px-1.5 py-0">{criticos} crít.</Badge>}
        </div>
      );
    },
  },
  {
    accessorKey: "estado",
    header: "Estado",
    cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; },
  },
];

export default function InspeccionesPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Inspecciones de Seguridad</h1>
            <p className="text-sm text-muted-foreground mt-0.5">Control y seguimiento</p>
          </div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Nueva inspección</Button>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Realizadas", value: MOCK.filter((i) => i.estado === "completada").length, icon: ClipboardCheckIcon, color: "text-green-500" },
            { label: "Hallazgos", value: MOCK.reduce((a, i) => a + i.hallazgos, 0), icon: AlertTriangleIcon, color: "text-amber-500" },
            { label: "Críticos", value: MOCK.reduce((a, i) => a + i.criticos, 0), icon: AlertTriangleIcon, color: "text-red-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2"><s.icon className="size-4 text-muted-foreground" /><p className="text-xs text-muted-foreground">{s.label}</p></div>
              <p className={`text-2xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4">
          <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input placeholder="Buscar por área..." className="pl-9 bg-card" value={(table.getColumn("area")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("area")?.setFilterValue(e.target.value)} />
        </div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>
              {table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron inspecciones.</TableCell></TableRow>)}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
