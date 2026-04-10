"use client";

import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { MODULE_REGISTRY } from "@/lib/modules/registry";

/**
 * Reusable index page for a module — shows its sub-areas as cards.
 * Used by every module's page.tsx.
 */
export function ModuleIndex({ moduleId }: { moduleId: string }) {
  const mod = MODULE_REGISTRY[moduleId];
  if (!mod) return null;

  const Icon = mod.nav.icon;
  const areas = mod.subnav ?? [];

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-4xl px-6 py-8 sm:px-8">
        {/* Header */}
        <div className="flex items-start gap-4 mb-8">
          <div className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-primary/10">
            <Icon className="size-5 text-primary" />
          </div>
          <div>
            <h1 className="text-xl font-semibold tracking-tight">
              {mod.nav.label}
            </h1>
            {areas.length > 0 && (
              <p className="text-sm text-muted-foreground mt-0.5">
                {areas.length} {areas.length === 1 ? "área" : "áreas"} de
                trabajo
              </p>
            )}
          </div>
        </div>

        {/* Sub-areas grid */}
        {areas.length > 0 ? (
          <div className="grid gap-3 sm:grid-cols-2">
            {areas.map((sub) => (
              <Link
                key={sub.path}
                href={sub.path}
                className="group flex items-center justify-between rounded-xl border border-border/40 bg-card p-5 transition-all hover:border-border hover:shadow-sm"
              >
                <div>
                  <h3 className="text-sm font-semibold group-hover:text-foreground transition-colors">
                    {sub.label}
                  </h3>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {mod.nav.label} &mdash; {sub.label}
                  </p>
                </div>
                <ArrowRight className="size-4 text-muted-foreground/30 transition-all group-hover:text-foreground group-hover:translate-x-0.5" />
              </Link>
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <div className="flex size-12 items-center justify-center rounded-2xl bg-muted mb-4">
              <Icon className="size-5 text-muted-foreground" />
            </div>
            <p className="text-sm font-medium">{mod.nav.label}</p>
            <p className="text-xs text-muted-foreground mt-1">
              Este módulo no tiene sub-áreas configuradas.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
