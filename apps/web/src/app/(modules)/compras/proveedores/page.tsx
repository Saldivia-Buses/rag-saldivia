"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, PlusIcon, SearchIcon, BuildingIcon, GlobeIcon, MailIcon, PhoneIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Proveedor = {
  id: string;
  razonSocial: string;
  cuit: string;
  rubro: string;
  pais: string;
  contacto: string;
  email: string;
  telefono: string;
  estado: "activo" | "suspendido" | "evaluacion";
  calificacion: "A" | "B" | "C";
};

const MOCK: Proveedor[] = [
  { id: "1", razonSocial: "Aceros Bragado S.A.", cuit: "30-55667788-0", rubro: "Aceros y metales", pais: "Argentina", contacto: "Martín Suárez", email: "ventas@acerosbragado.com", telefono: "2342-42-1234", estado: "activo", calificacion: "A" },
  { id: "2", razonSocial: "Lincoln Electric Argentina", cuit: "30-11223344-5", rubro: "Soldadura", pais: "Argentina", contacto: "Laura Méndez", email: "lmendez@lincoln.com.ar", telefono: "11-4567-8900", estado: "activo", calificacion: "A" },
  { id: "3", razonSocial: "Pinturas Colorín S.A.", cuit: "30-44556677-8", rubro: "Pinturas industriales", pais: "Argentina", contacto: "Diego Romero", email: "industrial@colorin.com.ar", telefono: "11-5678-9012", estado: "activo", calificacion: "B" },
  { id: "4", razonSocial: "ZF Group", cuit: "—", rubro: "Transmisiones", pais: "Alemania", contacto: "Hans Weber", email: "hweber@zf.com", telefono: "+49-7541-77-0", estado: "activo", calificacion: "A" },
  { id: "5", razonSocial: "FAINSA", cuit: "—", rubro: "Butacas transporte", pais: "España", contacto: "Jordi Puig", email: "jpuig@fainsa.com", telefono: "+34-93-123-4567", estado: "activo", calificacion: "A" },
  { id: "6", razonSocial: "Vidrios San Justo S.R.L.", cuit: "30-99887766-1", rubro: "Vidrios automotor", pais: "Argentina", contacto: "Pablo Rivas", email: "privas@vidriosanj.com", telefono: "11-6789-0123", estado: "evaluacion", calificacion: "C" },
  { id: "7", razonSocial: "Goma Plast S.A.", cuit: "30-33221100-9", rubro: "Burletes y gomas", pais: "Argentina", contacto: "Silvia Torres", email: "storres@gomaplast.com", telefono: "11-7890-1234", estado: "suspendido", calificacion: "C" },
];

const estadoBadge: Record<Proveedor["estado"], { label: string; color: "green" | "red" | "amber" }> = {
  activo: { label: "Activo", color: "green" },
  suspendido: { label: "Suspendido", color: "red" },
  evaluacion: { label: "En evaluación", color: "amber" },
};
const califColor: Record<Proveedor["calificacion"], string> = { A: "text-green-500", B: "text-amber-500", C: "text-red-500" };

const columns: ColumnDef<Proveedor>[] = [
  {
    accessorKey: "razonSocial",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Proveedor <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => (
      <div className="flex items-center gap-3">
        <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted"><BuildingIcon className="size-4 text-muted-foreground" /></div>
        <div><p className="text-sm font-medium">{row.original.razonSocial}</p><p className="text-xs text-muted-foreground">{row.original.cuit !== "—" ? row.original.cuit : row.original.pais}</p></div>
      </div>
    ),
  },
  { accessorKey: "rubro", header: "Rubro", cell: ({ row }) => <span className="text-sm">{row.original.rubro}</span> },
  {
    accessorKey: "pais",
    header: "País",
    cell: ({ row }) => <div className="flex items-center gap-1.5 text-sm"><GlobeIcon className="size-3.5 text-muted-foreground" />{row.original.pais}</div>,
  },
  {
    accessorKey: "calificacion",
    header: "Calif.",
    cell: ({ row }) => <span className={`text-sm font-bold ${califColor[row.original.calificacion]}`}>{row.original.calificacion}</span>,
  },
  {
    id: "contacto",
    header: "Contacto",
    cell: ({ row }) => (
      <div className="flex items-center gap-1">
        <Button variant="ghost" size="icon" className="size-7"><MailIcon className="size-3.5" /></Button>
        <Button variant="ghost" size="icon" className="size-7"><PhoneIcon className="size-3.5" /></Button>
      </div>
    ),
  },
  { accessorKey: "estado", header: "Estado", cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; } },
];

export default function ProveedoresPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Proveedores</h1><p className="text-sm text-muted-foreground mt-0.5">Padrón y calificación</p></div>
          <Button size="sm"><PlusIcon className="size-4 mr-1.5" />Nuevo proveedor</Button>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Activos", value: MOCK.filter((p) => p.estado === "activo").length, color: "text-green-500" },
            { label: "Internacionales", value: MOCK.filter((p) => p.pais !== "Argentina").length, color: "text-blue-500" },
            { label: "En evaluación", value: MOCK.filter((p) => p.estado === "evaluacion").length, color: "text-amber-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <p className="text-xs text-muted-foreground">{s.label}</p>
              <p className={`text-2xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
        <div className="relative mb-4"><SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" /><Input placeholder="Buscar proveedor..." className="pl-9 bg-card" value={(table.getColumn("razonSocial")?.getFilterValue() as string) ?? ""} onChange={(e) => table.getColumn("razonSocial")?.setFilterValue(e.target.value)} /></div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>{table.getHeaderGroups().map((hg) => (<TableRow key={hg.id}>{hg.headers.map((h) => (<TableHead key={h.id}>{h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}</TableHead>))}</TableRow>))}</TableHeader>
            <TableBody>{table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron proveedores.</TableCell></TableRow>)}</TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
