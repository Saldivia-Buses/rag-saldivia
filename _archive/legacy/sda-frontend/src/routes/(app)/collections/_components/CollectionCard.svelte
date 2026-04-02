<script lang="ts">
    import Badge from '$lib/components/ui/Badge.svelte';
    import Skeleton from '$lib/components/ui/Skeleton.svelte';
    import type { CollectionStats } from '$lib/server/gateway';

    interface Props {
        name: string;
        stats: CollectionStats | null;
        href: string;
    }
    let { name, stats, href }: Props = $props();
</script>

<a {href} class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-4
                  hover:border-[var(--accent)] transition-colors block group">
    <div class="flex items-start justify-between mb-3">
        <span class="text-sm font-semibold text-[var(--text)] truncate">{name}</span>
        {#if stats?.has_sparse}
            <Badge variant="blue">Sparse</Badge>
        {/if}
    </div>
    {#if stats}
        <div class="text-2xl font-bold text-[var(--text)] mb-1">
            {stats.entity_count.toLocaleString()}
        </div>
        <div class="text-xs text-[var(--text-faint)]">
            entidades{stats.index_type ? ` · ${stats.index_type}` : ''}
        </div>
    {:else}
        <Skeleton class="h-8 w-24 mb-1" />
        <Skeleton class="h-3 w-32" />
    {/if}
</a>
