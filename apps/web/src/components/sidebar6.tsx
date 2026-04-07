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
  FileText,
  Database,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth/store";
import { useEnabledModules } from "@/lib/modules/hooks";
import { MODULE_REGISTRY, CORE_NAV_ITEMS } from "@/lib/modules/registry";

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
  SidebarGroupLabel,
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
import { DarkModeToggle } from "@/components/dark-mode-toggle";
import { HeaderSearch } from "@/components/search-command";

const routeLabels: Record<string, string> = {
  "/inicio": "Inicio",
  "/chat": "Chat",
  "/collections": "Colecciones",
  "/documents": "Documentos",
  "/notifications": "Notificaciones",
  "/settings": "Mi cuenta",
  "/system-settings": "Configuración",
  "/fleet": "Flota",
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
            {i > 0 && <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/50" />}
            {isLast ? (
              <span className="font-medium text-foreground/70">{label}</span>
            ) : (
              <Link href={href} className="text-muted-foreground/50 hover:text-foreground/70 transition-colors">
                {label}
              </Link>
            )}
          </span>
        );
      })}
    </nav>
  );
}

// Base nav item - used by simple sidebars
type NavItem = {
  label: string;
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  href: string;
  isActive?: boolean;
  // Optional children for submenus (Sidebar3+)
  children?: NavItem[];
};

// Nav group with optional collapsible state
type NavGroup = {
  title: string;
  items: NavItem[];
  // Optional: default collapsed state (Sidebar2+)
  defaultOpen?: boolean;
};

// User data for footer (Sidebar6+)
type UserData = {
  name: string;
  email: string;
  avatar: string;
};

// Complete sidebar data structure
type SidebarData = {
  // Logo/branding (all sidebars)
  logo: {
    src: string;
    alt: string;
    title: string;
    description: string;
  };
  // Main navigation groups (all sidebars)
  navGroups: NavGroup[];
  // Footer navigation group (all sidebars)
  footerGroup: NavGroup;
  // User data for user footer (Sidebar6+)
  user?: UserData;
  // Workspaces for switcher (Sidebar7+)
  workspaces?: Array<{
    id: string;
    name: string;
    logo: string;
    plan: string;
  }>;
  // Currently active workspace (Sidebar7+)
  activeWorkspace?: string;
};

// Static sidebar parts
const sidebarLogo = {
  src: "/logo-placeholder.svg",
  alt: "SDA Framework",
  title: "SDA Framework",
  description: "Plataforma empresarial",
};

const footerGroup: NavGroup = {
  title: "Cuenta",
  items: [
    { label: "Configuracion", icon: Settings, href: "/system-settings" },
    { label: "Ayuda", icon: HelpCircle, href: "#" },
  ],
};

/**
 * Builds the dynamic nav groups from core items + enabled modules.
 */
function useSidebarData(): SidebarData {
  const user = useAuthStore((s) => s.user);
  const { data: enabledModules = [] } = useEnabledModules();

  // Core items — always visible
  const coreItems: NavItem[] = [
    { label: "Inicio", icon: LayoutDashboard, href: "/inicio" },
    { label: "Chat", icon: MessageSquare, href: "/chat" },
    { label: "Documentos", icon: FileText, href: "/documents" },
    { label: "Colecciones", icon: Database, href: "/collections" },
    { label: "Notificaciones", icon: Bell, href: "/notifications" },
  ];

  // Module items — only enabled modules for this tenant
  const moduleItems: NavItem[] = enabledModules
    .map((m) => MODULE_REGISTRY[m.id])
    .filter(Boolean)
    .filter((m) => !["chat", "rag", "notifications", "ingest"].includes(m.id)) // core already shown above
    .sort((a, b) => a.nav.position - b.nav.position)
    .map((m) => ({
      label: m.nav.label,
      icon: m.nav.icon,
      href: m.nav.path,
    }));

  const navGroups: NavGroup[] = [
    { title: "Principal", defaultOpen: true, items: coreItems },
  ];

  if (moduleItems.length > 0) {
    navGroups.push({ title: "Modulos", defaultOpen: true, items: moduleItems });
  }

  return {
    logo: sidebarLogo,
    navGroups,
    footerGroup,
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
          <div className="flex aspect-square size-8 items-center justify-center rounded-sm bg-primary">
            <img
              src={logo.src}
              alt={logo.alt}
              className="size-6 text-primary-foreground invert dark:invert-0"
            />
          </div>
          <div className="flex flex-col gap-0.5 leading-none group-data-[collapsible=icon]:hidden">
            <span className="font-medium">{logo.title}</span>
            <span className="text-xs text-muted-foreground">
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
  const isActive = pathname === item.href;
  const hasChildren = item.children && item.children.length > 0;

  if (!hasChildren) {
    return (
      <SidebarMenuItem>
        <SidebarMenuButton isActive={isActive} render={<Link href={item.href} prefetch />}>
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
            <Avatar className="size-8 rounded-lg">
              <AvatarImage src={user.avatar} alt={user.name} />
              <AvatarFallback className="rounded-lg">
                {user.name
                  .split(" ")
                  .map((n) => n[0])
                  .join("")}
              </AvatarFallback>
            </Avatar>
            <div className="grid flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden">
              <span className="truncate font-medium">{user.name}</span>
              <span className="truncate text-xs text-muted-foreground">
                {user.email}
              </span>
            </div>
            <ChevronsUpDown className="ml-auto size-4 group-data-[collapsible=icon]:hidden" />
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
            side="bottom"
            align="end"
            sideOffset={4}
          >
            <div className="flex items-center gap-2 px-1.5 py-1 text-left text-sm">
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
          {data.navGroups.map((group) => (
            <SidebarGroup key={group.title}>
              <SidebarGroupLabel>{group.title}</SidebarGroupLabel>
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
          <SidebarGroupLabel>{data.footerGroup.title}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {data.footerGroup.items.map((item) => (
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
  const pathname = usePathname();
  const pageLabel = routeLabels[pathname] || pathname.replace("/", "");
  return (
    <SidebarProvider open={false} className={cn("!h-svh !min-h-0 overflow-hidden", className)}>
      <AppSidebar />
      <SidebarInset className="flex flex-col h-svh overflow-hidden !bg-sidebar">
        {/* Header — same gray as sidebar */}
        <header className="flex h-14 shrink-0 items-center gap-3 px-4">
          <NavBreadcrumb />
          <div className="flex-1" />
          <DarkModeToggle />
        </header>
        {/* Search — inline expanding, fixed at viewport center */}
        <HeaderSearch />
        {/* Workspace — darker bg with rounded top corners, inset shadow for depth in light mode */}
        <div className="flex flex-1 flex-col min-h-0 bg-background rounded-t-2xl overflow-hidden shadow-[inset_1px_1px_4px_0px_rgba(0,0,0,0.15)] dark:shadow-none">
          {children}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
};

export { Sidebar6 };
