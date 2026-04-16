"use client";

import Link from "next/link";
import { MODULE_REGISTRY, type ModuleManifest } from "@/lib/modules/registry";
import { useAuthStore } from "@/lib/auth/store";
import {
  ArrowRight,
  MessageSquare,
  Database,
  Sparkles,
} from "lucide-react";

const MODULE_CATEGORIES: { title: string; description: string; range: [number, number] }[] = [
  { title: "Operaciones", description: "Producción, calidad e ingeniería", range: [20, 39] },
  { title: "Soporte", description: "Compras, administración y recursos", range: [40, 49] },
  { title: "Inteligencia", description: "IA y análisis avanzado", range: [90, 99] },
];

function getGreeting(): string {
  const h = new Date().getHours();
  if (h < 12) return "Buenos días";
  if (h < 19) return "Buenas tardes";
  return "Buenas noches";
}

function QuickAction({
  href,
  icon: Icon,
  label,
  sublabel,
}: {
  href: string;
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  label: string;
  sublabel: string;
}) {
  return (
    <Link
      href={href}
      className="group flex items-center gap-3 rounded-xl border border-border/40 bg-card px-4 py-3 transition-all hover:border-border hover:shadow-sm"
    >
      <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10">
        <Icon className="size-4 text-primary" />
      </div>
      <div className="min-w-0">
        <p className="text-sm font-medium">{label}</p>
        <p className="text-xs text-muted-foreground">{sublabel}</p>
      </div>
    </Link>
  );
}

function ModuleCard({ module }: { module: ModuleManifest }) {
  const Icon = module.nav.icon;
  return (
    <Link
      href={module.nav.path}
      className="group relative flex flex-col rounded-xl border border-border/40 bg-card p-5 transition-all hover:border-border hover:shadow-sm"
    >
      <div className="flex items-start justify-between mb-3">
        <div className="flex size-10 items-center justify-center rounded-xl bg-primary/10">
          <Icon className="size-5 text-primary" />
        </div>
        <ArrowRight className="size-4 text-muted-foreground/30 transition-all group-hover:text-foreground group-hover:translate-x-0.5" />
      </div>
      <h3 className="text-sm font-semibold mb-0.5">{module.nav.label}</h3>
      {module.subnav && module.subnav.length > 0 && (
        <p className="text-xs text-muted-foreground leading-relaxed">
          {module.subnav.map((s) => s.label).join(" · ")}
        </p>
      )}
    </Link>
  );
}

function CategorySection({
  title,
  description,
  modules,
}: {
  title: string;
  description: string;
  modules: ModuleManifest[];
}) {
  if (modules.length === 0) return null;
  return (
    <section>
      <div className="mb-3">
        <h2 className="text-sm font-semibold text-foreground">{title}</h2>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {modules.map((m) => (
          <ModuleCard key={m.id} module={m} />
        ))}
      </div>
    </section>
  );
}

export default function InicioPage() {
  const user = useAuthStore((s) => s.user);

  // Show every module from the static registry — no backend dependency, no
  // hydration flicker. Per-tenant gating moves into module pages themselves
  // when a finer-grained mechanism is in place.
  const allModules = Object.values(MODULE_REGISTRY).sort(
    (a, b) => a.nav.position - b.nav.position,
  );

  const firstName = user?.name?.split(" ")[0] ?? "Usuario";

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        {/* Greeting */}
        <div className="mb-8">
          <h1 className="text-xl font-semibold tracking-tight">
            {getGreeting()}, {firstName}
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            Accedé a tus áreas de trabajo y herramientas.
          </p>
        </div>

        {/* Quick Actions */}
        <div className="grid gap-3 sm:grid-cols-3 mb-10">
          <QuickAction
            href="/chat"
            icon={MessageSquare}
            label="Chat"
            sublabel="Consultar al asistente"
          />
          <QuickAction
            href="/collections"
            icon={Database}
            label="Colecciones"
            sublabel="Bases de conocimiento"
          />
          <QuickAction
            href="/astro"
            icon={Sparkles}
            label="Astro"
            sublabel="Análisis inteligente"
          />
        </div>

        {/* Module Categories */}
        {allModules.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <div className="flex size-12 items-center justify-center rounded-2xl bg-muted mb-4">
              <Database className="size-5 text-muted-foreground" />
            </div>
            <p className="text-sm font-medium">Sin módulos habilitados</p>
            <p className="text-xs text-muted-foreground mt-1">
              Contactá al administrador para habilitar módulos en tu organización.
            </p>
          </div>
        ) : (
          <div className="space-y-8">
            {MODULE_CATEGORIES.map((cat) => {
              const catModules = allModules.filter(
                (m) => m.nav.position >= cat.range[0] && m.nav.position <= cat.range[1]
              );
              return (
                <CategorySection
                  key={cat.title}
                  title={cat.title}
                  description={cat.description}
                  modules={catModules}
                />
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
