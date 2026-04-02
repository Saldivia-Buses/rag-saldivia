import { TruckIcon } from "lucide-react";

export default function FleetPage() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4">
      <div className="flex size-16 items-center justify-center rounded-full bg-muted">
        <TruckIcon className="size-8 text-muted-foreground" />
      </div>
      <div className="text-center">
        <h2 className="text-lg font-semibold">Flota</h2>
        <p className="text-sm text-muted-foreground mt-1">
          Módulo de gestión de flota. Próximamente.
        </p>
      </div>
    </div>
  );
}
