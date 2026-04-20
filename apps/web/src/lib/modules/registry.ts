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
  BusFront,
  Landmark,
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
      "/produccion/centros",
    ],
    subnav: [
      { path: "/produccion/ordenes", label: "Órdenes" },
      { path: "/produccion/seguimiento", label: "Seguimiento" },
      { path: "/produccion/pcp", label: "PCP" },
      { path: "/produccion/preentrega", label: "Preentrega" },
      { path: "/produccion/centros", label: "Centros de producción" },
    ],
  },
  calidad: {
    id: "calidad",
    nav: { label: "Calidad", icon: ClipboardCheck, path: "/calidad", position: 31 },
    routes: [
      "/calidad",
      "/calidad/inspecciones",
      "/calidad/no-conformidades",
      "/calidad/auditorias",
      "/calidad/documentos-controlados",
      "/calidad/planes-accion",
      "/calidad/indicadores",
      "/calidad/scorecards",
      "/calidad/trazabilidad",
      "/calidad/postventa",
      "/calidad/sgc",
    ],
    subnav: [
      { path: "/calidad/inspecciones", label: "Inspecciones" },
      { path: "/calidad/no-conformidades", label: "No Conformidades" },
      { path: "/calidad/auditorias", label: "Auditorías" },
      { path: "/calidad/documentos-controlados", label: "Documentos controlados" },
      { path: "/calidad/planes-accion", label: "Planes de acción" },
      { path: "/calidad/indicadores", label: "Indicadores" },
      { path: "/calidad/scorecards", label: "Scorecards de proveedores" },
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
      "/ingenieria/producto/secciones",
      "/ingenieria/producto/productos",
      "/ingenieria/producto/atributos",
      "/ingenieria/producto/chasis-marcas",
      "/ingenieria/producto/chasis-modelos",
      "/ingenieria/desarrollo",
      "/ingenieria/definicion",
      "/ingenieria/legal",
    ],
    subnav: [
      { path: "/ingenieria/producto", label: "Producto" },
      { path: "/ingenieria/producto/secciones", label: "Secciones" },
      { path: "/ingenieria/producto/productos", label: "Catálogo de productos" },
      { path: "/ingenieria/producto/atributos", label: "Atributos" },
      { path: "/ingenieria/producto/chasis-marcas", label: "Marcas de chasis" },
      { path: "/ingenieria/producto/chasis-modelos", label: "Modelos de chasis" },
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
      "/mantenimiento/combustible",
      "/mantenimiento/taller/vehiculos",
      "/mantenimiento/taller/incidentes",
      "/mantenimiento/activos",
    ],
    subnav: [
      { path: "/mantenimiento/preventivo", label: "Preventivo" },
      { path: "/mantenimiento/correctivo", label: "Correctivo" },
      { path: "/mantenimiento/equipos", label: "Equipos" },
      { path: "/mantenimiento/combustible", label: "Combustible" },
      { path: "/mantenimiento/activos", label: "Activos" },
      { path: "/mantenimiento/taller/vehiculos", label: "Vehículos de clientes" },
      { path: "/mantenimiento/taller/incidentes", label: "Incidentes vehiculares" },
    ],
  },

  // ── Soporte ──────────────────────────────────────────────
  ventas: {
    id: "ventas",
    nav: { label: "Ventas", icon: FileText, path: "/ventas", position: 39 },
    routes: [
      "/ventas",
      "/ventas/cotizaciones",
      "/ventas/ordenes",
    ],
    subnav: [
      { path: "/ventas/cotizaciones", label: "Cotizaciones" },
      { path: "/ventas/ordenes", label: "Órdenes de venta" },
    ],
  },
  compras: {
    id: "compras",
    nav: { label: "Compras", icon: ShoppingCart, path: "/compras", position: 40 },
    routes: [
      "/compras",
      "/compras/ordenes",
      "/compras/proveedores",
      "/compras/calificaciones",
      "/compras/listas-precios",
      "/compras/abastecimiento",
      "/compras/comex",
    ],
    subnav: [
      { path: "/compras/ordenes", label: "Órdenes de Compra" },
      { path: "/compras/proveedores", label: "Proveedores" },
      { path: "/compras/calificaciones", label: "Calificaciones" },
      { path: "/compras/listas-precios", label: "Listas de precios" },
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
      "/administracion/facturacion/notas",
      "/administracion/pagos",
      "/administracion/contable",
      "/administracion/reclamos",
      "/administracion/comunicaciones",
      "/administracion/calendario",
      "/administracion/encuestas",
      "/administracion/almacen",
      "/administracion/almacen/herramientas",
      "/administracion/almacen/movimientos",
      "/administracion/almacen/costos",
      "/administracion/almacen/bodegas",
      "/administracion/sugerencias",
    ],
    subnav: [
      { path: "/administracion/facturacion", label: "Facturación" },
      { path: "/administracion/facturacion/notas", label: "Notas de comprobantes" },
      { path: "/administracion/pagos", label: "Pagos" },
      { path: "/administracion/contable", label: "Contabilidad" },
      { path: "/administracion/reclamos", label: "Reclamos de pagos" },
      { path: "/administracion/comunicaciones", label: "Comunicaciones" },
      { path: "/administracion/calendario", label: "Calendario" },
      { path: "/administracion/encuestas", label: "Encuestas" },
      { path: "/administracion/almacen/movimientos", label: "Movimientos de stock" },
      { path: "/administracion/almacen/costos", label: "Costos de artículos" },
      { path: "/administracion/almacen/bodegas", label: "Bodegas" },
      { path: "/administracion/almacen/herramientas", label: "Herramientas" },
      { path: "/administracion/sugerencias", label: "Sugerencias" },
    ],
  },
  tesoreria: {
    id: "tesoreria",
    nav: { label: "Tesorería", icon: Landmark, path: "/tesoreria", position: 42 },
    routes: [
      "/tesoreria",
      "/tesoreria/cuentas-bancarias",
      "/tesoreria/cajas",
      "/tesoreria/reconciliaciones",
      "/tesoreria/recuentos",
      "/tesoreria/importaciones",
      "/tesoreria/cartera-historica",
      "/tesoreria/recibos",
    ],
    subnav: [
      { path: "/tesoreria/cuentas-bancarias", label: "Cuentas bancarias" },
      { path: "/tesoreria/cajas", label: "Cajas" },
      { path: "/tesoreria/reconciliaciones", label: "Reconciliaciones" },
      { path: "/tesoreria/recuentos", label: "Recuentos de caja" },
      { path: "/tesoreria/importaciones", label: "Importaciones bancarias" },
      { path: "/tesoreria/cartera-historica", label: "Cheques históricos" },
      { path: "/tesoreria/recibos", label: "Recibos" },
    ],
  },
  rrhh: {
    id: "rrhh",
    nav: { label: "RRHH", icon: Users, path: "/rrhh", position: 43 },
    routes: [
      "/rrhh",
      "/rrhh/legajos",
      "/rrhh/licencias",
      "/rrhh/capacitaciones",
      "/rrhh/asistencia",
    ],
    subnav: [
      { path: "/rrhh/legajos", label: "Legajos" },
      { path: "/rrhh/licencias", label: "Licencias" },
      { path: "/rrhh/capacitaciones", label: "Capacitaciones" },
      { path: "/rrhh/asistencia", label: "Asistencia" },
    ],
  },
  seguridad: {
    id: "seguridad",
    nav: { label: "Higiene y Seguridad", icon: ShieldCheck, path: "/seguridad", position: 44 },
    routes: [
      "/seguridad",
      "/seguridad/inspecciones",
      "/seguridad/medicina",
      "/seguridad/incidentes",
      "/seguridad/agentes-riesgo",
      "/seguridad/exposiciones-riesgo",
    ],
    subnav: [
      { path: "/seguridad/inspecciones", label: "Inspecciones" },
      { path: "/seguridad/medicina", label: "Medicina Laboral" },
      { path: "/seguridad/incidentes", label: "Incidentes" },
      { path: "/seguridad/agentes-riesgo", label: "Agentes de riesgo" },
      { path: "/seguridad/exposiciones-riesgo", label: "Exposiciones de riesgo" },
    ],
  },

  // ── Inteligencia ─────────────────────────────────────────
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
