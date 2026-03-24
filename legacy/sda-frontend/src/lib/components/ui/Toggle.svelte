<!-- services/sda-frontend/src/lib/components/ui/Toggle.svelte -->
<script lang="ts">
    import { toggleMode } from 'mode-watcher';

    interface Props {
        label?: string;
        size?: 'sm' | 'md';
        class?: string;
    }
    let { label, size = 'md', class: className = '' }: Props = $props();

    const sizes = {
        sm: 'text-xs px-2 py-1 h-6',
        md: 'text-sm px-2.5 py-1.5 h-8',
    };
</script>

<button
    onclick={toggleMode}
    title={label ?? 'Alternar modo oscuro/claro'}
    class="inline-flex items-center gap-1.5 rounded-[var(--radius-sm)] bg-[var(--bg-surface)] border border-[var(--border)] text-[var(--text-muted)] hover:border-[var(--accent)] hover:text-[var(--text)] transition-colors {sizes[size]} {className}"
>
    <span class="w-3.5 h-3.5" aria-hidden="true">
        <!-- Sun icon (light mode indicator) -->
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="8" cy="8" r="3"/>
            <path d="M8 1v2M8 13v2M1 8h2M13 8h2M3.05 3.05l1.41 1.41M11.54 11.54l1.41 1.41M3.05 12.95l1.41-1.41M11.54 4.46l1.41-1.41"/>
        </svg>
    </span>
    {#if label}<span>{label}</span>{/if}
</button>
