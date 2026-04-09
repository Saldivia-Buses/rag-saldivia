"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, SearchIcon, ClipboardCheckIcon, ShieldCheckIcon, AlertTriangleIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type PreentregaUnidad = {
  id: string;
  chasis: string;
  modelo: string;
  cliente: string;
  itemsAprobados: number;
  itemsRechazados: number;
  itemsPendientes: number;
  totalItems: number;
  inspector: string;
  estado: "aprobado" | "rechazado" | "pendiente";
  fechaEntrega: string;
};

const fmtDate = (d: string) => new Date(d).toLocaleDateString("es-AR", { day: "2-digit", month: "short", year: "numeric" });

const MOCK: PreentregaUnidad[] = [
  { id: "1", chasis: "9BRS38000R0001890", modelo: "Saldivia Interurbano 380", cliente: "Empresa San José S.R.L.", itemsAprobados: 42, itemsRechazados: 0, itemsPendientes: 3, totalItems: 45, inspector: "Laura Fernández", estado: "pendiente", fechaEntrega: "2026-04-20" },
  { id: "2", chasis: "9BRS38000R0001892", modelo: "Saldivia Interurbano 380", cliente: "Empresa San José S.R.L.", itemsAprobados: 45, itemsRechazados: 0, itemsPendientes: 0, totalItems: 45, inspector: "Laura Fernández", estado: "aprobado", fechaEntrega: "2026-04-18" },
  { id: "3", chasis: "9BRS32000R0002050", modelo: "Saldivia Urbano 320", cliente: "Municipalidad de Córdoba", itemsAprobados: 38, itemsRechazados: 4, itemsPendientes: 3, totalItems: 45, inspector: "Pablo Acosta", estado: "rechazado", fechaEntrega: "2026-04-25" },
  { id: "4", chasis: "9BRS42000R0001200", modelo: "Saldivia Gran Turismo 420", cliente: "Flecha Bus S.A.", itemsAprobados: 48, itemsRechazados: 0, itemsPendientes: 0, totalItems: 48, inspector: "Laura Fernández", estado: "aprobado", fechaEntrega: "2026-04-12" },
  { id: "5", chasis: "9BRS42000R0001201", modelo: "Saldivia Gran Turismo 420", cliente: "Flecha Bus S.A.", itemsAprobados: 44, itemsRechazados: 1, itemsPendientes: 3, totalItems: 48, inspector: "Pablo Acosta", estado: "pendiente", fechaEntrega: "2026-04-15" },
  { id: "6", chasis: "9BRS32000R0002051", modelo: "Saldivia Urbano 320", cliente: "Transporte Automotor La Estrella", itemsAprobados: 30, itemsRechazados: 2, itemsPendientes: 13, totalItems: 45, inspector: "Laura Fernández", estado: "pendiente", fechaEntrega: "2026-05-10" },
  { id: "7", chasis: "9BRS50000R0000802", modelo: "Saldivia Doble Piso 500DD", cliente: "Andesmar S.A.", itemsAprobados: 50, itemsRechazados: 0, itemsPendientes: 2, totalItems: 52, inspector: "Pablo Acosta", estado: "pendiente", fechaEntrega: "2026-05-20" },
  { id: "8", chasis: "9BRS18000R0003205", modelo: "Saldivia Minibus 180", cliente: "Municipalidad de Neuquén", itemsAprobados: 40, itemsRechazados: 0, itemsPendientes: 0, totalItems: 40, inspector: "Laura Fernández", estado: "aprobado", fechaEntrega: "2026-04-10" },
];

const estadoBadge: Record<PreentregaUnidad["estado"], { label: string; color: "green" | "red" | "amber" }> = {
  aprobado: { label: "Aprobado", color: "green" },
  rechazado: { label: "Rechazado", color: "red" },
  pendiente: { label: "Pendiente", color: "amber" },
};

const columns: ColumnDef<PreentregaUnidad>[] = [
  {
    accessorKey: "chasis",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Chasis <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <div><p className="text-sm font-medium font-mono">{row.original.chasis}</p><p className="text-xs text-muted-foreground">{row.original.modelo}</p></div>,
  },
  {
    accessorKey: "cliente",
    header: "Cliente",
    cell: ({ row }) => <p className="text-sm">{row.original.cliente}</p>,
  },
  {
    accessorKey: "itemsAprobados",
    header: "Checklist",
    cell: ({ row }) => {
      const { itemsAprobados, itemsRechazados, itemsPendientes, totalItems } = row.original;
      const pct = Math.round((itemsAprobados / totalItems) * 100);
      return (
        <div className="min-w-[160px]">
          <div className="flex items-center gap-2 mb-1">
            <Progress value={pct} className="flex-1" />
            <span className="text-xs font-mono text-muted-foreground w-8 text-right">{pct}%</span>
          </div>
          <div className="flex gap-2 text-[10px]">
            <span className="text-green-500">{itemsAprobados} ok</span>
            {itemsRechazados > 0 && <span className="text-red-500">{itemsRechazados} rech.</span>}
            {itemsPendientes > 0 && <span className="text-amber-500">{itemsPendientes} pend.</span>}
          </div>
        </div>
      );
    },
  },
  {
    accessorKey: "inspector",
    header: "Inspector",
    cell: ({ row }) => <p className="text-sm">{row.original.inspector}</p>,
  },
  {
    accessorKey: "estado",
    header: "Estado",
    cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; },
  },
  {
    accessorKey: "fechaEntrega",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Entrega <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => {
      const d = new Date(row.original.fechaEntrega);
      const days = Math.ceil((d.getTime() - Date.now()) / 86400000);
      const color = row.original.estado === "aprobado" ? "" : days <= 3 ? "text-red-500" : days <= 7 ? "text-amber-500" : "";
      return <p className={`text-sm ${color}`}>{fmtDate(row.original.fechaEntrega)}</p>;
    },
  },
];

export default function ProduccionPreentregaPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  const aprobadas = MOCK.filter((u) => u.estado === "aprobado").length;
  const rechazadas = MOCK.filter((u) => u.estado === "rechazado").length;
  const pendientes = MOCK.filter((u) => u.estado === "pendiente").length;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Inspeccion Preentrega</h1><p className="text-sm text-muted-foreground mt-0.5">Checklist de calidad previo a la entrega al cliente</p></div>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Aprobadas", value: aprobadas, icon: ShieldCheckIcon, color: "text-green-500" },
            { label: "Rechazadas", value: rechazadas, icon: AlertTriangleIcon, color: "text-red-500" },
            { label: "Pendientes", value: pendientes, icon: ClipboardCheckIcon, color: "text-amber-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2"><s.icon className="size-4 text-muted-foreground" /><p className="text-xs text-muted-foreground">{s.label}</p></div>
              <p className={`text-xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar por chasis o cliente..." className="pl-9 bg-card" value={(table.getColumn("chasis")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("chasis")?.setFilterValue(e.target.value)} /></div>
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
