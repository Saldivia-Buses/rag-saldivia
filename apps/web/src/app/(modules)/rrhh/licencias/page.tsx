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
} from "lucide-react";
import { useState } from "react";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
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

type Licencia = {
  id: string;
  empleado: { nombre: string; apellido: string; legajo: string };
  tipo: "vacaciones" | "enfermedad" | "examen" | "maternidad" | "otro";
  desde: string;
  hasta: string;
  dias: number;
  estado: "pendiente" | "aprobada" | "rechazada" | "en_curso";
  motivo: string;
};

const MOCK_LICENCIAS: Licencia[] = [
  { id: "1", empleado: { nombre: "Carlos", apellido: "Gómez", legajo: "L-001" }, tipo: "vacaciones", desde: "2026-04-14", hasta: "2026-04-25", dias: 10, estado: "aprobada", motivo: "Vacaciones anuales" },
  { id: "2", empleado: { nombre: "Ana", apellido: "Rodríguez", legajo: "L-004" }, tipo: "maternidad", desde: "2026-03-01", hasta: "2026-05-30", dias: 90, estado: "en_curso", motivo: "Licencia por maternidad" },
  { id: "3", empleado: { nombre: "Roberto", apellido: "Fernández", legajo: "L-005" }, tipo: "enfermedad", desde: "2026-04-07", hasta: "2026-04-09", dias: 3, estado: "aprobada", motivo: "Certificado médico" },
  { id: "4", empleado: { nombre: "Lucía", apellido: "García", legajo: "L-006" }, tipo: "examen", desde: "2026-04-20", hasta: "2026-04-21", dias: 2, estado: "pendiente", motivo: "Examen final UTN" },
  { id: "5", empleado: { nombre: "Valentina", apellido: "Torres", legajo: "L-008" }, tipo: "vacaciones", desde: "2026-05-01", hasta: "2026-05-14", dias: 10, estado: "pendiente", motivo: "Vacaciones" },
  { id: "6", empleado: { nombre: "María", apellido: "López", legajo: "L-002" }, tipo: "enfermedad", desde: "2026-03-28", hasta: "2026-03-29", dias: 2, estado: "aprobada", motivo: "Consulta médica" },
  { id: "7", empleado: { nombre: "Juan", apellido: "Martínez", legajo: "L-003" }, tipo: "vacaciones", desde: "2026-04-01", hasta: "2026-04-04", dias: 4, estado: "rechazada", motivo: "Vacaciones — conflicto con entrega" },
];

const estadoBadge: Record<Licencia["estado"], { label: string; color: "amber" | "green" | "red" | "blue" }> = {
  pendiente: { label: "Pendiente", color: "amber" },
  aprobada: { label: "Aprobada", color: "green" },
  rechazada: { label: "Rechazada", color: "red" },
  en_curso: { label: "En curso", color: "blue" },
};

const tipoBadge: Record<Licencia["tipo"], string> = {
  vacaciones: "Vacaciones",
  enfermedad: "Enfermedad",
  examen: "Examen",
  maternidad: "Maternidad",
  otro: "Otro",
};

const columns: ColumnDef<Licencia>[] = [
  {
    accessorKey: "empleado",
    header: ({ column }) => (
      <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>
        Empleado
        <ArrowUpDownIcon className="ml-1 size-3.5" />
      </Button>
    ),
    sortingFn: (a, b) => a.original.empleado.apellido.localeCompare(b.original.empleado.apellido),
    filterFn: (row, _id, filterValue: string) => {
      const e = row.original.empleado;
      const full = `${e.nombre} ${e.apellido} ${e.legajo}`.toLowerCase();
      return full.includes(filterValue.toLowerCase());
    },
    cell: ({ row }) => {
      const e = row.original.empleado;
      const initials = e.nombre[0] + e.apellido[0];
      return (
        <div className="flex items-center gap-3">
          <Avatar className="size-8">
            <AvatarFallback className="text-xs">{initials}</AvatarFallback>
          </Avatar>
          <div>
            <p className="text-sm font-medium">{e.nombre} {e.apellido}</p>
            <p className="text-xs text-muted-foreground">{e.legajo}</p>
          </div>
        </div>
      );
    },
  },
  {
    accessorKey: "tipo",
    header: "Tipo",
    cell: ({ row }) => (
      <Badge variant="secondary">{tipoBadge[row.original.tipo]}</Badge>
    ),
  },
  {
    accessorKey: "desde",
    header: ({ column }) => (
      <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>
        Período
        <ArrowUpDownIcon className="ml-1 size-3.5" />
      </Button>
    ),
    cell: ({ row }) => {
      const desde = new Date(row.original.desde).toLocaleDateString("es-AR", { day: "2-digit", month: "short" });
      const hasta = new Date(row.original.hasta).toLocaleDateString("es-AR", { day: "2-digit", month: "short" });
      return (
        <div className="flex items-center gap-2 text-sm">
          <CalendarIcon className="size-3.5 text-muted-foreground" />
          <span>{desde} — {hasta}</span>
          <span className="text-xs text-muted-foreground">({row.original.dias}d)</span>
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
  {
    accessorKey: "motivo",
    header: "Motivo",
    cell: ({ row }) => (
      <span className="text-sm text-muted-foreground line-clamp-1 max-w-[200px]">
        {row.original.motivo}
      </span>
    ),
  },
];

export default function LicenciasPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);

  const table = useReactTable({
    data: MOCK_LICENCIAS,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    state: { sorting, columnFilters },
  });

  const pendientes = MOCK_LICENCIAS.filter((l) => l.estado === "pendiente").length;
  const enCurso = MOCK_LICENCIAS.filter((l) => l.estado === "en_curso").length;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Licencias y Ausencias</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              {pendientes > 0 && <span className="text-amber-500 font-medium">{pendientes} pendiente{pendientes > 1 ? "s" : ""}</span>}
              {pendientes > 0 && enCurso > 0 && " · "}
              {enCurso > 0 && <span className="text-blue-500 font-medium">{enCurso} en curso</span>}
              {pendientes === 0 && enCurso === 0 && "Todo al día"}
            </p>
          </div>
          <Button size="sm">
            <PlusIcon className="size-4 mr-1.5" />
            Nueva licencia
          </Button>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-4 gap-3 mb-6">
          {(
            [
              { label: "Pendientes", value: pendientes, color: "text-amber-500" },
              { label: "En curso", value: enCurso, color: "text-blue-500" },
              { label: "Aprobadas", value: MOCK_LICENCIAS.filter((l) => l.estado === "aprobada").length, color: "text-green-500" },
              { label: "Rechazadas", value: MOCK_LICENCIAS.filter((l) => l.estado === "rechazada").length, color: "text-red-500" },
            ] as const
          ).map((stat) => (
            <div key={stat.label} className="rounded-xl border border-border/40 bg-card p-4">
              <p className="text-xs text-muted-foreground">{stat.label}</p>
              <p className={`text-2xl font-semibold ${stat.color}`}>{stat.value}</p>
            </div>
          ))}
        </div>

        {/* Search */}
        <div className="relative mb-4">
          <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input
            placeholder="Buscar por empleado..."
            className="pl-9 bg-card"
            value={(table.getColumn("empleado")?.getFilterValue() as string) ?? ""}
            onChange={(e) => table.getColumn("empleado")?.setFilterValue(e.target.value)}
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
                    No se encontraron licencias.
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
