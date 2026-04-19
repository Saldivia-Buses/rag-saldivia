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

interface ChassisModelDetail {
  id: string;
  tenant_id: string;
  brand_id: string;
  model_code: string;
  description: string;
  traction: string;
  engine_location: string;
  active: boolean;
  created_at: string;
  updated_at: string;
  brand_name: string;
}

export default function ChassisModelDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data: model, isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "manufacturing", "chassis-models", id] as const,
    queryFn: () => api.get<ChassisModelDetail>(`/v1/erp/manufacturing/chassis-models/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando modelo de chasis" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-3xl px-6 py-8 sm:px-8">
        <Link
          href="/ingenieria/producto/chasis-modelos"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a modelos de chasis
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {model && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">{model.description}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  {model.brand_name} · <span className="font-mono">{model.model_code}</span>
                </p>
              </div>
              <Badge variant={model.active ? "default" : "secondary"}>
                {model.active ? "Activo" : "Inactivo"}
              </Badge>
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Tracción</div>
                <div className="mt-1 font-mono text-sm">{model.traction || "—"}</div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Ubicación motor</div>
                <div className="mt-1 font-mono text-sm">{model.engine_location || "—"}</div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Creado</div>
                <div className="mt-1 font-mono text-sm">{fmtDate(model.created_at)}</div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Actualizado</div>
                <div className="mt-1 font-mono text-sm">{fmtDate(model.updated_at)}</div>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
