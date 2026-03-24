<script lang="ts">
    import { MessageSquare, BookOpen, Upload, Users, ClipboardList, Settings, LogOut } from 'lucide-svelte';
    import SidebarItem from './SidebarItem.svelte';

    interface Props {
        role: string;
        areaId: number;
    }

    let { role, areaId }: Props = $props();

    let isAdmin = $derived(role === 'admin');
    let isManager = $derived(role === 'admin' || role === 'area_manager');
    let loggingOut = $state(false);

    async function handleLogout() {
        loggingOut = true;
        try {
            await fetch('/api/auth/session', { method: 'DELETE' });
            window.location.href = '/login';
        } catch {
            // Even on network error, clear local state and redirect
            window.location.href = '/login';
        }
    }
</script>

<nav class="w-10 min-h-screen bg-[#0a0f1e] border-r border-[#1e293b]
            flex flex-col items-center py-2 gap-1">

    <!-- Logo badge -->
    <div class="bg-[#6366f1] text-white font-bold text-[9px] w-[26px] h-5
                flex items-center justify-center rounded mb-1.5">
        SDA
    </div>

    <SidebarItem href="/chat" label="Chat" icon={MessageSquare} />
    <SidebarItem href="/collections" label="Colecciones" icon={BookOpen} />

    {#if isManager}
        <SidebarItem href="/admin/users" label={isAdmin ? 'Admin global' : 'Admin área'} icon={Users} />
    {/if}

    {#if isAdmin}
        <SidebarItem href="/audit" label="Auditoría" icon={ClipboardList} />
    {/if}

    <!-- Settings + Logout at bottom -->
    <div class="mt-auto flex flex-col items-center gap-1">
        <SidebarItem href="/settings" label="Configuración" icon={Settings} />
        <button
            onclick={handleLogout}
            disabled={loggingOut}
            class="relative group flex items-center justify-center w-7 h-6 rounded
                   transition-colors hover:bg-[#1e293b] disabled:opacity-50"
            title="Cerrar sesión"
        >
            <LogOut size={14} class="text-[#64748b] {loggingOut ? 'animate-pulse' : ''}" />
            <span class="absolute left-full ml-2 px-2 py-1 bg-[#1e293b] text-[#e2e8f0] text-xs
                         rounded opacity-0 group-hover:opacity-100 pointer-events-none whitespace-nowrap z-50">
                Cerrar sesión
            </span>
        </button>
    </div>
</nav>
