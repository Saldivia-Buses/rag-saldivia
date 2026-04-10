"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, CogIcon, CalendarIcon, WrenchIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Equipo = {
  id: string;
  codigo: string;
  nombre: string;
  ubicacion: string;
  tipo: string;
  estado: "operativo" | "mantenimiento" | "fuera_servicio";
  ultimoMant: string;
  proximoMant: string;
  horasUso: number;
};

const MOCK: Equipo[] = [
  { id: "1", codigo: "SOL-001", nombre: "Soldadora MIG Lincoln 350", ubicacion: "Línea 1", tipo: "Soldadura", estado: "operativo", ultimoMant: "2026-03-20", proximoMant: "2026-06-20", horasUso: 4520 },
  { id: "2", codigo: "COR-001", nombre: "Cortadora plasma Hypertherm", ubicacion: "Sector corte", tipo: "Corte", estado: "operativo", ultimoMant: "2026-03-01", proximoMant: "2026-06-01", horasUso: 3200 },
  { id: "3", codigo: "PNT-001", nombre: "Cabina de pintura Blowtherm", ubicacion: "Pintura", tipo: "Pintura", estado: "mantenimiento", ultimoMant: "2026-04-05", proximoMant: "2026-04-12", horasUso: 6800 },
  { id: "4", codigo: "PUE-001", nombre: "Puente grúa 10tn", ubicacion: "Ensamble", tipo: "Izaje", estado: "operativo", ultimoMant: "2026-02-15", proximoMant: "2026-05-15", horasUso: 8900 },
  { id: "5", codigo: "COM-001", nombre: "Compresor Atlas Copco GA30", ubicacion: "Sala compresores", tipo: "Neumática", estado: "operativo", ultimoMant: "2026-03-10", proximoMant: "2026-06-10", horasUso: 12000 },
  { id: "6", codigo: "TOR-001", nombre: "Torno CNC Romi", ubicacion: "Mecanizado", tipo: "Mecanizado", estado: "fuera_servicio", ultimoMant: "2026-01-15", proximoMant: "—", horasUso: 15600 },
  { id: "7", codigo: "AUT-001", nombre: "Autoelevador Toyota 2.5tn", ubicacion: "Depósito", tipo: "Transporte", estado: "operativo", ultimoMant: "2026-03-28", proximoMant: "2026-04-28", horasUso: 5400 },
];

const estadoBadge: Record<Equipo["estado"], { label: string; color: "green" | "amber" | "red" }> = {
  operativo: { label: "Operativo", color: "green" },
  mantenimiento: { label: "En mantenim.", color: "amber" },
  fuera_servicio: { label: "Fuera de servicio", color: "red" },
};

const columns: ColumnDef<Equipo>[] = [
  {
    accessorKey: "codigo",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Equipo <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => (
      <div className="flex items-center gap-3">
        <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted"><CogIcon className="size-4 text-muted-foreground" /></div>
        <div><p className="text-sm font-medium">{row.original.nombre}</p><p className="text-xs text-muted-foreground font-mono">{row.original.codigo} · {row.original.tipo}</p></div>
      </div>
    ),
  },
  { accessorKey: "ubicacion", header: "Ubicación", cell: ({ row }) => <span className="text-sm">{row.original.ubicacion}</span> },
  {
    accessorKey: "horasUso",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Horas uso <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <span className="text-sm font-mono">{row.original.horasUso.toLocaleString("es-AR")} hs</span>,
  },
  {
    accessorKey: "proximoMant",
    header: "Próx. mantenimiento",
    cell: ({ row }) => {
      const p = row.original.proximoMant;
      if (p === "—") return <span className="text-sm text-muted-foreground">—</span>;
      const date = new Date(p);
      const days = Math.ceil((date.getTime() - Date.now()) / 86400000);
      const color = days <= 0 ? "text-red-500 font-medium" : days <= 14 ? "text-amber-500" : "";
      return <div className="flex items-center gap-2 text-sm"><CalendarIcon className={`size-3.5 ${days <= 0 ? "text-red-500" : days <= 14 ? "text-amber-500" : "text-muted-foreground"}`} /><span className={color}>{date.toLocaleDateString("es-AR", { day: "2-digit", month: "short" })}</span></div>;
    },
  },
  { accessorKey: "estado", header: "Estado", cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; } },
];

export default function EquiposPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Equipos</h1><p className="text-sm text-muted-foreground mt-0.5">Registro y estado de equipamiento</p></div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Nuevo equipo</Button>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Operativos", value: MOCK.filter((e) => e.estado === "operativo").length, icon: CogIcon, color: "text-green-500" },
            { label: "En mantenimiento", value: MOCK.filter((e) => e.estado === "mantenimiento").length, icon: WrenchIcon, color: "text-amber-500" },
            { label: "Fuera de servicio", value: MOCK.filter((e) => e.estado === "fuera_servicio").length, icon: CogIcon, color: "text-red-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2"><s.icon className="size-4 text-muted-foreground" /><p className="text-xs text-muted-foreground">{s.label}</p></div>
              <p className={`text-2xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar equipo..." className="pl-9 bg-card" value={(table.getColumn("codigo")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("codigo")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>
              {table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron equipos.</TableCell></TableRow>)}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
