"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, CalendarIcon, AlertTriangleIcon, MapPinIcon, ShieldAlertIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Incidente = {
  id: string;
  fecha: string;
  tipo: "accidente" | "incidente" | "casi_accidente" | "enfermedad_profesional";
  gravedad: "leve" | "moderado" | "grave" | "fatal";
  area: string;
  descripcion: string;
  diasPerdidos: number;
  estado: "abierto" | "investigacion" | "cerrado";
  involucrado: string;
};

const MOCK: Incidente[] = [
  { id: "INC-012", fecha: "2026-04-07", tipo: "incidente", gravedad: "leve", area: "Línea de soldadura", descripcion: "Salpicadura menor en antebrazo", diasPerdidos: 0, estado: "cerrado", involucrado: "Carlos Gómez" },
  { id: "INC-011", fecha: "2026-03-28", tipo: "accidente", gravedad: "moderado", area: "Sector de corte", descripcion: "Corte en mano con amoladora", diasPerdidos: 5, estado: "cerrado", involucrado: "Diego Pérez" },
  { id: "INC-010", fecha: "2026-03-15", tipo: "casi_accidente", gravedad: "leve", area: "Patio de maniobras", descripcion: "Caída de carga suspendida (sin lesionados)", diasPerdidos: 0, estado: "cerrado", involucrado: "—" },
  { id: "INC-009", fecha: "2026-02-20", tipo: "accidente", gravedad: "grave", area: "Planta de ensamble", descripcion: "Caída de altura desde plataforma", diasPerdidos: 30, estado: "investigacion", involucrado: "Roberto Fernández" },
  { id: "INC-008", fecha: "2026-01-10", tipo: "incidente", gravedad: "leve", area: "Depósito", descripcion: "Derrame de solvente", diasPerdidos: 0, estado: "cerrado", involucrado: "—" },
];

const gravedadBadge: Record<Incidente["gravedad"], { label: string; color: "green" | "amber" | "red" | "gray" }> = {
  leve: { label: "Leve", color: "green" },
  moderado: { label: "Moderado", color: "amber" },
  grave: { label: "Grave", color: "red" },
  fatal: { label: "Fatal", color: "red" },
};
const estadoBadge: Record<Incidente["estado"], { label: string; color: "amber" | "blue" | "green" }> = {
  abierto: { label: "Abierto", color: "amber" },
  investigacion: { label: "Investigación", color: "blue" },
  cerrado: { label: "Cerrado", color: "green" },
};
const tipoLabel: Record<Incidente["tipo"], string> = { accidente: "Accidente", incidente: "Incidente", casi_accidente: "Casi accidente", enfermedad_profesional: "Enf. profesional" };

const columns: ColumnDef<Incidente>[] = [
  {
    accessorKey: "id",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>ID <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <div><p className="text-sm font-medium font-mono">{row.original.id}</p><p className="text-xs text-muted-foreground">{tipoLabel[row.original.tipo]}</p></div>,
  },
  {
    accessorKey: "fecha",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Fecha <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <div className="flex items-center gap-2 text-sm"><CalendarIcon className="size-3.5 text-muted-foreground" />{new Date(row.original.fecha).toLocaleDateString("es-AR", { day: "2-digit", month: "short" })}</div>,
  },
  {
    accessorKey: "area",
    header: "Área",
    cell: ({ row }) => <div className="flex items-center gap-2 text-sm"><MapPinIcon className="size-3.5 text-muted-foreground" />{row.original.area}</div>,
  },
  {
    accessorKey: "descripcion",
    header: "Descripción",
    cell: ({ row }) => <span className="text-sm text-muted-foreground line-clamp-1 max-w-[250px]">{row.original.descripcion}</span>,
  },
  {
    accessorKey: "gravedad",
    header: "Gravedad",
    cell: ({ row }) => { const g = gravedadBadge[row.original.gravedad]; return <Badge variant="outline" color={g.color}>{g.label}</Badge>; },
  },
  {
    accessorKey: "diasPerdidos",
    header: "Días perdidos",
    cell: ({ row }) => {
      const d = row.original.diasPerdidos;
      return <span className={`text-sm font-medium ${d > 0 ? "text-red-500" : "text-muted-foreground"}`}>{d > 0 ? d : "—"}</span>;
    },
  },
  {
    accessorKey: "estado",
    header: "Estado",
    cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; },
  },
];

export default function IncidentesPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  const diasPerdidos = MOCK.reduce((a, i) => a + i.diasPerdidos, 0);
  const abiertos = MOCK.filter((i) => i.estado !== "cerrado").length;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Registro de Incidentes</h1><p className="text-sm text-muted-foreground mt-0.5">Accidentes, incidentes y casi-accidentes</p></div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Reportar incidente</Button>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Total registros", value: MOCK.length, icon: ShieldAlertIcon, color: "" },
            { label: "Abiertos", value: abiertos, icon: AlertTriangleIcon, color: "text-amber-500" },
            { label: "Días perdidos (año)", value: diasPerdidos, icon: CalendarIcon, color: "text-red-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2"><s.icon className="size-4 text-muted-foreground" /><p className="text-xs text-muted-foreground">{s.label}</p></div>
              <p className={`text-2xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar por área o descripción..." className="pl-9 bg-card" value={(table.getColumn("area")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("area")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>
              {table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron incidentes.</TableCell></TableRow>)}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
