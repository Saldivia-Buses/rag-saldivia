<script lang="ts">
    import Sidebar from '$lib/components/layout/Sidebar.svelte';
    import ToastContainer from '$lib/components/ui/ToastContainer.svelte';
    import { toastStore } from '$lib/stores/toast.svelte';

    interface Props { data: any; children: any; }
    let { data, children }: Props = $props();

    function onBoundaryError(error: unknown, reset: () => void) {
        const msg = error instanceof Error ? error.message : String(error);
        toastStore.error(`Error inesperado: ${msg}`);
        console.error('[app boundary]', error);
    }
</script>

<div class="flex h-screen overflow-hidden bg-[var(--bg-base)]">
    <Sidebar
        role={data.user.role}
        userName={data.user.name ?? data.user.email}
        userEmail={data.user.email}
    />

    <main class="flex-1 overflow-auto min-w-0">
        <svelte:boundary onerror={onBoundaryError}>
            {@render children()}
            {#snippet failed(error, reset)}
                <div class="flex items-center justify-center h-64 p-8">
                    <div class="text-center">
                        <div class="text-4xl mb-3">⚠️</div>
                        <p class="text-sm font-semibold text-[var(--text)] mb-1">Algo salió mal</p>
                        <p class="text-xs text-[var(--text-muted)] mb-4">{error instanceof Error ? error.message : String(error)}</p>
                        <button
                            onclick={reset}
                            class="text-xs text-[var(--accent)] hover:underline"
                        >Reintentar</button>
                    </div>
                </div>
            {/snippet}
        </svelte:boundary>
    </main>
</div>

<ToastContainer />
