"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, SearchIcon, WrenchIcon, PauseCircleIcon, CheckCircle2Icon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type UnidadProduccion = {
  id: string;
  chasis: string;
  modelo: string;
  op: string;
  etapa: "soldadura" | "pintura" | "ensamble" | "terminacion" | "inspeccion";
  diasEnEtapa: number;
  operador: string;
  estado: "en_proceso" | "detenida" | "completada";
};

const MOCK: UnidadProduccion[] = [
  { id: "1", chasis: "9BRS38200R0001542", modelo: "Saldivia Gran Turismo 420", op: "OP-2026-042", etapa: "ensamble", diasEnEtapa: 3, operador: "Carlos Méndez", estado: "en_proceso" },
  { id: "2", chasis: "9BRS38200R0001543", modelo: "Saldivia Gran Turismo 420", op: "OP-2026-042", etapa: "soldadura", diasEnEtapa: 8, operador: "Miguel Ríos", estado: "en_proceso" },
  { id: "3", chasis: "9BRS32000R0002101", modelo: "Saldivia Urbano 320", op: "OP-2026-041", etapa: "pintura", diasEnEtapa: 2, operador: "Federico Álvarez", estado: "en_proceso" },
  { id: "4", chasis: "9BRS32000R0002102", modelo: "Saldivia Urbano 320", op: "OP-2026-041", etapa: "soldadura", diasEnEtapa: 12, operador: "Roberto Gómez", estado: "detenida" },
  { id: "5", chasis: "9BRS38000R0001890", modelo: "Saldivia Interurbano 380", op: "OP-2026-040", etapa: "inspeccion", diasEnEtapa: 1, operador: "Laura Fernández", estado: "en_proceso" },
  { id: "6", chasis: "9BRS38000R0001891", modelo: "Saldivia Interurbano 380", op: "OP-2026-040", etapa: "terminacion", diasEnEtapa: 4, operador: "Martín Pereyra", estado: "en_proceso" },
  { id: "7", chasis: "9BRS32000R0002103", modelo: "Saldivia Urbano 320", op: "OP-2026-041", etapa: "ensamble", diasEnEtapa: 6, operador: "Diego Sánchez", estado: "en_proceso" },
  { id: "8", chasis: "9BRS18000R0003201", modelo: "Saldivia Minibus 180", op: "OP-2026-036", etapa: "pintura", diasEnEtapa: 15, operador: "Hernán Quiroga", estado: "detenida" },
  { id: "9", chasis: "9BRS38000R0001892", modelo: "Saldivia Interurbano 380", op: "OP-2026-040", etapa: "inspeccion", diasEnEtapa: 1, operador: "Laura Fernández", estado: "completada" },
  { id: "10", chasis: "9BRS50000R0000801", modelo: "Saldivia Doble Piso 500DD", op: "OP-2026-039", etapa: "soldadura", diasEnEtapa: 5, operador: "Néstor Villalba", estado: "en_proceso" },
];

const etapaBadge: Record<UnidadProduccion["etapa"], { label: string; color: "amber" | "indigo" | "blue" | "teal" | "green" }> = {
  soldadura: { label: "Soldadura", color: "amber" },
  pintura: { label: "Pintura", color: "indigo" },
  ensamble: { label: "Ensamble", color: "blue" },
  terminacion: { label: "Terminacion", color: "teal" },
  inspeccion: { label: "Inspeccion", color: "green" },
};

const estadoBadge: Record<UnidadProduccion["estado"], { label: string; color: "blue" | "red" | "green" }> = {
  en_proceso: { label: "En Proceso", color: "blue" },
  detenida: { label: "Detenida", color: "red" },
  completada: { label: "Completada", color: "green" },
};

const columns: ColumnDef<UnidadProduccion>[] = [
  {
    accessorKey: "chasis",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Chasis / VIN <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <div><p className="text-sm font-medium font-mono">{row.original.chasis}</p><p className="text-xs text-muted-foreground">{row.original.op}</p></div>,
  },
  {
    accessorKey: "modelo",
    header: "Modelo",
    cell: ({ row }) => <p className="text-sm">{row.original.modelo}</p>,
  },
  {
    accessorKey: "etapa",
    header: "Etapa actual",
    cell: ({ row }) => { const e = etapaBadge[row.original.etapa]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; },
  },
  {
    accessorKey: "diasEnEtapa",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Dias <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => {
      const d = row.original.diasEnEtapa;
      const color = d >= 10 ? "text-red-500 font-semibold" : d >= 5 ? "text-amber-500" : "";
      return <p className={`text-sm font-mono text-center ${color}`}>{d}d</p>;
    },
  },
  {
    accessorKey: "operador",
    header: "Operador",
    cell: ({ row }) => <p className="text-sm">{row.original.operador}</p>,
  },
  {
    accessorKey: "estado",
    header: "Estado",
    cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; },
  },
];

export default function ProduccionSeguimientoPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  const enProceso = MOCK.filter((u) => u.estado === "en_proceso").length;
  const detenidas = MOCK.filter((u) => u.estado === "detenida").length;
  const completadas = MOCK.filter((u) => u.estado === "completada").length;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Seguimiento de Unidades</h1><p className="text-sm text-muted-foreground mt-0.5">Control en tiempo real por unidad en planta</p></div>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "En proceso", value: enProceso, icon: WrenchIcon, color: "text-blue-500" },
            { label: "Detenidas", value: detenidas, icon: PauseCircleIcon, color: "text-red-500" },
            { label: "Completadas", value: completadas, icon: CheckCircle2Icon, color: "text-green-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2"><s.icon className="size-4 text-muted-foreground" /><p className="text-xs text-muted-foreground">{s.label}</p></div>
              <p className={`text-xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar por chasis, modelo u operador..." className="pl-9 bg-card" value={(table.getColumn("chasis")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("chasis")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>{table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron unidades.</TableCell></TableRow>)}</TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
