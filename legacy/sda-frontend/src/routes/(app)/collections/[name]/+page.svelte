<script lang="ts">
    import DeleteModal from '../_components/DeleteModal.svelte';
    import { toastStore } from '$lib/stores/toast.svelte';
    import { goto } from '$app/navigation';
    import { Trash2, FileText, Database } from 'lucide-svelte';

    let { data } = $props();

    let showDelete = $state(false);

    async function handleDelete() {
        const res = await fetch(`/api/collections/${data.name}`, { method: 'DELETE' });
        if (!res.ok) {
            const body = await res.json().catch(() => ({}));
            throw new Error((body as any).message ?? `Error ${res.status}`);
        }
        toastStore.success(`Colección "${data.name}" eliminada.`);
        goto('/collections');
    }
</script>

<div class="p-6 max-w-2xl">
    <div class="flex items-center gap-3 mb-6">
        <a href="/collections" class="text-[var(--text-faint)] hover:text-[var(--text-muted)] text-sm transition-colors">
            ← Colecciones
        </a>
        <h1 class="text-lg font-semibold text-[var(--text)]">{data.name}</h1>
    </div>

    {#if data.error}
        <p class="text-sm text-[var(--danger)] mb-4">{data.error}</p>
    {/if}

    <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-5 mb-4">
        <div class="grid grid-cols-2 gap-6 text-sm">
            <div>
                <div class="flex items-center gap-1.5 text-[var(--text-faint)] text-xs mb-1.5">
                    <Database size={12} />
                    Entidades
                </div>
                <div class="text-[var(--text)] font-semibold text-xl">
                    {data.stats?.entity_count?.toLocaleString() ?? '—'}
                </div>
            </div>
            <div>
                <div class="flex items-center gap-1.5 text-[var(--text-faint)] text-xs mb-1.5">
                    <FileText size={12} />
                    Documentos
                </div>
                <div class="text-[var(--text)] font-semibold text-xl">
                    {data.stats?.document_count?.toLocaleString() ?? '—'}
                </div>
            </div>
        </div>
    </div>

    <div class="flex items-center gap-3">
        <a
            href="/chat"
            class="inline-flex items-center px-4 py-2 text-sm bg-[var(--accent)] text-white
                   rounded-lg hover:opacity-90 transition-opacity"
        >
            Consultar esta colección
        </a>
        <button
            onclick={() => showDelete = true}
            class="inline-flex items-center gap-1.5 px-3 py-2 text-sm text-[var(--danger)]
                   border border-[var(--border)] rounded-lg hover:border-[var(--danger)]
                   hover:bg-red-950/20 transition-colors"
        >
            <Trash2 size={14} />
            Eliminar
        </button>
    </div>
</div>

<DeleteModal bind:open={showDelete} name={data.name} onconfirm={handleDelete} />
