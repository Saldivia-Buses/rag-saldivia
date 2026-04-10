"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, CogIcon, WeightIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Pieza = {
  id: string;
  partNumber: string;
  descripcion: string;
  material: string;
  pesoKg: number;
  proveedor: string;
  revision: string;
  estado: "aprobado" | "revision" | "pendiente";
};

const MOCK: Pieza[] = [
  { id: "1", partNumber: "SB-STR-001", descripcion: "Larguero principal chasis — perfil C 250x80x6", material: "Acero SAE 1020", pesoKg: 48.5, proveedor: "Aceros Bragado S.A.", revision: "R3", estado: "aprobado" },
  { id: "2", partNumber: "SB-STR-002", descripcion: "Travesaño central reforzado", material: "Acero SAE 1020", pesoKg: 12.3, proveedor: "Aceros Bragado S.A.", revision: "R2", estado: "aprobado" },
  { id: "3", partNumber: "SB-PNL-010", descripcion: "Panel lateral exterior — chapa estampada", material: "Chapa naval 3mm", pesoKg: 22.0, proveedor: "Aceros Bragado S.A.", revision: "R4", estado: "aprobado" },
  { id: "4", partNumber: "SB-VDR-020", descripcion: "Parabrisas laminado curvo", material: "Vidrio laminado 6mm", pesoKg: 18.7, proveedor: "Vidrios San Justo S.R.L.", revision: "R1", estado: "aprobado" },
  { id: "5", partNumber: "SB-INT-030", descripcion: "Portaequipaje superior — módulo 1200mm", material: "Aluminio 6063-T5", pesoKg: 3.8, proveedor: "Aluar S.A.", revision: "R2", estado: "revision" },
  { id: "6", partNumber: "SB-PIS-040", descripcion: "Piso antideslizante — plancha completa", material: "Madera multilaminada + resina fenólica", pesoKg: 35.0, proveedor: "Placacentro S.A.", revision: "R1", estado: "aprobado" },
  { id: "7", partNumber: "SB-ASI-050", descripcion: "Soporte butaca urbana — base soldada", material: "Acero SAE 1010", pesoKg: 4.2, proveedor: "Metalúrgica del Sur", revision: "R0", estado: "pendiente" },
  { id: "8", partNumber: "SB-ELE-060", descripcion: "Arnés eléctrico principal — tablero", material: "Cobre / PVC autoextinguible", pesoKg: 8.5, proveedor: "Cables Prysmian", revision: "R3", estado: "revision" },
  { id: "9", partNumber: "SB-CLI-070", descripcion: "Conducto A/C — tramo techo", material: "Fibra de vidrio", pesoKg: 6.1, proveedor: "Plásticos Rafaela S.A.", revision: "R1", estado: "aprobado" },
  { id: "10", partNumber: "SB-PNT-080", descripcion: "Máscara frontal — molde principal", material: "PRFV (poliéster reforzado)", pesoKg: 14.0, proveedor: "Plásticos Rafaela S.A.", revision: "R2", estado: "pendiente" },
];

const estadoBadge: Record<Pieza["estado"], { label: string; color: "green" | "amber" | "gray" }> = {
  aprobado: { label: "Aprobado", color: "green" },
  revision: { label: "En revisión", color: "amber" },
  pendiente: { label: "Pendiente", color: "gray" },
};

const columns: ColumnDef<Pieza>[] = [
  {
    accessorKey: "partNumber",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Part Number <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => (
      <div className="flex items-center gap-3">
        <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted"><CogIcon className="size-4 text-muted-foreground" /></div>
        <div><p className="text-sm font-medium font-mono">{row.original.partNumber}</p><p className="text-xs text-muted-foreground line-clamp-1">{row.original.descripcion}</p></div>
      </div>
    ),
  },
  { accessorKey: "material", header: "Material", cell: ({ row }) => <span className="text-sm">{row.original.material}</span> },
  {
    accessorKey: "pesoKg",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Peso <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <div className="flex items-center gap-1.5 text-sm"><WeightIcon className="size-3.5 text-muted-foreground" /><span className="font-mono">{row.original.pesoKg.toFixed(1)} kg</span></div>,
  },
  { accessorKey: "proveedor", header: "Proveedor", cell: ({ row }) => <span className="text-sm">{row.original.proveedor}</span> },
  {
    accessorKey: "revision",
    header: "Rev.",
    cell: ({ row }) => <span className="text-sm font-mono">{row.original.revision}</span>,
  },
  { accessorKey: "estado", header: "Estado", cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; } },
];

export default function IngenieriaDefinicionPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  const pesoTotal = MOCK.reduce((a, p) => a + p.pesoKg, 0);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Definición de Producto</h1><p className="text-sm text-muted-foreground mt-0.5">Especificaciones técnicas y BOM</p></div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Nueva pieza</Button>
        </div>
        <div className="grid grid-cols-4 gap-3 mb-6">
          {[
            { label: "Total piezas", value: MOCK.length, color: "text-blue-500" },
            { label: "Aprobadas", value: MOCK.filter((p) => p.estado === "aprobado").length, color: "text-green-500" },
            { label: "En revisión", value: MOCK.filter((p) => p.estado === "revision").length, color: "text-amber-500" },
            { label: "Peso total", value: `${pesoTotal.toFixed(1)} kg`, color: "text-indigo-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <p className="text-xs text-muted-foreground">{s.label}</p>
              <p className={`text-2xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar pieza o part number..." className="pl-9 bg-card" value={(table.getColumn("partNumber")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("partNumber")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>{table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron piezas.</TableCell></TableRow>)}</TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
