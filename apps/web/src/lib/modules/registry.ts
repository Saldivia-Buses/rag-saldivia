/**
 * Module registry — maps module IDs to their frontend manifest.
 *
 * This is the static definition of what each module contributes to the UI.
 * The list of ENABLED modules for a tenant comes from the backend
 * (via useEnabledModules hook). This registry just says "if module X is
 * enabled, here's what it adds to the nav."
 */

import {
  Truck,
  HardHat,
  Users,
  GitBranch,
  Calendar,
  BarChart3,
  MessageSquare,
  FileText,
  Database,
  type LucideIcon,
} from "lucide-react";

export interface ModuleManifest {
  id: string;
  nav: {
    label: string;
    icon: LucideIcon;
    path: string;
    position: number;
  };
  routes: string[];
}

export const MODULE_REGISTRY: Record<string, ModuleManifest> = {
  fleet: {
    id: "fleet",
    nav: { label: "Flota", icon: Truck, path: "/fleet", position: 40 },
    routes: [
      "/fleet",
      "/fleet/vehicles",
      "/fleet/drivers",
      "/fleet/maintenance",
      "/fleet/routes",
    ],
  },
  construction: {
    id: "construction",
    nav: { label: "Obra", icon: HardHat, path: "/construction", position: 41 },
    routes: [
      "/construction",
      "/construction/projects",
      "/construction/safety",
      "/construction/blueprints",
    ],
  },
  crm: {
    id: "crm",
    nav: { label: "CRM", icon: Users, path: "/crm", position: 30 },
    routes: ["/crm", "/crm/contacts", "/crm/pipeline"],
  },
  workflows: {
    id: "workflows",
    nav: {
      label: "Workflows",
      icon: GitBranch,
      path: "/workflows",
      position: 31,
    },
    routes: ["/workflows"],
  },
  calendar: {
    id: "calendar",
    nav: {
      label: "Calendario",
      icon: Calendar,
      path: "/calendar",
      position: 32,
    },
    routes: ["/calendar"],
  },
  reports: {
    id: "reports",
    nav: {
      label: "Reportes",
      icon: BarChart3,
      path: "/reports",
      position: 33,
    },
    routes: ["/reports"],
  },
  feedback: {
    id: "feedback",
    nav: {
      label: "Calidad IA",
      icon: MessageSquare,
      path: "/feedback",
      position: 50,
    },
    routes: ["/feedback"],
  },
};

/**
 * Core nav items — always visible, not module-dependent.
 */
export const CORE_NAV_ITEMS = [
  {
    label: "Chat",
    icon: MessageSquare,
    path: "/chat",
    position: 0,
  },
  {
    label: "Documentos",
    icon: FileText,
    path: "/documents",
    position: 10,
  },
  {
    label: "Colecciones",
    icon: Database,
    path: "/collections",
    position: 20,
  },
];
