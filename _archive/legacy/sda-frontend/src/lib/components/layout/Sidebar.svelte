<script lang="ts">
    import { page } from '$app/stores';
    import { toggleMode } from 'mode-watcher';
    import {
        MessageSquare, BookOpen, Upload, Users,
        Building2, Shield, ClipboardList, Settings,
        LogOut, ChevronLeft, ChevronRight, LayoutDashboard,
        Sun, Moon
    } from 'lucide-svelte';

    interface Props {
        role: string;
        userName: string;
        userEmail: string;
    }

    let { role, userName, userEmail }: Props = $props();

    let collapsed = $state(false);
    let loggingOut = $state(false);

    let isAdmin    = $derived(role === 'admin');
    let isManager  = $derived(role === 'admin' || role === 'area_manager');

    type NavItem = {
        href: string;
        label: string;
        icon: any;
        adminOnly?: boolean;
        managerOnly?: boolean;
    };

    const mainNav: NavItem[] = [
        { href: '/chat',        label: 'Chat',        icon: MessageSquare },
        { href: '/collections', label: 'Colecciones', icon: BookOpen },
        { href: '/upload',      label: 'Documentos',  icon: Upload },
    ];

    const adminNav: NavItem[] = [
        { href: '/admin/users',       label: 'Usuarios',     icon: Users,           managerOnly: true },
        { href: '/admin/areas',       label: 'Áreas',        icon: Building2,       managerOnly: true },
        { href: '/admin/permissions', label: 'Permisos',     icon: Shield,          adminOnly: true },
        { href: '/admin/rag-config',  label: 'RAG Config',   icon: Settings,        adminOnly: true },
        { href: '/admin/system',      label: 'Sistema',      icon: LayoutDashboard, adminOnly: true },
        { href: '/audit',             label: 'Auditoría',    icon: ClipboardList,   adminOnly: true },
    ];

    function isActive(href: string) {
        return $page.url.pathname.startsWith(href);
    }

    async function handleLogout() {
        loggingOut = true;
        try {
            await fetch('/api/auth/session', { method: 'DELETE' });
        } finally {
            window.location.href = '/login';
        }
    }
</script>

<nav
    style="width: {collapsed ? 'var(--sidebar-collapsed)' : 'var(--sidebar-width)'}"
    class="
        flex-shrink-0 h-screen
        bg-[var(--sidebar-bg)] border-r border-[var(--border)]
        flex flex-col
        transition-[width] duration-200 ease-in-out
        overflow-hidden
    "
