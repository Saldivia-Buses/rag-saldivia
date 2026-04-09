"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, SearchIcon, CalendarIcon, TrendingUpIcon, TrendingDownIcon, DollarSignIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Asiento = {
  id: string;
  fecha: string;
  numero: string;
  concepto: string;
  tipo: "ingreso" | "egreso" | "ajuste";
  cuenta: string;
  debe: number;
  haber: number;
};

const fmt = (n: number) => n === 0 ? "—" : new Intl.NumberFormat("es-AR", { style: "currency", currency: "ARS", maximumFractionDigits: 0 }).format(n);

const MOCK: Asiento[] = [
  { id: "1", fecha: "2026-04-08", numero: "AS-1201", concepto: "Cobro factura Tandil", tipo: "ingreso", cuenta: "Banco Nación", debe: 45000000, haber: 0 },
  { id: "2", fecha: "2026-04-08", numero: "AS-1201", concepto: "Cobro factura Tandil", tipo: "ingreso", cuenta: "Deudores por ventas", debe: 0, haber: 45000000 },
  { id: "3", fecha: "2026-04-05", numero: "AS-1200", concepto: "Pago energía eléctrica", tipo: "egreso", cuenta: "Gastos servicios", debe: 8900000, haber: 0 },
  { id: "4", fecha: "2026-04-05", numero: "AS-1200", concepto: "Pago energía eléctrica", tipo: "egreso", cuenta: "Banco Nación", debe: 0, haber: 8900000 },
  { id: "5", fecha: "2026-04-01", numero: "AS-1199", concepto: "Liquidación sueldos marzo", tipo: "egreso", cuenta: "Sueldos y jornales", debe: 42000000, haber: 0 },
  { id: "6", fecha: "2026-04-01", numero: "AS-1199", concepto: "Liquidación sueldos marzo", tipo: "egreso", cuenta: "Banco Nación", debe: 0, haber: 42000000 },
  { id: "7", fecha: "2026-03-31", numero: "AS-1198", concepto: "Ajuste tipo de cambio", tipo: "ajuste", cuenta: "Diferencia de cambio", debe: 1200000, haber: 0 },
  { id: "8", fecha: "2026-03-31", numero: "AS-1198", concepto: "Ajuste tipo de cambio", tipo: "ajuste", cuenta: "Deudores por ventas (USD)", debe: 0, haber: 1200000 },
];

const tipoBadge: Record<Asiento["tipo"], { label: string; color: "green" | "red" | "gray" }> = {
  ingreso: { label: "Ingreso", color: "green" },
  egreso: { label: "Egreso", color: "red" },
  ajuste: { label: "Ajuste", color: "gray" },
};

const columns: ColumnDef<Asiento>[] = [
  {
    accessorKey: "fecha",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Fecha <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <div className="flex items-center gap-2 text-sm"><CalendarIcon className="size-3.5 text-muted-foreground" />{new Date(row.original.fecha).toLocaleDateString("es-AR", { day: "2-digit", month: "short" })}</div>,
  },
  { accessorKey: "numero", header: "Asiento", cell: ({ row }) => <span className="text-sm font-mono">{row.original.numero}</span> },
  {
    accessorKey: "concepto",
    header: "Concepto",
    cell: ({ row }) => <div><p className="text-sm">{row.original.concepto}</p><p className="text-xs text-muted-foreground">{row.original.cuenta}</p></div>,
  },
  { accessorKey: "tipo", header: "Tipo", cell: ({ row }) => { const t = tipoBadge[row.original.tipo]; return <Badge variant="outline" color={t.color}>{t.label}</Badge>; } },
  {
    accessorKey: "debe",
    header: () => <div className="text-right">Debe</div>,
    cell: ({ row }) => <div className={`text-right text-sm font-mono ${row.original.debe > 0 ? "font-medium" : "text-muted-foreground"}`}>{fmt(row.original.debe)}</div>,
  },
  {
    accessorKey: "haber",
    header: () => <div className="text-right">Haber</div>,
    cell: ({ row }) => <div className={`text-right text-sm font-mono ${row.original.haber > 0 ? "font-medium" : "text-muted-foreground"}`}>{fmt(row.original.haber)}</div>,
  },
];

export default function ContablePage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  const totalDebe = MOCK.reduce((a, m) => a + m.debe, 0);
  const totalHaber = MOCK.reduce((a, m) => a + m.haber, 0);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Contabilidad</h1><p className="text-sm text-muted-foreground mt-0.5">Libro diario — asientos contables</p></div>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Total Debe", value: fmt(totalDebe), icon: TrendingUpIcon, color: "text-blue-500" },
            { label: "Total Haber", value: fmt(totalHaber), icon: TrendingDownIcon, color: "text-blue-500" },
            { label: "Balance", value: fmt(totalDebe - totalHaber), icon: DollarSignIcon, color: totalDebe === totalHaber ? "text-green-500" : "text-red-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2"><s.icon className="size-4 text-muted-foreground" /><p className="text-xs text-muted-foreground">{s.label}</p></div>
              <p className={`text-xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar por concepto..." className="pl-9 bg-card" value={(table.getColumn("concepto")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("concepto")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>
              {table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron asientos.</TableCell></TableRow>)}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
