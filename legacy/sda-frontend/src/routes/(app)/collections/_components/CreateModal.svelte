<script lang="ts">
    import Modal from '$lib/components/ui/Modal.svelte';
    import Input from '$lib/components/ui/Input.svelte';
    import Button from '$lib/components/ui/Button.svelte';

    interface Props {
        open?: boolean;
        oncreate?: (name: string, schema: string) => Promise<void> | void;
        onclose?: () => void;
    }
    let { open = $bindable(false), oncreate, onclose }: Props = $props();

    let name = $state('');
    let schema = $state('default');
    let errorMsg = $state('');
    let loading = $state(false);

    const SCHEMAS = [
        { value: 'default', label: 'Default (dense vectors)' },
        { value: 'sparse', label: 'Sparse (hybrid search)' },
    ];

    function validate(): string {
        if (!name.trim()) return 'El nombre es requerido.';
        if (!/^[a-z0-9_-]+$/i.test(name.trim())) return 'Solo letras, números, guiones y guiones bajos.';
        if (name.trim().length > 64) return 'Máximo 64 caracteres.';
        return '';
    }

    async function handleSubmit() {
        errorMsg = validate();
        if (errorMsg) return;
        loading = true;
        try {
            await oncreate?.(name.trim(), schema);
            open = false;
            name = '';
            schema = 'default';
            errorMsg = '';
        } catch (e: any) {
            errorMsg = e.message ?? 'Error al crear la colección.';
        } finally {
            loading = false;
        }
    }

    function handleClose() {
        name = '';
        schema = 'default';
        errorMsg = '';
        open = false;
        onclose?.();
    }

    function handleNameKeydown(e: KeyboardEvent) {
        if (e.key === 'Enter') {
            handleSubmit();
        }
    }
</script>

<Modal bind:open title="Nueva colección" onclose={handleClose} size="sm">
    <div class="space-y-4">
        <div>
            <label for="collection-name" class="block text-xs font-medium text-[var(--text-muted)] mb-1.5">
                Nombre
            </label>
            <input
                id="collection-name"
                bind:value={name}
                type="text"
                placeholder="ej: documentos-legales"
                onkeydown={handleNameKeydown}
                disabled={loading}
                class="w-full px-3 py-2 text-sm rounded-[var(--radius-md)]
                       bg-[var(--bg-surface)] border text-[var(--text)]
                       placeholder:text-[var(--text-faint)]
                       focus:outline-none focus:ring-2 focus:ring-[var(--accent)] focus:border-[var(--accent)]
                       disabled:opacity-50 disabled:cursor-not-allowed
                       transition-colors border-[var(--border)]"
            />
            {#if errorMsg}
                <p class="text-xs text-[var(--danger)] mt-1.5">{errorMsg}</p>
            {/if}
        </div>
        <div>
            <label for="collection-schema" class="block text-xs font-medium text-[var(--text-muted)] mb-1.5">
                Schema
            </label>
            <select
                id="collection-schema"
                bind:value={schema}
                disabled={loading}
                class="w-full bg-[var(--bg-surface)] border border-[var(--border)] rounded-lg
                       px-3 py-2 text-sm text-[var(--text)] focus:outline-none
                       focus:ring-2 focus:ring-[var(--accent)] focus:border-[var(--accent)]
                       disabled:opacity-50 disabled:cursor-not-allowed
                       transition-colors"
            >
                {#each SCHEMAS as s (s.value)}
                    <option value={s.value}>{s.label}</option>
                {/each}
            </select>
        </div>
    </div>

    {#snippet footer()}
        <Button variant="ghost" onclick={handleClose} disabled={loading}>Cancelar</Button>
        <Button onclick={handleSubmit} disabled={loading}>
            {loading ? 'Creando...' : 'Crear colección'}
        </Button>
    {/snippet}
</Modal>
