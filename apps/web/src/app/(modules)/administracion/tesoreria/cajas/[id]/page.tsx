"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate } from "@/lib/erp/format";
import type { CashRegister, CashCount } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function CashRegisterDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const registerQ = useQuery({
    queryKey: erpKeys.cashRegister(id),
    queryFn: () => api.get<CashRegister>(`/v1/erp/treasury/cash-registers/${id}`),
    enabled: !!id,
  });

  const countsQ = useQuery({
    queryKey: erpKeys.cashCounts({ cash_register_id: id, page_size: "200" }),
    queryFn: () =>
      api.get<{ cash_counts: CashCount[] }>(
        `/v1/erp/treasury/cash-counts?cash_register_id=${id}&page_size=200`,
      ),
    select: (d) => d.cash_counts,
    enabled: !!id,
  });

  if (registerQ.error)
    return <ErrorState message="Error cargando caja" onRetry={() => window.location.reload()} />;

  const register = registerQ.data;
  const counts = countsQ.data ?? [];
  const totalDiff = counts.reduce((s, c) => s + Number(c.difference ?? 0), 0);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <Link
          href="/administracion/tesoreria/cajas"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a cajas
        </Link>

        {registerQ.isLoading && <Skeleton className="h-48 w-full" />}

        {register && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">{register.name}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Creada el <span className="font-mono">{fmtDate(register.created_at)}</span>
                </p>
              </div>
              <Badge variant={register.active ? "default" : "secondary"}>
                {register.active ? "Activa" : "Inactiva"}
              </Badge>
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-3">
              <Metric label="Arqueos" value={String(counts.length)} />
              <Metric
                label="Diferencia acumulada"
                value={totalDiff.toFixed(2)}
              />
              <Metric label="Cuenta contable" value={register.account_id ? "Vinculada" : "—"} />
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">
              Arqueos ({counts.length})
            </h2>
            <div className="mb-6 overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[130px]">Fecha</TableHead>
                    <TableHead className="w-[140px] text-right">Esperado</TableHead>
                    <TableHead className="w-[140px] text-right">Contado</TableHead>
                    <TableHead className="w-[140px] text-right">Diferencia</TableHead>
                    <TableHead>Notas</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {counts.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} className="h-20 text-center text-sm text-muted-foreground">
                        Sin arqueos registrados.
                      </TableCell>
                    </TableRow>
                  )}
                  {counts.map((c) => (
                    <TableRow key={c.id}>
                      <TableCell className="font-mono text-xs text-muted-foreground">
                        {c.date ? fmtDate(c.date) : "—"}
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">
                        {c.expected != null ? Number(c.expected).toFixed(2) : "—"}
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">
                        {c.counted != null ? Number(c.counted).toFixed(2) : "—"}
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">
                        {c.difference != null ? Number(c.difference).toFixed(2) : "—"}
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground">{c.notes || "—"}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 font-mono text-sm">{value}</div>
    </div>
  );
}
