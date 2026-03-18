<script lang="ts">
    import { page } from '$app/stores';

    interface Props {
        href: string;
        label: string;
        icon: any;
    }

    let { href, label, icon: Icon }: Props = $props();

    let active = $derived($page.url.pathname.startsWith(href));
</script>

<a
    {href}
    class="relative group flex items-center justify-center w-7 h-6 rounded
           transition-colors {active
               ? 'bg-[#1e293b] border-l-2 border-[#6366f1]'
               : 'hover:bg-[#1e293b]'}"
    title={label}
>
    <Icon size={14} class="text-{active ? '[#a5b4fc]' : '[#64748b]'}" />
    <!-- Tooltip -->
    <span class="absolute left-full ml-2 px-2 py-1 bg-[#1e293b] text-[#e2e8f0] text-xs
                 rounded opacity-0 group-hover:opacity-100 pointer-events-none whitespace-nowrap z-50">
        {label}
    </span>
</a>
