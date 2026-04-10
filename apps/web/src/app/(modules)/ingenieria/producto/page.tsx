"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, BusIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Producto = {
  id: string;
  codigo: string;
  nombre: string;
  tipo: "urbano" | "larga_distancia" | "escolar";
  plataforma: string;
  version: string;
  estado: "activo" | "desarrollo" | "descontinuado";
};

const MOCK: Producto[] = [
  { id: "1", codigo: "SB-420", nombre: "Saldivia Urbano 420", tipo: "urbano", plataforma: "Mercedes-Benz OF 1721", version: "v3.2", estado: "activo" },
  { id: "2", codigo: "SB-500LD", nombre: "Saldivia Gran Turismo 500", tipo: "larga_distancia", plataforma: "Scania K 410", version: "v2.1", estado: "activo" },
  { id: "3", codigo: "SB-320E", nombre: "Saldivia Escolar 320", tipo: "escolar", plataforma: "Mercedes-Benz OF 1418", version: "v1.4", estado: "activo" },
  { id: "4", codigo: "SB-450DD", nombre: "Saldivia Doble Piso 450", tipo: "larga_distancia", plataforma: "Scania K 440", version: "v1.0", estado: "desarrollo" },
  { id: "5", codigo: "SB-380U", nombre: "Saldivia Urbano 380", tipo: "urbano", plataforma: "Agrale MT 15.0", version: "v4.0", estado: "activo" },
  { id: "6", codigo: "SB-350A", nombre: "Saldivia Articulado 350", tipo: "urbano", plataforma: "Volvo B340M", version: "v0.8", estado: "desarrollo" },
  { id: "7", codigo: "SB-280", nombre: "Saldivia Midibus 280", tipo: "urbano", plataforma: "Mercedes-Benz LO 916", version: "v2.6", estado: "descontinuado" },
  { id: "8", codigo: "SB-460LD", nombre: "Saldivia Premium 460", tipo: "larga_distancia", plataforma: "Mercedes-Benz O 500 RSD", version: "v1.8", estado: "activo" },
];

const estadoBadge: Record<Producto["estado"], { label: string; color: "green" | "amber" | "red" }> = {
  activo: { label: "Activo", color: "green" },
  desarrollo: { label: "En desarrollo", color: "amber" },
  descontinuado: { label: "Descontinuado", color: "red" },
};

const tipoBadge: Record<Producto["tipo"], { label: string; color: "blue" | "indigo" | "teal" }> = {
  urbano: { label: "Urbano", color: "blue" },
  larga_distancia: { label: "Larga distancia", color: "indigo" },
  escolar: { label: "Escolar", color: "teal" },
};

const columns: ColumnDef<Producto>[] = [
  {
    accessorKey: "codigo",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Código <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => (
      <div className="flex items-center gap-3">
        <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted"><BusIcon className="size-4 text-muted-foreground" /></div>
        <div><p className="text-sm font-medium font-mono">{row.original.codigo}</p><p className="text-xs text-muted-foreground">{row.original.nombre}</p></div>
      </div>
    ),
  },
  { accessorKey: "tipo", header: "Tipo", cell: ({ row }) => { const t = tipoBadge[row.original.tipo]; return <Badge variant="outline" color={t.color}>{t.label}</Badge>; } },
  { accessorKey: "plataforma", header: "Plataforma", cell: ({ row }) => <span className="text-sm">{row.original.plataforma}</span> },
  {
    accessorKey: "version",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Versión <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <span className="text-sm font-mono">{row.original.version}</span>,
  },
  { accessorKey: "estado", header: "Estado", cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; } },
];

export default function IngenieriaProductoPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Catálogo de Producto</h1><p className="text-sm text-muted-foreground mt-0.5">Modelos de carrocería Saldivia</p></div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Nuevo modelo</Button>
        </div>
        <div className="grid grid-cols-4 gap-3 mb-6">
          {[
            { label: "Modelos activos", value: MOCK.filter((p) => p.estado === "activo").length, color: "text-green-500" },
            { label: "En desarrollo", value: MOCK.filter((p) => p.estado === "desarrollo").length, color: "text-amber-500" },
            { label: "Urbanos", value: MOCK.filter((p) => p.tipo === "urbano").length, color: "text-blue-500" },
            { label: "Larga distancia", value: MOCK.filter((p) => p.tipo === "larga_distancia").length, color: "text-indigo-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <p className="text-xs text-muted-foreground">{s.label}</p>
              <p className={`text-2xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar modelo..." className="pl-9 bg-card" value={(table.getColumn("codigo")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("codigo")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>{table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron modelos.</TableCell></TableRow>)}</TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
