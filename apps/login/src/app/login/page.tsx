import { LoginForm } from "@/components/login-form";

export default function LoginPage() {
  return (
    <main className="flex min-h-screen items-center justify-center">
      <div className="flex flex-col items-center gap-6">
        {/* Logo — branded, isolated, no system hints */}
        <div className="flex flex-col items-center gap-2">
          <div className="text-2xl font-bold tracking-tight text-primary">
            SDA
          </div>
          <p className="text-sm text-muted-foreground">Framework</p>
        </div>

        <h1 className="text-xl font-semibold">Iniciar sesion</h1>

        <LoginForm />
      </div>
    </main>
  );
}
