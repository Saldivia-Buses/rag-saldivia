"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, CalendarIcon, PackageIcon, DollarSignIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type OC = {
  id: string;
  numero: string;
  proveedor: string;
  descripcion: string;
  fecha: string;
  entrega: string;
  total: number;
  moneda: "ARS" | "USD";
  estado: "borrador" | "enviada" | "confirmada" | "recibida_parcial" | "completa" | "cancelada";
};

const fmtARS = (n: number) => new Intl.NumberFormat("es-AR", { style: "currency", currency: "ARS", maximumFractionDigits: 0 }).format(n);
const fmtUSD = (n: number) => new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 0 }).format(n);

const MOCK: OC[] = [
  { id: "1", numero: "OC-2026-089", proveedor: "Aceros Bragado S.A.", descripcion: "Chapa naval 3mm x 1500mm — 200 chapas", fecha: "2026-04-08", entrega: "2026-04-22", total: 18500000, moneda: "ARS", estado: "enviada" },
  { id: "2", numero: "OC-2026-088", proveedor: "Lincoln Electric", descripcion: "Alambre ER70S-6 + gas Ar/CO2", fecha: "2026-04-05", entrega: "2026-04-12", total: 6700000, moneda: "ARS", estado: "confirmada" },
  { id: "3", numero: "OC-2026-087", proveedor: "Pinturas Colorín", descripcion: "Epoxi + PU línea transporte", fecha: "2026-04-03", entrega: "2026-04-15", total: 4200000, moneda: "ARS", estado: "recibida_parcial" },
  { id: "4", numero: "OC-2026-086", proveedor: "ZF Group (Alemania)", descripcion: "Caja automática ZF 6HP", fecha: "2026-03-28", entrega: "2026-05-15", total: 42000, moneda: "USD", estado: "confirmada" },
  { id: "5", numero: "OC-2026-085", proveedor: "Vidrios San Justo", descripcion: "Parabrisas laminado — modelo SB420", fecha: "2026-03-25", entrega: "2026-04-05", total: 3800000, moneda: "ARS", estado: "completa" },
  { id: "6", numero: "OC-2026-084", proveedor: "Butacas FAINSA (España)", descripcion: "Butacas urbanas x 44 unidades", fecha: "2026-03-20", entrega: "2026-05-30", total: 28000, moneda: "USD", estado: "enviada" },
];

const estadoBadge: Record<OC["estado"], { label: string; color: "gray" | "blue" | "indigo" | "amber" | "green" | "red" }> = {
  borrador: { label: "Borrador", color: "gray" },
  enviada: { label: "Enviada", color: "blue" },
  confirmada: { label: "Confirmada", color: "indigo" },
  recibida_parcial: { label: "Parcial", color: "amber" },
  completa: { label: "Completa", color: "green" },
  cancelada: { label: "Cancelada", color: "red" },
};

const columns: ColumnDef<OC>[] = [
  {
    accessorKey: "numero",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>OC <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <p className="text-sm font-medium font-mono">{row.original.numero}</p>,
  },
  {
    accessorKey: "proveedor",
    header: "Proveedor",
    cell: ({ row }) => <div><p className="text-sm font-medium">{row.original.proveedor}</p><p className="text-xs text-muted-foreground line-clamp-1">{row.original.descripcion}</p></div>,
  },
  {
    accessorKey: "entrega",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Entrega <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => {
      const d = new Date(row.original.entrega);
      const days = Math.ceil((d.getTime() - Date.now()) / 86400000);
      const color = days <= 0 ? "text-red-500" : days <= 7 ? "text-amber-500" : "";
      return <div className="flex items-center gap-2 text-sm"><CalendarIcon className="size-3.5 text-muted-foreground" /><span className={color}>{d.toLocaleDateString("es-AR", { day: "2-digit", month: "short" })}</span></div>;
    },
  },
  {
    accessorKey: "total",
    header: () => <div className="text-right">Total</div>,
    cell: ({ row }) => <div className="text-right text-sm font-medium font-mono">{row.original.moneda === "USD" ? fmtUSD(row.original.total) : fmtARS(row.original.total)}</div>,
  },
  { accessorKey: "estado", header: "Estado", cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; } },
];

export default function OrdenesCompraPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Órdenes de Compra</h1><p className="text-sm text-muted-foreground mt-0.5">Emisión y seguimiento de OC</p></div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Nueva OC</Button>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Pendientes", value: MOCK.filter((o) => !["completa", "cancelada"].includes(o.estado)).length, icon: PackageIcon, color: "text-blue-500" },
            { label: "Total ARS pendiente", value: fmtARS(MOCK.filter((o) => o.moneda === "ARS" && !["completa", "cancelada"].includes(o.estado)).reduce((a, o) => a + o.total, 0)), icon: DollarSignIcon, color: "text-amber-500" },
            { label: "Total USD pendiente", value: fmtUSD(MOCK.filter((o) => o.moneda === "USD" && !["completa", "cancelada"].includes(o.estado)).reduce((a, o) => a + o.total, 0)), icon: DollarSignIcon, color: "text-green-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2"><s.icon className="size-4 text-muted-foreground" /><p className="text-xs text-muted-foreground">{s.label}</p></div>
              <p className={`text-xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar por proveedor..." className="pl-9 bg-card" value={(table.getColumn("proveedor")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("proveedor")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>{table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron órdenes.</TableCell></TableRow>)}</TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
