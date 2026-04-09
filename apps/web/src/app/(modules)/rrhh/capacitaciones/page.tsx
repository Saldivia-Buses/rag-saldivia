"use client";

import {
  type ColumnDef,
  type ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  type SortingState,
  useReactTable,
} from "@tanstack/react-table";
import {
  ArrowUpDownIcon,
  PlusIcon,
  SearchIcon,
  CalendarIcon,
  UsersIcon,
  BookOpenIcon,
} from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Progress } from "@/components/ui/progress";

type Capacitacion = {
  id: string;
  titulo: string;
  tipo: "obligatoria" | "opcional" | "induccion";
  instructor: string;
  fecha: string;
  duracion: string;
  asistentes: number;
  totalConvocados: number;
  estado: "programada" | "en_curso" | "completada" | "cancelada";
  sector: string;
};

const MOCK_CAPACITACIONES: Capacitacion[] = [
  { id: "1", titulo: "Seguridad e Higiene — Actualización Anual", tipo: "obligatoria", instructor: "Ing. Marcos Vidal", fecha: "2026-04-15", duracion: "4 hs", asistentes: 42, totalConvocados: 45, estado: "programada", sector: "Todos" },
  { id: "2", titulo: "Soldadura MIG/MAG — Nivel Avanzado", tipo: "opcional", instructor: "Carlos Gómez", fecha: "2026-04-10", duracion: "8 hs", asistentes: 8, totalConvocados: 10, estado: "en_curso", sector: "Producción" },
  { id: "3", titulo: "Inducción — Nuevos Ingresos Abril", tipo: "induccion", instructor: "RRHH", fecha: "2026-04-01", duracion: "2 hs", asistentes: 3, totalConvocados: 3, estado: "completada", sector: "Todos" },
  { id: "4", titulo: "ISO 9001:2025 — Auditorías Internas", tipo: "obligatoria", instructor: "Ext. Bureau Veritas", fecha: "2026-04-22", duracion: "16 hs", asistentes: 0, totalConvocados: 12, estado: "programada", sector: "Calidad" },
  { id: "5", titulo: "Manejo de Autoelevador", tipo: "obligatoria", instructor: "Ext. Safety First", fecha: "2026-03-28", duracion: "6 hs", asistentes: 15, totalConvocados: 15, estado: "completada", sector: "Producción" },
  { id: "6", titulo: "Primeros Auxilios y RCP", tipo: "obligatoria", instructor: "Dr. Paula Ríos", fecha: "2026-05-05", duracion: "4 hs", asistentes: 0, totalConvocados: 30, estado: "programada", sector: "Todos" },
  { id: "7", titulo: "Excel Avanzado — Reportes", tipo: "opcional", instructor: "Ext. Educación IT", fecha: "2026-03-15", duracion: "12 hs", asistentes: 6, totalConvocados: 8, estado: "cancelada", sector: "Administración" },
];

const estadoBadge: Record<Capacitacion["estado"], { label: string; color: "amber" | "blue" | "green" | "red" }> = {
  programada: { label: "Programada", color: "amber" },
  en_curso: { label: "En curso", color: "blue" },
  completada: { label: "Completada", color: "green" },
  cancelada: { label: "Cancelada", color: "red" },
};

const tipoBadge: Record<Capacitacion["tipo"], { label: string; color: "red" | "gray" | "indigo" }> = {
  obligatoria: { label: "Obligatoria", color: "red" },
  opcional: { label: "Opcional", color: "gray" },
  induccion: { label: "Inducción", color: "indigo" },
};

