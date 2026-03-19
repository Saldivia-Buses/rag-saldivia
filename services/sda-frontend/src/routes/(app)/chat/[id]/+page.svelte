<script lang="ts">
    import { onMount } from 'svelte';
    import { ChatStore } from '$lib/stores/chat.svelte';
    import { MessageSquare, Send, RefreshCw } from 'lucide-svelte';

    let { data } = $props();
    const chat = new ChatStore();

    let input = $state('');
    let selectedCollection = $state(data.session.collection);

    onMount(() => {
        chat.collection = data.session.collection;
        chat.crossdoc = data.session.crossdoc;
        if (data.session.messages?.length) {
            chat.loadMessages(data.session.messages);
        }
    });

    async function sendMessage() {
        if (!input.trim() || chat.streaming) return;
        const query = input;
        input = '';

        chat.addUserMessage(query);
        chat.startStream();

        try {
            const resp = await fetch(`/api/chat/stream/${data.session.id}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    query,
                    collection_names: [selectedCollection],
                    crossdoc: chat.crossdoc,
                }),
            });

            if (!resp.ok) {
                chat.appendToken('\n[Error: no se pudo conectar con el servidor]');
                return;
            }

            const reader = resp.body!.getReader();
            const decoder = new TextDecoder();
            let buffer = '';

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                // stream: true preserves multi-byte sequences across chunks
                buffer += decoder.decode(value, { stream: true });
                const lines = buffer.split('\n');
                // Keep last (potentially incomplete) line in buffer
                buffer = lines.pop() ?? '';
                // Parse SSE events
                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        const dataStr = line.slice(6).trim();
                        if (dataStr === '[DONE]') continue;
                        try {
                            const event = JSON.parse(dataStr);
                            if (event.choices?.[0]?.delta?.content) {
                                chat.appendToken(event.choices[0].delta.content);
                            }
                            if (event.citations) {
                                chat.setSources(event.citations.map((c: any) => ({
                                    document: c.source_name ?? c.document ?? '',
                                    page: c.page,
                                    excerpt: c.content ?? c.excerpt ?? '',
                                })));
                            }
                        } catch { /* ignore parse errors on partial chunks */ }
                    }
                }
            }
        } catch (err) {
            console.error('Stream error:', err);
            chat.appendToken('\n[Error: conexión interrumpida]');
        } finally {
            chat.finalizeStream();
        }
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    }
</script>

<div class="flex h-screen overflow-hidden">

    <!-- Panel izquierdo: historial -->
    <div class="w-40 bg-[var(--bg-surface)] border-r border-[var(--border)] flex flex-col p-2 gap-1 overflow-y-auto">
        <div class="text-xs text-[var(--text-faint)] font-semibold uppercase tracking-wide mb-1">
            Historial
        </div>
        <a href="/chat" data-sveltekit-preload-data="false" class="flex items-center gap-1.5 text-[var(--accent)] text-xs mb-2 hover:underline">
            <MessageSquare size={10} /> Nueva consulta
        </a>
        {#each data.history as session}
            <a
                href="/chat/{session.id}"
                data-sveltekit-preload-data="false"
                class="bg-[var(--bg-surface)] rounded p-1.5 block hover:bg-[var(--border)] transition-colors
                       {session.id === data.session.id ? 'border-l-2 border-[var(--accent)]' : ''}"
            >
                <div class="text-xs text-[var(--text-muted)] font-medium truncate">{session.title}</div>
                <div class="text-xs text-[var(--text-faint)] mt-0.5">{session.updated_at.slice(0,10)}</div>
            </a>
        {/each}
    </div>

    <!-- Panel central: conversación -->
    <div class="flex-1 flex flex-col border-r border-[var(--border)] min-w-0">
        <!-- Header -->
        <div class="flex items-center gap-2 px-3 py-2 border-b border-[var(--border)] text-xs">
            <select
                bind:value={selectedCollection}
                class="bg-[var(--bg-surface)] border border-[var(--border)] rounded px-2 py-0.5 text-[var(--text)]"
            >
                {#each data.collections as col}
                    <option value={col}>{col}</option>
                {/each}
            </select>
            <label class="flex items-center gap-1 text-[var(--text-muted)] cursor-pointer">
                <input type="checkbox" bind:checked={chat.crossdoc} class="accent-[var(--accent)]" />
                Crossdoc
            </label>
        </div>

        <!-- Messages -->
        <div class="flex-1 overflow-y-auto p-3 flex flex-col gap-3">
            {#each chat.messages as msg}
                {#if msg.role === 'user'}
                    <div class="flex justify-end">
                        <div class="bg-[var(--accent)] rounded-lg rounded-tr-sm px-3 py-2 max-w-[70%]">
                            <p class="text-xs text-white">{msg.content}</p>
                        </div>
                    </div>
                {:else}
                    <div class="flex gap-2">
                        <div class="w-5 h-5 bg-[var(--accent)] rounded-full flex-shrink-0 mt-0.5"></div>
                        <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-lg rounded-tl-sm px-3 py-2 max-w-[80%]">
                            <p class="text-xs text-[var(--text)] leading-relaxed whitespace-pre-wrap">
                                {msg.content}
                            </p>
                        </div>
                    </div>
                {/if}
            {/each}

            <!-- Streaming -->
            {#if chat.streaming}
                <div class="flex gap-2">
                    <div class="w-5 h-5 bg-[var(--accent)] rounded-full flex-shrink-0 mt-0.5 animate-pulse"></div>
                    <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-lg rounded-tl-sm px-3 py-2 max-w-[80%]">
                        <p class="text-xs text-[var(--text)] leading-relaxed whitespace-pre-wrap">
                            {chat.streamingContent}<span class="animate-pulse">▋</span>
                        </p>
                    </div>
                </div>
            {/if}
        </div>

        <!-- Input -->
        <div class="p-3 border-t border-[var(--border)]">
            <div class="flex gap-2 bg-[var(--bg-surface)] border border-[var(--border)] rounded-lg px-3 py-2">
                <textarea
                    bind:value={input}
                    onkeydown={handleKeydown}
                    rows={1}
                    placeholder="Escribí tu consulta..."
                    class="flex-1 bg-transparent text-xs text-[var(--text)] placeholder-[var(--text-faint)]
                           resize-none outline-none"
                ></textarea>
                <button
                    onclick={sendMessage}
                    disabled={chat.streaming || !input.trim()}
                    class="text-[var(--accent)] hover:text-[var(--accent-hover)] disabled:opacity-40 transition-colors"
                >
                    {#if chat.streaming}
                        <RefreshCw size={16} class="animate-spin" />
                    {:else}
                        <Send size={16} />
                    {/if}
                </button>
            </div>
        </div>
    </div>

    <!-- Panel derecho: fuentes -->
    <div class="w-48 bg-[var(--bg-surface)] p-3 overflow-y-auto">
        <div class="text-xs text-[var(--text-faint)] font-semibold uppercase tracking-wide mb-2">
            Fuentes ({chat.sources.length})
        </div>
        {#each chat.sources as source, i}
            <div class="bg-[var(--bg-surface)] rounded p-2 mb-2 border-l-2
                        {i === 0 ? 'border-[var(--accent)]' : i === 1 ? 'border-[var(--accent-hover)]' : 'border-[var(--border)]'}">
                <div class="text-xs text-[var(--accent)] font-semibold truncate">{source.document}</div>
                {#if source.page}
                    <div class="text-xs text-[var(--text-faint)]">p. {source.page}</div>
                {/if}
                <div class="text-xs text-[var(--text-muted)] mt-1 line-clamp-3">{source.excerpt}</div>
            </div>
        {/each}
    </div>

</div>
