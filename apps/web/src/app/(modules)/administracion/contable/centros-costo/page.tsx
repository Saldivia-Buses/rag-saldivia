"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";

import { api } from "@/lib/api/client";

interface CostCenter {
  id: string;
  code: string;
  name: string;
}

export default function CentrosCostoListPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["cost-centers"],
    queryFn: async () => {
      const res = await api.get<{ items: CostCenter[] }>(
        "/v1/erp/accounting/cost-centers?page_size=200",
      );
      return res.items ?? [];
    },
  });

  return (
    <div className="space-y-4 p-6">
      <header>
        <h1 className="text-2xl font-semibold">Centros de costo</h1>
        <p className="text-muted-foreground text-sm">
          Plan de centros de costo contables.
        </p>
      </header>

      {isLoading && <p className="text-muted-foreground">Cargando…</p>}
      {error && (
        <p className="text-destructive text-sm">
          No se pudieron cargar los centros de costo.
        </p>
      )}
      {data && data.length === 0 && (
        <p className="text-muted-foreground">Sin centros de costo registrados.</p>
      )}
      {data && data.length > 0 && (
        <table className="w-full text-sm tabular-nums">
          <thead className="text-muted-foreground border-b text-left">
            <tr>
              <th className="py-2 pr-4 font-medium">Código</th>
              <th className="py-2 pr-4 font-medium">Nombre</th>
            </tr>
          </thead>
          <tbody>
            {data.map((c) => (
              <tr key={c.id} className="hover:bg-muted/40 border-b">
                <td className="py-2 pr-4">
                  <Link
                    className="hover:underline"
                    href={`/administracion/contable/centros-costo/${c.id}`}
                  >
                    {c.code}
                  </Link>
                </td>
                <td className="py-2 pr-4">{c.name}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
