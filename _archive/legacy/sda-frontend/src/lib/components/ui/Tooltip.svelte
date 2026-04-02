<!-- services/sda-frontend/src/lib/components/ui/Tooltip.svelte -->
<script lang="ts">
    interface Props {
        text: string;
        position?: 'top' | 'bottom' | 'left' | 'right';
        children: any;
    }
    let { text, position = 'top', children }: Props = $props();

    const positionClasses = {
        top:    'bottom-full left-1/2 -translate-x-1/2 mb-1.5',
        bottom: 'top-full left-1/2 -translate-x-1/2 mt-1.5',
        left:   'right-full top-1/2 -translate-y-1/2 mr-1.5',
        right:  'left-full top-1/2 -translate-y-1/2 ml-1.5',
    };
</script>

<div class="relative inline-flex group">
    {@render children()}
    <div
        role="tooltip"
        class="pointer-events-none absolute {positionClasses[position]}
               z-50 px-2 py-1 text-xs rounded-[var(--radius-sm)]
               bg-[var(--bg-surface)] border border-[var(--border)]
               text-[var(--text-muted)] whitespace-nowrap shadow-sm
               opacity-0 group-hover:opacity-100 group-focus-within:opacity-100 transition-opacity duration-150"
    >
        {text}
    </div>
</div>
