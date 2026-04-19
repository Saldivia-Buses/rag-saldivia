"use client";

import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { EntitySearchResult } from "@/lib/erp/types";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { SearchIcon } from "lucide-react";

export type EntityType = "customer" | "supplier" | "employee";

interface EntityPickerProps {
  type: EntityType;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelect: (entity: EntitySearchResult) => void;
  title?: string;
}

const defaultTitleFor: Record<EntityType, string> = {
  supplier: "Seleccionar proveedor",
  customer: "Seleccionar cliente",
  employee: "Seleccionar empleado",
};

export function EntityPicker({ type, open, onOpenChange, onSelect, title }: EntityPickerProps) {
  const [query, setQuery] = useState("");
  const [debounced, setDebounced] = useState("");

  useEffect(() => {
    const t = setTimeout(() => setDebounced(query.trim()), 250);
    return () => clearTimeout(t);
  }, [query]);

  function close() {
    setQuery("");
    setDebounced("");
    onOpenChange(false);
  }

  function pick(entity: EntitySearchResult) {
    onSelect(entity);
    close();
  }

  const { data, isLoading, error } = useQuery({
    queryKey: erpKeys.entities(type, debounced),
    queryFn: () =>
      api.get<{ entities: EntitySearchResult[]; total: number }>(
        `/v1/erp/entities?type=${type}&search=${encodeURIComponent(debounced)}&page_size=50`,
      ),
    enabled: open,
    select: (d) => d.entities,
  });

  return (
    <Dialog open={open} onOpenChange={(v) => !v && close()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{title ?? defaultTitleFor[type]}</DialogTitle>
        </DialogHeader>
        <div className="relative">
          <SearchIcon className="pointer-events-none absolute left-2.5 top-2.5 size-4 text-muted-foreground" />
          <Input
            autoFocus
            placeholder="Buscar por nombre o código…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="pl-8"
          />
        </div>
        <div className="max-h-[50vh] overflow-y-auto rounded-md border border-border/40">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[100px]">Código</TableHead>
                <TableHead>Nombre</TableHead>
                <TableHead className="w-[180px]">Contacto</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={3}>
                    <Skeleton className="h-20 w-full" />
                  </TableCell>
                </TableRow>
              )}
              {!isLoading && (!data || data.length === 0) && (
                <TableRow>
                  <TableCell colSpan={3} className="h-20 text-center text-sm text-muted-foreground">
                    {debounced ? "Sin resultados" : "Escribí para buscar…"}
                  </TableCell>
                </TableRow>
              )}
              {!isLoading &&
                data?.map((e) => (
                  <TableRow
                    key={e.id}
                    className="cursor-pointer hover:bg-muted/40"
                    onClick={() => pick(e)}
                  >
                    <TableCell className="font-mono text-sm">{e.code}</TableCell>
                    <TableCell className="text-sm">{e.name}</TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                      {e.email ?? e.phone ?? "—"}
                    </TableCell>
                  </TableRow>
                ))}
            </TableBody>
          </Table>
        </div>
        {error && (
          <p className="text-sm text-destructive">Error cargando entidades.</p>
        )}
      </DialogContent>
    </Dialog>
  );
}
