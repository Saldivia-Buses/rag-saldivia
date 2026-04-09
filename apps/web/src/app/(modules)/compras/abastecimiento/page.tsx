"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, SearchIcon, PackageIcon, AlertTriangleIcon, TrendingDownIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Material = {
  id: string;
  codigo: string;
  nombre: string;
  unidad: string;
  stockActual: number;
  stockMinimo: number;
  stockMaximo: number;
  ultimaCompra: string;
  proveedor: string;
  estado: "ok" | "bajo" | "critico" | "sin_stock";
};

const MOCK: Material[] = [
  { id: "1", codigo: "MAT-001", nombre: "Chapa naval 3mm x 1500mm", unidad: "chapas", stockActual: 120, stockMinimo: 50, stockMaximo: 300, ultimaCompra: "2026-04-08", proveedor: "Aceros Bragado", estado: "ok" },
  { id: "2", codigo: "MAT-002", nombre: "Alambre ER70S-6 1.2mm", unidad: "kg", stockActual: 180, stockMinimo: 100, stockMaximo: 500, ultimaCompra: "2026-04-05", proveedor: "Lincoln Electric", estado: "ok" },
  { id: "3", codigo: "MAT-003", nombre: "Pintura epoxi gris", unidad: "litros", stockActual: 35, stockMinimo: 50, stockMaximo: 200, ultimaCompra: "2026-03-20", proveedor: "Colorín", estado: "bajo" },
  { id: "4", codigo: "MAT-004", nombre: "Perfil C 80x40x2mm", unidad: "barras", stockActual: 8, stockMinimo: 20, stockMaximo: 100, ultimaCompra: "2026-03-15", proveedor: "Aceros Bragado", estado: "critico" },
  { id: "5", codigo: "MAT-005", nombre: "Burletes de puerta", unidad: "metros", stockActual: 0, stockMinimo: 50, stockMaximo: 200, ultimaCompra: "2026-02-28", proveedor: "Goma Plast", estado: "sin_stock" },
  { id: "6", codigo: "MAT-006", nombre: "Tornillos M10x30 gr 8.8", unidad: "unidades", stockActual: 2400, stockMinimo: 500, stockMaximo: 5000, ultimaCompra: "2026-03-10", proveedor: "Bulonera Central", estado: "ok" },
  { id: "7", codigo: "MAT-007", nombre: "Gas Ar/CO2 80/20", unidad: "tubos", stockActual: 3, stockMinimo: 5, stockMaximo: 15, ultimaCompra: "2026-04-01", proveedor: "AGA", estado: "bajo" },
];

const estadoBadge: Record<Material["estado"], { label: string; color: "green" | "amber" | "red" | "gray" }> = {
  ok: { label: "OK", color: "green" },
  bajo: { label: "Bajo", color: "amber" },
  critico: { label: "Crítico", color: "red" },
  sin_stock: { label: "Sin stock", color: "red" },
};

const columns: ColumnDef<Material>[] = [
  {
    accessorKey: "nombre",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Material <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <div><p className="text-sm font-medium">{row.original.nombre}</p><p className="text-xs text-muted-foreground font-mono">{row.original.codigo} · {row.original.unidad}</p></div>,
  },
  {
    accessorKey: "stockActual",
    header: "Stock",
    cell: ({ row }) => {
      const m = row.original;
      const pct = m.stockMaximo > 0 ? Math.round((m.stockActual / m.stockMaximo) * 100) : 0;
      return (
        <div className="w-28">
          <div className="flex items-center justify-between text-xs mb-1">
            <span className="font-medium">{m.stockActual}</span>
            <span className="text-muted-foreground">/ {m.stockMaximo}</span>
          </div>
          <Progress value={pct} className="h-1.5" />
        </div>
      );
    },
  },
  { accessorKey: "proveedor", header: "Proveedor", cell: ({ row }) => <span className="text-sm">{row.original.proveedor}</span> },
  { accessorKey: "estado", header: "Estado", cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; } },
];

export default function AbastecimientoPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Abastecimiento</h1><p className="text-sm text-muted-foreground mt-0.5">Control de stock de materiales</p></div>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Materiales", value: MOCK.length, icon: PackageIcon, color: "" },
            { label: "Stock bajo", value: MOCK.filter((m) => m.estado === "bajo").length, icon: TrendingDownIcon, color: "text-amber-500" },
            { label: "Crítico / sin stock", value: MOCK.filter((m) => m.estado === "critico" || m.estado === "sin_stock").length, icon: AlertTriangleIcon, color: "text-red-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2"><s.icon className="size-4 text-muted-foreground" /><p className="text-xs text-muted-foreground">{s.label}</p></div>
              <p className={`text-2xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar material..." className="pl-9 bg-card" value={(table.getColumn("nombre")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("nombre")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>{table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron materiales.</TableCell></TableRow>)}</TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
