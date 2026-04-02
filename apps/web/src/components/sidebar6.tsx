"use client";

import {
  BadgeCheck,
  BarChart3,
  Briefcase,
  ChevronRight,
  ChevronsUpDown,
  ClipboardList,
  Clock3,
  FileText,
  Folder,
  Globe2,
  HelpCircle,
  LayoutDashboard,
  LogOut,
  Search,
  Settings,
  Sparkles,
  Star,
  User,
  Users,
} from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInput,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { cn } from "@/lib/utils";

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

// Shared sidebar data - works with all sidebar variations
const sidebarData: SidebarData = {
  logo: {
    src: "https://deifkwefumgah.cloudfront.net/shadcnblocks/block/logos/shadcnblocks-logo.svg",
    alt: "Shadcnblocks",
    title: "Shadcnblocks",
    description: "Build your app",
  },
  navGroups: [
    {
      title: "Overview",
      defaultOpen: true,
      items: [
        {
          label: "Dashboard",
          icon: LayoutDashboard,
          href: "#",
          isActive: true,
        },
        { label: "Tasks", icon: ClipboardList, href: "#" },
        { label: "Roadmap", icon: BarChart3, href: "#" },
      ],
    },
    {
      title: "Projects",
      defaultOpen: true,
      items: [
        {
          label: "Active Projects",
          icon: Briefcase,
          href: "#",
          children: [
            { label: "Project Alpha", icon: FileText, href: "#" },
            { label: "Project Beta", icon: FileText, href: "#" },
            { label: "Project Gamma", icon: FileText, href: "#" },
          ],
        },
        {
          label: "Archived",
          icon: Folder,
          href: "#",
          children: [
            { label: "2024 Archive", icon: FileText, href: "#" },
            { label: "2023 Archive", icon: FileText, href: "#" },
          ],
        },
      ],
    },
    {
      title: "Team",
      defaultOpen: false,
      items: [
        { label: "Members", icon: Users, href: "#" },
        { label: "Sprints", icon: Clock3, href: "#" },
        { label: "Approvals", icon: BadgeCheck, href: "#" },
        { label: "Reviews", icon: Star, href: "#" },
      ],
    },
    {
      title: "Workspace",
      defaultOpen: false,
      items: [
        { label: "Integrations", icon: Globe2, href: "#" },
        { label: "Automations", icon: Sparkles, href: "#" },
      ],
    },
  ],
  footerGroup: {
    title: "Support",
    items: [
      { label: "Help Center", icon: HelpCircle, href: "#" },
      { label: "Settings", icon: Settings, href: "#" },
    ],
  },
  user: {
    name: "John Doe",
    email: "john@example.com",
    avatar:
      "https://deifkwefumgah.cloudfront.net/shadcnblocks/block/avatar-1.webp",
  },
  workspaces: [
    {
      id: "1",
      name: "Shadcnblocks",
      logo: "https://deifkwefumgah.cloudfront.net/shadcnblocks/block/logos/shadcnblocks-logo.svg",
      plan: "Enterprise",
    },
    {
      id: "2",
      name: "Shadcn Templates",
      logo: "https://deifkwefumgah.cloudfront.net/shadcnblocks/block/logos/shadcnblocks-logo.svg",
      plan: "Startup",
    },
    {
      id: "3",
      name: "Shadcn Components",
      logo: "https://deifkwefumgah.cloudfront.net/shadcnblocks/block/logos/shadcnblocks-logo.svg",
      plan: "Free",
    },
  ],
  activeWorkspace: "1",
};

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
          <div className="flex flex-col gap-0.5 leading-none">
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

const SearchForm = () => {
  return (
    <form>
      <SidebarGroup className="py-0">
        <SidebarGroupContent className="relative">
          <Label htmlFor="search" className="sr-only">
            Search
          </Label>
          <SidebarInput id="search" placeholder="Search..." className="pl-8" />
          <Search className="pointer-events-none absolute top-1/2 left-2 size-4 -translate-y-1/2 opacity-50 select-none" />
        </SidebarGroupContent>
      </SidebarGroup>
    </form>
  );
};

const NavMenuItem = ({ item }: { item: NavItem }) => {
  const Icon = item.icon;
  const hasChildren = item.children && item.children.length > 0;

  if (!hasChildren) {
    return (
      <SidebarMenuItem>
        <SidebarMenuButton isActive={item.isActive} render={<a href={item.href} />}><Icon className="size-4" /><span>{item.label}</span></SidebarMenuButton>
      </SidebarMenuItem>
    );
  }

  return (
    <Collapsible defaultOpen={item.isActive} className="group/collapsible" render={<SidebarMenuItem />}><CollapsibleTrigger render={<SidebarMenuButton isActive={item.isActive} />}><Icon className="size-4" /><span>{item.label}</span><ChevronRight className="ml-auto size-4 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" /></CollapsibleTrigger><CollapsibleContent>
                <SidebarMenuSub>
                  {item.children!.map((child) => (
                    <SidebarMenuSubItem key={child.label}>
                      <SidebarMenuSubButton isActive={child.isActive} render={<a href={child.href} />}>{child.label}</SidebarMenuSubButton>
                    </SidebarMenuSubItem>
                  ))}
                </SidebarMenuSub>
              </CollapsibleContent></Collapsible>
  );
};

const NavUser = ({ user }: { user: UserData }) => {
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger render={<SidebarMenuButton size="lg" className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground" />}><Avatar className="size-8 rounded-lg">
                                  <AvatarImage src={user.avatar} alt={user.name} />
                                  <AvatarFallback className="rounded-lg">
                                    {user.name
                                      .split(" ")
                                      .map((n) => n[0])
                                      .join("")}
                                  </AvatarFallback>
                                </Avatar><div className="grid flex-1 text-left text-sm leading-tight">
                                  <span className="truncate font-medium">{user.name}</span>
                                  <span className="truncate text-xs text-muted-foreground">
                                    {user.email}
                                  </span>
                                </div><ChevronsUpDown className="ml-auto size-4" /></DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
            side="bottom"
            align="end"
            sideOffset={4}
          >
            <DropdownMenuLabel className="p-0 font-normal">
              <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
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
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem>
              <User className="mr-2 size-4" />
              Account
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem>
              <LogOut className="mr-2 size-4" />
              Log out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
};

const AppSidebar = ({ ...props }: React.ComponentProps<typeof Sidebar>) => {
  return (
    <Sidebar {...props}>
      <SidebarHeader>
        <SidebarLogo logo={sidebarData.logo} />
        <SearchForm />
      </SidebarHeader>
      <SidebarContent className="overflow-hidden">
        <ScrollArea className="min-h-0 flex-1">
          {sidebarData.navGroups.map((group) => (
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
        {sidebarData.user && <NavUser user={sidebarData.user} />}
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
};

interface Sidebar6Props {
  className?: string;
  children?: React.ReactNode;
}

const Sidebar6 = ({ className, children }: Sidebar6Props) => {
  return (
    <SidebarProvider className={cn("!h-svh !min-h-0 overflow-hidden", className)}>
      <AppSidebar />
      <SidebarInset className="flex flex-col h-svh overflow-hidden">
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator
            orientation="vertical"
            className="mr-2 data-[orientation=vertical]:h-4"
          />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink href="#">Overview</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator className="hidden md:block" />
              <BreadcrumbItem>
                <BreadcrumbPage>Dashboard</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col min-h-0">
          {children}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
};

export { Sidebar6 };
