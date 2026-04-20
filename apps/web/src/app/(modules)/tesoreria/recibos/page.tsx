"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";

import { api } from "@/lib/api/client";

interface Receipt {
  id: string;
  number: string;
  entity_name: string | null;
  amount: number | string;
  issued_at: string;
}

export default function RecibosListPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["receipts"],
    queryFn: async () => {
      const res = await api.get<{ items: Receipt[] }>(
        "/v1/erp/treasury/receipts?page_size=100",
      );
      return res.items ?? [];
    },
  });

  return (
    <div className="space-y-4 p-6">
      <header>
        <h1 className="text-2xl font-semibold">Recibos</h1>
        <p className="text-muted-foreground text-sm">
          Recibos emitidos por tesorería.
        </p>
      </header>

      {isLoading && <p className="text-muted-foreground">Cargando…</p>}
      {error && (
        <p className="text-destructive text-sm">
          No se pudieron cargar los recibos.
        </p>
      )}
      {data && data.length === 0 && (
        <p className="text-muted-foreground">Sin recibos registrados.</p>
      )}
      {data && data.length > 0 && (
        <table className="w-full text-sm tabular-nums">
          <thead className="text-muted-foreground border-b text-left">
            <tr>
              <th className="py-2 pr-4 font-medium">Número</th>
              <th className="py-2 pr-4 font-medium">Entidad</th>
              <th className="py-2 pr-4 text-right font-medium">Monto</th>
              <th className="py-2 pr-4 font-medium">Fecha</th>
            </tr>
          </thead>
          <tbody>
            {data.map((r) => (
              <tr key={r.id} className="hover:bg-muted/40 border-b">
                <td className="py-2 pr-4">
                  <Link className="hover:underline" href={`/tesoreria/recibos/${r.id}`}>
                    {r.number}
                  </Link>
                </td>
                <td className="py-2 pr-4">{r.entity_name ?? "—"}</td>
                <td className="py-2 pr-4 text-right">
                  {typeof r.amount === "number"
                    ? r.amount.toLocaleString("es-AR", { minimumFractionDigits: 2 })
                    : r.amount}
                </td>
                <td className="py-2 pr-4">{r.issued_at?.slice(0, 10)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
