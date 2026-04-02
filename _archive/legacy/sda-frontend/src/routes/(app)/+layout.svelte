<script lang="ts">
    import Sidebar from '$lib/components/layout/Sidebar.svelte';
    import ToastContainer from '$lib/components/ui/ToastContainer.svelte';
    import { preferencesStore } from '$lib/stores/preferences.svelte';
    import { crossdoc } from '$lib/stores/crossdoc.svelte';
    import { toastStore } from '$lib/stores/toast.svelte';

    interface Props { data: any; children: any; }
    let { data, children }: Props = $props();

    // Guard para no re-disparar el mismo toast en cada navegación
    const shownNotifications = new Set<string>();

    $effect(() => {
        if (data.preferences) {
            preferencesStore.init(data.preferences);
            // Hidratar CrossdocStore con preferencias del usuario
            crossdoc.options.vdbTopK = data.preferences.vdb_top_k;
            crossdoc.options.rerankerTopK = data.preferences.reranker_top_k;
            crossdoc.options.maxSubQueries = data.preferences.max_sub_queries;
            crossdoc.options.followUpRetries = data.preferences.follow_up_retries;
            crossdoc.options.showDecomposition = data.preferences.show_decomposition;
        }
    });

    $effect(() => {
        for (const notif of data.notifications ?? []) {
            if (!shownNotifications.has(notif.message)) {
                shownNotifications.add(notif.message);
                if (notif.type === 'alert') {
                    toastStore.warning(notif.message);
                } else {
                    toastStore.info(notif.message);
                }
            }
        }
    });

    function handleBoundaryError(error: unknown, reset: () => void) {
        console.error('[boundary error]', error);
    }
</script>

<div class="flex h-screen overflow-hidden bg-[var(--bg-base)]">
    <Sidebar
        role={data.user.role}
        userName={data.user.name ?? data.user.email}
        userEmail={data.user.email}
    />

    <main class="flex-1 overflow-auto min-w-0">
        <svelte:boundary onerror={handleBoundaryError}>
            {@render children()}
            {#snippet failed(error, reset)}
                <div class="flex flex-col items-center justify-center h-full gap-4 p-8 text-center">
                    <p class="text-[var(--danger)] font-medium">Algo salió mal</p>
                    <p class="text-sm text-[var(--text-muted)]">{(error as Error).message ?? 'Error desconocido'}</p>
                    <button
                        onclick={reset}
                        class="px-4 py-2 text-sm bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-md)] hover:border-[var(--accent)] text-[var(--text-muted)] hover:text-[var(--text)] transition-colors"
                    >
                        Reintentar
                    </button>
                </div>
            {/snippet}
        </svelte:boundary>
    </main>
</div>

<ToastContainer />
