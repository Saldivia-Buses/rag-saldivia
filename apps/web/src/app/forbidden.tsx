import Link from "next/link";

export default function Forbidden() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="flex flex-col items-center gap-4 text-center">
        <h1 className="text-2xl font-semibold">Acceso denegado</h1>
        <p className="text-muted-foreground">
          No tenes permisos para acceder a este recurso.
        </p>
        <Link
          href="/chat"
          className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          Volver al inicio
        </Link>
      </div>
    </div>
  );
}
