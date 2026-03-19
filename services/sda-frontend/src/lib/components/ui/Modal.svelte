<script lang="ts">
    interface Props {
        open?: boolean;
        title?: string;
        onclose?: () => void;
        size?: 'sm' | 'md' | 'lg';
        children: any;
        footer?: any;
    }

    let {
        open = $bindable(false),
        title,
        onclose,
        size = 'md',
        children,
        footer,
    }: Props = $props();

    const sizes = { sm: 'max-w-sm', md: 'max-w-md', lg: 'max-w-2xl' };

    function close() {
        open = false;
        onclose?.();
    }

    function onKeydown(e: KeyboardEvent) {
        if (e.key === 'Escape') close();
    }

    function onBackdropClick(e: MouseEvent) {
        if (e.target === e.currentTarget) close();
    }
</script>

<svelte:window onkeydown={open ? onKeydown : undefined} />

{#if open}
    <!-- Backdrop -->
    <div
        class="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4"
        role="dialog"
        aria-modal="true"
        aria-label={title}
        onclick={onBackdropClick}
    >
        <!-- Panel -->
        <div class="
            bg-[var(--bg-surface)] border border-[var(--border)]
            rounded-[var(--radius-lg)] shadow-2xl w-full {sizes[size]}
        ">
            {#if title}
                <div class="flex items-center justify-between px-5 py-4 border-b border-[var(--border)]">
                    <h2 class="text-sm font-semibold text-[var(--text)]">{title}</h2>
                    <button
                        onclick={close}
                        class="text-[var(--text-faint)] hover:text-[var(--text)] transition-colors p-1 rounded"
                        aria-label="Cerrar"
                    >✕</button>
                </div>
            {/if}

            <div class="px-5 py-4">
                {@render children()}
            </div>

            {#if footer}
                <div class="px-5 py-4 border-t border-[var(--border)] flex justify-end gap-2">
                    {@render footer()}
                </div>
            {/if}
        </div>
    </div>
{/if}