>
    <!-- Header -->
    {#if collapsed}
        <div class="flex flex-col items-center py-2 border-b border-[var(--border)] min-h-[var(--topbar-height)]">
            <button
                onclick={() => collapsed = !collapsed}
                class="w-8 h-8 flex items-center justify-center rounded-[var(--radius-md)]
                       text-[var(--text-faint)] hover:text-[var(--text)] hover:bg-[var(--bg-hover)] transition-colors"
                title="Expandir"
            >
                <ChevronRight size={16} />
            </button>
        </div>
    {:else}
        <div class="flex items-center gap-2.5 px-3 py-3.5 border-b border-[var(--border)] min-h-[var(--topbar-height)]">
            <div class="w-7 h-7 bg-[var(--accent)] rounded-lg flex items-center justify-center text-white font-bold text-xs flex-shrink-0">
                S
            </div>
            <div class="overflow-hidden flex-1">
                <div class="text-sm font-bold text-[var(--text)] whitespace-nowrap">SDA</div>
                <div class="text-[10px] text-[var(--text-faint)] whitespace-nowrap">Saldivia Buses</div>
            </div>
            <button
                onclick={() => collapsed = !collapsed}
                class="text-[var(--text-faint)] hover:text-[var(--text)] hover:bg-[var(--bg-hover)] transition-colors flex-shrink-0 p-2 rounded-[var(--radius-sm)]"
                title="Colapsar"
            >
                <ChevronLeft size={16} />
            </button>
        </div>
    {/if}

    <!-- Nav -->
    <div class="flex-1 overflow-y-auto py-2 px-2 flex flex-col gap-0.5">
        {#if !collapsed}
            <div class="text-[9px] font-bold text-[var(--text-faint)] uppercase tracking-wider px-2 pt-2 pb-1">
                Principal
            </div>
        {/if}

        {#each mainNav as item}
            <a
                href={item.href}
                title={collapsed ? item.label : undefined}
                class="
                    flex items-center gap-2.5 px-2.5 py-2 rounded-[var(--radius-md)]
                    text-sm transition-colors group relative
                    {isActive(item.href)
                        ? 'bg-[var(--bg-surface)] text-[var(--text)] border-l-2 border-[var(--accent)] pl-[9px]'
                        : 'text-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text)]'}
                "
            >
                <item.icon size={16} class="flex-shrink-0" />
                {#if !collapsed}
                    <span class="whitespace-nowrap">{item.label}</span>
                {:else}
                    <span class="
                        absolute left-full ml-2 px-2 py-1 rounded-[var(--radius-sm)]
                        bg-[var(--bg-surface)] border border-[var(--border)]
                        text-xs text-[var(--text)] whitespace-nowrap
                        opacity-0 group-hover:opacity-100 pointer-events-none z-50
                        transition-opacity
                    ">{item.label}</span>
                {/if}
            </a>
        {/each}

        {#if isManager}
            {#if !collapsed}
                <div class="text-[9px] font-bold text-[var(--text-faint)] uppercase tracking-wider px-2 pt-4 pb-1">
                    Administración
                </div>
            {:else}
                <div class="h-px bg-[var(--border)] mx-2 my-2"></div>
            {/if}

            {#each adminNav as item}
                {#if (item.adminOnly && isAdmin) || (item.managerOnly && isManager)}
                    <a
                        href={item.href}
                        title={collapsed ? item.label : undefined}
                        class="
                            flex items-center gap-2.5 px-2.5 py-2 rounded-[var(--radius-md)]
                            text-sm transition-colors group relative
                            {isActive(item.href)
                                ? 'bg-[var(--bg-surface)] text-[var(--text)] border-l-2 border-[var(--accent)] pl-[9px]'
                                : 'text-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text)]'}
                        "
                    >
                        <item.icon size={16} class="flex-shrink-0" />
                        {#if !collapsed}
                            <span class="whitespace-nowrap">{item.label}</span>
                        {:else}
                            <span class="
                                absolute left-full ml-2 px-2 py-1 rounded-[var(--radius-sm)]
                                bg-[var(--bg-surface)] border border-[var(--border)]
                                text-xs text-[var(--text)] whitespace-nowrap
                                opacity-0 group-hover:opacity-100 pointer-events-none z-50
                                transition-opacity
                            ">{item.label}</span>
                        {/if}
                    </a>
                {/if}
            {/each}
        {/if}

        {#if !collapsed}
            <div class="text-[9px] font-bold text-[var(--text-faint)] uppercase tracking-wider px-2 pt-4 pb-1">
                Cuenta
            </div>
        {:else}
            <div class="h-px bg-[var(--border)] mx-2 my-2"></div>
        {/if}

        <a href="/settings"
           title={collapsed ? 'Configuración' : undefined}
           class="
               flex items-center gap-2.5 px-2.5 py-2 rounded-[var(--radius-md)]
               text-sm transition-colors group relative
               {isActive('/settings')
                   ? 'bg-[var(--bg-surface)] text-[var(--text)]'
                   : 'text-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text)]'}
           "
        >
            <Settings size={16} class="flex-shrink-0" />
            {#if !collapsed}
                <span>Configuración</span>
            {:else}
                <span class="absolute left-full ml-2 px-2 py-1 rounded-[var(--radius-sm)] bg-[var(--bg-surface)] border border-[var(--border)] text-xs text-[var(--text)] whitespace-nowrap opacity-0 group-hover:opacity-100 pointer-events-none z-50 transition-opacity">Configuración</span>
            {/if}
        </a>
    </div>

    <!-- Footer: avatar + theme toggle + logout -->
    <div class="border-t border-[var(--border)] p-2">
        <!-- Theme toggle -->
        <button
            onclick={toggleMode}
            title="Cambiar tema"
            class="
                flex items-center gap-2.5 w-full px-2.5 py-2 rounded-[var(--radius-md)]
                text-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text)]
                transition-colors text-sm mb-1
            "
        >
            <Sun size={16} class="flex-shrink-0 dark:hidden" />
            <Moon size={16} class="flex-shrink-0 hidden dark:block" />
            {#if !collapsed}
                <span>Cambiar tema</span>
            {/if}
        </button>

        <!-- User -->
        <div class="flex items-center gap-2.5 px-2.5 py-2">
            <div class="w-6 h-6 bg-[var(--accent)] rounded-full flex items-center justify-center text-white text-xs font-bold flex-shrink-0">
                {userName.charAt(0).toUpperCase()}
            </div>
            {#if !collapsed}
                <div class="flex-1 overflow-hidden">
                    <div class="text-xs font-semibold text-[var(--text)] truncate">{userName}</div>
                    <div class="text-[10px] text-[var(--text-faint)] truncate">{userEmail}</div>
                </div>
                <button
                    onclick={handleLogout}
                    disabled={loggingOut}
                    title="Cerrar sesión"
                    class="text-[var(--text-faint)] hover:text-[var(--danger)] transition-colors disabled:opacity-50 flex-shrink-0"
                >
                    <LogOut size={14} />
                </button>
            {/if}
        </div>
    </div>
</nav>
