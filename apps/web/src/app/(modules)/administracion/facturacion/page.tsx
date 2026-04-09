"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, CalendarIcon, DollarSignIcon, FileTextIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Factura = {
  id: string;
  numero: string;
  tipo: "A" | "B" | "C" | "E";
  cliente: string;
  cuit: string;
  fecha: string;
  vencimiento: string;
  total: number;
  estado: "emitida" | "cobrada" | "vencida" | "anulada";
};

const fmt = (n: number) => new Intl.NumberFormat("es-AR", { style: "currency", currency: "ARS", maximumFractionDigits: 0 }).format(n);

const MOCK: Factura[] = [
  { id: "1", numero: "0001-00004521", tipo: "A", cliente: "Municipalidad de Tandil", cuit: "30-12345678-9", fecha: "2026-04-08", vencimiento: "2026-05-08", total: 45000000, estado: "emitida" },
  { id: "2", numero: "0001-00004520", tipo: "A", cliente: "Provincia de Buenos Aires", cuit: "30-98765432-1", fecha: "2026-04-05", vencimiento: "2026-05-05", total: 120000000, estado: "emitida" },
  { id: "3", numero: "0001-00004519", tipo: "A", cliente: "ERSA Urbano", cuit: "30-55667788-0", fecha: "2026-03-28", vencimiento: "2026-04-28", total: 78000000, estado: "cobrada" },
  { id: "4", numero: "0001-00004518", tipo: "B", cliente: "Cooperativa El Rápido", cuit: "30-11223344-5", fecha: "2026-03-20", vencimiento: "2026-04-20", total: 35000000, estado: "cobrada" },
  { id: "5", numero: "0001-00004517", tipo: "A", cliente: "Empresa San José", cuit: "30-44556677-8", fecha: "2026-03-15", vencimiento: "2026-04-15", total: 62000000, estado: "vencida" },
  { id: "6", numero: "0001-00004516", tipo: "E", cliente: "Transporte del Oeste (Chile)", cuit: "99-00112233-4", fecha: "2026-03-10", vencimiento: "2026-04-10", total: 95000000, estado: "cobrada" },
];

const estadoBadge: Record<Factura["estado"], { label: string; color: "blue" | "green" | "red" | "gray" }> = {
  emitida: { label: "Emitida", color: "blue" },
  cobrada: { label: "Cobrada", color: "green" },
  vencida: { label: "Vencida", color: "red" },
  anulada: { label: "Anulada", color: "gray" },
};

const columns: ColumnDef<Factura>[] = [
  {
    accessorKey: "numero",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Factura <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <div><p className="text-sm font-medium font-mono">{row.original.numero}</p><p className="text-xs text-muted-foreground">Tipo {row.original.tipo}</p></div>,
  },
  {
    accessorKey: "cliente",
    header: "Cliente",
    cell: ({ row }) => <div><p className="text-sm font-medium">{row.original.cliente}</p><p className="text-xs text-muted-foreground">{row.original.cuit}</p></div>,
  },
  {
    accessorKey: "fecha",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Fecha <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <div className="flex items-center gap-2 text-sm"><CalendarIcon className="size-3.5 text-muted-foreground" />{new Date(row.original.fecha).toLocaleDateString("es-AR", { day: "2-digit", month: "short" })}</div>,
  },
  {
    accessorKey: "vencimiento",
    header: "Vencimiento",
    cell: ({ row }) => {
      const d = new Date(row.original.vencimiento);
      const days = Math.ceil((d.getTime() - Date.now()) / 86400000);
      const color = row.original.estado === "vencida" ? "text-red-500 font-medium" : days <= 7 ? "text-amber-500" : "";
      return <span className={`text-sm ${color}`}>{d.toLocaleDateString("es-AR", { day: "2-digit", month: "short" })}</span>;
    },
  },
  {
    accessorKey: "total",
    header: () => <div className="text-right">Total</div>,
    cell: ({ row }) => <div className="text-right text-sm font-medium font-mono">{fmt(row.original.total)}</div>,
  },
  { accessorKey: "estado", header: "Estado", cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; } },
];

export default function FacturacionPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  const totalEmitido = MOCK.filter((f) => f.estado === "emitida").reduce((a, f) => a + f.total, 0);
  const totalCobrado = MOCK.filter((f) => f.estado === "cobrada").reduce((a, f) => a + f.total, 0);
  const totalVencido = MOCK.filter((f) => f.estado === "vencida").reduce((a, f) => a + f.total, 0);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Facturación</h1><p className="text-sm text-muted-foreground mt-0.5">Emisión y seguimiento de facturas</p></div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Nueva factura</Button>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Pendiente de cobro", value: fmt(totalEmitido), icon: FileTextIcon, color: "text-blue-500" },
            { label: "Cobrado (mes)", value: fmt(totalCobrado), icon: DollarSignIcon, color: "text-green-500" },
            { label: "Vencido", value: fmt(totalVencido), icon: DollarSignIcon, color: "text-red-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2"><s.icon className="size-4 text-muted-foreground" /><p className="text-xs text-muted-foreground">{s.label}</p></div>
              <p className={`text-xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar por cliente..." className="pl-9 bg-card" value={(table.getColumn("cliente")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("cliente")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>
              {table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron facturas.</TableCell></TableRow>)}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
