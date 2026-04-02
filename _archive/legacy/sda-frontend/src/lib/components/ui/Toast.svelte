<script lang="ts">
    import { toastStore, type Toast } from '$lib/stores/toast.svelte';

    interface Props { toast: Toast; }
    let { toast }: Props = $props();

    const styles = {
        success: 'bg-[var(--success-bg)] border-[var(--success)] text-[var(--success)]',
        error:   'bg-[var(--danger-bg)] border-[var(--danger)] text-[var(--danger)]',
        warning: 'bg-[var(--warning-bg)] border-[var(--warning)] text-[var(--warning)]',
        info:    'bg-[var(--info-bg)] border-[var(--info)] text-[var(--info)]',
    };

    const icons = { success: '✓', error: '✕', warning: '⚠', info: 'ℹ' };
</script>

<div class="
    flex items-start gap-3 px-4 py-3 rounded-[var(--radius-md)]
    border shadow-lg min-w-72 max-w-sm text-sm
    {styles[toast.type]}
">
    <span class="font-bold text-base leading-none mt-0.5">{icons[toast.type]}</span>
    <p class="flex-1 leading-snug">{toast.message}</p>
    <button
        onclick={() => toastStore.dismiss(toast.id)}
        class="opacity-60 hover:opacity-100 transition-opacity leading-none mt-0.5"
        aria-label="Cerrar notificación"
    >✕</button>
</div>
