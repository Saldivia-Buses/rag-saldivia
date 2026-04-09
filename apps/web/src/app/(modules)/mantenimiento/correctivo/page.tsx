"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, CalendarIcon, AlertTriangleIcon, WrenchIcon, ClockIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type OrdenCorrectiva = {
  id: string;
  equipo: string;
  codigoEquipo: string;
  falla: string;
  prioridad: "baja" | "media" | "alta" | "critica";
  reportada: string;
  resuelta: string | null;
  responsable: string;
  estado: "abierta" | "en_curso" | "resuelta" | "esperando_repuesto";
};

const MOCK: OrdenCorrectiva[] = [
  { id: "MC-145", equipo: "Torno CNC Romi", codigoEquipo: "TOR-001", falla: "Falla en husillo — vibración excesiva", prioridad: "critica", reportada: "2026-04-02", resuelta: null, responsable: "Roberto Fernández", estado: "esperando_repuesto" },
  { id: "MC-144", equipo: "Cabina de pintura", codigoEquipo: "PNT-001", falla: "Extractor no arranca", prioridad: "alta", reportada: "2026-04-05", resuelta: null, responsable: "Roberto Fernández", estado: "en_curso" },
  { id: "MC-143", equipo: "Soldadora MIG Lincoln", codigoEquipo: "SOL-001", falla: "Arrastre de alambre irregular", prioridad: "media", reportada: "2026-04-03", resuelta: "2026-04-04", responsable: "Carlos Gómez", estado: "resuelta" },
  { id: "MC-142", equipo: "Autoelevador Toyota", codigoEquipo: "AUT-001", falla: "Pérdida de aceite hidráulico", prioridad: "alta", reportada: "2026-03-28", resuelta: "2026-03-30", responsable: "Roberto Fernández", estado: "resuelta" },
  { id: "MC-141", equipo: "Compresor Atlas Copco", codigoEquipo: "COM-001", falla: "Presión baja — válvula check", prioridad: "media", reportada: "2026-03-20", resuelta: "2026-03-22", responsable: "Roberto Fernández", estado: "resuelta" },
];

const prioridadBadge: Record<OrdenCorrectiva["prioridad"], { label: string; color: "gray" | "amber" | "red" | "red" }> = {
  baja: { label: "Baja", color: "gray" },
  media: { label: "Media", color: "amber" },
  alta: { label: "Alta", color: "red" },
  critica: { label: "Crítica", color: "red" },
};
const estadoBadge: Record<OrdenCorrectiva["estado"], { label: string; color: "amber" | "blue" | "green" | "indigo" }> = {
  abierta: { label: "Abierta", color: "amber" },
  en_curso: { label: "En curso", color: "blue" },
  resuelta: { label: "Resuelta", color: "green" },
  esperando_repuesto: { label: "Esp. repuesto", color: "indigo" },
};

const columns: ColumnDef<OrdenCorrectiva>[] = [
  {
    accessorKey: "id",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Orden <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <p className="text-sm font-medium font-mono">{row.original.id}</p>,
  },
  {
    accessorKey: "equipo",
    header: "Equipo",
    cell: ({ row }) => <div><p className="text-sm font-medium">{row.original.equipo}</p><p className="text-xs text-muted-foreground font-mono">{row.original.codigoEquipo}</p></div>,
  },
  { accessorKey: "falla", header: "Falla", cell: ({ row }) => <span className="text-sm line-clamp-1 max-w-[250px]">{row.original.falla}</span> },
  { accessorKey: "prioridad", header: "Prioridad", cell: ({ row }) => { const p = prioridadBadge[row.original.prioridad]; return <Badge variant="outline" color={p.color}>{p.label}</Badge>; } },
  {
    accessorKey: "reportada",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Reportada <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <div className="flex items-center gap-2 text-sm"><CalendarIcon className="size-3.5 text-muted-foreground" />{new Date(row.original.reportada).toLocaleDateString("es-AR", { day: "2-digit", month: "short" })}</div>,
  },
  { accessorKey: "responsable", header: "Responsable", cell: ({ row }) => <span className="text-sm">{row.original.responsable}</span> },
  { accessorKey: "estado", header: "Estado", cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; } },
];

export default function CorrectivoPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  const abiertas = MOCK.filter((o) => o.estado !== "resuelta").length;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Mantenimiento Correctivo</h1><p className="text-sm text-muted-foreground mt-0.5">Fallas y reparaciones</p></div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Reportar falla</Button>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Abiertas", value: abiertas, icon: AlertTriangleIcon, color: "text-amber-500" },
            { label: "Esp. repuesto", value: MOCK.filter((o) => o.estado === "esperando_repuesto").length, icon: ClockIcon, color: "text-indigo-500" },
            { label: "Resueltas (mes)", value: MOCK.filter((o) => o.estado === "resuelta").length, icon: WrenchIcon, color: "text-green-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2"><s.icon className="size-4 text-muted-foreground" /><p className="text-xs text-muted-foreground">{s.label}</p></div>
              <p className={`text-2xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar por equipo o falla..." className="pl-9 bg-card" value={(table.getColumn("equipo")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("equipo")?.setFilterValue(e.target.value)} /></div>
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
