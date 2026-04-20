"use client";

import Link from "next/link";
import {
  Bell,
  ChevronRight,
  ChevronsUpDown,
  HelpCircle,
  LayoutDashboard,
  LogOut,
  MessageSquare,
  Settings,
  User,
  Database,
  Megaphone,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth/store";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarProvider,
} from "@/components/ui/sidebar";
import { cn } from "@/lib/utils";
import { usePathname } from "next/navigation";
import { HeaderSearch } from "@/components/search-command";

const routeLabels: Record<string, string> = {
  "/inicio": "Inicio",
  "/chat": "Chat",
  "/collections": "Colecciones",
  "/documents": "Documentos",
  "/notifications": "Notificaciones",
  "/settings": "Mi cuenta",
  "/system-settings": "Configuración",
  "/produccion": "Producción",
  "/produccion/ordenes": "Órdenes",
  "/produccion/seguimiento": "Seguimiento",
  "/produccion/pcp": "PCP",
  "/produccion/preentrega": "Preentrega",
  "/calidad": "Calidad",
  "/calidad/inspecciones": "Inspecciones",
  "/calidad/no-conformidades": "No Conformidades",
  "/calidad/trazabilidad": "Trazabilidad",
  "/calidad/postventa": "Postventa",
  "/calidad/sgc": "SGC",
  "/ingenieria": "Ingeniería",
  "/ingenieria/producto": "Producto",
  "/ingenieria/desarrollo": "Desarrollo",
  "/ingenieria/definicion": "Definición",
  "/ingenieria/legal": "Legal y Técnica",
  "/manufactura": "Manufactura",
  "/manufactura/unidades": "Unidades",
  "/manufactura/controles": "Controles de Producción",
  "/manufactura/certificaciones": "Certificaciones",
  "/mantenimiento": "Mantenimiento",
  "/mantenimiento/preventivo": "Preventivo",
  "/mantenimiento/correctivo": "Correctivo",
  "/mantenimiento/equipos": "Equipos",
  "/compras": "Compras",
  "/compras/ordenes": "Órdenes de Compra",
  "/compras/proveedores": "Proveedores",
  "/compras/abastecimiento": "Abastecimiento",
  "/compras/comex": "Comex",
  "/administracion": "Administración",
  "/administracion/facturacion": "Facturación",
  "/administracion/pagos": "Pagos",
  "/administracion/contable": "Contabilidad",
  "/rrhh": "RRHH",
  "/rrhh/legajos": "Legajos",
  "/rrhh/licencias": "Licencias",
  "/rrhh/capacitaciones": "Capacitaciones",
  "/seguridad": "Higiene y Seguridad",
  "/seguridad/inspecciones": "Inspecciones",
  "/seguridad/medicina": "Medicina Laboral",
  "/seguridad/incidentes": "Incidentes",
  "/feedback": "Calidad IA",
};

function NavBreadcrumb() {
  const pathname = usePathname();
  const segments = pathname.split("/").filter(Boolean);

  if (segments.length === 0) return null;

  return (
    <nav aria-label="breadcrumb" className="flex items-center gap-1.5 text-sm">
      {segments.map((seg, i) => {
        const href = "/" + segments.slice(0, i + 1).join("/");
        const label = routeLabels[href] || seg.charAt(0).toUpperCase() + seg.slice(1);
        const isLast = i === segments.length - 1;

        return (
          <span key={href} className="flex items-center gap-1.5">
            {i > 0 && <ChevronRight className="size-3 text-muted-foreground/40" />}
            {isLast ? (
              <span className="font-medium text-foreground">{label}</span>
            ) : (
              <Link href={href} className="text-muted-foreground hover:text-foreground transition-colors">
                {label}
              </Link>
            )}
          </span>
        );
      })}
    </nav>
  );
}

type NavItem = {
  label: string;
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  href: string;
  isActive?: boolean;
  children?: NavItem[];
};

