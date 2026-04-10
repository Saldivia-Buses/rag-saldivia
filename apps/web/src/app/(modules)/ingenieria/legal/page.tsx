"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, ShieldCheckIcon, CalendarIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Certificacion = {
  id: string;
  numero: string;
  tipo: "homologacion" | "cnrt" | "vtv" | "exportacion";
  modelo: string;
  autoridad: string;
  fechaEmision: string;
  fechaVencimiento: string;
  estado: "vigente" | "por_vencer" | "vencido";
};

const MOCK: Certificacion[] = [
  { id: "1", numero: "HMLG-2024-0156", tipo: "homologacion", modelo: "SB-420", autoridad: "INTI", fechaEmision: "2024-03-15", fechaVencimiento: "2027-03-15", estado: "vigente" },
  { id: "2", numero: "CNRT-2025-0892", tipo: "cnrt", modelo: "SB-420", autoridad: "CNRT", fechaEmision: "2025-01-10", fechaVencimiento: "2026-01-10", estado: "vigente" },
  { id: "3", numero: "HMLG-2023-0098", tipo: "homologacion", modelo: "SB-500LD", autoridad: "INTI", fechaEmision: "2023-06-20", fechaVencimiento: "2026-06-20", estado: "por_vencer" },
  { id: "4", numero: "CNRT-2025-1034", tipo: "cnrt", modelo: "SB-500LD", autoridad: "CNRT", fechaEmision: "2025-02-28", fechaVencimiento: "2026-02-28", estado: "vigente" },
  { id: "5", numero: "EXP-2024-0045", tipo: "exportacion", modelo: "SB-460LD", autoridad: "INTI / Aduana", fechaEmision: "2024-08-10", fechaVencimiento: "2026-08-10", estado: "por_vencer" },
  { id: "6", numero: "VTV-2025-3321", tipo: "vtv", modelo: "SB-380U", autoridad: "RTO Buenos Aires", fechaEmision: "2025-11-05", fechaVencimiento: "2026-11-05", estado: "vigente" },
  { id: "7", numero: "HMLG-2022-0201", tipo: "homologacion", modelo: "SB-280", autoridad: "INTI", fechaEmision: "2022-04-12", fechaVencimiento: "2025-04-12", estado: "vencido" },
  { id: "8", numero: "CNRT-2024-0567", tipo: "cnrt", modelo: "SB-320E", autoridad: "CNRT", fechaEmision: "2024-09-01", fechaVencimiento: "2025-09-01", estado: "vencido" },
  { id: "9", numero: "EXP-2025-0012", tipo: "exportacion", modelo: "SB-500LD", autoridad: "INTI / Aduana", fechaEmision: "2025-03-18", fechaVencimiento: "2027-03-18", estado: "vigente" },
  { id: "10", numero: "HMLG-2025-0310", tipo: "homologacion", modelo: "SB-320E", autoridad: "INTI", fechaEmision: "2025-07-22", fechaVencimiento: "2028-07-22", estado: "vigente" },
];

const estadoBadge: Record<Certificacion["estado"], { label: string; color: "green" | "amber" | "red" }> = {
  vigente: { label: "Vigente", color: "green" },
  por_vencer: { label: "Por vencer", color: "amber" },
  vencido: { label: "Vencido", color: "red" },
};

const tipoBadge: Record<Certificacion["tipo"], { label: string; color: "blue" | "indigo" | "teal" | "gray" }> = {
  homologacion: { label: "Homologación", color: "blue" },
  cnrt: { label: "CNRT", color: "indigo" },
  vtv: { label: "VTV", color: "teal" },
  exportacion: { label: "Exportación", color: "gray" },
};

const fmtDate = (iso: string) => new Date(iso).toLocaleDateString("es-AR", { day: "2-digit", month: "short", year: "numeric" });

const columns: ColumnDef<Certificacion>[] = [
  {
    accessorKey: "numero",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Certificado <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => (
      <div className="flex items-center gap-3">
        <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted"><ShieldCheckIcon className="size-4 text-muted-foreground" /></div>
        <div><p className="text-sm font-medium font-mono">{row.original.numero}</p><p className="text-xs text-muted-foreground">{row.original.modelo}</p></div>
      </div>
    ),
  },
  { accessorKey: "tipo", header: "Tipo", cell: ({ row }) => { const t = tipoBadge[row.original.tipo]; return <Badge variant="outline" color={t.color}>{t.label}</Badge>; } },
  { accessorKey: "autoridad", header: "Autoridad", cell: ({ row }) => <span className="text-sm">{row.original.autoridad}</span> },
  {
    accessorKey: "fechaEmision",
    header: "Emisión",
    cell: ({ row }) => <div className="flex items-center gap-1.5 text-sm"><CalendarIcon className="size-3.5 text-muted-foreground" />{fmtDate(row.original.fechaEmision)}</div>,
  },
  {
    accessorKey: "fechaVencimiento",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Vencimiento <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => {
      const d = new Date(row.original.fechaVencimiento);
      const days = Math.ceil((d.getTime() - Date.now()) / 86400000);
      const color = days <= 0 ? "text-red-500" : days <= 90 ? "text-amber-500" : "";
      return <div className="flex items-center gap-1.5 text-sm"><CalendarIcon className="size-3.5 text-muted-foreground" /><span className={color}>{fmtDate(row.original.fechaVencimiento)}</span></div>;
    },
  },
  { accessorKey: "estado", header: "Estado", cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; } },
];

export default function IngenieriaLegalPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Legal y Técnica</h1><p className="text-sm text-muted-foreground mt-0.5">Certificaciones y homologaciones</p></div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Nueva certificación</Button>
        </div>
        <div className="grid grid-cols-4 gap-3 mb-6">
          {[
            { label: "Vigentes", value: MOCK.filter((c) => c.estado === "vigente").length, color: "text-green-500" },
            { label: "Por vencer", value: MOCK.filter((c) => c.estado === "por_vencer").length, color: "text-amber-500" },
            { label: "Vencidos", value: MOCK.filter((c) => c.estado === "vencido").length, color: "text-red-500" },
            { label: "Total certificados", value: MOCK.length, color: "text-blue-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <p className="text-xs text-muted-foreground">{s.label}</p>
              <p className={`text-2xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar certificado..." className="pl-9 bg-card" value={(table.getColumn("numero")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("numero")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>{table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron certificaciones.</TableCell></TableRow>)}</TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
