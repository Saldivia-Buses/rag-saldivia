"use client";

import { type ColumnDef, type ColumnFiltersState, flexRender, getCoreRowModel, getFilteredRowModel, getSortedRowModel, type SortingState, useReactTable } from "@tanstack/react-table";
import { ArrowUpDownIcon, SearchIcon, GlobeIcon, ShipIcon, CalendarIcon, DollarSignIcon } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type Importacion = {
  id: string;
  despacho: string;
  proveedor: string;
  origen: string;
  descripcion: string;
  incoterm: string;
  valorFOB: number;
  fechaEmbarque: string;
  etaArgentina: string;
  estado: "cotizacion" | "orden_enviada" | "produccion" | "embarcado" | "en_aduana" | "liberado";
};

const fmtUSD = (n: number) => new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 0 }).format(n);

const MOCK: Importacion[] = [
  { id: "1", despacho: "IMP-2026-012", proveedor: "ZF Group", origen: "Alemania", descripcion: "Caja ZF 6HP x 2 unidades", incoterm: "FOB", valorFOB: 84000, fechaEmbarque: "2026-04-20", etaArgentina: "2026-05-25", estado: "produccion" },
  { id: "2", despacho: "IMP-2026-011", proveedor: "FAINSA", origen: "España", descripcion: "Butacas urbanas x 44", incoterm: "CIF", valorFOB: 28000, fechaEmbarque: "2026-04-15", etaArgentina: "2026-05-20", estado: "embarcado" },
  { id: "3", despacho: "IMP-2026-010", proveedor: "Voith Turbo", origen: "Alemania", descripcion: "Retarder hidráulico", incoterm: "FOB", valorFOB: 18500, fechaEmbarque: "2026-03-10", etaArgentina: "2026-04-15", estado: "en_aduana" },
  { id: "4", despacho: "IMP-2026-009", proveedor: "Webasto", origen: "Alemania", descripcion: "Climatizadores techo x 3", incoterm: "CIF", valorFOB: 36000, fechaEmbarque: "2026-02-20", etaArgentina: "2026-03-28", estado: "liberado" },
];

const estadoBadge: Record<Importacion["estado"], { label: string; color: "gray" | "blue" | "indigo" | "amber" | "green" | "teal" }> = {
  cotizacion: { label: "Cotización", color: "gray" },
  orden_enviada: { label: "OC enviada", color: "blue" },
  produccion: { label: "En producción", color: "indigo" },
  embarcado: { label: "Embarcado", color: "teal" },
  en_aduana: { label: "En aduana", color: "amber" },
  liberado: { label: "Liberado", color: "green" },
};

const columns: ColumnDef<Importacion>[] = [
  {
    accessorKey: "despacho",
    header: ({ column }) => <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>Despacho <ArrowUpDownIcon className="ml-1 size-3.5" /></Button>,
    cell: ({ row }) => <p className="text-sm font-medium font-mono">{row.original.despacho}</p>,
  },
  {
    accessorKey: "proveedor",
    header: "Proveedor",
    cell: ({ row }) => <div><p className="text-sm font-medium">{row.original.proveedor}</p><p className="text-xs text-muted-foreground">{row.original.origen} · {row.original.incoterm}</p></div>,
  },
  { accessorKey: "descripcion", header: "Descripción", cell: ({ row }) => <span className="text-sm line-clamp-1 max-w-[200px]">{row.original.descripcion}</span> },
  {
    accessorKey: "valorFOB",
    header: () => <div className="text-right">Valor FOB</div>,
    cell: ({ row }) => <div className="text-right text-sm font-medium font-mono">{fmtUSD(row.original.valorFOB)}</div>,
  },
  {
    accessorKey: "etaArgentina",
    header: "ETA Argentina",
    cell: ({ row }) => {
      const d = new Date(row.original.etaArgentina);
      const days = Math.ceil((d.getTime() - Date.now()) / 86400000);
      return <div className="flex items-center gap-2 text-sm"><ShipIcon className="size-3.5 text-muted-foreground" /><span>{d.toLocaleDateString("es-AR", { day: "2-digit", month: "short" })}</span>{days > 0 && <span className="text-xs text-muted-foreground">({days}d)</span>}</div>;
    },
  },
  { accessorKey: "estado", header: "Estado", cell: ({ row }) => { const e = estadoBadge[row.original.estado]; return <Badge variant="outline" color={e.color}>{e.label}</Badge>; } },
];

export default function ComexPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({ data: MOCK, columns, onSortingChange: setSorting, onColumnFiltersChange: setColumnFilters, getCoreRowModel: getCoreRowModel(), getSortedRowModel: getSortedRowModel(), getFilteredRowModel: getFilteredRowModel(), state: { sorting, columnFilters } });

  const totalFOB = MOCK.reduce((a, i) => a + i.valorFOB, 0);
  const enTransito = MOCK.filter((i) => ["embarcado", "en_aduana"].includes(i.estado)).length;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div><h1 className="text-xl font-semibold tracking-tight">Comercio Exterior</h1><p className="text-sm text-muted-foreground mt-0.5">Importaciones y despachos</p></div>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Operaciones activas", value: MOCK.filter((i) => i.estado !== "liberado").length, icon: GlobeIcon, color: "text-blue-500" },
            { label: "En tránsito", value: enTransito, icon: ShipIcon, color: "text-teal-500" },
            { label: "Valor FOB total", value: fmtUSD(totalFOB), icon: DollarSignIcon, color: "" },
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
            <TableBody>{table.getRowModel().rows.length ? table.getRowModel().rows.map((row) => (<TableRow key={row.id}>{row.getVisibleCells().map((cell) => (<TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>))}</TableRow>)) : (<TableRow><TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">No se encontraron importaciones.</TableCell></TableRow>)}</TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