type NavGroup = {
  title: string;
  items: NavItem[];
  defaultOpen?: boolean;
};

type UserData = {
  name: string;
  email: string;
  avatar: string;
};

type SidebarData = {
  logo: {
    src: string;
    alt: string;
    title: string;
    description: string;
  };
  navGroups: NavGroup[];
  footerItems: NavItem[];
  user?: UserData;
};

const sidebarLogo = {
  src: "/logo-placeholder.svg",
  alt: "SDA Framework",
  title: "SDA Framework",
  description: "Plataforma empresarial",
};

const footerItems: NavItem[] = [
  { label: "Configuracion", icon: Settings, href: "/system-settings" },
  { label: "Ayuda", icon: HelpCircle, href: "#" },
];

function useSidebarData(): SidebarData {
  const user = useAuthStore((s) => s.user);

  // Sidebar shows only core navigation. ERP modules are accessed from /inicio
  // (cards grid) so the sidebar stays compact regardless of tenant config.
  const coreItems: NavItem[] = [
    { label: "Inicio", icon: LayoutDashboard, href: "/inicio" },
    { label: "Chat", icon: MessageSquare, href: "/chat" },
    // /documents page does not exist yet — re-add the entry once apps/web/src/app/(core)/documents/page.tsx lands.
    { label: "Colecciones", icon: Database, href: "/collections" },
    { label: "Notificaciones", icon: Bell, href: "/notifications" },
    { label: "Sugerencias y bugs", icon: Megaphone, href: "/administracion/sugerencias" },
  ];

  const navGroups: NavGroup[] = [
    { title: "Principal", defaultOpen: true, items: coreItems },
  ];

  return {
    logo: sidebarLogo,
    navGroups,
    footerItems,
    user: user
      ? { name: user.name, email: user.email, avatar: "" }
      : { name: "Usuario", email: "", avatar: "" },
  };
}

const SidebarLogo = ({ logo }: { logo: SidebarData["logo"] }) => {
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton size="lg">
          <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary">
            <img
              src={logo.src}
              alt={logo.alt}
              className="size-5 text-primary-foreground invert dark:invert-0"
            />
          </div>
          <div className="flex flex-col gap-0.5 leading-none group-data-[collapsible=icon]:hidden">
            <span className="font-semibold text-sm">{logo.title}</span>
            <span className="text-[11px] text-muted-foreground">
              {logo.description}
            </span>
          </div>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
};

const NavMenuItem = ({ item }: { item: NavItem }) => {
  const pathname = usePathname();
  const Icon = item.icon;
  const isActive = pathname === item.href || pathname.startsWith(item.href + "/");
  const hasChildren = item.children && item.children.length > 0;

  if (!hasChildren) {
    return (
      <SidebarMenuItem>
        <SidebarMenuButton
          isActive={isActive}
          render={<Link href={item.href} prefetch />}
          className={cn(
            isActive && "bg-white/[0.06] text-foreground font-medium",
            !isActive && "text-muted-foreground hover:text-foreground hover:bg-white/[0.04]",
          )}
        >
          <Icon className="size-4" />
          <span>{item.label}</span>
        </SidebarMenuButton>
      </SidebarMenuItem>
    );
  }

  return (
    <Collapsible defaultOpen={item.isActive} className="group/collapsible" render={<SidebarMenuItem />}>
      <CollapsibleTrigger render={<SidebarMenuButton isActive={item.isActive} />}>
        <Icon className="size-4" />
        <span>{item.label}</span>
        <ChevronRight className="ml-auto size-4 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
      </CollapsibleTrigger>
      <CollapsibleContent>
        <SidebarMenuSub>
          {item.children!.map((child) => (
            <SidebarMenuSubItem key={child.label}>
              <SidebarMenuSubButton isActive={child.isActive} render={<Link href={child.href} prefetch />}>
                {child.label}
              </SidebarMenuSubButton>
            </SidebarMenuSubItem>
          ))}
        </SidebarMenuSub>
      </CollapsibleContent>
    </Collapsible>
  );
};

