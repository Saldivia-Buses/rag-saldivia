<script lang="ts">
    import { MessageSquare, BookOpen, Upload, Users, ClipboardList, Settings } from 'lucide-svelte';
    import SidebarItem from './SidebarItem.svelte';

    interface Props {
        role: string;
        areaId: number;
    }

    let { role, areaId }: Props = $props();

    let isAdmin = $derived(role === 'admin');
    let isManager = $derived(role === 'admin' || role === 'area_manager');
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

    <!-- Settings at bottom -->
    <div class="mt-auto">
        <SidebarItem href="/settings" label="Configuración" icon={Settings} />
    </div>
</nav>
