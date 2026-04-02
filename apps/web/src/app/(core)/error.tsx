"use client";

import { Button } from "@/components/ui/button";
import { AlertTriangleIcon } from "lucide-react";

export default function CoreError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4 p-8">
      <div className="flex size-16 items-center justify-center rounded-full bg-destructive/10">
        <AlertTriangleIcon className="size-8 text-destructive" />
      </div>
      <div className="text-center">
        <h2 className="text-lg font-semibold">Algo salió mal</h2>
        <p className="text-sm text-muted-foreground mt-1 max-w-md">
          {error.message || "Ocurrió un error inesperado."}
        </p>
      </div>
      <Button onClick={reset} variant="outline">
        Reintentar
      </Button>
    </div>
  );
}
