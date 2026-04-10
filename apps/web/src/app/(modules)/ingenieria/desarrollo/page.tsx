"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, FlaskConicalIcon, CalendarIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Proyecto = {
  id: string;
  codigo: string;
  modelo: string;
  descripcion: string;
  fase: "diseno" | "prototipo" | "validacion" | "produccion";
  progreso: number;
  responsable: string;
  fechaObjetivo: string;
};

const MOCK: Proyecto[] = [
  { id: "1", codigo: "DEV-2026-001", modelo: "SB-450DD", descripcion: "Desarrollo carrocería doble piso para Scania K 440", fase: "prototipo", progreso: 45, responsable: "Ing. Carlos Benedetti", fechaObjetivo: "2026-09-15" },
  { id: "2", codigo: "DEV-2026-002", modelo: "SB-350A", descripcion: "Bus articulado 18m para corredores BRT", fase: "diseno", progreso: 20, responsable: "Ing. María Fernández", fechaObjetivo: "2027-03-01" },
  { id: "3", codigo: "DEV-2026-003", modelo: "SB-420", descripcion: "Rediseño interior accesibilidad universal v3.3", fase: "validacion", progreso: 85, responsable: "Ing. Lucas Peralta", fechaObjetivo: "2026-06-30" },
  { id: "4", codigo: "DEV-2026-004", modelo: "SB-500LD", descripcion: "Integración sistema entretenimiento a bordo", fase: "prototipo", progreso: 55, responsable: "Ing. Sofía Aguirre", fechaObjetivo: "2026-08-15" },
  { id: "5", codigo: "DEV-2025-018", modelo: "SB-380U", descripcion: "Adaptación chasis Agrale MT 15.0 LE", fase: "produccion", progreso: 100, responsable: "Ing. Roberto Mansilla", fechaObjetivo: "2026-04-01" },
  { id: "6", codigo: "DEV-2026-005", modelo: "SB-320E", descripcion: "Sistema de seguridad escolar (GPS + cámaras + alarma)", fase: "diseno", progreso: 10, responsable: "Ing. Andrea Suárez", fechaObjetivo: "2026-12-01" },
  { id: "7", codigo: "DEV-2026-006", modelo: "SB-460LD", descripcion: "Optimización aerodinámica para reducción consumo 8%", fase: "validacion", progreso: 70, responsable: "Ing. Carlos Benedetti", fechaObjetivo: "2026-07-20" },
];

const faseBadge: Record<Proyecto["fase"], { label: string; color: "gray" | "blue" | "amber" | "green" }> = {
  diseno: { label: "Diseño", color: "gray" },
  prototipo: { label: "Prototipo", color: "blue" },
  validacion: { label: "Validación", color: "amber" },
  produccion: { label: "Producción", color: "green" },
};

const columns: ColumnDef<Proyecto>[] = [
  {
    accessorKey: "codigo",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Proyecto <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => (
      <div className="flex items-center gap-3">
        <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted"><FlaskConicalIcon className="size-4 text-muted-foreground" /></div>
        <div><p className="text-sm font-medium font-mono">{row.original.codigo}</p><p className="text-xs text-muted-foreground">{row.original.modelo}</p></div>
      </div>
    ),
  },
  { accessorKey: "descripcion", header: "Descripción", cell: ({ row }) => <p className="text-sm line-clamp-1 max-w-[240px]">{row.original.descripcion}</p> },
  { accessorKey: "fase", header: "Fase", cell: ({ row }) => { const f = faseBadge[row.original.fase]; return <Badge variant="outline" color={f.color}>{f.label}</Badge>; } },
  {
    accessorKey: "progreso",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Avance <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => (
      <div className="flex items-center gap-2 min-w-[120px]">
        <Progress value={row.original.progreso} className="flex-1" />
        <span className="text-xs font-mono text-muted-foreground w-8 text-right">{row.original.progreso}%</span>
      </div>
    ),
  },
  { accessorKey: "responsable", header: "Responsable", cell: ({ row }) => <span className="text-sm">{row.original.responsable}</span> },
  {
    accessorKey: "fechaObjetivo",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Objetivo <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => {
      const d = new Date(row.original.fechaObjetivo);
      const days = Math.ceil((d.getTime() - Date.now()) / 86400000);
      const color = days <= 0 ? "text-red-500" : days <= 30 ? "text-amber-500" : "";
      return <div className="flex items-center gap-1.5 text-sm"><CalendarIcon className="size-3.5 text-muted-foreground" /><span className={color}>{d.toLocaleDateString("es-AR", { day: "2-digit", month: "short", year: "numeric" })}</span></div>;
    },
  },
];

export default function IngenieriaDesarrolloPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  const avgProgress = Math.round(MOCK.reduce((a, p) => a + p.progreso, 0) / MOCK.length);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Desarrollo</h1><p className="text-sm text-muted-foreground mt-0.5">Proyectos de ingeniería en curso</p></div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Nuevo proyecto</Button>
        </div>
        <div className="grid grid-cols-4 gap-3 mb-6">
          {[
            { label: "Proyectos activos", value: MOCK.filter((p) => p.fase !== "produccion").length, color: "text-blue-500" },
            { label: "En prototipo", value: MOCK.filter((p) => p.fase === "prototipo").length, color: "text-indigo-500" },
            { label: "En validación", value: MOCK.filter((p) => p.fase === "validacion").length, color: "text-amber-500" },
            { label: "Avance promedio", value: `${avgProgress}%`, color: "text-green-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <p className="text-xs text-muted-foreground">{s.label}</p>
              <p className={`text-2xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar proyecto..." className="pl-9 bg-card" value={(table.getColumn("codigo")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("codigo")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>{table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron proyectos.</TableCell></TableRow>)}</TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
