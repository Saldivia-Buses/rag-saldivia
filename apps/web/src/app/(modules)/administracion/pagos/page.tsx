"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, CalendarIcon, DollarSignIcon, CreditCardIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Pago = {
  id: string;
  proveedor: string;
  concepto: string;
  fecha: string;
  vencimiento: string;
  monto: number;
  metodo: "transferencia" | "cheque" | "efectivo" | "echeq";
  estado: "pendiente" | "pagado" | "vencido" | "programado";
};

const fmt = (n: number) => new Intl.NumberFormat("es-AR", { style: "currency", currency: "ARS", maximumFractionDigits: 0 }).format(n);

const MOCK: Pago[] = [
  { id: "PAG-210", proveedor: "Aceros Bragado S.A.", concepto: "Chapa naval 3mm — lote 45", fecha: "2026-04-08", vencimiento: "2026-04-22", monto: 18500000, metodo: "echeq", estado: "programado" },
  { id: "PAG-209", proveedor: "Pinturas Colorín", concepto: "Pintura epoxi + poliuretano", fecha: "2026-04-05", vencimiento: "2026-04-20", monto: 4200000, metodo: "transferencia", estado: "pendiente" },
  { id: "PAG-208", proveedor: "EDENOR", concepto: "Factura energía marzo 2026", fecha: "2026-04-01", vencimiento: "2026-04-15", monto: 8900000, metodo: "transferencia", estado: "pagado" },
  { id: "PAG-207", proveedor: "AySA", concepto: "Agua marzo 2026", fecha: "2026-04-01", vencimiento: "2026-04-15", monto: 350000, metodo: "transferencia", estado: "pagado" },
  { id: "PAG-206", proveedor: "Lincoln Electric", concepto: "Alambre soldadura + gas", fecha: "2026-03-25", vencimiento: "2026-04-10", monto: 6700000, metodo: "cheque", estado: "vencido" },
  { id: "PAG-205", proveedor: "Sueldos", concepto: "Liquidación marzo 2026", fecha: "2026-03-31", vencimiento: "2026-04-05", monto: 42000000, metodo: "transferencia", estado: "pagado" },
];

const estadoBadge: Record<Pago["estado"], { label: string; color: "amber" | "green" | "red" | "blue" }> = {
  pendiente: { label: "Pendiente", color: "amber" },
  pagado: { label: "Pagado", color: "green" },
  vencido: { label: "Vencido", color: "red" },
  programado: { label: "Programado", color: "blue" },
};

const columns: ColumnDef<Pago>[] = [
  {
    accessorKey: "id",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Pago <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <p className="text-sm font-medium font-mono">{row.original.id}</p>,
  },
  {
    accessorKey: "proveedor",
    header: "Proveedor / Concepto",
    cell: ({ row }) => <div><p className="text-sm font-medium">{row.original.proveedor}</p><p className="text-xs text-muted-foreground line-clamp-1">{row.original.concepto}</p></div>,
  },
  {
    accessorKey: "vencimiento",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Vencimiento <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => {
      const d = new Date(row.original.vencimiento);
      const days = Math.ceil((d.getTime() - Date.now()) / 86400000);
      const color = row.original.estado === "vencido" ? "text-red-500 font-medium" : days <= 5 ? "text-amber-500" : "";
      return <div className="flex items-center gap-2 text-sm"><CalendarIcon className="size-3.5 text-muted-foreground" /><span className={color}>{d.toLocaleDateString("es-AR", { day: "2-digit", month: "short" })}</span></div>;
    },
  },
  {
    accessorKey: "metodo",
    header: "Método",
    cell: ({ row }) => <Badge variant="secondary">{row.original.metodo === "echeq" ? "eCheq" : row.original.metodo.charAt(0).toUpperCase() + row.original.metodo.slice(1)}</Badge>,
  },
  {
    accessorKey: "monto",
    header: () => <div className="text-right">Monto</div>,
    cell: ({ row }) => <div className="text-right text-sm font-medium font-mono">{fmt(row.original.monto)}</div>,
  },
  { accessorKey: "estado", header: "Estado", cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; } },
];

export default function PagosPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  const totalPendiente = MOCK.filter((p) => p.estado === "pendiente" || p.estado === "programado").reduce((a, p) => a + p.monto, 0);
  const totalPagado = MOCK.filter((p) => p.estado === "pagado").reduce((a, p) => a + p.monto, 0);
  const totalVencido = MOCK.filter((p) => p.estado === "vencido").reduce((a, p) => a + p.monto, 0);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Pagos</h1><p className="text-sm text-muted-foreground mt-0.5">Cuentas a pagar y pagos realizados</p></div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Nuevo pago</Button>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Pendiente", value: fmt(totalPendiente), icon: CreditCardIcon, color: "text-amber-500" },
            { label: "Pagado (mes)", value: fmt(totalPagado), icon: DollarSignIcon, color: "text-green-500" },
            { label: "Vencido", value: fmt(totalVencido), icon: DollarSignIcon, color: "text-red-500" },
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
            <TableBody>
              {table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron pagos.</TableCell></TableRow>)}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
