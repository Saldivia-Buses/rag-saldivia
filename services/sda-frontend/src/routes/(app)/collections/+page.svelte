<script lang="ts">
    let { data } = $props();
</script>

<div class="p-6">
    <h1 class="text-lg font-semibold text-[var(--text)] mb-4">Colecciones</h1>
    <div class="grid grid-cols-3 gap-4">
        {#each data.collections as name}
            <a href="/collections/{name}"
               class="bg-[var(--bg-base)] border border-[var(--border)] rounded-lg p-4
                      hover:border-[var(--accent)] transition-colors block">
                <div class="text-sm font-medium text-[var(--text)] mb-2">{name}</div>
                {#if data.stats[name]}
                    <div class="text-xs text-[var(--text-faint)]">
                        {data.stats[name].entity_count?.toLocaleString() ?? '—'} entidades
                    </div>
                {/if}
            </a>
        {/each}
    </div>
</div>
