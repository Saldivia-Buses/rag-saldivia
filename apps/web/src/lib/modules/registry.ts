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
  BusFront,
  type LucideIcon,
} from "lucide-react";

export interface SubRoute {
  path: string;
  label: string;
  icon?: LucideIcon;
}

export interface ModuleManifest {
  id: string;
  nav: {
    label: string;
    icon: LucideIcon;
    path: string;
    position: number;
  };
  routes: string[];
  /** Sub-navigation items shown inside the module. Tenant can disable individual sub-routes via config. */
  subnav?: SubRoute[];
}

export const MODULE_REGISTRY: Record<string, ModuleManifest> = {
  // ── Manufactura ──────────────────────────────────────────
  manufactura: {
    id: "manufactura",
    nav: { label: "Manufactura", icon: BusFront, path: "/manufactura", position: 20 },
    routes: [
      "/manufactura",
      "/manufactura/unidades",
      "/manufactura/controles",
      "/manufactura/certificaciones",
    ],
    subnav: [
      { path: "/manufactura/unidades", label: "Unidades" },
      { path: "/manufactura/controles", label: "Controles" },
      { path: "/manufactura/certificaciones", label: "Certificaciones" },
    ],
  },

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
    subnav: [
      { path: "/produccion/ordenes", label: "Órdenes" },
      { path: "/produccion/seguimiento", label: "Seguimiento" },
      { path: "/produccion/pcp", label: "PCP" },
      { path: "/produccion/preentrega", label: "Preentrega" },
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
    subnav: [
      { path: "/calidad/inspecciones", label: "Inspecciones" },
      { path: "/calidad/no-conformidades", label: "No Conformidades" },
      { path: "/calidad/trazabilidad", label: "Trazabilidad" },
      { path: "/calidad/postventa", label: "Postventa" },
      { path: "/calidad/sgc", label: "SGC" },
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
    subnav: [
      { path: "/ingenieria/producto", label: "Producto" },
      { path: "/ingenieria/desarrollo", label: "Desarrollo" },
      { path: "/ingenieria/definicion", label: "Definición" },
      { path: "/ingenieria/legal", label: "Legal y Técnica" },
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
    subnav: [
      { path: "/mantenimiento/preventivo", label: "Preventivo" },
      { path: "/mantenimiento/correctivo", label: "Correctivo" },
      { path: "/mantenimiento/equipos", label: "Equipos" },
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
    subnav: [
      { path: "/compras/ordenes", label: "Órdenes de Compra" },
      { path: "/compras/proveedores", label: "Proveedores" },
      { path: "/compras/abastecimiento", label: "Abastecimiento" },
      { path: "/compras/comex", label: "Comex" },
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
      "/administracion/sugerencias",
    ],
    subnav: [
      { path: "/administracion/facturacion", label: "Facturación" },
      { path: "/administracion/pagos", label: "Pagos" },
      { path: "/administracion/contable", label: "Contabilidad" },
      { path: "/administracion/sugerencias", label: "Sugerencias" },
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
    subnav: [
      { path: "/rrhh/legajos", label: "Legajos" },
      { path: "/rrhh/licencias", label: "Licencias" },
      { path: "/rrhh/capacitaciones", label: "Capacitaciones" },
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
    subnav: [
      { path: "/seguridad/inspecciones", label: "Inspecciones" },
      { path: "/seguridad/medicina", label: "Medicina Laboral" },
      { path: "/seguridad/incidentes", label: "Incidentes" },
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
    label: "Colecciones",
    icon: Database,
    path: "/collections",
    position: 10,
  },
];
