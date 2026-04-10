"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, CalendarIcon, WrenchIcon, CheckCircleIcon, ClockIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type OrdenPreventiva = {
  id: string;
  equipo: string;
  codigoEquipo: string;
  tarea: string;
  frecuencia: string;
  programada: string;
  responsable: string;
  estado: "pendiente" | "en_curso" | "completada" | "vencida";
};

const MOCK: OrdenPreventiva[] = [
  { id: "MP-301", equipo: "Compresor Atlas Copco", codigoEquipo: "COM-001", tarea: "Cambio filtros + aceite", frecuencia: "Trimestral", programada: "2026-04-10", responsable: "Roberto Fernández", estado: "en_curso" },
  { id: "MP-302", equipo: "Puente grúa 10tn", codigoEquipo: "PUE-001", tarea: "Inspección cables y frenos", frecuencia: "Trimestral", programada: "2026-04-15", responsable: "Roberto Fernández", estado: "pendiente" },
  { id: "MP-303", equipo: "Soldadora MIG Lincoln", codigoEquipo: "SOL-001", tarea: "Limpieza antorcha + calibración", frecuencia: "Mensual", programada: "2026-04-20", responsable: "Carlos Gómez", estado: "pendiente" },
  { id: "MP-300", equipo: "Autoelevador Toyota", codigoEquipo: "AUT-001", tarea: "Service completo 500hs", frecuencia: "Semestral", programada: "2026-04-05", responsable: "Roberto Fernández", estado: "completada" },
  { id: "MP-299", equipo: "Cortadora plasma", codigoEquipo: "COR-001", tarea: "Cambio consumibles", frecuencia: "Mensual", programada: "2026-04-01", responsable: "Roberto Fernández", estado: "completada" },
  { id: "MP-298", equipo: "Cabina de pintura", codigoEquipo: "PNT-001", tarea: "Cambio filtros cabina", frecuencia: "Quincenal", programada: "2026-03-25", responsable: "Diego Pérez", estado: "vencida" },
];

const estadoBadge: Record<OrdenPreventiva["estado"], { label: string; color: "amber" | "blue" | "green" | "red" }> = {
  pendiente: { label: "Pendiente", color: "amber" },
  en_curso: { label: "En curso", color: "blue" },
  completada: { label: "Completada", color: "green" },
  vencida: { label: "Vencida", color: "red" },
};

const columns: ColumnDef<OrdenPreventiva>[] = [
  {
    accessorKey: "id",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Orden <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <div><p className="text-sm font-medium font-mono">{row.original.id}</p><p className="text-xs text-muted-foreground">{row.original.frecuencia}</p></div>,
  },
  {
    accessorKey: "equipo",
    header: "Equipo",
    cell: ({ row }) => <div><p className="text-sm font-medium">{row.original.equipo}</p><p className="text-xs text-muted-foreground font-mono">{row.original.codigoEquipo}</p></div>,
  },
  { accessorKey: "tarea", header: "Tarea", cell: ({ row }) => <span className="text-sm">{row.original.tarea}</span> },
  {
    accessorKey: "programada",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Programada <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <div className="flex items-center gap-2 text-sm"><CalendarIcon className="size-3.5 text-muted-foreground" />{new Date(row.original.programada).toLocaleDateString("es-AR", { day: "2-digit", month: "short" })}</div>,
  },
  { accessorKey: "responsable", header: "Responsable", cell: ({ row }) => <span className="text-sm">{row.original.responsable}</span> },
  { accessorKey: "estado", header: "Estado", cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; } },
];

export default function PreventivoPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Mantenimiento Preventivo</h1><p className="text-sm text-muted-foreground mt-0.5">Órdenes programadas</p></div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Nueva orden</Button>
        </div>
        <div className="grid grid-cols-4 gap-3 mb-6">
          {([
            { label: "Pendientes", value: MOCK.filter((o) => o.estado === "pendiente").length, icon: ClockIcon, color: "text-amber-500" },
            { label: "En curso", value: MOCK.filter((o) => o.estado === "en_curso").length, icon: WrenchIcon, color: "text-blue-500" },
            { label: "Completadas", value: MOCK.filter((o) => o.estado === "completada").length, icon: CheckCircleIcon, color: "text-green-500" },
            { label: "Vencidas", value: MOCK.filter((o) => o.estado === "vencida").length, icon: CalendarIcon, color: "text-red-500" },
          ] as const).map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2"><s.icon className="size-4 text-muted-foreground" /><p className="text-xs text-muted-foreground">{s.label}</p></div>
              <p className={`text-2xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar por equipo..." className="pl-9 bg-card" value={(table.getColumn("equipo")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("equipo")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>
              {table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron órdenes.</TableCell></TableRow>)}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
