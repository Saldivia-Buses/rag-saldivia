<!-- services/sda-frontend/src/lib/components/layout/Topbar.svelte -->
<!-- TODO Fase 5: implementar topbar con breadcrumbs, acciones contextuales -->
<script lang="ts">
    interface Props {
        title?: string;
        class?: string;
    }
    let { title, class: className = '' }: Props = $props();
</script>

{#if title}
    <header class="h-12 flex items-center px-4 border-b border-[var(--border)] bg-[var(--bg-surface)] {className}">
        <h1 class="text-sm font-medium text-[var(--text)]">{title}</h1>
    </header>
{/if}
