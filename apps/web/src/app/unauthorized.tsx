import Link from "next/link";

export default function Unauthorized() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="flex flex-col items-center gap-4 text-center">
        <h1 className="text-2xl font-semibold">Sesion expirada</h1>
        <p className="text-muted-foreground">
          Tu sesion ha expirado. Por favor, inicia sesion nuevamente.
        </p>
        <Link
          href="/login"
          className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          Iniciar sesion
        </Link>
      </div>
    </div>
  );
}
