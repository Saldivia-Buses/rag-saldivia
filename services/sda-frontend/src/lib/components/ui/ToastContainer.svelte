<script lang="ts">
    import { toastStore } from '$lib/stores/toast.svelte';
    import Toast from './Toast.svelte';
</script>

<!-- Portal de toasts — esquina inferior derecha -->
<div
    class="fixed bottom-5 right-5 z-[100] flex flex-col gap-2 pointer-events-none"
    aria-live="polite"
    aria-atomic="false"
>
    {#each toastStore.toasts as toast (toast.id)}
        <div class="pointer-events-auto">
            <Toast {toast} />
        </div>
    {/each}
</div>
