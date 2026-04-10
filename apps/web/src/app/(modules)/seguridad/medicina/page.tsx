"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, CalendarIcon, HeartPulseIcon, ClockIcon } from "lucide-react";
import { useState } from "react";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Examen = {
  id: string;
  empleado: { nombre: string; apellido: string; legajo: string };
  tipo: "preocupacional" | "periodico" | "egreso" | "reingreso";
  fecha: string;
  vencimiento: string;
  resultado: "apto" | "apto_restricciones" | "no_apto" | "pendiente";
  observaciones: string;
};

const MOCK: Examen[] = [
  { id: "EX-081", empleado: { nombre: "Carlos", apellido: "Gómez", legajo: "L-001" }, tipo: "periodico", fecha: "2026-03-15", vencimiento: "2027-03-15", resultado: "apto", observaciones: "" },
  { id: "EX-082", empleado: { nombre: "María", apellido: "López", legajo: "L-002" }, tipo: "periodico", fecha: "2026-02-20", vencimiento: "2027-02-20", resultado: "apto_restricciones", observaciones: "Restricción carga >15kg" },
  { id: "EX-083", empleado: { nombre: "Roberto", apellido: "Fernández", legajo: "L-005" }, tipo: "periodico", fecha: "2026-01-10", vencimiento: "2027-01-10", resultado: "apto", observaciones: "" },
  { id: "EX-084", empleado: { nombre: "Lucía", apellido: "García", legajo: "L-006" }, tipo: "periodico", fecha: "2025-12-05", vencimiento: "2026-12-05", resultado: "apto", observaciones: "" },
  { id: "EX-085", empleado: { nombre: "Valentina", apellido: "Torres", legajo: "L-008" }, tipo: "preocupacional", fecha: "2023-05-02", vencimiento: "2026-05-02", resultado: "apto", observaciones: "" },
  { id: "EX-086", empleado: { nombre: "Diego", apellido: "Pérez", legajo: "L-007" }, tipo: "egreso", fecha: "2026-04-01", vencimiento: "-", resultado: "apto", observaciones: "Egreso por renuncia" },
  { id: "EX-087", empleado: { nombre: "Juan", apellido: "Martínez", legajo: "L-003" }, tipo: "periodico", fecha: "2026-04-20", vencimiento: "2027-04-20", resultado: "pendiente", observaciones: "Turno agendado" },
];

const resultadoBadge: Record<Examen["resultado"], { label: string; color: "green" | "amber" | "red" | "gray" }> = {
  apto: { label: "Apto", color: "green" },
  apto_restricciones: { label: "Apto c/restricciones", color: "amber" },
  no_apto: { label: "No apto", color: "red" },
  pendiente: { label: "Pendiente", color: "gray" },
};
const tipoLabel: Record<Examen["tipo"], string> = { preocupacional: "Preocupacional", periodico: "Periódico", egreso: "Egreso", reingreso: "Reingreso" };

const columns: ColumnDef<Examen>[] = [
  {
    accessorKey: "empleado",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Empleado <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    sortingFn: (a, b) => a.original.empleado.apellido.localeCompare(b.original.empleado.apellido),
    filterFn: (row, _id, value: string) => `${row.original.empleado.nombre} ${row.original.empleado.apellido} ${row.original.empleado.legajo}`.toLowerCase().includes(value.toLowerCase()),
    cell: ({ row }) => {
      const e = row.original.empleado;
      return (
        <div className="flex items-center gap-3">
          <Avatar className="size-8"><AvatarFallback className="text-xs">{e.nombre[0]}{e.apellido[0]}</AvatarFallback></Avatar>
          <div><p className="text-sm font-medium">{e.nombre} {e.apellido}</p><p className="text-xs text-muted-foreground">{e.legajo}</p></div>
        </div>
      );
    },
  },
  { accessorKey: "tipo", header: "Tipo", cell: ({ row }) => <Badge variant="secondary">{tipoLabel[row.original.tipo]}</Badge> },
  {
    accessorKey: "fecha",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Fecha <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <div className="flex items-center gap-2 text-sm"><CalendarIcon className="size-3.5 text-muted-foreground" />{new Date(row.original.fecha).toLocaleDateString("es-AR", { day: "2-digit", month: "short", year: "numeric" })}</div>,
  },
  {
    accessorKey: "vencimiento",
    header: "Vencimiento",
    cell: ({ row }) => {
      const v = row.original.vencimiento;
      if (v === "-") return <span className="text-sm text-muted-foreground">N/A</span>;
      const vDate = new Date(v);
      const days = Math.ceil((vDate.getTime() - Date.now()) / 86400000);
      const color = days <= 0 ? "text-red-500 font-medium" : days <= 60 ? "text-amber-500" : "";
      return <div className="flex items-center gap-2 text-sm"><ClockIcon className={`size-3.5 ${days <= 0 ? "text-red-500" : days <= 60 ? "text-amber-500" : "text-muted-foreground"}`} /><span className={color}>{vDate.toLocaleDateString("es-AR", { day: "2-digit", month: "short", year: "numeric" })}</span></div>;
    },
  },
  { accessorKey: "resultado", header: "Resultado", cell: ({ row }) => { const r = resultadoBadge[row.original.resultado]; return <Badge variant="outline" color={r.color}>{r.label}</Badge>; } },
];

export default function MedicinaPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Medicina Laboral</h1><p className="text-sm text-muted-foreground mt-0.5">Exámenes médicos y aptitud</p></div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Nuevo examen</Button>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Aptos", value: MOCK.filter((e) => e.resultado === "apto").length, icon: HeartPulseIcon, color: "text-green-500" },
            { label: "Con restricciones", value: MOCK.filter((e) => e.resultado === "apto_restricciones").length, icon: HeartPulseIcon, color: "text-amber-500" },
            { label: "Pendientes", value: MOCK.filter((e) => e.resultado === "pendiente").length, icon: ClockIcon, color: "text-gray-400" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2"><s.icon className="size-4 text-muted-foreground" /><p className="text-xs text-muted-foreground">{s.label}</p></div>
              <p className={`text-2xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar por empleado..." className="pl-9 bg-card" value={(table.getColumn("empleado")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("empleado")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>
              {table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron exámenes.</TableCell></TableRow>)}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