const NavUser = ({ user }: { user: UserData }) => {
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <SidebarMenuButton
                size="lg"
                className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
              />
            }
          >
            <Avatar className="size-7 rounded-lg">
              <AvatarImage src={user.avatar} alt={user.name} />
              <AvatarFallback className="rounded-lg text-xs">
                {user.name
                  .split(" ")
                  .map((n) => n[0])
                  .join("")}
              </AvatarFallback>
            </Avatar>
            <div className="grid flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden">
              <span className="truncate font-medium text-sm">{user.name}</span>
              <span className="truncate text-[11px] text-muted-foreground">
                {user.email}
              </span>
            </div>
            <ChevronsUpDown className="ml-auto size-3.5 text-muted-foreground group-data-[collapsible=icon]:hidden" />
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
            side="bottom"
            align="end"
            sideOffset={4}
          >
            <div className="flex items-center gap-2 px-1.5 py-1.5 text-left text-sm">
              <Avatar className="size-8 rounded-lg">
                <AvatarImage src={user.avatar} alt={user.name} />
                <AvatarFallback className="rounded-lg">
                  {user.name
                    .split(" ")
                    .map((n) => n[0])
                    .join("")}
                </AvatarFallback>
              </Avatar>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-medium">{user.name}</span>
                <span className="truncate text-xs text-muted-foreground">
                  {user.email}
                </span>
              </div>
            </div>
            <DropdownMenuSeparator />
            <Link href="/settings">
              <DropdownMenuItem>
                <User className="mr-2 size-4" />
                Mi cuenta
              </DropdownMenuItem>
            </Link>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => {
                useAuthStore.getState().logout().then(() => {
                  window.location.href = "/login";
                });
              }}
            >
              <LogOut className="mr-2 size-4" />
              Cerrar sesion
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
};

const AppSidebar = ({ ...props }: React.ComponentProps<typeof Sidebar>) => {
  const data = useSidebarData();

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <SidebarLogo logo={data.logo} />
      </SidebarHeader>
      <SidebarContent className="overflow-hidden">
        <ScrollArea className="min-h-0 flex-1">
          {data.navGroups.map((group, i) => (
            <SidebarGroup key={group.title} className={i > 0 ? "mt-2" : ""}>
              {/* No labels — spacing as visual separator */}
              <SidebarGroupContent>
                <SidebarMenu>
                  {group.items.map((item) => (
                    <NavMenuItem key={item.label} item={item} />
                  ))}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          ))}
        </ScrollArea>
      </SidebarContent>
      <SidebarFooter>
        <SidebarGroup className="py-0">
          <SidebarGroupContent>
            <SidebarMenu>
              {data.footerItems.map((item) => (
                <NavMenuItem key={item.label} item={item} />
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        {data.user && <NavUser user={data.user} />}
      </SidebarFooter>
    </Sidebar>
  );
};

interface Sidebar6Props {
  className?: string;
  children?: React.ReactNode;
}

const Sidebar6 = ({ className, children }: Sidebar6Props) => {
  return (
    <SidebarProvider open={false} className={cn("!h-svh !min-h-0 overflow-hidden", className)}>
      <AppSidebar />
      <SidebarInset className="flex flex-col h-svh overflow-hidden !bg-sidebar">
        {/* Header */}
        <header className="flex h-14 shrink-0 items-center gap-4 px-5">
          <NavBreadcrumb />
          <div className="flex-1" />
        </header>
        {/* Search — centrado fixed en viewport, animación de expand */}
        <HeaderSearch />
        {/* Workspace — content area with rounded top */}
        <div className="flex flex-1 flex-col min-h-0 bg-background rounded-t-2xl overflow-hidden shadow-[inset_1px_1px_4px_0px_rgba(0,0,0,0.12)] dark:shadow-none">
          {children}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
};

export { Sidebar6 };
