/**
 * Module registry — maps module IDs to their frontend manifest.
 *
 * Based on Saldivia Buses organigram Rev. 28 (2025).
 * Each module maps to a real area of the company.
 *
 * The list of ENABLED modules for a tenant comes from the backend
 * (via useEnabledModules hook). This registry says "if module X is
 * enabled, here's what it adds to the nav."
 */

import {
  Factory,
  ClipboardCheck,
  Ruler,
  Wrench,
  ShoppingCart,
  DollarSign,
  Users,
  ShieldCheck,
  MessageSquare,
  FileText,
  Database,
  Star,
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
  // ── Operaciones ──────────────────────────────────────────
  produccion: {
    id: "produccion",
    nav: { label: "Producción", icon: Factory, path: "/produccion", position: 30 },
    routes: [
      "/produccion",
      "/produccion/ordenes",
      "/produccion/seguimiento",
      "/produccion/pcp",
      "/produccion/preentrega",
    ],
  },
  calidad: {
    id: "calidad",
    nav: { label: "Calidad", icon: ClipboardCheck, path: "/calidad", position: 31 },
    routes: [
      "/calidad",
      "/calidad/inspecciones",
      "/calidad/no-conformidades",
      "/calidad/trazabilidad",
      "/calidad/postventa",
      "/calidad/sgc",
    ],
  },
  ingenieria: {
    id: "ingenieria",
    nav: { label: "Ingeniería", icon: Ruler, path: "/ingenieria", position: 32 },
    routes: [
      "/ingenieria",
      "/ingenieria/producto",
      "/ingenieria/desarrollo",
      "/ingenieria/definicion",
      "/ingenieria/legal",
    ],
  },
  mantenimiento: {
    id: "mantenimiento",
    nav: { label: "Mantenimiento", icon: Wrench, path: "/mantenimiento", position: 33 },
    routes: [
      "/mantenimiento",
      "/mantenimiento/preventivo",
      "/mantenimiento/correctivo",
      "/mantenimiento/equipos",
    ],
  },

  // ── Soporte ──────────────────────────────────────────────
  compras: {
    id: "compras",
    nav: { label: "Compras", icon: ShoppingCart, path: "/compras", position: 40 },
    routes: [
      "/compras",
      "/compras/ordenes",
      "/compras/proveedores",
      "/compras/abastecimiento",
      "/compras/comex",
    ],
  },
  administracion: {
    id: "administracion",
    nav: { label: "Administración", icon: DollarSign, path: "/administracion", position: 41 },
    routes: [
      "/administracion",
      "/administracion/facturacion",
      "/administracion/pagos",
      "/administracion/contable",
    ],
  },
  rrhh: {
    id: "rrhh",
    nav: { label: "RRHH", icon: Users, path: "/rrhh", position: 42 },
    routes: [
      "/rrhh",
      "/rrhh/legajos",
      "/rrhh/licencias",
      "/rrhh/capacitaciones",
    ],
  },
  seguridad: {
    id: "seguridad",
    nav: { label: "Higiene y Seguridad", icon: ShieldCheck, path: "/seguridad", position: 43 },
    routes: [
      "/seguridad",
      "/seguridad/inspecciones",
      "/seguridad/medicina",
      "/seguridad/incidentes",
    ],
  },

  // ── Inteligencia ─────────────────────────────────────────
  astro: {
    id: "astro",
    nav: { label: "Astro", icon: Star, path: "/astro", position: 90 },
    routes: ["/astro"],
  },
  feedback: {
    id: "feedback",
    nav: { label: "Calidad IA", icon: MessageSquare, path: "/feedback", position: 91 },
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
