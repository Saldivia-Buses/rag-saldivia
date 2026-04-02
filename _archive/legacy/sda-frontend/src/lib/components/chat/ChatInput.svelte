<script lang="ts">
    import { Send, Square } from 'lucide-svelte';
    import CrossdocSettingsPopover from './CrossdocSettingsPopover.svelte';

    interface Props {
        streaming: boolean;
        crossdoc: boolean;
        onsubmit: (query: string) => void;
        onstop: () => void;
        oncrossdoctoggle: () => void;
    }

    let { streaming, crossdoc, onsubmit, onstop, oncrossdoctoggle }: Props = $props();

    let input = $state('');

    function handleSubmit() {
        const query = input.trim();
        if (!query || streaming) return;
        input = '';
        onsubmit(query);
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSubmit();
        }
    }
</script>

<div class="p-3 border-t border-[var(--border)] flex-shrink-0">
    <div
        class="flex gap-2 bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-md)]
               px-3 py-2 focus-within:border-[var(--accent)] transition-colors"
    >
        <textarea
            bind:value={input}
            onkeydown={handleKeydown}
            rows={1}
            placeholder="Escribí tu consulta..."
            disabled={streaming}
            class="flex-1 bg-transparent text-xs text-[var(--text)] placeholder-[var(--text-faint)]
                   resize-none outline-none disabled:opacity-60"
            style="max-height: 120px; overflow-y: auto;"
        ></textarea>

        <div class="flex items-center gap-2 flex-shrink-0">
            <CrossdocSettingsPopover active={crossdoc} ontoggle={oncrossdoctoggle} />

            {#if streaming}
                <button
                    onclick={onstop}
                    title="Detener generación"
                    class="text-[var(--danger)] hover:opacity-80 transition-opacity"
                >
                    <Square size={14} fill="currentColor" />
                </button>
            {:else}
                <button
                    onclick={handleSubmit}
                    disabled={!input.trim()}
                    title="Enviar (Enter)"
                    class="text-[var(--accent)] hover:text-[var(--accent-hover)]
                           disabled:opacity-40 transition-colors"
                >
                    <Send size={16} />
                </button>
            {/if}
        </div>
    </div>
</div>
