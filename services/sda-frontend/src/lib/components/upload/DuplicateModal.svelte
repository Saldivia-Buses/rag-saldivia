<script lang="ts">
    interface Props {
        filename: string;
        collection: string;
        state: 'completed' | 'stalled';
        indexedAt?: string | null;
        pages?: number | null;
        onConfirm: () => void;
        onCancel: () => void;
    }

    let { filename, collection, state, indexedAt, pages, onConfirm, onCancel }: Props = $props();

    function formatDate(iso: string | null | undefined): string {
        if (!iso) return 'fecha desconocida';
        const d = new Date(iso);
        const diff = Math.floor((Date.now() - d.getTime()) / 1000);
        if (diff < 3600) return `hace ${Math.floor(diff / 60)} minutos`;
        if (diff < 86400) return `hace ${Math.floor(diff / 3600)} horas`;
        return `hace ${Math.floor(diff / 86400)} días`;
    }

    const isCompleted = state === 'completed';
    const title = isCompleted
        ? `"${filename}" ya fue indexado`
        : `"${filename}" intentó indexarse y falló`;
    const confirmLabel = isCompleted ? 'Subir de todas formas' : 'Reintentar ahora';
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
    <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-xl p-6 w-full max-w-sm shadow-xl">
        <div class="flex items-start gap-3 mb-4">
            <span class="text-xl">{isCompleted ? '⚠️' : '❌'}</span>
            <p class="text-sm font-medium text-[var(--text)]">{title}</p>
        </div>

        <div class="text-xs text-[var(--text-muted)] space-y-1 mb-6 ml-8">
            <p>Colección: <span class="text-[var(--text)]">{collection}</span></p>
            {#if indexedAt}
                <p>{isCompleted ? 'Indexado' : 'Último intento'}: <span class="text-[var(--text)]">{formatDate(indexedAt)}</span></p>
            {/if}
            {#if pages}
                <p>Páginas: <span class="text-[var(--text)]">{pages}</span></p>
            {/if}
        </div>

        <div class="flex gap-2">
            <button
                onclick={onCancel}
                class="flex-1 py-2 text-sm border border-[var(--border)] rounded-lg
                       text-[var(--text-muted)] hover:border-[var(--text-muted)] transition-colors"
            >
                Cancelar
            </button>
            <button
                onclick={onConfirm}
                class="flex-1 py-2 text-sm font-medium rounded-lg transition-opacity
                       {isCompleted
                           ? 'bg-[var(--accent)] text-white hover:opacity-90'
                           : 'bg-amber-500 text-white hover:opacity-90'}"
            >
                {confirmLabel}
            </button>
        </div>
    </div>
</div>
