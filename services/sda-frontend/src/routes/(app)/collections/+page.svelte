<script lang="ts">
    import CollectionCard from './_components/CollectionCard.svelte';
    import CreateModal from './_components/CreateModal.svelte';
    import { collectionsStore } from '$lib/stores/collections.svelte';
    import { toastStore } from '$lib/stores/toast.svelte';
    import { invalidateAll } from '$app/navigation';
    import { Plus } from 'lucide-svelte';
    import type { PageData } from './$types';
    import type { CollectionStats } from '$lib/server/gateway';

    let { data }: { data: PageData } = $props();

    // Hydrate the client-side store from server data (for Fase 17 CommandPalette)
    $effect(() => {
        collectionsStore.init(data.collections);
    });

    let showCreate = $state(false);

    async function handleCreate(name: string, schema: string) {
        await collectionsStore.create(name, schema);
        toastStore.success(`Colección "${name}" creada.`);
        await invalidateAll();
    }

    function getStats(name: string): CollectionStats | null {
        return (data.stats as Record<string, CollectionStats | null>)[name] ?? null;
    }
</script>

<div class="p-6">
    <div class="flex items-center justify-between mb-5">
        <h1 class="text-lg font-semibold text-[var(--text)]">Colecciones</h1>
        <button
            onclick={() => showCreate = true}
            class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm
                   bg-[var(--accent)] text-white rounded-lg hover:opacity-90 transition-opacity"
        >
            <Plus size={14} />
            Nueva colección
        </button>
    </div>

    {#if data.error}
        <p class="text-sm text-[var(--danger)]">{data.error}</p>
    {:else if data.collections.length === 0}
        <p class="text-sm text-[var(--text-muted)]">No hay colecciones. Creá una para empezar.</p>
    {:else}
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {#each data.collections as name (name)}
                <CollectionCard
                    {name}
                    stats={getStats(name)}
                    href="/collections/{name}"
                />
            {/each}
        </div>
    {/if}
</div>

<CreateModal bind:open={showCreate} oncreate={handleCreate} />
