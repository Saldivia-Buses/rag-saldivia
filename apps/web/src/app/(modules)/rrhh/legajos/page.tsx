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
  MoreHorizontalIcon,
  MailIcon,
  PhoneIcon,
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

type Empleado = {
  id: string;
  legajo: string;
  nombre: string;
  apellido: string;
  dni: string;
  email: string;
  telefono: string;
  puesto: string;
  sector: string;
  fechaIngreso: string;
  estado: "activo" | "licencia" | "baja";
};

const MOCK_EMPLEADOS: Empleado[] = [
  { id: "1", legajo: "L-001", nombre: "Carlos", apellido: "Gómez", dni: "30.456.789", email: "cgomez@saldivia.com", telefono: "11-4567-8901", puesto: "Soldador", sector: "Producción", fechaIngreso: "2019-03-15", estado: "activo" },
  { id: "2", legajo: "L-002", nombre: "María", apellido: "López", dni: "32.123.456", email: "mlopez@saldivia.com", telefono: "11-5678-9012", puesto: "Inspectora de Calidad", sector: "Calidad", fechaIngreso: "2020-07-01", estado: "activo" },
  { id: "3", legajo: "L-003", nombre: "Juan", apellido: "Martínez", dni: "28.789.012", email: "jmartinez@saldivia.com", telefono: "11-6789-0123", puesto: "Jefe de Planta", sector: "Producción", fechaIngreso: "2015-01-10", estado: "activo" },
  { id: "4", legajo: "L-004", nombre: "Ana", apellido: "Rodríguez", dni: "35.234.567", email: "arodriguez@saldivia.com", telefono: "11-7890-1234", puesto: "Administrativa", sector: "Administración", fechaIngreso: "2021-09-20", estado: "licencia" },
  { id: "5", legajo: "L-005", nombre: "Roberto", apellido: "Fernández", dni: "27.567.890", email: "rfernandez@saldivia.com", telefono: "11-8901-2345", puesto: "Electricista", sector: "Mantenimiento", fechaIngreso: "2018-04-05", estado: "activo" },
  { id: "6", legajo: "L-006", nombre: "Lucía", apellido: "García", dni: "33.890.123", email: "lgarcia@saldivia.com", telefono: "11-9012-3456", puesto: "Ingeniera de Producto", sector: "Ingeniería", fechaIngreso: "2022-02-14", estado: "activo" },
  { id: "7", legajo: "L-007", nombre: "Diego", apellido: "Pérez", dni: "29.345.678", email: "dperez@saldivia.com", telefono: "11-0123-4567", puesto: "Pintor", sector: "Producción", fechaIngreso: "2017-11-28", estado: "baja" },
  { id: "8", legajo: "L-008", nombre: "Valentina", apellido: "Torres", dni: "36.678.901", email: "vtorres@saldivia.com", telefono: "11-1234-5678", puesto: "Compradora", sector: "Compras", fechaIngreso: "2023-05-08", estado: "activo" },
];

const estadoBadge: Record<Empleado["estado"], { label: string; color: "green" | "amber" | "red" }> = {
  activo: { label: "Activo", color: "green" },
  licencia: { label: "En licencia", color: "amber" },
  baja: { label: "Baja", color: "red" },
};

const columns: ColumnDef<Empleado>[] = [
  {
    accessorKey: "nombre",
    header: ({ column }) => (
      <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>
        Empleado
        <ArrowUpDownIcon className="ml-1 size-3.5" />
      </Button>
    ),
    cell: ({ row }) => {
      const e = row.original;
      const initials = e.nombre[0] + e.apellido[0];
      return (
        <div className="flex items-center gap-3">
          <Avatar className="size-8">
            <AvatarFallback className="text-xs">{initials}</AvatarFallback>
          </Avatar>
          <div>
            <p className="text-sm font-medium">{e.nombre} {e.apellido}</p>
            <p className="text-xs text-muted-foreground">{e.legajo} · DNI {e.dni}</p>
          </div>
        </div>
      );
    },
  },
  {
    accessorKey: "puesto",
    header: "Puesto",
    cell: ({ row }) => (
      <div>
        <p className="text-sm">{row.original.puesto}</p>
        <p className="text-xs text-muted-foreground">{row.original.sector}</p>
      </div>
    ),
  },
  {
    accessorKey: "fechaIngreso",
    header: ({ column }) => (
      <Button variant="ghost" size="sm" onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>
        Ingreso
        <ArrowUpDownIcon className="ml-1 size-3.5" />
      </Button>
    ),
    cell: ({ row }) => {
      const date = new Date(row.getValue("fechaIngreso") as string);
      return <span className="text-sm">{date.toLocaleDateString("es-AR")}</span>;
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
    id: "contacto",
    header: "Contacto",
    cell: () => (
      <div className="flex items-center gap-1">
        <Button variant="ghost" size="icon" className="size-7">
          <MailIcon className="size-3.5" />
        </Button>
        <Button variant="ghost" size="icon" className="size-7">
          <PhoneIcon className="size-3.5" />
        </Button>
      </div>
    ),
  },
  {
    id: "actions",
    cell: () => (
      <DropdownMenu>
        <DropdownMenuTrigger render={<Button variant="ghost" size="icon" className="size-7" />}>
          <MoreHorizontalIcon className="size-3.5" />
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem>Ver legajo</DropdownMenuItem>
          <DropdownMenuItem>Editar</DropdownMenuItem>
          <DropdownMenuItem>Documentación</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    ),
  },
];

export default function LegajosPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);

  const table = useReactTable({
    data: MOCK_EMPLEADOS,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    state: { sorting, columnFilters },
  });

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Legajos de Personal</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              {MOCK_EMPLEADOS.length} empleados registrados
            </p>
          </div>
          <Button size="sm">
            <PlusIcon className="size-4 mr-1.5" />
            Nuevo empleado
          </Button>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Activos", value: MOCK_EMPLEADOS.filter((e) => e.estado === "activo").length, color: "text-green-500" },
            { label: "En licencia", value: MOCK_EMPLEADOS.filter((e) => e.estado === "licencia").length, color: "text-amber-500" },
            { label: "Baja", value: MOCK_EMPLEADOS.filter((e) => e.estado === "baja").length, color: "text-red-500" },
          ].map((stat) => (
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
            placeholder="Buscar por nombre..."
            className="pl-9 bg-card"
            value={(table.getColumn("nombre")?.getFilterValue() as string) ?? ""}
            onChange={(e) => table.getColumn("nombre")?.setFilterValue(e.target.value)}
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
                    No se encontraron empleados.
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
