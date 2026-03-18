<script lang="ts">
    let { data } = $props();
</script>

<div class="p-6">
    <h1 class="text-lg font-semibold text-[#e2e8f0] mb-4">Colecciones</h1>
    <div class="grid grid-cols-3 gap-4">
        {#each data.collections as name}
            <a href="/collections/{name}"
               class="bg-[#0f172a] border border-[#1e293b] rounded-lg p-4
                      hover:border-[#6366f1] transition-colors block">
                <div class="text-sm font-medium text-[#e2e8f0] mb-2">{name}</div>
                {#if data.stats[name]}
                    <div class="text-xs text-[#475569]">
                        {data.stats[name].entity_count?.toLocaleString() ?? '—'} entidades
                    </div>
                {/if}
            </a>
        {/each}
    </div>
</div>
