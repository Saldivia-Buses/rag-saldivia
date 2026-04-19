"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { CostCenter } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";

export default function CostCenterDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const centerQ = useQuery({
    queryKey: erpKeys.costCenter(id),
    queryFn: () => api.get<CostCenter>(`/v1/erp/accounting/cost-centers/${id}`),
    enabled: !!id,
  });

  const siblingsQ = useQuery({
    queryKey: [...erpKeys.all, "accounting", "cost-centers"] as const,
    queryFn: () =>
      api.get<{ cost_centers: CostCenter[] }>("/v1/erp/accounting/cost-centers"),
    select: (d) => d.cost_centers,
    enabled: !!id,
  });

  if (centerQ.error)
    return <ErrorState message="Error cargando centro de costo" onRetry={() => window.location.reload()} />;

  const center = centerQ.data;
  const siblings = siblingsQ.data ?? [];
  const parent = center?.id
    ? siblings.find((s) => s.id === (center as unknown as { parent_id?: string }).parent_id)
    : null;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-4xl px-6 py-8 sm:px-8">
        <Link
          href="/administracion/contable?tab=cost-centers"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a contable
        </Link>

        {centerQ.isLoading && <Skeleton className="h-48 w-full" />}

        {center && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">{center.name}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Código <span className="font-mono">{center.code}</span>
                  {parent ? ` · Padre ${parent.code} · ${parent.name}` : ""}
                </p>
              </div>
              <Badge variant={center.active ? "default" : "secondary"}>
                {center.active ? "Activo" : "Inactivo"}
              </Badge>
            </div>

            <div className="rounded-xl border border-border/40 bg-card px-4 py-3 text-sm">
              <p className="text-muted-foreground">
                Los asientos contables pueden imputarse a este centro de costo desde la solapa
                Diario de <Link href="/administracion/contable" className="underline">contable</Link>.
                El detalle de asientos por centro está pendiente de un filtro backend
                (<span className="font-mono">entries?cost_center_id=</span>) — seguimiento en
                la próxima sesión.
              </p>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
