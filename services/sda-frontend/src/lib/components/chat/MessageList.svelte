<script lang="ts">
    import { ChevronDown } from 'lucide-svelte';
    import { isNearBottom } from '$lib/utils/scroll';
    import MarkdownRenderer from './MarkdownRenderer.svelte';

    interface Message {
        role: 'user' | 'assistant';
        content: string;
        timestamp: string;
    }

    interface Props {
        messages: Message[];
        streaming: boolean;
        streamingContent: string;
    }

    let { messages, streaming, streamingContent }: Props = $props();

    let scrollEl = $state<HTMLDivElement | null>(null);
    let showScrollButton = $state(false);

    function handleScroll() {
        if (!scrollEl) return;
        showScrollButton = !isNearBottom(scrollEl);
    }

    function scrollToBottom() {
        if (!scrollEl) return;
        scrollEl.scrollTo({ top: scrollEl.scrollHeight, behavior: 'smooth' });
        showScrollButton = false;
    }

    // Auto-scroll: solo si el usuario ya está cerca del fondo
    $effect(() => {
        // Dependencias que disparan el efecto:
        void messages.length;
        void streamingContent.length;

        if (!scrollEl) return;
        if (isNearBottom(scrollEl)) {
            scrollEl.scrollTo({ top: scrollEl.scrollHeight, behavior: 'smooth' });
        }
    });
</script>

<div class="relative flex-1 overflow-hidden">
    <div
        bind:this={scrollEl}
        onscroll={handleScroll}
        class="h-full overflow-y-auto p-3 flex flex-col gap-3"
    >
        {#each messages as msg (msg.timestamp)}
            {#if msg.role === 'user'}
                <div class="flex justify-end">
                    <div class="bg-[var(--accent)] rounded-lg rounded-tr-sm px-3 py-2 max-w-[70%]">
                        <p class="text-xs text-white whitespace-pre-wrap">{msg.content}</p>
                    </div>
                </div>
            {:else}
                <div class="flex gap-2">
                    <div class="w-5 h-5 bg-[var(--accent)] rounded-full flex-shrink-0 mt-0.5"></div>
                    <div class="bg-[var(--bg-surface)] border border-[var(--border)]
                                rounded-lg rounded-tl-sm px-3 py-2 max-w-[80%] min-w-0">
                        <MarkdownRenderer content={msg.content} />
                    </div>
                </div>
            {/if}
        {/each}

        <!-- Mensaje en streaming -->
        {#if streaming}
            <div class="flex gap-2">
                <div class="w-5 h-5 bg-[var(--accent)] rounded-full animate-pulse flex-shrink-0 mt-0.5"></div>
                <div class="bg-[var(--bg-surface)] border border-[var(--border)]
                            rounded-lg rounded-tl-sm px-3 py-2 max-w-[80%] min-w-0">
                    <MarkdownRenderer content={streamingContent} />
                    <span class="text-[var(--text-faint)] animate-pulse">▋</span>
                </div>
            </div>
        {/if}
    </div>

    <!-- Botón "Ir al fondo" — aparece cuando el usuario scrolleó arriba -->
    {#if showScrollButton}
        <button
            onclick={scrollToBottom}
            title="Ir al fondo"
            class="absolute bottom-4 right-4 bg-[var(--bg-surface)] border border-[var(--border)]
                   rounded-full p-2 shadow-lg text-[var(--text-muted)] hover:text-[var(--text)]
                   hover:bg-[var(--bg-hover)] transition-colors z-10"
        >
            <ChevronDown size={16} />
        </button>
    {/if}
</div>