const columns: ColumnDef<Capacitacion>[] = [
  {
    accessorKey: "titulo",
    header: ({ column }) => (
      <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>
        Capacitación
        <ArrowUpDownIcon className="ml-1 size-3.5" />
      </Button>
    ),
    cell: ({ row }) => {
      const c = row.original;
      const tipo = tipoBadge[c.tipo];
      return (
        <div>
          <div className="flex items-center gap-2">
            <p className="text-sm font-medium">{c.titulo}</p>
            <Badge variant="secondary" color={tipo.color} className="text-[10px] px-1.5 py-0">{tipo.label}</Badge>
          </div>
          <p className="text-xs text-muted-foreground mt-0.5">{c.instructor} · {c.sector}</p>
        </div>
      );
    },
  },
  {
    accessorKey: "fecha",
    header: ({ column }) => (
      <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>
        Fecha
        <ArrowUpDownIcon className="ml-1 size-3.5" />
      </Button>
    ),
    cell: ({ row }) => {
      const c = row.original;
      return (
        <div className="flex items-center gap-2 text-sm">
          <CalendarIcon className="size-3.5 text-muted-foreground" />
          <span>{new Date(c.fecha).toLocaleDateString("es-AR", { day: "2-digit", month: "short" })}</span>
          <span className="text-xs text-muted-foreground">{c.duracion}</span>
        </div>
      );
    },
  },
  {
    accessorKey: "asistentes",
    header: "Asistencia",
    cell: ({ row }) => {
      const c = row.original;
      const pct = c.totalConvocados > 0 ? Math.round((c.asistentes / c.totalConvocados) * 100) : 0;
      return (
        <div className="w-32">
          <div className="flex items-center justify-between text-xs mb-1">
            <span className="text-muted-foreground">{c.asistentes}/{c.totalConvocados}</span>
            <span className="font-medium">{pct}%</span>
          </div>
          <Progress value={pct} className="h-1.5" />
        </div>
      );
    },
  },
  {
    accessorKey: "estado",
    header: "Estado",
    cell: ({ row }) => {
      const est = estadoBadge[row.original.estado];
      return <Badge variant="outline" color={est.color}>{est.label}</Badge>;
    },
  },
];

export default function CapacitacionesPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);

  const table = useReactTable({
    data: MOCK_CAPACITACIONES,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    state: { sorting, columnFilters },
  });

  const completadas = MOCK_CAPACITACIONES.filter((c) => c.estado === "completada").length;
  const programadas = MOCK_CAPACITACIONES.filter((c) => c.estado === "programada").length;
  const totalHoras = MOCK_CAPACITACIONES.filter((c) => c.estado === "completada").reduce((acc, c) => acc + parseInt(c.duracion), 0);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Capacitaciones</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              Plan de formación y entrenamientos
            </p>
          </div>
          <Button size="sm">
            <PlusIcon className="size-4 mr-1.5" />
            Nueva capacitación
          </Button>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Completadas", value: completadas, icon: BookOpenIcon, sub: `${totalHoras} hs dictadas` },
            { label: "Programadas", value: programadas, icon: CalendarIcon, sub: "Próximas" },
            { label: "Personas formadas", value: MOCK_CAPACITACIONES.filter((c) => c.estado === "completada").reduce((acc, c) => acc + c.asistentes, 0), icon: UsersIcon, sub: "Este mes" },
          ].map((stat) => (
            <div key={stat.label} className="rounded-xl border border-border/40 bg-card p-4">
              <div className="flex items-center gap-2 mb-2">
                <stat.icon className="size-4 text-muted-foreground" />
                <p className="text-xs text-muted-foreground">{stat.label}</p>
              </div>
              <p className="text-2xl font-semibold">{stat.value}</p>
              <p className="text-xs text-muted-foreground">{stat.sub}</p>
            </div>
          ))}
        </div>

        {/* Search */}
        <div className="relative mb-4">
          <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input
            placeholder="Buscar capacitación..."
            className="pl-9 bg-card"
            value={(table.getColumn("titulo")?.getFilterValue() as string) ?? ""}
            onChange={(e) => table.getColumn("titulo")?.setFilterValue(e.target.value)}
          />
        </div>

        {/* Table */}
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <TableHead key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(header.column.columnDef.header, header.getContext())}
                    </TableHead>
                  ))}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {table.getRowModel().rows.length ? (
                table.getRowModel().rows.map((row) => (
                  <TableRow key={row.id}>
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id}>
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">
                    No se encontraron capacitaciones.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
