"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";

interface ScorecardDetail {
  id: string;
  tenant_id: string;
  supplier_id: string;
  period: string;
  total_receipts: number;
  accepted_qty: string | null;
  rejected_qty: string | null;
  total_demerits: number;
  quality_score: string | null;
  created_at: string;
  supplier_name: string;
}

function scoreVariant(score: number): "default" | "secondary" | "destructive" {
  if (score >= 80) return "default";
  if (score >= 60) return "secondary";
  return "destructive";
}

export default function ScorecardDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data: card, isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "quality", "supplier-scorecards", id] as const,
    queryFn: () => api.get<ScorecardDetail>(`/v1/erp/quality/supplier-scorecards/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando scorecard" onRetry={() => window.location.reload()} />;

  const score = card?.quality_score ? Number(card.quality_score) : 0;
  const accepted = card?.accepted_qty ? Number(card.accepted_qty) : 0;
  const rejected = card?.rejected_qty ? Number(card.rejected_qty) : 0;
  const totalProcessed = accepted + rejected;
  const rejectionRate = totalProcessed > 0 ? (rejected / totalProcessed) * 100 : 0;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-4xl px-6 py-8 sm:px-8">
        <Link
          href="/calidad/scorecards"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a scorecards
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {card && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">
                  <Link
                    href={`/compras/proveedores/${card.supplier_id}`}
                    className="hover:underline"
                  >
                    {card.supplier_name}
                  </Link>
                </h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Período <span className="font-mono">{card.period}</span>
                  {" · "}registrado {fmtDate(card.created_at)}
                </p>
              </div>
              <Badge variant={scoreVariant(score)}>
                Score {score.toFixed(1)}
              </Badge>
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Recepciones</div>
                <div className="mt-1 font-mono text-sm">{card.total_receipts}</div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Aceptado</div>
                <div className="mt-1 font-mono text-sm">{accepted.toFixed(2)}</div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Rechazado</div>
                <div className="mt-1 font-mono text-sm text-destructive">
                  {rejected.toFixed(2)}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Demeritos</div>
                <div className="mt-1 font-mono text-sm">{card.total_demerits}</div>
              </div>
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Tasa de rechazo</div>
                <div className="mt-1 font-mono text-sm">
                  {rejectionRate.toFixed(2)}%
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Score de calidad</div>
                <div className="mt-1 font-mono text-sm">
                  {score.toFixed(2)} / 100
                </div>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
