<script lang="ts">
    import Modal from '$lib/components/ui/Modal.svelte';
    import Button from '$lib/components/ui/Button.svelte';

    interface Props {
        open?: boolean;
        name?: string;
        onconfirm?: () => Promise<void> | void;
        onclose?: () => void;
    }
    let { open = $bindable(false), name = '', onconfirm, onclose }: Props = $props();

    let loading = $state(false);

    async function handleConfirm() {
        loading = true;
        try {
            await onconfirm?.();
            open = false;
        } catch {
            // error handled by caller (toast)
        } finally {
            loading = false;
        }
    }

    function handleClose() {
        open = false;
        onclose?.();
    }
</script>

<Modal bind:open title="Eliminar colección" onclose={handleClose} size="sm">
    <p class="text-sm text-[var(--text-muted)]">
        ¿Estás seguro de que querés eliminar
        <span class="font-semibold text-[var(--text)]">"{name}"</span>?
        Esta acción no se puede deshacer.
    </p>

    {#snippet footer()}
        <Button variant="ghost" onclick={handleClose} disabled={loading}>Cancelar</Button>
        <Button variant="danger" onclick={handleConfirm} disabled={loading}>
            {loading ? 'Eliminando...' : 'Eliminar'}
        </Button>
    {/snippet}
</Modal>
